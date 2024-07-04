package trusted_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/http/middleware"
	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/http/middleware/trusted"
	"github.com/StasMerzlyakov/go-metrics/internal/server/domain"
	"github.com/go-resty/resty/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestSubnetCheckOk(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	defer logger.Sync()

	suga := logger.Sugar()
	domain.SetMainLogger(suga)

	mw, err := trusted.NewTrustedSubnetCheckMW("127.0.0.1/24")
	require.NoError(t, err)

	mux := http.NewServeMux()

	mockHandler := NewMockHandler(ctrl)

	mockHandler.EXPECT().ServeHTTP(gomock.Any(), gomock.Any()).Times(1)

	mux.Handle("/test", middleware.Conveyor(mockHandler, mw))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	req := resty.New().R()
	req.Method = http.MethodPost
	req.Header.Add(trusted.IPXRealHeader, "127.0.0.1")

	req.URL = srv.URL + "/test"

	resp, err := req.Send()
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())
}

func TestSubnetCheckForbitten(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	defer logger.Sync()

	suga := logger.Sugar()
	domain.SetMainLogger(suga)

	mw, err := trusted.NewTrustedSubnetCheckMW("127.0.0.1/24")
	require.NoError(t, err)

	mux := http.NewServeMux()

	mockHandler := NewMockHandler(ctrl)

	mockHandler.EXPECT().ServeHTTP(gomock.Any(), gomock.Any()).Times(0)

	mux.Handle("/test", middleware.Conveyor(mockHandler, mw))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	req := resty.New().R()
	req.Method = http.MethodPost
	req.Header.Add(trusted.IPXRealHeader, "192.168.0.1")

	req.URL = srv.URL + "/test"

	resp, err := req.Send()
	require.Nil(t, err)
	require.Equal(t, http.StatusForbidden, resp.StatusCode())
}

func TestSubnetNoCheck(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	defer logger.Sync()

	suga := logger.Sugar()
	domain.SetMainLogger(suga)

	mw, err := trusted.NewTrustedSubnetCheckMW("")
	require.NoError(t, err)

	mux := http.NewServeMux()

	mockHandler := NewMockHandler(ctrl)

	mockHandler.EXPECT().ServeHTTP(gomock.Any(), gomock.Any()).Times(1)

	mux.Handle("/test", middleware.Conveyor(mockHandler, mw))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	req := resty.New().R()
	req.Method = http.MethodPost
	req.Header.Add(trusted.IPXRealHeader, "192.168.0.1")

	req.URL = srv.URL + "/test"

	resp, err := req.Send()
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())
}
