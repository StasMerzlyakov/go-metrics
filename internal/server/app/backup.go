package app

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/StasMerzlyakov/go-metrics/internal/server/domain"
)

func NewBackup(storage AllMetricsStorage, formatter BackupFormatter, checker MetricsChecker) *backUper {
	return &backUper{
		storage:   storage,
		formatter: formatter,
		checker:   checker,
	}
}

type backUper struct {
	storage   AllMetricsStorage
	formatter BackupFormatter
	checker   MetricsChecker
}

func (bU *backUper) RestoreBackUp(ctx context.Context) error {
	metrics, err := bU.formatter.Read(ctx)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			panic(err)
		}
		return nil
	}

	// Проверка данных
	for _, m := range metrics {
		err := bU.checker.CheckMetrics(&m)
		if err != nil {
			return err
		}
	}

	return bU.storage.SetAllMetrics(ctx, metrics)
}

func (bU *backUper) DoBackUp(ctx context.Context) error {
	logger := domain.GetMainLogger()
	action := domain.GetAction(1)
	metrics, err := bU.storage.GetAllMetrics(ctx)
	if err != nil {
		return fmt.Errorf("backup err - getAllMetrics return err: %w", err)
	}
	if len(metrics) == 0 {
		return nil
	}
	err = bU.formatter.Write(ctx, metrics)
	if err != nil {
		logger.Errorf(action, "error", fmt.Sprintf("backup error - %s", err.Error()))
		return fmt.Errorf("write backup err: %w", err)
	}
	return nil
}
