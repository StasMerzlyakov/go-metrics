package retriable_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/StasMerzlyakov/go-metrics/internal/agent/mocks"
	"github.com/StasMerzlyakov/go-metrics/internal/agent/retriable"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestInvoker1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mLog := mocks.NewMockLogger(ctrl)
	mLog.EXPECT().Infow(gomock.Any(), gomock.Any()).AnyTimes()

	conf := &retriable.RetriableInvokerConf{
		RetriableErr:    io.EOF,
		FirstRetryDelay: time.Duration(time.Second),
		DelayIncrement:  time.Duration(2 * time.Second),
		RetryCount:      4,
	}

	testCases := []struct {
		name                    string
		retriableError          error
		invocationFnError       error
		expectedInvokationCount int
	}{
		{
			"retriable",
			io.EOF,
			io.EOF,
			4,
		},
		{
			"is_not_retriable",
			io.ErrClosedPipe,
			io.EOF,
			1,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()

			invokeCount := 0
			fn := func(ctx context.Context) error {
				defer func() {
					invokeCount++
				}()
				return fmt.Errorf("wrap error %w", test.invocationFnError)
			}

			conf.RetriableErr = test.retriableError
			invoker := retriable.CreateRetriableInvokerConf(conf, mLog)

			maxTestDuration := maxInvokationDuration(conf)
			startTime := time.Now()
			err := invoker.Invoke(fn, ctx)
			assert.Equal(t, test.expectedInvokationCount, invokeCount)
			assert.True(t, errors.Is(err, test.invocationFnError))
			assert.True(t, time.Since(startTime) < maxTestDuration+time.Second) // добавим секунду на накладные расходы
		})
	}
}

func maxInvokationDuration(conf *retriable.RetriableInvokerConf) time.Duration {
	if conf.RetryCount == 0 {
		return 0
	}
	n := time.Duration(conf.RetryCount - 1)
	// по формуле арифметической прогрессии
	maxTime := n * (conf.FirstRetryDelay + conf.FirstRetryDelay + time.Duration(n-1)*conf.DelayIncrement) / 2
	return maxTime
}

func TestInvokerCancellation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mLog := mocks.NewMockLogger(ctrl)
	mLog.EXPECT().Infow(gomock.Any(), gomock.Any()).AnyTimes()

	conf := &retriable.RetriableInvokerConf{
		RetriableErr:    io.EOF,
		FirstRetryDelay: time.Duration(time.Second),
		DelayIncrement:  time.Duration(2 * time.Second),
		RetryCount:      4,
	}

	testCases := []struct {
		name              string
		retriableError    error
		invocationFnError error
	}{
		{
			"retriable",
			io.EOF,
			io.EOF,
		},
		{
			"is_not_retriable",
			io.ErrClosedPipe,
			io.EOF,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			fn := func(ctx context.Context) error {
				return fmt.Errorf("wrap error %w", test.invocationFnError)
			}

			conf.RetriableErr = test.retriableError
			invoker := retriable.CreateRetriableInvokerConf(conf, mLog)
			ctx, cancelFn := context.WithTimeout(context.Background(), time.Millisecond*500)
			startTime := time.Now()
			err := invoker.Invoke(fn, ctx)
			cancelFn()
			assert.Error(t, err)
			assert.True(t, time.Since(startTime) < time.Second) // добавим секунду на накладные расходы
		})
	}
}
