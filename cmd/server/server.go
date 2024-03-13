package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/StasMerzlyakov/go-metrics/internal/config"
	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/fs/formatter"
	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/http/handler"
	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/http/middleware"
	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/http/middleware/compress"
	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/http/middleware/logging"
	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/storage/memory"
	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/storage/postgres"
	"github.com/StasMerzlyakov/go-metrics/internal/server/app"
	"github.com/go-chi/chi/v5"

	"go.uber.org/zap"
)

type Server interface {
	Start(ctx context.Context)
	Shutdown(ctx context.Context)
}

func createMiddleWareList(log *zap.SugaredLogger) []func(http.Handler) http.Handler {
	return []func(http.Handler) http.Handler{
		logging.NewLoggingResponseMW(log),
		compress.NewCompressGZIPResponseMW(log), //compress.NewCompressGZIPBufferResponseMW(log),
		compress.NewUncompressGZIPRequestMW(log),
		logging.NewLoggingRequestMW(log),
	}
}

type FullStorage interface {
	app.Storage
	app.Pinger
	Bootstrap(ctx context.Context) error
	Close(ctx context.Context) error
}

func main() {

	// -------- Контекст сервера ---------
	srvCtx, cancelFn := context.WithCancel(context.Background())

	// -------- Конфигурация ----------
	srvConf, err := config.LoadServerConfig()
	if err != nil {
		panic(err)
	}

	// -------- Логгер ---------------
	logger, err := zap.NewDevelopment()
	if err != nil {
		// вызываем панику, если ошибка
		panic("cannot initialize zap")
	}
	defer logger.Sync()

	sugarLog := logger.Sugar()

	// -------- Хранилище -------------
	var storage FullStorage

	if srvConf.DatabaseDSN != "" {
		storage = postgres.NewStorage(sugarLog, srvConf.DatabaseDSN)
	} else {
		storage = memory.NewStorage()
	}

	defer storage.Close(srvCtx)

	if err := storage.Bootstrap(srvCtx); err != nil {
		panic(err)
	}

	// -------- Бэкап ------------
	backupFomratter := formatter.NewJSON(sugarLog, srvConf.FileStoragePath)
	backUper := app.NewBackup(sugarLog, storage, backupFomratter)

	if srvConf.Restore {
		// восстановленгие бэкапа
		if err := backUper.RestoreBackUp(srvCtx); err != nil {
			panic(err)
		}
	}

	// проверяем - нужен ли синхронный бэкап
	doSyncBackup := srvConf.StoreInterval == 0

	if !doSyncBackup {
		// запускаем фоновый процесс
		go func() {
			storeInterval := time.Duration(srvConf.StoreInterval) * time.Second
			var ticker = time.NewTicker(storeInterval)
			defer ticker.Stop()
			for {
				select {
				case <-srvCtx.Done():
					sugarLog.Infow("Run", "msg", "backup finished")
					return
				case <-ticker.C:
					if err := backUper.DoBackUp(srvCtx); err != nil {
						sugarLog.Fatalw("DoBackUp", "msg", err.Error())
					}
				}
			}
		}()

	}

	// ---------- Http сервер -----------
	httpHandler := chi.NewMux()

	// мидлы
	mwList := createMiddleWareList(sugarLog)
	middleware.Add(httpHandler, mwList...)

	// операции с метриками
	metricApp := app.NewMetrics(storage)
	handler.AddMetricOperations(httpHandler, metricApp, sugarLog)

	// административные операции
	adminApp := app.NewAdminApp(sugarLog, storage)
	handler.AddAdminOperations(httpHandler, adminApp, sugarLog)

	// запускаем http-сервер
	srv := &http.Server{
		Addr:        srvConf.URL,
		Handler:     httpHandler,
		ReadTimeout: 0,
		IdleTimeout: 0,
	}

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			sugarLog.Fatalw("ListenAndServe", "msg", err.Error())
			panic(err)
		}
	}()

	// --------------- Обрабатываем остановку сервера --------------
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)

	defer func() {
		cancelFn()
		srv.Shutdown(srvCtx)
	}()
	<-exit
}
