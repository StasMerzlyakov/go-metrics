package app

import (
	"context"

	"github.com/StasMerzlyakov/go-metrics/internal/server/domain"
)

//go:generate mockgen -destination "./generated_mocks_test.go" -package ${GOPACKAGE}_test . Pinger,AllMetricsStorage,BackupFormatter,Storage

type Pinger interface {
	Ping(ctx context.Context) error
}

type AllMetricsStorage interface {
	SetAllMetrics(ctx context.Context, in []domain.Metrics) error
	GetAllMetrics(ctx context.Context) ([]domain.Metrics, error)
}

type BackupFormatter interface {
	Write(ctx context.Context, in []domain.Metrics) error
	Read(ctx context.Context) ([]domain.Metrics, error)
}

type Storage interface {
	SetAllMetrics(ctx context.Context, marr []domain.Metrics) error
	GetAllMetrics(ctx context.Context) ([]domain.Metrics, error)
	Set(ctx context.Context, m *domain.Metrics) error
	Add(ctx context.Context, m *domain.Metrics) error
	SetMetrics(ctx context.Context, metric []domain.Metrics) error
	AddMetrics(ctx context.Context, metric []domain.Metrics) error
	Get(ctx context.Context, id string, mType domain.MetricType) (*domain.Metrics, error)
}
