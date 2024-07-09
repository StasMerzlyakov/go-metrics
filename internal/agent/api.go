package agent

import "context"

//go:generate mockgen -destination "./generated_mocks_test.go" -package ${GOPACKAGE}_test . ResultSender,MetricStorage,Logger
type ResultSender interface {
	SendMetrics(ctx context.Context, metrics []Metrics) error
	Stop()
}

type MetricStorage interface {
	Refresh() error
	GetMetrics() []Metrics
}

type Logger interface {
	Infow(msg string, keysAndValues ...any)
}
