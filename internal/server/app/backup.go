package app

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/StasMerzlyakov/go-metrics/internal/server/domain"
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

func NewBackup(storage AllMetricsStorage, formatter BackupFormatter) *backUper {
	return &backUper{
		storage:   storage,
		formatter: formatter,
	}
}

type backUper struct {
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
	logger := domain.GetMainLogger()
	action := domain.GetAction(1)
	metrics, err := bU.storage.GetAllMetrics(ctx)
	if err != nil {
		return err
	}
	err = bU.formatter.Write(ctx, metrics)
	if err != nil {
		logger.Errorf(action, "error", fmt.Sprintf("backup error - %s", err.Error()))
		return err
	}
	return nil
}
