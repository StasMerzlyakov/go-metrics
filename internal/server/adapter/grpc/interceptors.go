package grpc

import (
	"context"
	"time"

	"github.com/StasMerzlyakov/go-metrics/internal/server/domain"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// EncrichWithRequestIDInterceptor Добавляет к запросу RequestID и устанавливает в контекст логгер
func EncrichWithRequestIDInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		logger := domain.GetMainLogger()
		requestUUID := uuid.New()
		enrichedCtx := domain.EnrichWithRequestIDLogger(ctx, requestUUID, logger)
		return handler(enrichedCtx, req)
	}
}

// LoggingRequestInfoInteceptor выводит инфрмацию о времени выполнения и статусе возврата
func LoggingRequestInfoInteceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		log := domain.GetCtxLogger(ctx)
		start := time.Now()

		resp, err = handler(ctx, req)
		duration := time.Since(start)

		if err != nil {
			if e, ok := status.FromError(err); ok {
				log.Infow("requestStatus",
					"duration", duration,
					"status", e.Code())
			} else {
				log.Errorw("requestStatus",
					"duration", duration,
					"err", "status unknown",
				)
			}
		} else {
			log.Infow("requestStatus",
				"duration", duration,
				"status", codes.OK,
			)
		}
		return
	}
}
