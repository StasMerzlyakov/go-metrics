package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/StasMerzlyakov/go-metrics/internal/config"
	"github.com/StasMerzlyakov/go-metrics/internal/server"
	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/fs/formatter"
	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/http/handler"
	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/http/middleware"
	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/http/middleware/compress"
	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/http/middleware/logging"
	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/storage/memory"
	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/storage/postgres"
	"github.com/StasMerzlyakov/go-metrics/internal/server/app"
	"github.com/StasMerzlyakov/go-metrics/internal/server/domain"
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

func main() {

	// Конфигурация
	srvConf, err := config.LoadServerConfig()
	if err != nil {
		panic(err)
	}

	// Создаем логгер
	logger, err := zap.NewDevelopment()
	if err != nil {
		// вызываем панику, если ошибка
		panic("cannot initialize zap")
	}
	defer logger.Sync()

	sugarLog := logger.Sugar()

	// Сборка сервера
	memStorage := memory.NewStorage()

	httpHandler := chi.NewMux()

	// мидлы
	mwList := createMiddleWareList(sugarLog)
	middleware.Add(httpHandler, mwList...)

	metricApp := app.NewMetrics(memStorage)
	handler.AddMetricOperations(httpHandler, metricApp, sugarLog)

	pgStorage := postgres.NewStorage(sugarLog, srvConf.DatabaseDSN)

	adminApp := app.NewAdminApp(sugarLog, pgStorage)
	handler.AddAdminOperations(httpHandler, adminApp, sugarLog)

	// бэкап
	backupFomratter := formatter.NewJSON(sugarLog, srvConf.FileStoragePath)
	backup := app.NewBackup(sugarLog, memStorage, backupFomratter)

	// проверяем - нужен ли синхронный бэкап
	doSyncBackup := srvConf.StoreInterval == 0

	if doSyncBackup {
		sugarLog.Warnf("NewMetricsServer", "msg", "backup work in sync mode")
		// синхронный бэкап реализован через мехинизм листенеров изменений
		//  (изменение данных может происходить и не только через http)
		metricApp.AddListener(func(ctx context.Context, m *domain.Metrics) error {
			return backup.DoBackUp(ctx)
		})
	}

	resources := []server.StartStopListener{pgStorage}

	var server Server = server.NewMetricsServer(srvConf,
		sugarLog,
		httpHandler,
		backup,
		resources...)

	// Запуск сервера
	ctx, cancelFn := context.WithCancel(context.Background())
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)

	server.Start(ctx)
	defer func() {
		cancelFn()
		shutdownCtx := context.TODO()
		server.Shutdown(shutdownCtx)
	}()
	<-exit
}
