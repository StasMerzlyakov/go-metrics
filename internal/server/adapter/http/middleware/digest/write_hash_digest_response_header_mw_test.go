package digest_test

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/http/middleware"
	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/http/middleware/digest"
	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/http/mocks"
	"github.com/go-resty/resty/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestHash256Header(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHandler := mocks.NewMockHandler(ctrl)

	mockHandler.EXPECT().ServeHTTP(gomock.Any(), gomock.Any()).DoAndReturn(
		func(w http.ResponseWriter, req *http.Request) {
			w.Write([]byte("hello from server"))
		}).AnyTimes()

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	defer logger.Sync()

	suga := logger.Sugar()

	testKey := "key"
	checkHashMW := digest.NewWriteHashDigestResponseHeaderMW(suga, testKey)

	srv := httptest.NewServer(middleware.Conveyor(mockHandler, checkHashMW))
	defer srv.Close()

	req := resty.New().R().
		SetHeader("Content-Type", "text/plain; charset=UTF-8").
		SetBody("hello world")

	resp, err := req.Post(srv.URL)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())

	digestHex := resp.Header().Get("HashSHA256")
	require.NotEmpty(t, digestHex)

	digest, err := hex.DecodeString(digestHex)
	require.NoError(t, err)

	hasher := hmac.New(sha256.New, []byte(testKey))

	hasher.Write(resp.Body())

	expected := hasher.Sum(nil)

	require.True(t, bytes.Equal(expected, digest))
}
