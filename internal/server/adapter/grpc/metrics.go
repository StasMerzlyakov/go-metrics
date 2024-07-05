package grpc

import (
	"context"

	pb "github.com/StasMerzlyakov/go-metrics/internal/proto"
	"github.com/StasMerzlyakov/go-metrics/internal/server/domain"
	"google.golang.org/grpc/status"
)

func NewGRPCAdapter(mApp MetricApp) *adapter {
	return &adapter{
		mApp: mApp,
	}
}

type adapter struct {
	pb.UnimplementedMetricsServer // хак какой-то!!
	mApp                          MetricApp
}

func (ad *adapter) Update(ctx context.Context, req *pb.MetricsRequest) (*pb.MetricsResponse, error) {
	var metrics []domain.Metrics

	for _, mtr := range req.GetMetrics() {

		switch mtr.Type {
		case pb.Metric_GAUGE:
			metric := domain.Metrics{
				ID:    mtr.Name,
				MType: domain.GaugeType,
				Value: domain.ValuePtr(mtr.Value),
			}
			metrics = append(metrics, metric)
		default:
			metric := domain.Metrics{
				ID:    mtr.Name,
				MType: domain.CounterType,
				Delta: domain.DeltaPtr(mtr.Delta),
			}
			metrics = append(metrics, metric)
		}
	}

	if err := ad.mApp.UpdateAll(ctx, metrics); err != nil {
		code := MapDomainErrorToGRPCCodeErr(err)
		return nil, status.Error(code, "")
	}

	return nil, nil
}
