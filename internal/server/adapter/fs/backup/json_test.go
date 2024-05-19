package backup_test

import (
	"context"
	"os"
	"reflect"
	"testing"

	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/fs/backup"
	"github.com/StasMerzlyakov/go-metrics/internal/server/domain"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func getLogger() *zap.SugaredLogger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		// вызываем панику, если ошибка
		panic("cannot initialize zap")
	}

	return logger.Sugar()
}

var toWrite = []domain.Metrics{
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

func TestJsonFormatter(t *testing.T) {

	logger := getLogger()
	domain.SetMainLogger(logger)
	ctx := context.Background()

	// Файл не указан
	jF := backup.NewJSON("")
	restored, err := jF.Read(ctx)
	require.True(t, len(restored) == 0)
	require.ErrorIs(t, os.ErrNotExist, err)

	err = jF.Write(ctx, nil)
	require.ErrorIs(t, os.ErrNotExist, err)

	// Проверяем сохранение и восстановление
	tmpDir := os.TempDir()
	file, err := os.CreateTemp(tmpDir, "json_backup_test*")

	require.NoError(t, err)

	defer os.Remove(file.Name())

	fileName := file.Name()
	jF = backup.NewJSON(fileName)
	restored, err = jF.Read(ctx)
	require.Error(t, err) // EOF
	require.True(t, len(restored) == 0)

	err = jF.Write(ctx, toWrite)
	require.NoError(t, err) // EOF

	restored, err = jF.Read(ctx)
	require.NoError(t, err)
	require.True(t, reflect.DeepEqual(toWrite, restored))
}
