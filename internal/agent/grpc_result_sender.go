package agent

import (
	"context"

	"github.com/StasMerzlyakov/go-metrics/internal/config"
	pb "github.com/StasMerzlyakov/go-metrics/internal/proto"

	"github.com/StasMerzlyakov/go-metrics/internal/server/domain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewGRPCResultSender(conf *config.AgentConfiguration) *grpcResultSender {

	conn, err := grpc.NewClient(conf.ServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}

	client := pb.NewMetricsClient(conn)

	return &grpcResultSender{
		conn:   conn,
		client: client,
	}

}

type grpcResultSender struct {
	conn   *grpc.ClientConn
	client pb.MetricsClient
}

func (h *grpcResultSender) SendMetrics(ctx context.Context, metrics []Metrics) error {

	var pbMetrics []*pb.Metric

	for _, m := range metrics {
		switch m.MType {
		case MetricType(domain.CounterType):
			pbM := &pb.Metric{
				Name:  m.ID,
				Type:  pb.Metric_COUNTER,
				Delta: *m.Delta,
			}
			pbMetrics = append(pbMetrics, pbM)
		case MetricType(domain.GaugeType):
			pbM := &pb.Metric{
				Name:  m.ID,
				Type:  pb.Metric_GAUGE,
				Value: *m.Value,
			}
			pbMetrics = append(pbMetrics, pbM)
		}
	}

	mr := &pb.MetricsRequest{
		Metrics: pbMetrics,
	}

	_, err := h.client.Update(ctx, mr)
	return err
}

func (h *grpcResultSender) Stop() {
	h.conn.Close()
}
