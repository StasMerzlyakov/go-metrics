package retriable

import (
	"context"
	"errors"
	"time"
)

//go:generate mockgen -destination "../mocks/$GOFILE" -package mocks . Invoker,Logger
type Invoker interface {
	Invoke(ctx context.Context, args ...any) error
}

type Logger interface {
	Infow(msg string, keysAndValues ...any)
}

type InvokableFunc func(ctx context.Context, args ...any) error

func (f InvokableFunc) Invoke(ctx context.Context, args ...any) error {
	return f(ctx, args)
}

type RetriableInvokerConf struct {
	RetriableError  error
	FirstRetryDelay time.Duration
	DelayIncrement  time.Duration
	RetryCount      int
}

func DefaultConf(retriableError error) *RetriableInvokerConf {
	return &RetriableInvokerConf{
		RetriableError:  retriableError,
		FirstRetryDelay: time.Duration(time.Second),
		DelayIncrement:  time.Duration(2 * time.Second),
		RetryCount:      4,
	}
}

type retriableInvoker struct {
	RetriableInvokerConf
	logger Logger
	fn     InvokableFunc
}

func (r *retriableInvoker) Invoke(ctx context.Context, args ...any) error {
	var err error
	iter := 1

	for {
		r.logger.Infow("Invoke", "iteration", iter, "status", "start")
		err = r.fn(ctx, args...)
		if err == nil {
			r.logger.Infow("Invoke", "iteration", iter, "status", "ok")
			return nil
		}

		if !errors.Is(err, r.RetriableError) || iter == r.RetryCount {
			r.logger.Infow("Invoke", "iteration", iter, "status", "err", "msg", err.Error())
			return err
		}
		nextInvokation := r.FirstRetryDelay + time.Duration(iter-1)*r.DelayIncrement
		select {
		case <-ctx.Done():
			r.logger.Infow("Invoke", "status", "err", "msg", "context cancelled")
			return ctx.Err()
		case <-time.After(nextInvokation):
			iter++
			continue
		}
	}
}

func CreateRetriableFn(retriableError error, logger Logger, fn InvokableFunc) InvokableFunc {
	return CreateRetriableFnConf(DefaultConf(retriableError), logger, fn)
}

func CreateRetriableFnConf(r *RetriableInvokerConf, logger Logger, fn InvokableFunc) InvokableFunc {
	invoker := CreateRetriableInvokerConf(r, logger, fn)
	return invoker.Invoke
}

func CreateRetriableInvoker(retriableError error, logger Logger, fn InvokableFunc) Invoker {
	return CreateRetriableInvokerConf(DefaultConf(retriableError), logger, fn)
}

func CreateRetriableInvokerConf(r *RetriableInvokerConf, logger Logger, fn InvokableFunc) Invoker {
	invoker := &retriableInvoker{
		RetriableInvokerConf: *r,
		logger:               logger,
		fn:                   fn,
	}
	return invoker
}
