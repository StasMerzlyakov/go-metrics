package domain

// Metrics используется для передачи данных о метриках при взаимодействии с сервисом сбора метрик.
type Metrics struct {
	ID    string     `json:"id"`              // имя метрики
	MType MetricType `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64     `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64   `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

type MetricType string

// Допустимые значения типа метрики
const (
	GaugeType   MetricType = "gauge"
	CounterType MetricType = "counter"
)
