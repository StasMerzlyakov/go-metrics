package handler

import (
	"context"

	"github.com/StasMerzlyakov/go-metrics/internal/server/domain"
)

//go:generate mockgen -destination "./generated_mocks_test.go" -package ${GOPACKAGE}_test . AdminApp,MetricApp

type AdminApp interface {
	Ping(ctx context.Context) error
}

type MetricApp interface {
	GetAllMetrics(ctx context.Context) ([]domain.Metrics, error)
	Get(ctx context.Context, metricType domain.MetricType, name string) (*domain.Metrics, error)
	UpdateAll(ctx context.Context, mtr []domain.Metrics) error
	Update(ctx context.Context, mtr *domain.Metrics) (*domain.Metrics, error)
}
