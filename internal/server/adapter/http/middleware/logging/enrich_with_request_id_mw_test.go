package logging_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/http/middleware"
	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/http/middleware/logging"
	"github.com/StasMerzlyakov/go-metrics/internal/server/domain"
	"github.com/go-resty/resty/v2"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestEnrichMW(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	defer logger.Sync()

	suga := logger.Sugar()
	domain.SetMainLogger(suga)

	mux := http.NewServeMux()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockHandler := NewMockHandler(ctrl)

	mockHandler.EXPECT().ServeHTTP(gomock.Any(), gomock.Any()).DoAndReturn(
		func(w http.ResponseWriter, r *http.Request) {

			ctx := r.Context()
			_, ok := ctx.Value(domain.LoggerKey).(domain.Logger)
			require.True(t, ok)

		}).Times(1)

	enrMW := logging.EncrichWithRequestIDMW()

	mux.Handle("/test", middleware.Conveyor(mockHandler, enrMW))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	req := resty.New().R()
	req.Method = http.MethodPost

	req.URL = srv.URL + "/test"

	resp, err := req.Send()
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())
}
