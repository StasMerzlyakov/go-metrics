package agent_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/StasMerzlyakov/go-metrics/internal/agent"
	"github.com/StasMerzlyakov/go-metrics/internal/config"
	"github.com/golang/mock/gomock"
)

func TestAgentOk(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSender := NewMockResultSender(ctrl)

	mockSender.EXPECT().SendMetrics(gomock.Any(), gomock.Any()).Return(nil).MaxTimes(2).MinTimes(1)

	mockStorage := NewMockMetricStorage(ctrl)

	mockStorage.EXPECT().GetMetrics().Return([]agent.Metrics{}).MaxTimes(2).MinTimes(1)
	mockStorage.EXPECT().Refresh().MinTimes(2).MaxTimes(4) // ждем 3 секунды, poolInterval 1 секунда

	config := config.AgentConfiguration{
		PollInterval:   1,
		ReportInterval: 2,
	}

	client := agent.Create(&config, mockSender, mockStorage)

	ctx, fn := context.WithTimeout(context.Background(), 3*time.Second)
	defer fn()

	client.Start(ctx)

	client.Wait()
}

func TestAgentErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSender := NewMockResultSender(ctrl)

	mockSender.EXPECT().SendMetrics(gomock.Any(), gomock.Any()).Return(errors.New("test err")).MaxTimes(2).MinTimes(1)

	mockStorage := NewMockMetricStorage(ctrl)

	mockStorage.EXPECT().GetMetrics().Return([]agent.Metrics{}).MaxTimes(2).MinTimes(1)
	mockStorage.EXPECT().Refresh().Return(errors.New("test err")).MinTimes(2).MaxTimes(4) // ждем 3 секунды, poolInterval 1 секунда

	config := config.AgentConfiguration{
		PollInterval:   1,
		ReportInterval: 2,
	}

	client := agent.Create(&config, mockSender, mockStorage)

	ctx, fn := context.WithTimeout(context.Background(), 3*time.Second)
	defer fn()

	client.Start(ctx)

	client.Wait()
}
