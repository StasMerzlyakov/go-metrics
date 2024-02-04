package agent

import (
	"context"
	"fmt"
	"github.com/StasMerzlyakov/go-metrics/internal/storage"
	"sync"
	"time"
)

type Configuration struct {
	ServerAddr        string
	ContentType       string
	PollIntervalSec   int
	ReportIntervalSec int
	// TODO - http timeout
}

type Agent interface {
	Wait()
}

func CreateAgent(ctx context.Context, config Configuration) (Agent, error) {
	agent := &agent{
		metricsSource: NewRuntimeMetricsSource(),
		resultSender:  NewHTTPResultSender(config.ServerAddr, config.ContentType),
		gaugeStorage:  storage.NewMemoryFloat64Storage(),
		poolCounter:   0,
	}
	go agent.PoolMetrics(ctx, config.PollIntervalSec)
	go agent.ReportMetrics(ctx, config.ReportIntervalSec)
	agent.wg.Add(2)
	return agent, nil
}

type agent struct {
	metricsSource MetricsSource
	resultSender  ResultSender
	gaugeStorage  storage.MetricsStorage[float64]
	poolCounter   int64
	wg            sync.WaitGroup
}

func (a *agent) Wait() {
	a.wg.Wait()
}

func (a *agent) PoolMetrics(ctx context.Context, pollIntervalSec int) {
	counter := 0
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("[%v] PoolMetrics DONE\n", time.Now())
			a.wg.Done()
			return
		default:
			time.Sleep(1 * time.Second) // Будем просыпаться каждую секунду для проверки ctx
			counter++
			if counter == pollIntervalSec {
				counter = 0
				for k, v := range a.metricsSource.PollMetrics() {
					a.gaugeStorage.Set(k, v)
				}
				a.poolCounter = a.metricsSource.PollCount()
				fmt.Printf("[%v] PoolMetrics - %v\n", time.Now(), a.poolCounter)
			}
		}
	}
}

func (a *agent) ReportMetrics(ctx context.Context, reportIntervalSec int) {
	counter := 0
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("[%v] ReportMetrics DONE\n", time.Now())
			a.wg.Done()
			return
		default:
			time.Sleep(1 * time.Second) // Будем просыпаться каждую секунду для проверки ctx
			counter++
			if counter == reportIntervalSec {
				counter = 0
				for _, key := range a.gaugeStorage.Keys() {
					val, _ := a.gaugeStorage.Get(key)
					_ = a.resultSender.SendGauge(key, val)
				}
				_ = a.resultSender.SendCounter("PoolCount", a.poolCounter)
				fmt.Printf("[%v] ReportMetrics success - %v\n", time.Now(), a.poolCounter)
			}
		}
	}
}
