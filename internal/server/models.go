package server

// Обмен http<->http_adapter
type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

// Обмен http_adapter <-> metric_controller
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
