package digest_test

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
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

func TestCheckHashDigestRequestMW_1_Header_Not_Exists(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHandler := NewMockHandler(ctrl)

	mockHandler.EXPECT().ServeHTTP(gomock.Any(), gomock.Any()).DoAndReturn(
		func(w http.ResponseWriter, req *http.Request) {
			_, err := io.ReadAll(req.Body)
			if err != nil && err != io.EOF {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
			defer req.Body.Close()
		}).AnyTimes()

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	defer logger.Sync()

	suga := logger.Sugar()
	domain.SetMainLogger(suga)

	checkHashMW := digest.NewCheckHashDigestRequestMW("")

	srv := httptest.NewServer(middleware.Conveyor(mockHandler, checkHashMW))
	defer srv.Close()

	req := resty.New().R().
		SetHeader("Content-Type", "text/plain; charset=UTF-8").
		SetBody("hello world")

	resp, err := req.Post(srv.URL)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode())
}

func TestCheckHashDigestRequestMW_2_Header_Is_Not_Hex(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHandler := NewMockHandler(ctrl)

	mockHandler.EXPECT().ServeHTTP(gomock.Any(), gomock.Any()).DoAndReturn(
		func(w http.ResponseWriter, req *http.Request) {
			_, err := io.ReadAll(req.Body)
			if err != nil && err != io.EOF {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
			defer req.Body.Close()
		}).AnyTimes()

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	defer logger.Sync()

	suga := logger.Sugar()

	domain.SetMainLogger(suga)

	checkHashMW := digest.NewCheckHashDigestRequestMW("")

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

func TestCheckHashDigestRequestMW_3_Digest_Mistmach(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHandler := NewMockHandler(ctrl)

	mockHandler.EXPECT().ServeHTTP(gomock.Any(), gomock.Any()).DoAndReturn(
		func(w http.ResponseWriter, req *http.Request) {
			_, err := io.ReadAll(req.Body)
			if err != nil && err != io.EOF {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
			defer req.Body.Close()
		}).AnyTimes()

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	defer logger.Sync()

	suga := logger.Sugar()

	domain.SetMainLogger(suga)

	checkHashMW := digest.NewCheckHashDigestRequestMW("")

	srv := httptest.NewServer(middleware.Conveyor(mockHandler, checkHashMW))
	defer srv.Close()

	hasher := sha256.New()
	sum := hasher.Sum([]byte("hello world"))

	req := resty.New().R().
		SetHeader("Content-Type", "text/plain; charset=UTF-8").
		SetHeader("HashSHA256", hex.EncodeToString(sum)).
		SetBody("hello world")

	resp, err := req.Post(srv.URL)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode())
}

func TestCheckHashDigestRequestMW_4_Digest_OK(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHandler := NewMockHandler(ctrl)

	mockHandler.EXPECT().ServeHTTP(gomock.Any(), gomock.Any()).DoAndReturn(
		func(w http.ResponseWriter, req *http.Request) {
			_, err := io.ReadAll(req.Body)
			if err != nil && err != io.EOF {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
			defer req.Body.Close()
		}).AnyTimes()

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	defer logger.Sync()

	suga := logger.Sugar()

	domain.SetMainLogger(suga)

	testKey := "testKey"
	checkHashMW := digest.NewCheckHashDigestRequestMW(testKey)

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

func TestCheckHashDigestRequestMW_4_Json_OK(t *testing.T) {

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	defer logger.Sync()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHandler := NewMockHandler(ctrl)

	mockHandler.EXPECT().ServeHTTP(gomock.Any(), gomock.Any()).DoAndReturn(
		func(w http.ResponseWriter, req *http.Request) {
			defer req.Body.Close()

			var metrics []domain.Metrics
			if err := json.NewDecoder(req.Body).Decode(&metrics); err != nil {
				fullErr := fmt.Errorf("%w: json decode error - %v", domain.ErrDataFormat, err.Error())
				http.Error(w, fullErr.Error(), http.StatusBadRequest)
				return
			}
		}).AnyTimes()

	suga := logger.Sugar()

	domain.SetMainLogger(suga)

	testKey := "testKey"
	checkHashMW := digest.NewCheckHashDigestRequestMW(testKey)

	srv := httptest.NewServer(middleware.Conveyor(mockHandler, checkHashMW))
	defer srv.Close()

	input := []domain.Metrics{
		{
			ID:    "PoolCount",
			MType: domain.CounterType,
			Delta: domain.DeltaPtr(1),
		},
		{
			ID:    "RandomValue",
			MType: domain.GaugeType,
			Value: domain.ValuePtr(1.1),
		},
	}

	var buff bytes.Buffer
	err = json.NewEncoder(&buff).Encode(input)
	require.NoError(t, err)

	var buffBytes = buff.Bytes()

	hasher := hmac.New(sha256.New, []byte(testKey))
	hasher.Write(buffBytes)
	sum := hasher.Sum(nil)

	req := resty.New().R().
		SetHeader("Content-Type", "text/plain; charset=UTF-8").
		SetHeader("HashSHA256", hex.EncodeToString(sum)).
		SetBody(buffBytes)

	resp, err := req.Post(srv.URL)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
}
