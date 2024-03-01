package backuper

import (
	"context"
	"errors"
	"os"
	"sync"
	"time"

	"github.com/StasMerzlyakov/go-metrics/internal/config"
	"github.com/StasMerzlyakov/go-metrics/internal/server/domain"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
)

type Storage interface {
	SetAllMetrics(in []domain.Metrics) error
	GetAllMetrics() ([]domain.Metrics, error)
}

type BackupFormatter interface {
	Write([]domain.Metrics) error
	Read() ([]domain.Metrics, error)
}

func New(suga *zap.SugaredLogger, config *config.ServerConfiguration, storage Storage, formatter BackupFormatter) *backUper {
	doBackUp := config.FileStoragePath != ""
	backUper := &backUper{
		storage:          storage,
		suga:             suga,
		formatter:        formatter,
		storeIntervalSec: config.StoreIntervalSec,
		doBackup:         doBackUp,
	}

	if config.Restore {
		backUper.RestoreBackUp()
	}

	return backUper
}

type backUper struct {
	storage          Storage
	suga             *zap.SugaredLogger
	formatter        BackupFormatter
	wg               sync.WaitGroup
	storeIntervalSec uint
	doBackup         bool
}

func (bU *backUper) RestoreBackUp() error {
	metrics, err := bU.formatter.Read()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		panic(err)
	}

	bU.storage.SetAllMetrics(metrics)
	return nil
}

func (bU *backUper) DoBackUp() error {
	if bU.doBackup {
		metrics, err := bU.storage.GetAllMetrics()
		if err != nil {
			return err
		}
		err = bU.formatter.Write(metrics)
		if err != nil {
			return err
		}
		bU.suga.Infow("DoBackUp", "status", "ok", "msg", "backup is done")
		return nil
	} else {
		bU.suga.Warnw("DoBackUp", "status", "ok", "msg", "backup file is not specified")
	}
	return nil
}

func (bU *backUper) Run(ctx context.Context) error {
	bU.wg.Add(1)
	defer bU.wg.Done()

	if bU.doBackup && bU.storeIntervalSec > 0 {
		storeInterval := time.Duration(bU.storeIntervalSec) * time.Second
		for {
			select {
			case <-ctx.Done():
				logrus.Info("Run", "msg", "backup finished")
				return nil

			case <-time.After(storeInterval):
				bU.DoBackUp()
			}
		}
	}
	return nil
}

func (bU *backUper) WaitDone() {
	bU.wg.Wait()
}
