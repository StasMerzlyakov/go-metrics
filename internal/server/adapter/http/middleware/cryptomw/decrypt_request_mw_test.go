package cryptomw_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"io"
	"net/http"
	"net/http/httptest"
	reflect "reflect"

	"testing"

	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/http/middleware"
	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/http/middleware/cryptomw"
	"github.com/go-resty/resty/v2"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"gotest.tools/v3/assert"
)

func TestDecryptMW(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHandler := NewMockHandler(ctrl)

	testBytes := []byte("test string to check encription/decryption")

	mockHandler.EXPECT().ServeHTTP(gomock.Any(), gomock.Any()).DoAndReturn(
		func(w http.ResponseWriter, req *http.Request) {
			data, err := io.ReadAll(req.Body)
			if err != nil && err != io.EOF {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
			defer req.Body.Close()
			require.True(t, reflect.DeepEqual(testBytes, data))
		}).Times(1)

	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	require.NoError(t, err)

	encrypted, err := rsa.EncryptOAEP(sha512.New(), rand.Reader, &privateKey.PublicKey, testBytes, nil)
	require.NoError(t, err)

	decryptedMW := cryptomw.NewDecrytpMw(privateKey)

	srv := httptest.NewServer(middleware.Conveyor(mockHandler, decryptedMW))
	defer srv.Close()

	req := resty.New().R().
		SetHeader("Content-Type", "text/plain; charset=UTF-8").
		SetBody(encrypted)

	resp, err := req.Post(srv.URL)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
}
