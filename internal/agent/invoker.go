// Package agent содержит wrapper для реализации многократного вызова функции, при возникновении ошибки.
package agent

import (
	"context"
	"errors"
	"time"
)

type InvokableFn func(ctx context.Context) error

type ErrPreProcessFn func(err error) error

type RetriableInvokerConf struct {
	RetriableErr    error
	FirstRetryDelay time.Duration
	DelayIncrement  time.Duration
	RetryCount      int
	PreProccFn      ErrPreProcessFn
}

func DefaultErrPreProcFn(err error) error {
	return err
}

func DefaultConf(retriableErr error) *RetriableInvokerConf {
	return DefaultConfFn(retriableErr, DefaultErrPreProcFn)
}

func DefaultConfFn(retriableErr error, preProccFn ErrPreProcessFn) *RetriableInvokerConf {
	if preProccFn == nil {
		preProccFn = DefaultErrPreProcFn
	}
	return &RetriableInvokerConf{
		RetriableErr:    retriableErr,
		FirstRetryDelay: time.Duration(time.Second),
		DelayIncrement:  time.Duration(2 * time.Second),
		RetryCount:      4,
		PreProccFn:      preProccFn,
	}
}

type retriableInvoker struct {
	RetriableInvokerConf
	logger Logger
}

func (r *retriableInvoker) Invoke(fn InvokableFn, ctx context.Context) error {
	var err error
	iter := 1

	for {
		r.logger.Infow("Invoke", "iteration", iter, "status", "start")
		err = fn(ctx)
		if err == nil {
			r.logger.Infow("Invoke", "iteration", iter, "status", "ok")
			return nil
		}

		err = r.PreProccFn(err)

		if !errors.Is(err, r.RetriableErr) || iter == r.RetryCount {
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

func CreateRetriableInvoker(retriableError error, logger Logger) *retriableInvoker {
	return CreateRetriableInvokerConf(DefaultConf(retriableError), logger)
}

func CreateRetriableInvokerConf(r *RetriableInvokerConf, logger Logger) *retriableInvoker {
	if r.PreProccFn == nil {
		r.PreProccFn = DefaultErrPreProcFn
	}
	invoker := &retriableInvoker{
		RetriableInvokerConf: *r,
		logger:               logger,
	}
	return invoker
}
