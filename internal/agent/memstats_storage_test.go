package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemStatsSource(t *testing.T) {
	mm := &memStatsSource{}
	expectedKeys := map[string]MetricType{
		"Alloc":         GaugeType,
		"BuckHashSys":   GaugeType,
		"Frees":         GaugeType,
		"GCCPUFraction": GaugeType,
		"GCSys":         GaugeType,
		"HeapAlloc":     GaugeType,
		"HeapIdle":      GaugeType,
		"HeapInuse":     GaugeType,
		"HeapObjects":   GaugeType,
		"HeapReleased":  GaugeType,
		"HeapSys":       GaugeType,
		"LastGC":        GaugeType,
		"Lookups":       GaugeType,
		"MCacheInuse":   GaugeType,
		"MCacheSys":     GaugeType,
		"MSpanInuse":    GaugeType,
		"MSpanSys":      GaugeType,
		"Mallocs":       GaugeType,
		"NextGC":        GaugeType,
		"NumForcedGC":   GaugeType,
		"NumGC":         GaugeType,
		"OtherSys":      GaugeType,
		"PauseTotalNs":  GaugeType,
		"StackInuse":    GaugeType,
		"StackSys":      GaugeType,
		"Sys":           GaugeType,
		"TotalAlloc":    GaugeType,
		"RandomValue":   GaugeType,
		"PoolCount":     CounterType,
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

	var poolCount string

	for _, metric := range metrics {
		mType, ok := expectedKeys[metric.Name]
		require.Truef(t, ok, "pollMetrics doesn't contain key %v", metric.Name)
		require.Equalf(t, mType, metric.Type, "pollMetrics contain key %v with different type", metric.Name)

		if metric.Name == "PoolCount" {
			poolCount = metric.Value
		}

	}

	assert.Equal(t, "4", poolCount)
}
