package wrapper

import (
	"context"

	"github.com/StasMerzlyakov/go-metrics/internal/common/wrapper/retriable"
	"github.com/StasMerzlyakov/go-metrics/internal/server/domain"
	"go.uber.org/zap"
)

type Invoker interface {
	Invoke(fn retriable.InvokableFn, ctx context.Context) error
}

type Storage interface {
	SetAllMetrics(ctx context.Context, marr []domain.Metrics) error
	GetAllMetrics(ctx context.Context) ([]domain.Metrics, error)
	Set(ctx context.Context, m *domain.Metrics) error
	Add(ctx context.Context, m *domain.Metrics) error
	SetMetrics(ctx context.Context, metric []domain.Metrics) error
	AddMetrics(ctx context.Context, metric []domain.Metrics) error
	Get(ctx context.Context, id string, mType domain.MetricType) (*domain.Metrics, error)
	Ping(ctx context.Context) error
	Bootstrap(ctx context.Context) error
	Close(ctx context.Context) error
}

func NewRetriableWrapper(rConf *retriable.RetriableInvokerConf, logger *zap.SugaredLogger, internalStorage Storage) *storage {
	invoker := retriable.CreateRetriableInvokerConf(rConf, &zapLoggerWrapper{
		logger: logger,
	})
	return &storage{
		invoker:         invoker,
		internalStorage: internalStorage,
	}
}

var _ Storage = (*storage)(nil)

type storage struct {
	invoker         Invoker
	internalStorage Storage
}

func (s *storage) SetAllMetrics(ctx context.Context, metricses []domain.Metrics) error {
	fn := func(ctx context.Context) error {
		return s.internalStorage.SetAllMetrics(ctx, metricses)
	}
	return s.invoker.Invoke(fn, ctx)
}

func (s *storage) GetAllMetrics(ctx context.Context) ([]domain.Metrics, error) {
	var result []domain.Metrics
	fn := func(ctx context.Context) error {
		res, err := s.internalStorage.GetAllMetrics(ctx)
		if err != nil {
			return err
		}
		result = append(result, res...)
		return nil
	}
	err := s.invoker.Invoke(fn, ctx)
	return result, err
}

func (s *storage) Set(ctx context.Context, m *domain.Metrics) error {
	fn := func(ctx context.Context) error {
		return s.internalStorage.Set(ctx, m)
	}
	return s.invoker.Invoke(fn, ctx)
}

func (s *storage) Add(ctx context.Context, m *domain.Metrics) error {
	fn := func(ctx context.Context) error {
		return s.internalStorage.Add(ctx, m)
	}
	return s.invoker.Invoke(fn, ctx)
}

func (s *storage) SetMetrics(ctx context.Context, metricses []domain.Metrics) error {
	fn := func(ctx context.Context) error {
		return s.internalStorage.SetMetrics(ctx, metricses)
	}
	return s.invoker.Invoke(fn, ctx)
}

func (s *storage) AddMetrics(ctx context.Context, metricses []domain.Metrics) error {
	fn := func(ctx context.Context) error {
		return s.internalStorage.AddMetrics(ctx, metricses)
	}
	return s.invoker.Invoke(fn, ctx)
}

func (s *storage) Get(ctx context.Context, id string, mType domain.MetricType) (*domain.Metrics, error) {
	var m *domain.Metrics
	var err error
	fn := func(ctx context.Context) error {
		m, err = s.internalStorage.Get(ctx, id, mType)
		return err
	}
	err = s.invoker.Invoke(fn, ctx)
	if err != nil {
		return nil, err
	}
	return m, nil
}
func (s *storage) Ping(ctx context.Context) error {
	return s.invoker.Invoke(s.internalStorage.Ping, ctx)
}
func (s *storage) Bootstrap(ctx context.Context) error {
	return s.invoker.Invoke(s.internalStorage.Bootstrap, ctx)
}
func (s *storage) Close(ctx context.Context) error {
	return s.invoker.Invoke(s.internalStorage.Close, ctx)
}

type zapLoggerWrapper struct {
	logger *zap.SugaredLogger
}

func (z *zapLoggerWrapper) Infow(msg string, keysAndValues ...any) {
	z.logger.Infow(msg, keysAndValues...)
}
