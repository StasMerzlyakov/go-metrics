package agent

type MetricType string

const (
	GaugeType   MetricType = "Gauge"
	CounterType MetricType = "Counter"
)

type Metric struct {
	Name  string
	Type  MetricType
	Value string
}
