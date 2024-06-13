package agent

import "context"

//go:generate mockgen -destination "./mocks/$GOFILE" -package mocks . ResultSender,MetricStorage
type ResultSender interface {
	SendMetrics(ctx context.Context, metrics []Metrics) error
}

type MetricStorage interface {
	Refresh() error
	GetMetrics() []Metrics
}
