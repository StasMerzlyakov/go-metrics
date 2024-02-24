package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/StasMerzlyakov/go-metrics/internal/config"
	"github.com/StasMerzlyakov/go-metrics/internal/server"
	"github.com/StasMerzlyakov/go-metrics/internal/server/storage"
	"go.uber.org/zap"
)

type Server interface {
	Start(ctx context.Context)
	WaitDone()
}

func main() {

	logger, err := zap.NewDevelopment()
	if err != nil {
		// вызываем панику, если ошибка
		panic("cannot initialize zap")
	}
	defer logger.Sync()

	srvConf, err := config.LoadServerConfig()
	if err != nil {
		logger.Sugar().Fatalf("configuration error", "err", err.Error())
		panic("configuration error")
	}

	counterStorage := storage.NewMemoryInt64Storage()
	gougeStorage := storage.NewMemoryFloat64Storage()
	controller := server.NewMetricController(
		counterStorage,
		gougeStorage,
	)

	adapter := server.NewHttpAdapterHandler(controller)

	var server Server = server.CreateMeterServer(srvConf, logger, adapter)

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
