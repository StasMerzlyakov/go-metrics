package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/StasMerzlyakov/go-metrics/internal/storage"
)

type Agent interface {
	Wait()
}

func CreateAgent(ctx context.Context, config *Configuration) (Agent, error) {
	agent := &agent{
		metricsSource: NewRuntimeMetricsSource(),
		resultSender:  NewHTTPResultSender(config.ServerAddr),
		gaugeStorage:  storage.NewMemoryFloat64Storage(),
		poolCounter:   0,
	}
	go agent.PoolMetrics(ctx, config.PollInterval)
	go agent.ReportMetrics(ctx, config.ReportInterval)
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
				fmt.Printf("[%v] PoolMetrics [SUCCESS] (poolCounter %v)\n", time.Now(), a.poolCounter)
			}
		}
	}
}

func (a *agent) ReportMetrics(ctx context.Context, reportIntervalSec int) {
	counter := 0
MAIN:
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
					if err := a.resultSender.SendGauge(key, val); err != nil {
						fmt.Printf("[%v] ReportMetrics [ERROR] (poolCounter %v)\n", time.Now(), err)
						continue MAIN
					}
				}
				_ = a.resultSender.SendCounter("PoolCount", a.poolCounter)
				fmt.Printf("[%v] ReportMetrics [SUCCESS] (poolCounter %v)\n", time.Now(), a.poolCounter)
			}
		}
	}
}
