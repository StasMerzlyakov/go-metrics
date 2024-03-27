package agent

import "context"

//go:generate mockgen -destination "./mocks/$GOFILE" -package mocks . ResultSender
type ResultSender interface {
	SendMetrics(ctx context.Context, metrics []Metrics) error
}
