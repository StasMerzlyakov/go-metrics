package agent

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"syscall"

	"github.com/StasMerzlyakov/go-metrics/internal/common/errors/retriable"
	"github.com/sirupsen/logrus"
)

type ResultSender interface {
	SendMetrics(ctx context.Context, metrics []Metrics) error
}

func NewHTTPRetryableResultSender(rConf retriable.RetriableInvokerConf, resultSender ResultSender) *httpRetryableResultSender {

	return &httpRetryableResultSender{
		invokeFn: retriable.CreateRetriableFn(syscall.ECONNREFUSED, &logrusWrapper{}, wrap(resultSender.SendMetrics)),
	}
}

type logrusWrapper struct{}

func (logrusWrapper) Infow(msg string, keysAndValues ...any) {
	strTempl := "%v" + strings.Repeat(" %v", len(keysAndValues))
	arr := []any{msg}
	arr = append(arr, keysAndValues...)
	logrus.Infof(strTempl, arr...)
}

type httpRetryableResultSender struct {
	invokeFn retriable.InvokableFunc
}

func wrap(fn func(ctx context.Context, metrics []Metrics) error) retriable.InvokableFunc {
	return func(ctx context.Context, args ...any) error {
		var metrics []Metrics
		if len(args) != 1 {
			return fmt.Errorf("unexpected arguments length %v", len(args))
		}

		metrics, ok := args[0].([]Metrics)
		if !ok {
			return fmt.Errorf("unexpected argument type %v", reflect.TypeOf(args[0]))
		}

		return fn(ctx, metrics)
	}
}

func (h *httpRetryableResultSender) SendMetrics(ctx context.Context, metrics []Metrics) error {
	return h.invokeFn(ctx, metrics)
}
