package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/StasMerzlyakov/go-metrics/internal/config"
	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/fs/backup"
	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/http/handler"
	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/http/middleware"
	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/http/middleware/compress"
	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/http/middleware/digest"
	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/http/middleware/logging"
	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/http/middleware/retry"
	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/storage/memory"
	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/storage/postgres"
	"github.com/StasMerzlyakov/go-metrics/internal/server/app"
	"github.com/StasMerzlyakov/go-metrics/internal/server/domain"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"

	"go.uber.org/zap"
)

type Server interface {
	Start(ctx context.Context)
	Shutdown(ctx context.Context)
}

func createMiddleWareList(srvConf *config.ServerConfiguration) []func(http.Handler) http.Handler {
	var mwList []func(http.Handler) http.Handler
	mwList = append(mwList, logging.EncrichWithRequestIDMW())
	mwList = append(mwList, logging.NewLoggingResponseMW())
	if srvConf.Key != "" {
		mwList = append(mwList, digest.NewWriteHashDigestResponseHeaderMW(srvConf.Key))
	}
	mwList = append(mwList, compress.NewCompressGZIPBufferResponseMW())
	mwList = append(mwList, compress.NewUncompressGZIPRequestMW())

	mwList = append(mwList, logging.NewLoggingRequestMW())

	// при работе с Postgres добавляем retriable-обертку
	// функция, обрабатывающая ошибки; в нужных случаях выкидывает нужную ошибку(domain.ErrDBConnection)
	// на которую реагирует retriableWrapper
	pgErrPreProcFn := func(err error) error {
		if err == nil {
			return nil
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgerrcode.IsConnectionException(pgErr.Code) {
				return domain.ErrDBConnection
			}
		}

		var conErr *pgconn.ConnectError
		if errors.As(err, &conErr) {
			return domain.ErrDBConnection
		}
		return err
	}
	mwList = append(mwList, retry.NewRetriableRequestMWConf(
		time.Duration(time.Second), time.Duration(2*time.Second), 4, pgErrPreProcFn,
	))

	return mwList
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

	domain.SetMainLogger(sugarLog)

	// -------- Хранилище -------------
	var storage FullStorage

	if srvConf.DatabaseDSN != "" {
		storage = postgres.NewStorage(srvConf.DatabaseDSN)
	} else {
		storage = memory.NewStorage()
	}

	defer storage.Close(srvCtx)

	if err := storage.Bootstrap(srvCtx); err != nil {
		panic(err)
	}

	// -------- Бэкап ------------
	backupFomratter := backup.NewJSON(srvConf.FileStoragePath)
	backUper := app.NewBackup(storage, backupFomratter)

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
	mwList := createMiddleWareList(srvConf)
	middleware.Add(httpHandler, mwList...)

	// операции с метриками
	metricApp := app.NewMetrics(storage)

	var updateMWList []func(http.Handler) http.Handler
	if srvConf.Key != "" {
		updateMWList = append(updateMWList, digest.NewCheckHashDigestRequestMW(srvConf.Key))
	}

	handler.AddMetricOperations(httpHandler, metricApp, updateMWList...)

	// административные операции
	adminApp := app.NewAdminApp(storage)
	handler.AddAdminOperations(httpHandler, adminApp)

	// ppfod
	handler.AddPProfOperations(httpHandler)

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
