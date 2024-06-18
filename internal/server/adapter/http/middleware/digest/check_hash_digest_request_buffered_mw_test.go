package digest_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/http/middleware"
	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/http/middleware/digest"
	"github.com/StasMerzlyakov/go-metrics/internal/server/domain"
	"github.com/go-resty/resty/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestCheckHashDigestRequestBufferedMW_1_Header_Not_Exists(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHandler := NewMockHandler(ctrl)

	mockHandler.EXPECT().ServeHTTP(gomock.Any(), gomock.Any()).DoAndReturn(
		func(w http.ResponseWriter, req *http.Request) {
			// ничего не вычитываем и не закрываем - работает и так
		}).AnyTimes()

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	defer logger.Sync()

	suga := logger.Sugar()
	domain.SetMainLogger(suga)

	checkHashMW := digest.NewCheckHashDigestRequestBufferedMW("")

	srv := httptest.NewServer(middleware.Conveyor(mockHandler, checkHashMW))
	defer srv.Close()

	req := resty.New().R().
		SetHeader("Content-Type", "text/plain; charset=UTF-8").
		SetBody("hello world")

	resp, err := req.Post(srv.URL)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode())
}

func TestCheckHashDigestRequestBufferedMW_2_Header_Is_Not_Hex(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHandler := NewMockHandler(ctrl)

	mockHandler.EXPECT().ServeHTTP(gomock.Any(), gomock.Any()).DoAndReturn(
		func(w http.ResponseWriter, req *http.Request) {
			// ничего не вычитываем и не закрываем - работает и так
		}).AnyTimes()

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	defer logger.Sync()

	suga := logger.Sugar()
	domain.SetMainLogger(suga)

	checkHashMW := digest.NewCheckHashDigestRequestBufferedMW("")

	srv := httptest.NewServer(middleware.Conveyor(mockHandler, checkHashMW))
	defer srv.Close()

	req := resty.New().R().
		SetHeader("Content-Type", "text/plain; charset=UTF-8").
		SetHeader("HashSHA256", "hello world"). // не hex
		SetBody("hello world")

	resp, err := req.Post(srv.URL)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode())
}

func TestCheckHashDigestRequestBufferedMW_3_Digest_Mistmach(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHandler := NewMockHandler(ctrl)

	mockHandler.EXPECT().ServeHTTP(gomock.Any(), gomock.Any()).DoAndReturn(
		func(w http.ResponseWriter, req *http.Request) {
			// ничего не вычитываем и не закрываем - работает и так
		}).AnyTimes()

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	defer logger.Sync()

	suga := logger.Sugar()
	domain.SetMainLogger(suga)

	checkHashMW := digest.NewCheckHashDigestRequestBufferedMW("")

	srv := httptest.NewServer(middleware.Conveyor(mockHandler, checkHashMW))
	defer srv.Close()

	hasher := sha256.New()
	hasher.Write([]byte("hello world"))
	sum := hasher.Sum(nil)

	req := resty.New().R().
		SetHeader("Content-Type", "text/plain; charset=UTF-8").
		SetHeader("HashSHA256", hex.EncodeToString(sum)).
		SetBody("hello world")

	resp, err := req.Post(srv.URL)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode())
}

func TestCheckHashDigestRequestBufferedMW_4_Digest_OK(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHandler := NewMockHandler(ctrl)

	mockHandler.EXPECT().ServeHTTP(gomock.Any(), gomock.Any()).DoAndReturn(
		func(w http.ResponseWriter, req *http.Request) {
			// ничего не вычитываем и не закрываем - работает и так
		}).AnyTimes()

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	defer logger.Sync()

	suga := logger.Sugar()
	domain.SetMainLogger(suga)

	testKey := "testKey"
	checkHashMW := digest.NewCheckHashDigestRequestBufferedMW(testKey)

	srv := httptest.NewServer(middleware.Conveyor(mockHandler, checkHashMW))
	defer srv.Close()

	hasher := hmac.New(sha256.New, []byte(testKey))

	hasher.Write([]byte("hello world"))
	sum := hasher.Sum(nil)

	req := resty.New().R().
		SetHeader("Content-Type", "text/plain; charset=UTF-8").
		SetHeader("HashSHA256", hex.EncodeToString(sum)).
		SetBody("hello world")

	resp, err := req.Post(srv.URL)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
}
