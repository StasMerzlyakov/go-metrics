package server

import (
	"fmt"

	"github.com/StasMerzlyakov/go-metrics/internal/storage"
)

type MetricController interface {
	GetAllMetrics() MetricModel
	GetCounter(name string) (int64, bool)
	GetGaguge(name string) (float64, bool)
	AddCounter(name string, value int64)
	SetGauge(name string, value float64)
}

func NewMetricController(counterStorage storage.MetricsStorage[int64],
	gaugeStorage storage.MetricsStorage[float64]) MetricController {
	return &metricController{
		counterStorage: counterStorage,
		gaugeStorage:   gaugeStorage,
	}
}

type MetricsData struct {
	Type  string
	Name  string
	Value string
}

type MetricModel struct {
	Items []MetricsData
}

type metricController struct {
	counterStorage storage.MetricsStorage[int64]
	gaugeStorage   storage.MetricsStorage[float64]
}

func (mc *metricController) GetAllMetrics() MetricModel {
	items := MetricModel{}
	for _, k := range mc.counterStorage.Keys() {
		v, _ := mc.counterStorage.Get(k)
		items.Items = append(items.Items, MetricsData{
			"counter",
			k,
			fmt.Sprintf("%v", v),
		})
	}

	for _, k := range mc.gaugeStorage.Keys() {
		v, _ := mc.gaugeStorage.Get(k)
		items.Items = append(items.Items, MetricsData{
			"counter",
			k,
			fmt.Sprintf("%v", v),
		})
	}
	return items
}

func (mc *metricController) GetCounter(name string) (int64, bool) {
	return mc.counterStorage.Get(name)
}

func (mc *metricController) GetGaguge(name string) (float64, bool) {
	return mc.gaugeStorage.Get(name)
}

func (mc *metricController) AddCounter(name string, value int64) {
	mc.counterStorage.Add(name, value)
}

func (mc *metricController) SetGauge(name string, value float64) {
	mc.gaugeStorage.Set(name, value)
}
