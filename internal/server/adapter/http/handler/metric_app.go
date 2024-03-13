package handler

import (
	"context"

	"github.com/StasMerzlyakov/go-metrics/internal/server/domain"
	_ "github.com/golang/mock/gomock"        // обязательно, требуется в сгенерированных mock-файлах,
	_ "github.com/golang/mock/mockgen/model" // обязательно для корректного запуска mockgen
)

//go:generate mockgen -destination "../mocks/$GOFILE" -package mocks . MetricApp

type MetricApp interface {
	GetAllMetrics(ctx context.Context) ([]domain.Metrics, error)
	GetCounter(ctx context.Context, name string) (*domain.Metrics, error)
	GetGauge(ctx context.Context, name string) (*domain.Metrics, error)
	AddCounter(ctx context.Context, m *domain.Metrics) error
	SetGauge(ctx context.Context, m *domain.Metrics) error
	Update(ctx context.Context, mtr []domain.Metrics) error
}
