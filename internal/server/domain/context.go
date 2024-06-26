package domain

import (
	"context"

	"github.com/google/uuid"
)

type ContextKey string

const keyLogger = ContextKey("Logger")
const LoggerKeyRequestID = "requestID"

//go:generate mockgen -destination "./mocks/$GOFILE" -package mocks . Logger
type Logger interface {
	Debugw(msg string, keysAndValues ...any)
	Infow(msg string, keysAndValues ...any)
	Errorw(msg string, keysAndValues ...any)
}

func EnrichWithRequestIDLogger(ctx context.Context, requestID uuid.UUID, logger Logger) context.Context {
	requestIDLogger := &requestIDLogger{
		internalLogger: logger,
		requestID:      requestID.String(),
	}
	resultCtx := context.WithValue(ctx, keyLogger, requestIDLogger)
	return resultCtx
}

// GetCtxLogger возвращает логгер из контекста. Если не найден - то просто MainLogger
func GetCtxLogger(ctx context.Context) Logger {
	if v := ctx.Value(keyLogger); v != nil {
		lg, ok := v.(Logger)
		if !ok {
			return GetMainLogger()
		}
		return lg
	}
	return GetMainLogger()
}

var _ Logger = (*requestIDLogger)(nil)

type requestIDLogger struct {
	requestID      string
	internalLogger Logger
}

func (l *requestIDLogger) Debugw(msg string, keysAndValues ...any) {
	keysAndValues = append(keysAndValues, LoggerKeyRequestID, l.requestID)
	l.internalLogger.Debugw(msg, keysAndValues...)
}

func (l *requestIDLogger) Infow(msg string, keysAndValues ...any) {
	keysAndValues = append(keysAndValues, LoggerKeyRequestID, l.requestID)
	l.internalLogger.Infow(msg, keysAndValues...)
}

func (l *requestIDLogger) Errorw(msg string, keysAndValues ...any) {
	keysAndValues = append(keysAndValues, LoggerKeyRequestID, l.requestID)
	l.internalLogger.Infow(msg, keysAndValues...)
}
