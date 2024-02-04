package agent

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRuntimeMetrics(t *testing.T) {
	rm := &runtimeMetrics{}
	expectedKeys := []string{
		"Alloc",
		"BuckHashSys",
		"Frees",
		"GCCPUFraction",
		"GCSys",
		"HeapAlloc",
		"HeapIdle",
		"HeapInuse",
		"HeapObjects",
		"HeapReleased",
		"HeapSys",
		"LastGC",
		"Lookups",
		"MCacheInuse",
		"MCacheSys",
		"MSpanInuse",
		"MSpanSys",
		"Mallocs",
		"NextGC",
		"NumForcedGC",
		"NumGC",
		"OtherSys",
		"PauseTotalNs",
		"StackInuse",
		"StackSys",
		"Sys",
		"TotalAlloc",
		"RandomValue",
	}
	result := rm.PollMetrics()
	assert.Equal(t, len(expectedKeys), len(result))
	for _, expectedKey := range expectedKeys {
		_, ok := result[expectedKey]
		require.Truef(t, ok, "pollMetrics doesn't contain key %v", expectedKey)
	}

	rm.PollMetrics()
	rm.PollMetrics()
	rm.PollMetrics()
	assert.Equal(t, int64(4), rm.PollCount())
}
