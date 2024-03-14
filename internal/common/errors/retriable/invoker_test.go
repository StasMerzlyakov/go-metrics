package retriable_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/StasMerzlyakov/go-metrics/internal/common/errors/mocks"
	"github.com/StasMerzlyakov/go-metrics/internal/common/errors/retriable"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestInvoker1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mInv := mocks.NewMockInvoker(ctrl)

	mLog := mocks.NewMockLogger(ctrl)
	mLog.EXPECT().Infow(gomock.Any(), gomock.Any()).AnyTimes()

	conf := &retriable.RetriableInvokerConf{
		RetriableError:  io.EOF,
		FirstRetryDelay: time.Duration(time.Second),
		DelayIncrement:  time.Duration(2 * time.Second),
		RetryCount:      4,
	}

	testCases := []struct {
		name                    string
		retriableError          error
		invocationFnError       error
		args                    []any
		expectedInvokationCount int
	}{
		{
			"retriable",
			io.EOF,
			io.EOF,
			[]any{1, 2, 3, 4},
			4,
		},
		{
			"is_not_retriable",
			io.ErrClosedPipe,
			io.EOF,
			[]any{1, "2"},
			1,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			mInv.EXPECT().Invoke(gomock.Any(), gomock.Any()).DoAndReturn(
				func(ctx context.Context, args ...any) error {
					assert.Equal(t, len(test.args), len(args))
					for i, m := range args {
						assert.Equal(t, test.args[i], m)
					}
					return fmt.Errorf("wrap error %w", test.invocationFnError)
				}).Times(test.expectedInvokationCount)

			conf.RetriableError = test.retriableError
			invokableFn := retriable.CreateRetriableFnConf(conf, mLog, mInv.Invoke)

			maxTestDuration := maxInvokationDuration(conf)
			startTime := time.Now()
			err := invokableFn(ctx, test.args...)
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

	mInv := mocks.NewMockInvoker(ctrl)

	mLog := mocks.NewMockLogger(ctrl)
	mLog.EXPECT().Infow(gomock.Any(), gomock.Any()).AnyTimes()

	conf := &retriable.RetriableInvokerConf{
		RetriableError:  io.EOF,
		FirstRetryDelay: time.Duration(time.Second),
		DelayIncrement:  time.Duration(2 * time.Second),
		RetryCount:      4,
	}

	testCases := []struct {
		name              string
		retriableError    error
		invocationFnError error
		args              []any
	}{
		{
			"retriable",
			io.EOF,
			io.EOF,
			[]any{1, 2, 3, 4},
		},
		{
			"is_not_retriable",
			io.ErrClosedPipe,
			io.EOF,
			[]any{1, "2"},
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			mInv.EXPECT().Invoke(gomock.Any(), gomock.Any()).DoAndReturn(
				func(ctx context.Context, args ...any) error {
					assert.Equal(t, len(test.args), len(args))
					for i, m := range args {
						assert.Equal(t, test.args[i], m)
					}
					return fmt.Errorf("wrap error %w", test.invocationFnError)
				}).AnyTimes()

			conf.RetriableError = test.retriableError
			invokableFn := retriable.CreateRetriableFnConf(conf, mLog, mInv.Invoke)
			ctx, cancelFn := context.WithTimeout(context.Background(), time.Millisecond*500)
			startTime := time.Now()
			err := invokableFn(ctx, test.args...)
			cancelFn()
			assert.Error(t, err)
			assert.True(t, time.Since(startTime) < time.Second) // добавим секунду на накладные расходы
		})
	}
}
