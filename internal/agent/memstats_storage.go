package agent

import (
	"math/rand"
	"runtime"
	"sync/atomic"
)

func NewMemStatsStorage() *memStatsSource {
	return &memStatsSource{
		poolCounter:    0,
		memStatStorage: nil,
	}
}

type memStatsSource struct {
	poolCounter    int64
	memStatStorage map[string]float64
}

func (m *memStatsSource) Refresh() error {
	defer atomic.AddInt64(&m.poolCounter, 1)
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	m.memStatStorage = map[string]float64{
		"Alloc":         float64(memStats.Alloc),
		"BuckHashSys":   float64(memStats.BuckHashSys),
		"Frees":         float64(memStats.Frees),
		"GCCPUFraction": memStats.GCCPUFraction,
		"GCSys":         float64(memStats.GCSys),
		"HeapAlloc":     float64(memStats.HeapAlloc),
		"HeapIdle":      float64(memStats.HeapIdle),
		"HeapInuse":     float64(memStats.HeapInuse),
		"HeapObjects":   float64(memStats.HeapObjects),
		"HeapReleased":  float64(memStats.HeapReleased),
		"HeapSys":       float64(memStats.HeapSys),
		"LastGC":        float64(memStats.LastGC),
		"Lookups":       float64(memStats.Lookups),
		"MCacheInuse":   float64(memStats.MCacheInuse),
		"MCacheSys":     float64(memStats.MCacheSys),
		"MSpanInuse":    float64(memStats.MSpanInuse),
		"MSpanSys":      float64(memStats.MSpanSys),
		"Mallocs":       float64(memStats.Mallocs),
		"NextGC":        float64(memStats.NextGC),
		"NumForcedGC":   float64(memStats.NumForcedGC),
		"NumGC":         float64(memStats.NumGC),
		"OtherSys":      float64(memStats.OtherSys),
		"PauseTotalNs":  float64(memStats.PauseTotalNs),
		"StackInuse":    float64(memStats.StackInuse),
		"StackSys":      float64(memStats.StackSys),
		"Sys":           float64(memStats.Sys),
		"TotalAlloc":    float64(memStats.TotalAlloc),
		"RandomValue":   rand.Float64(),
	}
	return nil
}

func (m *memStatsSource) GetMetrics() []Metrics {
	var metrics []Metrics
	for k, v := range m.memStatStorage {
		value := v
		metrics = append(metrics, Metrics{
			ID:    k,
			MType: GaugeType,
			Value: &value,
		})
	}

	poolCount := m.poolCounter
	metrics = append(metrics, Metrics{
		ID:    "PollCount",
		MType: CounterType,
		Delta: &poolCount,
	})
	return metrics
}
