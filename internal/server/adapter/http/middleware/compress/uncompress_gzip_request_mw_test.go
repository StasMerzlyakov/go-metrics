package compress_test

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/http/middleware"
	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/http/middleware/compress"
	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/http/mocks"
	"github.com/go-resty/resty/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestUncompressGZIPRequestMW(t *testing.T) {

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	defer logger.Sync()

	suga := logger.Sugar()

	content := []byte("Hello World")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHandler := mocks.NewMockHandler(ctrl)

	mockHandler.EXPECT().ServeHTTP(gomock.Any(), gomock.Any()).DoAndReturn(
		func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			defer r.Body.Close()

			if err != nil && err != io.EOF {
				http.Error(w, "read body err", http.StatusInternalServerError)
			}

			if !bytes.Equal(content, body) {
				http.Error(w, "unexpected body err", http.StatusBadRequest)
			}
		}).AnyTimes()

	uncompressMW := compress.NewUncompressGZIPRequestMW(suga)

	srv := httptest.NewServer(middleware.Conveyor(mockHandler, uncompressMW))
	defer srv.Close()

	testCases := []struct {
		Name     string
		Compress bool
	}{
		{
			Name:     "compressed",
			Compress: true,
		},

		{
			Name:     "uncompressed",
			Compress: false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			req := resty.New().R().
				SetHeader("Content-Type", "text/plain; charset=UTF-8")
			if tt.Compress {
				req.SetHeader("Content-Encoding", "gzip")
				var buff bytes.Buffer
				gz, err := gzip.NewWriterLevel(&buff, gzip.BestCompression)
				require.NoError(t, err)
				gz.Write(content)
				gz.Close()
				req.SetBody(buff.Bytes())
			} else {
				req.SetBody(content)
			}

			resp, err := req.Post(srv.URL)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode())
		})
	}
}
