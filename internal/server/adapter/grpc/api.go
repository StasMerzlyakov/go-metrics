package grpc

import (
	"context"

	"github.com/StasMerzlyakov/go-metrics/internal/server/domain"
)

//go:generate mockgen -destination "./generated_mocks_test.go" -package ${GOPACKAGE}_test . MetricApp

type MetricApp interface {
	UpdateAll(ctx context.Context, mtr []domain.Metrics) error
}
