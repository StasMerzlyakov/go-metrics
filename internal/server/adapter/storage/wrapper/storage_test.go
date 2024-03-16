package wrapper_test

import (
	"context"
	"io"
	"testing"

	"github.com/StasMerzlyakov/go-metrics/internal/common/wrapper/retriable"
	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/storage/memory"
	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/storage/wrapper"
	"github.com/StasMerzlyakov/go-metrics/internal/server/domain"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestRetruableStorageStoreAndLoad(t *testing.T) {

	toLoad := []domain.Metrics{
		{MType: domain.CounterType, ID: "PollCount", Delta: domain.DeltaPtr(1)},
		{MType: domain.GaugeType, ID: "RandomValue", Value: domain.ValuePtr(1.123)},
		{MType: domain.GaugeType, ID: "Alloc", Value: domain.ValuePtr(1.123)},
		{MType: domain.GaugeType, ID: "BuckHashSys", Value: domain.ValuePtr(1.123)},
		{MType: domain.GaugeType, ID: "Frees", Value: domain.ValuePtr(1.123)},
		{MType: domain.GaugeType, ID: "GCCPUFraction", Value: domain.ValuePtr(1.123)},
		{MType: domain.GaugeType, ID: "GCSys", Value: domain.ValuePtr(1.123)},
		{MType: domain.GaugeType, ID: "HeapAlloc", Value: domain.ValuePtr(1.123)},
		{MType: domain.GaugeType, ID: "HeapIdle", Value: domain.ValuePtr(1.123)},
		{MType: domain.GaugeType, ID: "HeapInuse", Value: domain.ValuePtr(1.123)},
		{MType: domain.GaugeType, ID: "HeapObjects", Value: domain.ValuePtr(1.123)},
		{MType: domain.GaugeType, ID: "HeapReleased", Value: domain.ValuePtr(1.123)},
		{MType: domain.GaugeType, ID: "HeapSys", Value: domain.ValuePtr(1.123)},
		{MType: domain.GaugeType, ID: "LastGC", Value: domain.ValuePtr(1.123)},
		{MType: domain.GaugeType, ID: "Lookups", Value: domain.ValuePtr(1.123)},
		{MType: domain.GaugeType, ID: "MCacheInuse", Value: domain.ValuePtr(1.123)},
		{MType: domain.GaugeType, ID: "MCacheSys", Value: domain.ValuePtr(1.123)},
		{MType: domain.GaugeType, ID: "MSpanInuse", Value: domain.ValuePtr(1.123)},
		{MType: domain.GaugeType, ID: "MSpanSys", Value: domain.ValuePtr(1.123)},
		{MType: domain.GaugeType, ID: "Mallocs", Value: domain.ValuePtr(1.123)},
		{MType: domain.GaugeType, ID: "NextGC", Value: domain.ValuePtr(1.123)},
		{MType: domain.GaugeType, ID: "NumForcedGC", Value: domain.ValuePtr(1.123)},
		{MType: domain.GaugeType, ID: "NumGC", Value: domain.ValuePtr(1.123)},
		{MType: domain.GaugeType, ID: "OtherSys", Value: domain.ValuePtr(1.123)},
		{MType: domain.GaugeType, ID: "PauseTotalNs", Value: domain.ValuePtr(1.123)},
		{MType: domain.GaugeType, ID: "StackInuse", Value: domain.ValuePtr(1.123)},
		{MType: domain.GaugeType, ID: "StackSys", Value: domain.ValuePtr(1.123)},
		{MType: domain.GaugeType, ID: "Sys", Value: domain.ValuePtr(1.123)},
		{MType: domain.GaugeType, ID: "TotalAlloc", Value: domain.ValuePtr(1.123)},
	}

	ctx := context.Background()

	mStorage := memory.NewStorage()
	conf := retriable.DefaultConf(io.EOF)
	storage := wrapper.NewRetriableWrapper(conf, logger(), mStorage)

	out, err := storage.GetAllMetrics(ctx)
	require.NoError(t, err)
	require.True(t, len(out) == 0)

	err = storage.SetAllMetrics(ctx, toLoad)
	require.NoError(t, err)

	out, err = storage.GetAllMetrics(ctx)
	require.NoError(t, err)
	require.Equal(t, len(toLoad), len(out))
}

func TestRetruableStorageGaugeOperations(t *testing.T) {

	mStorage := memory.NewStorage()
	conf := retriable.DefaultConf(io.EOF)
	storage := wrapper.NewRetriableWrapper(conf, logger(), mStorage)

	GagueID := "NumGC"

	mConst := &domain.Metrics{
		ID:    GagueID,
		MType: domain.GaugeType,
		Value: domain.ValuePtr(2),
	}

	ctx := context.Background()

	ms, err := storage.Get(ctx, GagueID, domain.CounterType)
	require.NoError(t, err)
	require.Nil(t, ms)

	ms, err = storage.Get(ctx, GagueID, domain.GaugeType)
	require.NoError(t, err)
	require.Nil(t, ms)

	err = storage.Set(ctx, mConst)
	require.NoError(t, err)

	ms, err = storage.Get(ctx, GagueID, domain.GaugeType)
	require.NoError(t, err)
	require.NotNil(t, ms)
	require.Equal(t, ms.ID, GagueID)
	require.Equal(t, ms.MType, domain.GaugeType)
	require.NotNil(t, ms.Value)
	require.Equal(t, float64(2), *ms.Value)
	require.Nil(t, ms.Delta)

	ms, err = storage.Get(ctx, GagueID, domain.CounterType)
	require.NoError(t, err)
	require.Nil(t, ms)

	err = storage.Set(ctx, mConst)
	require.NoError(t, err)
	ms, err = storage.Get(ctx, GagueID, domain.GaugeType)
	require.NoError(t, err)
	require.NotNil(t, ms)
	require.Equal(t, ms.ID, GagueID)
	require.Equal(t, ms.MType, domain.GaugeType)
	require.NotNil(t, ms.Value)
	require.Equal(t, float64(2), *ms.Value)
	require.Nil(t, ms.Delta)

	err = storage.Add(ctx, mConst)
	require.NoError(t, err)
	ms, err = storage.Get(ctx, GagueID, domain.GaugeType)
	require.NoError(t, err)
	require.NotNil(t, ms)
	require.Equal(t, ms.ID, GagueID)
	require.Equal(t, ms.MType, domain.GaugeType)
	require.NotNil(t, ms.Value)
	require.Equal(t, float64(4), *ms.Value)
	require.Nil(t, ms.Delta)
}

func TestRetruableStorageCounterOperations(t *testing.T) {

	mStorage := memory.NewStorage()
	conf := retriable.DefaultConf(io.EOF)
	storage := wrapper.NewRetriableWrapper(conf, logger(), mStorage)

	CounterID := "PollCount"

	mConst := &domain.Metrics{
		ID:    CounterID,
		MType: domain.CounterType,
		Delta: domain.DeltaPtr(2),
	}

	ctx := context.Background()

	ms, err := storage.Get(ctx, CounterID, domain.CounterType)
	require.NoError(t, err)
	require.Nil(t, ms)

	ms, err = storage.Get(ctx, CounterID, domain.GaugeType)
	require.NoError(t, err)
	require.Nil(t, ms)

	err = storage.Set(ctx, mConst)
	require.NoError(t, err)

	ms, err = storage.Get(ctx, CounterID, domain.CounterType)
	require.NoError(t, err)
	require.NotNil(t, ms)
	require.Equal(t, ms.ID, CounterID)
	require.Equal(t, ms.MType, domain.CounterType)
	require.NotNil(t, ms.Delta)
	require.Equal(t, int64(2), *ms.Delta)
	require.Nil(t, ms.Value)

	ms, err = storage.Get(ctx, CounterID, domain.GaugeType)
	require.NoError(t, err)
	require.Nil(t, ms)

	err = storage.Set(ctx, mConst)
	require.NoError(t, err)
	ms, err = storage.Get(ctx, CounterID, domain.CounterType)
	require.NoError(t, err)
	require.NotNil(t, ms)
	require.Equal(t, ms.ID, CounterID)
	require.Equal(t, ms.MType, domain.CounterType)
	require.NotNil(t, ms.Delta)
	require.Equal(t, int64(2), *ms.Delta)
	require.Nil(t, ms.Value)

	err = storage.Add(ctx, mConst)
	require.NoError(t, err)
	ms, err = storage.Get(ctx, CounterID, domain.CounterType)
	require.NoError(t, err)
	require.NotNil(t, ms)
	require.Equal(t, ms.ID, CounterID)
	require.Equal(t, ms.MType, domain.CounterType)
	require.NotNil(t, ms.Delta)
	require.Equal(t, int64(4), *ms.Delta)
	require.Nil(t, ms.Value)
}

func TestAddMetrics(t *testing.T) {

	t.Run("gague", func(t *testing.T) {
		mStorage := memory.NewStorage()
		conf := retriable.DefaultConf(io.EOF)
		storage := wrapper.NewRetriableWrapper(conf, logger(), mStorage)

		GagueID := "NumGC"

		mConst := &domain.Metrics{
			ID:    GagueID,
			MType: domain.GaugeType,
			Value: domain.ValuePtr(2),
		}

		ctx := context.Background()

		ms, err := storage.Get(ctx, GagueID, domain.GaugeType)
		require.NoError(t, err)
		require.Nil(t, ms)

		err = storage.AddMetrics(ctx, []domain.Metrics{*mConst})
		require.NoError(t, err)

		ms, err = storage.Get(ctx, GagueID, domain.GaugeType)
		require.NoError(t, err)
		require.Equal(t, *mConst.Value, *ms.Value)

		err = storage.AddMetrics(ctx, []domain.Metrics{*mConst})
		require.NoError(t, err)

		ms, err = storage.Get(ctx, GagueID, domain.GaugeType)
		require.NoError(t, err)
		require.Equal(t, 2**mConst.Value, *ms.Value)
	})

	t.Run("counter", func(t *testing.T) {
		mStorage := memory.NewStorage()
		conf := retriable.DefaultConf(io.EOF)
		storage := wrapper.NewRetriableWrapper(conf, logger(), mStorage)

		CounterID := "NumGC"

		mConst := &domain.Metrics{
			ID:    CounterID,
			MType: domain.CounterType,
			Delta: domain.DeltaPtr(2),
		}

		ctx := context.Background()

		ms, err := storage.Get(ctx, CounterID, domain.CounterType)
		require.NoError(t, err)
		require.Nil(t, ms)

		err = storage.AddMetrics(ctx, []domain.Metrics{*mConst})
		require.NoError(t, err)

		ms, err = storage.Get(ctx, CounterID, domain.CounterType)
		require.NoError(t, err)
		require.Equal(t, *mConst.Delta, *ms.Delta)

		err = storage.AddMetrics(ctx, []domain.Metrics{*mConst})
		require.NoError(t, err)

		ms, err = storage.Get(ctx, CounterID, domain.CounterType)
		require.NoError(t, err)
		require.Equal(t, 2**mConst.Delta, *ms.Delta)
	})
}

func TestSetMetrics(t *testing.T) {

	t.Run("gague", func(t *testing.T) {
		mStorage := memory.NewStorage()
		conf := retriable.DefaultConf(io.EOF)
		storage := wrapper.NewRetriableWrapper(conf, logger(), mStorage)

		GagueID := "NumGC"

		mConst := &domain.Metrics{
			ID:    GagueID,
			MType: domain.GaugeType,
			Value: domain.ValuePtr(2),
		}

		ctx := context.Background()

		ms, err := storage.Get(ctx, GagueID, domain.GaugeType)
		require.NoError(t, err)
		require.Nil(t, ms)

		err = storage.SetMetrics(ctx, []domain.Metrics{*mConst})
		require.NoError(t, err)

		ms, err = storage.Get(ctx, GagueID, domain.GaugeType)
		require.NoError(t, err)
		require.Equal(t, *mConst.Value, *ms.Value)

		err = storage.SetMetrics(ctx, []domain.Metrics{*mConst})
		require.NoError(t, err)

		ms, err = storage.Get(ctx, GagueID, domain.GaugeType)
		require.NoError(t, err)
		require.Equal(t, *mConst.Value, *ms.Value)
	})

	t.Run("counter", func(t *testing.T) {
		mStorage := memory.NewStorage()
		conf := retriable.DefaultConf(io.EOF)
		storage := wrapper.NewRetriableWrapper(conf, logger(), mStorage)

		CounterID := "NumGC"

		mConst := &domain.Metrics{
			ID:    CounterID,
			MType: domain.CounterType,
			Delta: domain.DeltaPtr(2),
		}

		ctx := context.Background()

		ms, err := storage.Get(ctx, CounterID, domain.CounterType)
		require.NoError(t, err)
		require.Nil(t, ms)

		err = storage.SetMetrics(ctx, []domain.Metrics{*mConst})
		require.NoError(t, err)

		ms, err = storage.Get(ctx, CounterID, domain.CounterType)
		require.NoError(t, err)
		require.Equal(t, *mConst.Delta, *ms.Delta)

		err = storage.SetMetrics(ctx, []domain.Metrics{*mConst})
		require.NoError(t, err)

		ms, err = storage.Get(ctx, CounterID, domain.CounterType)
		require.NoError(t, err)
		require.Equal(t, *mConst.Delta, *ms.Delta)
	})
}

func logger() *zap.SugaredLogger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		// вызываем панику, если ошибка
		panic("cannot initialize zap")
	}
	defer logger.Sync()

	log := logger.Sugar()
	return log
}
