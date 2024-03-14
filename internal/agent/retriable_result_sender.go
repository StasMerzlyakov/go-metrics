package agent

import (
	"context"
	"strings"
	"syscall"

	"github.com/StasMerzlyakov/go-metrics/internal/common/errors/retriable"
	"github.com/sirupsen/logrus"
)

type Invoker interface {
	Invoke(fn retriable.InvokableFn, ctx context.Context) error
}

type ResultSender interface {
	SendMetrics(ctx context.Context, metrics []Metrics) error
}

func NewHTTPRetryableResultSender(rConf retriable.RetriableInvokerConf, resultSender ResultSender) *httpRetriableResultSender {
	conf := retriable.DefaultConf(syscall.ECONNREFUSED)
	return &httpRetriableResultSender{
		invoker: retriable.CreateRetriableInvokerConf(conf, &logrusWrapper{}),
		sender:  resultSender,
	}
}

type logrusWrapper struct{}

func (logrusWrapper) Infow(msg string, keysAndValues ...any) {
	strTempl := "%v" + strings.Repeat(" %v", len(keysAndValues))
	arr := []any{msg}
	arr = append(arr, keysAndValues...)
	logrus.Infof(strTempl, arr...)
}

type httpRetriableResultSender struct {
	invoker Invoker
	sender  ResultSender
}

func (h *httpRetriableResultSender) SendMetrics(ctx context.Context, metrics []Metrics) error {
	fn := func(ctx context.Context) error {
		return h.sender.SendMetrics(ctx, metrics)
	}
	return h.invoker.Invoke(fn, ctx)
}
