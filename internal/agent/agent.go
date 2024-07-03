// Package agent содежит конфигурацию и код агента
package agent

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/StasMerzlyakov/go-metrics/internal/config"
	"github.com/sirupsen/logrus"
)

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
	go a.pollMetrics(ctx)
	go a.reportMetrics(ctx)
	a.wg.Add(2)
}

func (a *agent) pollMetrics(ctx context.Context) {
	pollInterval := time.Duration(a.pollIntervalSec) * time.Second
	defer a.wg.Done()
	for {
		select {
		case <-ctx.Done():
			logrus.Info("PollMetrics DONE")
			return

		case <-time.After(pollInterval):
			if err := a.metricStorage.Refresh(); err != nil {
				logrus.Errorf("PollMetrics metrics error: %v", err)
			}
			logrus.Info("PollMetrics SUCCESS")
		}
	}
}

func (a *agent) reportMetrics(ctx context.Context) {
	reportInterval := time.Duration(a.reportIntervalSec) * time.Second

	for {
		select {
		case <-ctx.Done():
			log.Println("ReportMetrics DONE")
			a.wg.Done()
			return
		case <-time.After(reportInterval):
			metrics := a.metricStorage.GetMetrics()
			err := a.resultSender.SendMetrics(ctx, metrics)
			if err != nil {
				logrus.Infof("ReportMetrics ERROR: %v\n", err)
			} else {
				logrus.Info("ReportMetrics SUCCESS")
			}
		}
	}
}
