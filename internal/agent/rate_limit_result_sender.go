package agent

import (
	"context"
	"fmt"
	"sync"

	"github.com/StasMerzlyakov/go-metrics/internal/config"
	"github.com/sirupsen/logrus"
)

func NewLimitResultSender(conf *config.AgentConfiguration, resultSender ResultSender) *limitResultSender {
	return &limitResultSender{
		sender:    resultSender,
		batchSize: conf.BatchSize,
		rateLimit: conf.RateLimit,
		batchChan: make(chan Metrics),
	}
}

type limitResultSender struct {
	sender             ResultSender
	batchSize          int
	rateLimit          int
	batchChan          chan Metrics
	closeBatchChanOnce sync.Once
	startBatcherOnce   sync.Once
}

func (ls *limitResultSender) SendMetrics(ctx context.Context, metrics []Metrics) error {
	logrus.Infof("SendMetrics start")

	ls.startBatcherOnce.Do(func() {
		go ls.batcher(ctx)
	})

	for _, m := range metrics {
		select {
		case <-ctx.Done():
			ls.closeBatchChanOnce.Do(func() {
				close(ls.batchChan)
			})
			return ctx.Err()
		case ls.batchChan <- m:
			continue
		}
	}

	return nil
}

func (ls *limitResultSender) batcher(ctx context.Context) error {
	var batch []Metrics

	batchPool := make(chan []Metrics)
	defer close(batchPool)

	workerCount := ls.rateLimit
	for i := 0; i < workerCount; i++ {
		go ls.worker(ctx, fmt.Sprintf("worker %d", i), batchPool)
	}

	for {
		select {
		case <-ctx.Done():
			ls.closeBatchChanOnce.Do(func() {
				close(ls.batchChan)
			})
			return ctx.Err()
		case m, ok := <-ls.batchChan:
			if !ok {
				return fmt.Errorf("channel is close")
			}
			batch = append(batch, m)
			if len(batch) == ls.batchSize {
				batchPool <- batch
				batch = batch[:0]
			}
		}
	}
}

func (ls *limitResultSender) worker(ctx context.Context, name string, batchPool <-chan []Metrics) error {
	logrus.Infof("worker %v started", name)
	for metrics := range batchPool {
		logrus.Infof("worker %v send start", name)
		if err := ls.sender.SendMetrics(ctx, metrics); err != nil {
			logrus.Warnf("worker %v send error %v", name, err.Error())
		} else {
			logrus.Infof("worker %v send success", name)
		}

	}
	return nil
}
