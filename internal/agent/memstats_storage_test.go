package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemStatsSource(t *testing.T) {
	mm := &memStatsSource{}
	expectedKeys := map[string]MetricType{
		"Alloc":           GaugeType,
		"BuckHashSys":     GaugeType,
		"Frees":           GaugeType,
		"GCCPUFraction":   GaugeType,
		"GCSys":           GaugeType,
		"HeapAlloc":       GaugeType,
		"HeapIdle":        GaugeType,
		"HeapInuse":       GaugeType,
		"HeapObjects":     GaugeType,
		"HeapReleased":    GaugeType,
		"HeapSys":         GaugeType,
		"LastGC":          GaugeType,
		"Lookups":         GaugeType,
		"MCacheInuse":     GaugeType,
		"MCacheSys":       GaugeType,
		"MSpanInuse":      GaugeType,
		"MSpanSys":        GaugeType,
		"Mallocs":         GaugeType,
		"NextGC":          GaugeType,
		"NumForcedGC":     GaugeType,
		"NumGC":           GaugeType,
		"OtherSys":        GaugeType,
		"PauseTotalNs":    GaugeType,
		"StackInuse":      GaugeType,
		"StackSys":        GaugeType,
		"Sys":             GaugeType,
		"TotalAlloc":      GaugeType,
		"RandomValue":     GaugeType,
		"TotalMemory":     GaugeType,
		"FreeMemory":      GaugeType,
		"CPUutilization1": GaugeType,
		"PollCount":       CounterType,
	}
	err := mm.Refresh()
	require.NoError(t, err)
	err = mm.Refresh()
	require.NoError(t, err)
	err = mm.Refresh()
	require.NoError(t, err)
	err = mm.Refresh()
	require.NoError(t, err)

	metrics := mm.GetMetrics()

	require.Equal(t, len(expectedKeys), len(metrics))

	var pollCount int64

	for _, metric := range metrics {
		mType, ok := expectedKeys[metric.ID]
		require.Truef(t, ok, "pollMetrics doesn't contain key %v", metric.ID)
		require.Equalf(t, mType, metric.MType, "pollMetrics contain key %v with different type", metric.ID)

		if metric.ID == "PollCount" {
			pollCount = *metric.Delta
		}

	}

	assert.Equal(t, int64(4), pollCount)
}
