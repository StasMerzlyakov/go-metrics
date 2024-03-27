package agent_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/StasMerzlyakov/go-metrics/internal/agent"
	"github.com/StasMerzlyakov/go-metrics/internal/agent/mocks"
	"github.com/StasMerzlyakov/go-metrics/internal/config"
	"github.com/golang/mock/gomock"
)

func TestPoolResultSenderCancellation_1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSender := mocks.NewMockResultSender(ctrl)

	mockSender.EXPECT().SendMetrics(gomock.Any(), gomock.Any()).Return(nil).Times(1)

	cnf := &config.AgentConfiguration{
		BatchSize: 5,
		RateLimit: 1,
	}

	sender := agent.NewPoolResultSender(cnf, mockSender)

	ctx, cancelFn := context.WithCancel(context.Background())

	// BatchSize == 5, отправим BatchSize+1 записей, потом остановим процесс
	var metrics []agent.Metrics
	value := 1.0
	for i := 0; i < cnf.BatchSize+1; i++ {
		metrics = append(metrics, agent.Metrics{
			ID:    "HeapReleased",
			MType: agent.GaugeType,
			Value: &value,
		})
	}

	sender.SendMetrics(ctx, metrics)
	time.Sleep(1 * time.Second)
	cancelFn()
}

func TestPoolResultSenderCancellation_2(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSender := mocks.NewMockResultSender(ctrl)

	mockSender.EXPECT().SendMetrics(gomock.Any(), gomock.Any()).Return(nil).Times(2)

	cnf := &config.AgentConfiguration{
		BatchSize: 3,
		RateLimit: 1,
	}

	sender := agent.NewPoolResultSender(cnf, mockSender)

	ctx, cancelFn := context.WithCancel(context.Background())

	// BatchSize == 3, отправим 2*BatchSize записей, потом остановим процесс
	var metrics []agent.Metrics
	value := 1.0
	for i := 0; i < 2*cnf.BatchSize; i++ {
		metrics = append(metrics, agent.Metrics{
			ID:    "HeapReleased",
			MType: agent.GaugeType,
			Value: &value,
		})
	}

	sender.SendMetrics(ctx, metrics)
	time.Sleep(1 * time.Second)
	cancelFn()
}

func TestPoolResultSenderCancellation_3(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSender := mocks.NewMockResultSender(ctrl)

	mockSender.EXPECT().SendMetrics(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, ms []agent.Metrics) error {
			// ожидаем завершения процесса, rateLimit = 1 => остановим worker => остановится batcher пока контекст не будет остановлен по таймауту
			<-ctx.Done()
			return nil
		},
	).Times(1) // вызов должен быть только один раз!!

	cnf := &config.AgentConfiguration{
		BatchSize: 3,
		RateLimit: 1,
	}

	sender := agent.NewPoolResultSender(cnf, mockSender)

	ctx, cancelFn := context.WithTimeout(context.Background(), 2*time.Second)

	// BatchSize == 3, отправим 2*BatchSize записей, потом остановим процесс
	var metrics []agent.Metrics
	value := 1.0
	for i := 0; i < 2*cnf.BatchSize+1; i++ {
		metrics = append(metrics, agent.Metrics{
			ID:    "HeapReleased",
			MType: agent.GaugeType,
			Value: &value,
		})
	}

	sender.SendMetrics(ctx, metrics)
	cancelFn()
	time.Sleep(2 * time.Second)
}

func TestPoolResultSenderCancellation_4(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSender := mocks.NewMockResultSender(ctrl)

	rateLimit := 2

	mockSender.EXPECT().SendMetrics(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, ms []agent.Metrics) error {
			<-ctx.Done() // ожидаем завершения процесса
			return nil
		},
	).Times(rateLimit) // вызов два раза*/

	cnf := &config.AgentConfiguration{
		BatchSize: 3,
		RateLimit: rateLimit,
	}

	sender := agent.NewPoolResultSender(cnf, mockSender)

	ctx, cancelFn := context.WithCancel(context.Background())

	// BatchSize == 3, отправим 2*BatchSize записей, потом остановим процесс
	var metrics []agent.Metrics
	value := 1.0
	for i := 0; i < 2*cnf.BatchSize; i++ {
		metrics = append(metrics, agent.Metrics{
			ID:    fmt.Sprintf("HeapReleased_%v", i),
			MType: agent.GaugeType,
			Value: &value,
		})
	}

	sender.SendMetrics(ctx, metrics)
	time.Sleep(2 * time.Second)
	cancelFn()

	time.Sleep(1 * time.Second)

}
