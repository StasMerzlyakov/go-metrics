package agent

import (
	"context"
	"fmt"
	"sync"

	"github.com/StasMerzlyakov/go-metrics/internal/config"
	"github.com/sirupsen/logrus"
)

func NewPoolResultSender(conf *config.AgentConfiguration, resultSender ResultSender) *poolResultSender {
	return &poolResultSender{
		sender:    resultSender,
		batchSize: conf.BatchSize,
		rateLimit: conf.RateLimit,
		batchChan: make(chan Metrics),
	}
}

type poolResultSender struct {
	sender           ResultSender
	batchSize        int
	rateLimit        int
	batchChan        chan Metrics
	startBatcherOnce sync.Once
}

func (rs *poolResultSender) Stop() {
	rs.sender.Stop()
}

func (rs *poolResultSender) SendMetrics(ctx context.Context, metrics []Metrics) error {
	logrus.Infof("SendMetrics start")

	rs.startBatcherOnce.Do(func() {
		go rs.batcher(ctx)
	})

	for _, m := range metrics {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case rs.batchChan <- m:
			continue
		}
	}

	return nil
}

func (rs *poolResultSender) batcher(ctx context.Context) error {
	var batch []Metrics

	batchPool := make(chan []Metrics)
	defer close(batchPool)

	workerCount := rs.rateLimit
	for i := 0; i < workerCount; i++ {
		go rs.worker(ctx, fmt.Sprintf("worker %d", i), batchPool)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case m, ok := <-rs.batchChan:
			if !ok {
				return fmt.Errorf("channel is close")
			}
			batch = append(batch, m)
			if len(batch) == rs.batchSize {
				logrus.Infof("send batch")
				select {
				case <-ctx.Done():
					return ctx.Err()
				case batchPool <- batch:
					batch = batch[:0]
				}
			}
		}
	}
}

func (rs *poolResultSender) worker(ctx context.Context, name string, batchPool <-chan []Metrics) error {
	logrus.Infof("worker %v started", name)
	for metrics := range batchPool {
		logrus.Infof("worker %v send start", name)
		if err := rs.sender.SendMetrics(ctx, metrics); err != nil {
			logrus.Warnf("worker %v send error %v", name, err.Error())
		} else {
			logrus.Infof("worker %v send success", name)
		}
	}
	return nil
}
