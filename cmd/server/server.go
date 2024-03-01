package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/StasMerzlyakov/go-metrics/internal/config"
	"github.com/StasMerzlyakov/go-metrics/internal/server"
	"github.com/StasMerzlyakov/go-metrics/internal/server/controller/http/handler"
	"github.com/StasMerzlyakov/go-metrics/internal/server/controller/http/middleware/compress"
	"github.com/StasMerzlyakov/go-metrics/internal/server/controller/http/middleware/logging"
	backuper "github.com/StasMerzlyakov/go-metrics/internal/server/controller/timer/backup"
	"github.com/StasMerzlyakov/go-metrics/internal/server/controller/timer/backup/formatter"
	"github.com/StasMerzlyakov/go-metrics/internal/server/storage/memory"
	"github.com/StasMerzlyakov/go-metrics/internal/server/usecase"
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

	sugarLog := logger.Sugar()

	// Сборка сервера
	storage := memory.NewStorage()

	backupFomratter := formatter.NewJson(sugarLog, srvConf.FileStoragePath)
	backuper := backuper.New(sugarLog, srvConf, storage, backupFomratter)

	usecase := usecase.NewMetricUseCase(storage)

	if srvConf.StoreIntervalSec == 0 {
		// бэкапер прячу в usecase а не допустим в мидлы из соображений что изменение данных может инициализироваться из разных адаптеров
		usecase.SetSyncBackUper(backuper)
	}

	if srvConf.Restore {
		backuper.RestoreBackUp()
	}

	mwList := createMiddleWareList(sugarLog)
	httpHandler := handler.NewHTTP(usecase, sugarLog, mwList...)

	var server Server = server.NewMetricsServer(srvConf, sugarLog, httpHandler, backuper)

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
