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
	Infow(msg string, keysAndValues ...interface{})
}

type InvokableFunc func(ctx context.Context, args ...any) error

func (f InvokableFunc) Invoke(ctx context.Context, args ...any) error {
	return f(ctx, args)
}

type RetriableInvokerConf struct {
	RetriableError        error
	RepeateDelay          time.Duration
	RepeateIncrementDelay time.Duration
	MaxRepeateCount       int
}

func CreateRetriableFn(retriableError error, logger Logger, fn InvokableFunc) InvokableFunc {
	return CreateRetriableFnConf(&RetriableInvokerConf{
		RetriableError:        retriableError,
		RepeateDelay:          time.Duration(time.Second),
		RepeateIncrementDelay: time.Duration(2 * time.Second),
	}, logger, fn)
}

func CreateRetriableFnConf(r *RetriableInvokerConf, logger Logger, fn InvokableFunc) InvokableFunc {
	return func(ctx context.Context, args ...any) error {

		i := 0
		for {
			err := fn(ctx, args...)

			if err == nil {
				logger.Infow("RetriableFn", "status", "ok")
				return nil
			}

			if errors.Is(err, r.RetriableError) {
				if i == r.MaxRepeateCount {
					logger.Infow("RetriableFn", "status", "err", "msg", err.Error())
					return err
				}

				// Обрабатываемая ошибка
				logger.Infow("RetriableFn", "status", "retry", "msg", err.Error())

				var waitTime time.Duration = r.RepeateDelay + time.Duration(i)*r.RepeateIncrementDelay
				time.Sleep(waitTime)
				i++
				continue
			} else {
				logger.Infow("RetriableFn", "status", "err", "msg", err.Error())
				return err
			}
		}
	}
}
