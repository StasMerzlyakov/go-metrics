package server

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/StasMerzlyakov/go-metrics/internal/config"
	"go.uber.org/zap"
)

type Backuper interface {
	RestoreBackUp(ctx context.Context) error
	DoBackUp(ctx context.Context) error
}

type StartStopListener interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

func NewMetricsServer(
	config *config.ServerConfiguration,
	sugar *zap.SugaredLogger,
	httpHandler http.Handler,
	backUper Backuper,
	resources ...StartStopListener,
) *meterServer {

	sugar.Infow("ServerConfig", "config", config)

	var startStopListeners []StartStopListener

	startStopListeners = append(startStopListeners, resources...)

	// проверяем - нужен ли синхронный бэкап
	doSyncBackup := config.StoreInterval == 0

	if backUper != nil && !doSyncBackup {
		backupListener := &backupListener{
			backUper:                backUper,
			backaupStoreIntervalSec: config.StoreInterval,
			sugar:                   sugar,
			restore:                 config.Restore,
		}
		startStopListeners = append(startStopListeners, backupListener)
	}

	srvListener := &httpServerListener{
		srv: &http.Server{
			Addr:        config.URL,
			Handler:     httpHandler,
			ReadTimeout: 0,
			IdleTimeout: 0,
		},
		sugar: sugar,
	}

	startStopListeners = append(startStopListeners, srvListener)

	return &meterServer{
		startStopListeners: startStopListeners,
		sugar:              sugar,
	}
}

type meterServer struct {
	sugar              *zap.SugaredLogger
	startStopListeners []StartStopListener
}

func (s *meterServer) Start(startContext context.Context) {

	for _, lst := range s.startStopListeners {
		startStopListener := lst
		go func() {
			if err := startStopListener.Start(startContext); err != nil {
				panic(err)
			}
		}()
	}

	s.sugar.Infow("Server started")
}

func (s *meterServer) Shutdown(ctx context.Context) {

	var wg sync.WaitGroup

	// TODO - подумать над управлением зависимостями между листенерами и порядком вызова
	for _, lst := range s.startStopListeners {
		startStopListener := lst
		wg.Add(1)
		go func() {
			defer wg.Done()
			startStopListener.Stop(ctx)
		}()

	}

	wg.Wait()
	s.sugar.Infow("Shutdown", "msg", "complete")
}

type backupListener struct {
	backUper                Backuper
	backaupStoreIntervalSec uint
	sugar                   *zap.SugaredLogger
	restore                 bool
}

func (backRes *backupListener) Start(ctx context.Context) error {

	// restore backup
	if backRes.restore {
		if err := backRes.backUper.RestoreBackUp(ctx); err != nil {
			panic(err)
		}
	}

	storeInterval := time.Duration(backRes.backaupStoreIntervalSec) * time.Second
	for {
		select {
		case <-ctx.Done():
			backRes.sugar.Infow("Run", "msg", "backup finished")
			return nil

		case <-time.After(storeInterval):
			if err := backRes.backUper.DoBackUp(ctx); err != nil {
				backRes.sugar.Fatalw("DoBackUp", "msg", err.Error())
			}
		}
	}
}

func (backRes *backupListener) Stop(ctx context.Context) error {
	return nil
}

type httpServerListener struct {
	srv   *http.Server
	sugar *zap.SugaredLogger
}

func (hList *httpServerListener) Start(ctx context.Context) error {
	if err := hList.srv.ListenAndServe(); err != http.ErrServerClosed {
		hList.sugar.Fatalw("srv.ListenAndServe", "msg", err.Error())
		return err
	}
	return nil
}

func (hList *httpServerListener) Stop(ctx context.Context) error {
	return hList.srv.Shutdown(ctx)
}
