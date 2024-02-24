package agent

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/StasMerzlyakov/go-metrics/internal/config"
	"github.com/sirupsen/logrus"
)

type ResultSender interface {
	SendMetrics(metrics []Metric) error
}

type MetricStorage interface {
	Refresh() error
	GetMetrics() []Metric
}

func Create(config *config.AgentConfiguration,
	resultSender ResultSender,
	metricStorage MetricStorage,

) *agent {
	agent := &agent{
		metricStorage:     metricStorage,
		resultSender:      resultSender,
		pollIntervalSec:   config.PollInterval,
		reportIntervalSec: config.ReportInterval,
	}

	return agent
}

type agent struct {
	metricStorage     MetricStorage
	resultSender      ResultSender
	pollIntervalSec   int
	reportIntervalSec int
	wg                sync.WaitGroup
}

func (a *agent) Wait() {
	a.wg.Wait()
}

func (a *agent) Start(ctx context.Context) {
	go a.poolMetrics(ctx)
	go a.reportMetrics(ctx)
	a.wg.Add(2)
}

func (a *agent) poolMetrics(ctx context.Context) {
	var poolInterval time.Duration = time.Duration(a.pollIntervalSec) * time.Second

	timer := time.NewTimer(poolInterval)
	defer func() {
		timer.Stop()
	}()

	for {
		select {
		case <-ctx.Done():
			logrus.Info("PoolMetrics DONE")
			a.wg.Done()
			return

		case <-timer.C:
			if err := a.metricStorage.Refresh(); err != nil {
				logrus.Fatalf("PoolMetrics metrics error: %v", err)
			}
			logrus.Info("PoolMetrics SUCCESS")
		}
	}
}

func (a *agent) reportMetrics(ctx context.Context) {

	var reportInterval time.Duration = time.Duration(a.reportIntervalSec) * time.Second

	timer := time.NewTimer(reportInterval)
	defer func() {
		timer.Stop()
	}()

	for {
		select {
		case <-ctx.Done():
			log.Println("ReportMetrics DONE")
			a.wg.Done()
			return
		case <-timer.C:
			metrics := a.metricStorage.GetMetrics()
			err := a.resultSender.SendMetrics(metrics)
			if err != nil {
				log.Printf("ReportMetrics ERROR: %v\n", err)
			} else {
				logrus.Info("ReportMetrics SUCCESS")
			}
		}
	}
}
