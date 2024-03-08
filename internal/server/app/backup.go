package app

import (
	"context"
	"errors"
	"os"

	"github.com/StasMerzlyakov/go-metrics/internal/server/domain"
	"go.uber.org/zap"
)

//go:generate mockgen -destination "./mocks/$GOFILE" -package mocks . AllMetricsStorage,BackupFormatter
type AllMetricsStorage interface {
	SetAllMetrics(ctx context.Context, in []domain.Metrics) error
	GetAllMetrics(ctx context.Context) ([]domain.Metrics, error)
}

type BackupFormatter interface {
	Write(ctx context.Context, in []domain.Metrics) error
	Read(ctx context.Context) ([]domain.Metrics, error)
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

func (bU *backUper) RestoreBackUp(ctx context.Context) error {
	metrics, err := bU.formatter.Read(ctx)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		panic(err)
	}

	return bU.storage.SetAllMetrics(ctx, metrics)
}

func (bU *backUper) DoBackUp(ctx context.Context) error {

	metrics, err := bU.storage.GetAllMetrics(ctx)
	if err != nil {
		return err
	}
	err = bU.formatter.Write(ctx, metrics)
	if err != nil {
		return err
	}
	bU.suga.Infow("DoBackUp", "status", "ok", "msg", "backup is done")
	return nil
}
