package server

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/StasMerzlyakov/go-metrics/internal/config"
	"github.com/StasMerzlyakov/go-metrics/internal/server/domain"
	"go.uber.org/zap"
)

type Backuper interface {
	RestoreBackUp() error
	DoBackUp() error
}

type ChangeListenerHolder interface {
	AddListener(changeListener domain.ChangeListener)
}

func NewMetricsServer(
	config *config.ServerConfiguration,
	sugar *zap.SugaredLogger,
	httpHandler http.Handler,
	holder ChangeListenerHolder,
	backUper Backuper,
) *meterServer {

	sugar.Infow("ServerConfig", "config", config)

	// restore backup
	if backUper != nil && config.Restore {
		if err := backUper.RestoreBackUp(); err != nil {
			panic(err)
		}
	}

	doSyncBackup := config.StoreInterval == 0
	if doSyncBackup && backUper != nil {
		sugar.Warnw("NewMetricsServer", "msg", "backup work in sync mode")
		holder.AddListener(&backupSyncListener{
			backUper: backUper,
		})
	}

	return &meterServer{
		srv: &http.Server{
			Addr:        config.URL,
			Handler:     httpHandler,
			ReadTimeout: 0,
			IdleTimeout: 0,
		},
		doSyncBackup:            doSyncBackup,
		backaupStoreIntervalSec: config.StoreInterval,
		sugar:                   sugar,
		backUper:                backUper,
	}
}

type meterServer struct {
	sugar                   *zap.SugaredLogger
	srv                     *http.Server
	wg                      sync.WaitGroup
	backUper                Backuper
	startContext            context.Context
	backaupStoreIntervalSec uint
	doSyncBackup            bool
}

func (s *meterServer) ServeBackup(ctx context.Context) error {

	storeInterval := time.Duration(s.backaupStoreIntervalSec) * time.Second
	for {
		select {
		case <-ctx.Done():
			s.sugar.Infow("Run", "msg", "backup finished")
			return nil

		case <-time.After(storeInterval):
			if err := s.backUper.DoBackUp(); err != nil {
				s.sugar.Fatalw("DoBackUp", "msg", err.Error())
			}
		}
	}
}

func (s *meterServer) Start(startContext context.Context) {
	s.startContext = startContext
	s.wg.Add(2)
	go func() {
		defer s.wg.Done()
		if err := s.srv.ListenAndServe(); err != http.ErrServerClosed {
			s.sugar.Fatalw("srv.ListenAndServe", "msg", err.Error())
		}
	}()
	if s.backUper != nil && !s.doSyncBackup {
		go func() {
			defer s.wg.Done()

			if err := s.ServeBackup(startContext); err != nil {
				s.sugar.Fatalw("ServeBackup", "msg", err.Error())
			}
		}()
	}
	s.sugar.Infof("Server started")
}

func (s *meterServer) WaitDone() {
	s.srv.Shutdown(s.startContext)
	s.wg.Wait()
	s.sugar.Infof("WaitDone")
}

type backupSyncListener struct {
	backUper Backuper
}

func (bsl *backupSyncListener) Refresh(*domain.Metrics) error {
	return bsl.backUper.DoBackUp()
}
