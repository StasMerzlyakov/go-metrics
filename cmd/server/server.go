package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/StasMerzlyakov/go-metrics/internal/config"
	"github.com/StasMerzlyakov/go-metrics/internal/server"
	"github.com/StasMerzlyakov/go-metrics/internal/server/middleware"
	"github.com/StasMerzlyakov/go-metrics/internal/server/storage"
	"go.uber.org/zap"
)

type Server interface {
	Start(ctx context.Context)
	WaitDone()
}

func createMWList(log *zap.SugaredLogger) []middleware.MWHandlerFn {
	logResponsMW := middleware.NewLogResponseMW(log)
	logRequestMW := middleware.NewLogRequestMW(log)

	return []middleware.MWHandlerFn{
		logRequestMW,
		logResponsMW,
	}
}

func main() {

	// Конфигурация
	srvConf := config.LoadServerConfig()

	// Сборка сервера
	counterStorage := storage.NewMemoryInt64Storage()
	gougeStorage := storage.NewMemoryFloat64Storage()
	controller := server.NewMetricController(
		counterStorage,
		gougeStorage,
	)

	adapter := server.NewHttpAdapterHandler(controller, srvConf.Log)

	mwList := createMWList(srvConf.Log)
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
