package server

type MetricsData struct {
	Type  MetricType
	Name  string
	Value string
}

type MetricModel struct {
	Items []MetricsData
}

type MetricType string

const (
	GaugeType   MetricType = "gauge"
	CounterType MetricType = "counter"
)
