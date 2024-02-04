package agent

import "errors"

type MetricsSource interface {
	PollMetrics() map[string]float64
	PollCount() int64
}

var ErrServerInteraction = errors.New("server interaction error")

type ResultSender interface {
	SendGauge(name string, value float64) error
	SendCounter(name string, value int64) error
}
