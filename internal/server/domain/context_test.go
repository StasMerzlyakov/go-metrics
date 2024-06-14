package domain_test

import (
	"context"
	"testing"

	"github.com/StasMerzlyakov/go-metrics/internal/server/domain"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestEnrichContextRequestID(t *testing.T) {

	requestUUID := uuid.New()
	reqStr := requestUUID.String()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := NewMockLogger(ctrl)

	testLoggerFn := func(msg string, keysAndValues ...any) {
		// Проверяем что что при вызове метода логирования добавляется информация о пользователе и requstId
		requestIDIsChecked := false

		for id, v := range keysAndValues {
			switch v := v.(type) {
			case string:
				if v == domain.LoggerKeyRequestID {
					require.True(t, id+1 < len(keysAndValues), "requestID is not set")
					k := keysAndValues[id+1]
					id, ok := k.(string)
					require.True(t, ok, "requestID is not string")
					require.Equal(t, reqStr, id, "unexpecred requestID value")
					requestIDIsChecked = true
				}
			}
		}
		require.Truef(t, requestIDIsChecked, "requestID is not specified")
	}

	m.EXPECT().Infow(gomock.Any(), gomock.Any()).DoAndReturn(testLoggerFn).AnyTimes()

	m.EXPECT().Errorw(gomock.Any(), gomock.Any()).DoAndReturn(testLoggerFn).AnyTimes()

	ctx := context.Background()

	enrichedCtx := domain.EnrichWithRequestIDLogger(ctx, requestUUID, m)

	log := domain.GetCtxLogger(enrichedCtx)

	log.Errorw("test errorw", "msg", "hello")
	log.Infow("test errorw", "msg", "hello")
}
