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
		RetriableError:        io.EOF,
		RepeateDelay:          time.Duration(time.Second),
		RepeateIncrementDelay: time.Duration(2 * time.Second),
		MaxRepeateCount:       3,
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
			err := invokableFn(ctx, test.args...)
			assert.True(t, errors.Is(err, test.invocationFnError))
		})
	}

}
