package app_test

import (
	"context"
	"errors"
	"os"
	reflect "reflect"
	"testing"

	"github.com/StasMerzlyakov/go-metrics/internal/server/app"
	"github.com/StasMerzlyakov/go-metrics/internal/server/domain"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestBackupRestoreBackUp(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	frmt := NewMockBackupFormatter(ctrl)

	storage := NewMockAllMetricsStorage(ctrl)

	checker := NewMockMetricsChecker(ctrl)

	checker.EXPECT().CheckMetrics(gomock.Any()).DoAndReturn(func(m *domain.Metrics) error {
		return nil
	}).AnyTimes()

	t.Run("test ok", func(t *testing.T) {

		var m = []domain.Metrics{
			{
				ID:    "a1s_asd1_1",
				MType: domain.CounterType,
				Delta: domain.DeltaPtr(1),
			},
			{
				ID:    "a1s_asd1_2",
				MType: domain.GaugeType,
				Value: domain.ValuePtr(1),
			},
		}

		frmt.EXPECT().Read(gomock.Any()).Return(m, nil).Times(1)

		storage.EXPECT().SetAllMetrics(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, in []domain.Metrics) error {
				require.True(t, reflect.DeepEqual(m, in))
				return nil
			},
		).Times(1)

		backUper := app.NewBackup(storage, frmt, checker)
		backUper.RestoreBackUp(context.Background())
	})

	t.Run("test no metrics", func(t *testing.T) {

		frmt.EXPECT().Read(gomock.Any()).Return(nil, os.ErrNotExist).Times(1)

		storage.EXPECT().SetAllMetrics(gomock.Any(), gomock.Any()).Times(0)

		backUper := app.NewBackup(storage, frmt, checker)
		backUper.RestoreBackUp(context.Background())
	})

	t.Run("test panic", func(t *testing.T) {

		defer func() {
			if r := recover(); r == nil {
				t.Errorf("The code did not panic")
			}
		}()

		frmt.EXPECT().Read(gomock.Any()).Return(nil, errors.New("Any err")).Times(1)

		storage.EXPECT().SetAllMetrics(gomock.Any(), gomock.Any()).Times(0)

		backUper := app.NewBackup(storage, frmt, checker)
		backUper.RestoreBackUp(context.Background())
	})
}

func TestBackupDoBackUp(t *testing.T) {

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	defer logger.Sync()

	suga := logger.Sugar()
	domain.SetMainLogger(suga)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	frmt := NewMockBackupFormatter(ctrl)

	storage := NewMockAllMetricsStorage(ctrl)

	checker := NewMockMetricsChecker(ctrl)

	checker.EXPECT().CheckMetrics(gomock.Any()).DoAndReturn(func(m *domain.Metrics) error {
		return nil
	}).AnyTimes()

	t.Run("test err", func(t *testing.T) {

		expErr := errors.New("test error")

		storage.EXPECT().GetAllMetrics(gomock.Any()).Return(nil, expErr).Times(1)

		frmt.EXPECT().Write(gomock.Any(), gomock.Any()).Times(0)

		backUper := app.NewBackup(storage, frmt, checker)
		err := backUper.DoBackUp(context.Background())
		assert.ErrorIs(t, err, expErr)

	})

	t.Run("no data", func(t *testing.T) {

		storage.EXPECT().GetAllMetrics(gomock.Any()).Return([]domain.Metrics{}, nil).Times(1)

		frmt.EXPECT().Write(gomock.Any(), gomock.Any()).Times(0)

		backUper := app.NewBackup(storage, frmt, checker)
		err := backUper.DoBackUp(context.Background())
		assert.NoError(t, err)
	})

	t.Run("metrics exists", func(t *testing.T) {

		var m = []domain.Metrics{
			{
				ID:    "a1s_asd1_1",
				MType: domain.CounterType,
				Delta: domain.DeltaPtr(1),
			},
			{
				ID:    "a1s_asd1_2",
				MType: domain.GaugeType,
				Value: domain.ValuePtr(1),
			},
		}

		storage.EXPECT().GetAllMetrics(gomock.Any()).Return(m, nil).Times(1)

		frmt.EXPECT().Write(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, in []domain.Metrics) error {
			require.True(t, reflect.DeepEqual(m, in))
			return nil
		}).Times(1)

		backUper := app.NewBackup(storage, frmt, checker)
		err := backUper.DoBackUp(context.Background())
		assert.NoError(t, err)
	})

	t.Run("write error", func(t *testing.T) {

		var m = []domain.Metrics{
			{
				ID:    "a1s_asd1_1",
				MType: domain.CounterType,
				Delta: domain.DeltaPtr(1),
			},
			{
				ID:    "a1s_asd1_2",
				MType: domain.GaugeType,
				Value: domain.ValuePtr(1),
			},
		}

		storage.EXPECT().GetAllMetrics(gomock.Any()).Return(m, nil).Times(1)

		expErr := errors.New("test error")

		frmt.EXPECT().Write(gomock.Any(), gomock.Any()).Return(expErr).Times(1)

		backUper := app.NewBackup(storage, frmt, checker)
		err := backUper.DoBackUp(context.Background())
		assert.ErrorIs(t, err, expErr)
	})

}
