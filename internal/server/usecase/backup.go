package usecase

import (
	"errors"
	"os"

	"github.com/StasMerzlyakov/go-metrics/internal/server/domain"
	"go.uber.org/zap"
)

type AllMetricsStorage interface {
	SetAllMetrics(in []domain.Metrics) error
	GetAllMetrics() ([]domain.Metrics, error)
}

type BackupFormatter interface {
	Write([]domain.Metrics) error
	Read() ([]domain.Metrics, error)
}

func NewBackup(suga *zap.SugaredLogger, storage AllMetricsStorage, formatter BackupFormatter) *backUper {
	return &backUper{
		suga:      suga,
		storage:   storage,
		formatter: formatter,
	}
}

type backUper struct {
	suga      *zap.SugaredLogger
	storage   AllMetricsStorage
	formatter BackupFormatter
}

func (bU *backUper) RestoreBackUp() error {
	metrics, err := bU.formatter.Read()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		panic(err)
	}

	return bU.storage.SetAllMetrics(metrics)
}

func (bU *backUper) DoBackUp() error {

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
}
