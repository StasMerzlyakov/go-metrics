package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/StasMerzlyakov/go-metrics/internal/config"
	"github.com/StasMerzlyakov/go-metrics/internal/server"
	"github.com/StasMerzlyakov/go-metrics/internal/server/middleware/compress"
	"github.com/StasMerzlyakov/go-metrics/internal/server/middleware/logging"
	"github.com/StasMerzlyakov/go-metrics/internal/server/storage"
	"go.uber.org/zap"
)

type Server interface {
	Start(ctx context.Context)
	WaitDone()
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

	log := logger.Sugar()

	// Сборка сервера
	counterStorage := storage.NewMemoryInt64Storage()
	gougeStorage := storage.NewMemoryFloat64Storage()
	controller := server.NewMetricController(
		counterStorage,
		gougeStorage,
	)

	adapter := server.NewHTTPAdapterHandler(controller, log)

	mwList := createMiddleWareList(log)
	var server Server = server.CreateMeterServer(srvConf, adapter, mwList...)

	// Запуск сервера
	ctx, cancelFn := context.WithCancel(context.Background())
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)

	server.Start(ctx)
	defer func() {
		cancelFn()
		server.WaitDone()
	}()
	<-exit
}
