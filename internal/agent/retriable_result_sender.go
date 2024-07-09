package agent

import (
	"context"
	"strings"
	"syscall"

	"github.com/sirupsen/logrus"
)

type Invoker interface {
	Invoke(fn InvokableFn, ctx context.Context) error
}

func NewHTTPRetryableResultSender(rConf RetriableInvokerConf, resultSender ResultSender) *httpRetriableResultSender {
	conf := DefaultConf(syscall.ECONNREFUSED)
	return &httpRetriableResultSender{
		invoker: CreateRetriableInvokerConf(conf, &logrusWrapper{}),
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

func (h *httpRetriableResultSender) Stop() {
	h.sender.Stop()
}
