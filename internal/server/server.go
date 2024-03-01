package server

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/StasMerzlyakov/go-metrics/internal/config"
	"github.com/StasMerzlyakov/go-metrics/internal/server/domain"
	"github.com/sirupsen/logrus"
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

	// restore backup
	if backUper != nil && config.Restore {
		if err := backUper.RestoreBackUp(); err != nil {
			panic(err)
		}
	}

	doSyncBackup := config.BackupStoreIntervalSec == 0
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
		backaupStoreIntervalSec: config.BackupStoreIntervalSec,
		sugar:                   sugar,
	}
}

type meterServer struct {
	sugar                   *zap.SugaredLogger
	srv                     *http.Server
	wg                      sync.WaitGroup
	backuper                Backuper
	startContext            context.Context
	backaupStoreIntervalSec uint
	doSyncBackup            bool
}

func (s *meterServer) runBackup(ctx context.Context) error {
	for {
		storeInterval := time.Duration(s.backaupStoreIntervalSec) * time.Second
		for {
			select {
			case <-ctx.Done():
				logrus.Info("Run", "msg", "backup finished")
				return nil

			case <-time.After(storeInterval):
				s.backuper.DoBackUp()
			}
		}
	}
}

func (s *meterServer) Start(startContext context.Context) {
	s.startContext = startContext
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := s.srv.ListenAndServe(); err != http.ErrServerClosed {
			s.sugar.Fatalw("srv.ListenAndServe", "msg", err.Error())
		}
	}()
	if s.backuper != nil && !s.doSyncBackup {
		go func() {
			defer s.wg.Done()

			if err := s.runBackup(startContext); err != nil {
				s.sugar.Fatalw("backup error", "msg", err.Error())
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
