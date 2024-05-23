package compress_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/http/middleware"
	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/http/middleware/compress"
	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/http/mocks"
	"github.com/StasMerzlyakov/go-metrics/internal/server/domain"
	"github.com/go-resty/resty/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestCompressGZIPResponseMW(t *testing.T) {

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	defer logger.Sync()
	mux := http.NewServeMux()

	suga := logger.Sugar()
	domain.SetMainLogger(suga)

	compressMW := compress.NewCompressGZIPResponseMW()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mux.Handle("/json", middleware.Conveyor(createMockJSONHandler(ctrl), compressMW))
	mux.Handle("/html", middleware.Conveyor(createMockHTMLHandler(ctrl), compressMW))
	mux.Handle("/text", middleware.Conveyor(createMockTextHandler(ctrl), compressMW))

	srv := httptest.NewServer(mux)
	defer srv.Close()

	testCases := []struct {
		Name            string
		Path            string
		AcceptEncoding  string
		ContentEncoding bool
	}{
		{
			"json_gzip",
			"/json",
			"gzip",
			true,
		},
		{
			"json",
			"/json",
			"",
			false,
		},
		{
			"json_deflate",
			"/json",
			"deflate",
			false,
		},
		{
			"html_gzip",
			"/html",
			"gzip",
			true,
		},
		{
			"html",
			"/html",
			"",
			false,
		},
		{
			"html_deflate",
			"/html",
			"deflate",
			false,
		},
		{
			"text_gzip",
			"/text",
			"gzip",
			false,
		},
		{
			"text",
			"/text",
			"",
			false,
		},
		{
			"text_deflate",
			"/text",
			"deflate",
			false,
		},
	}

	req := resty.New().R()
	req.Method = http.MethodPost

	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			req.URL = srv.URL + tt.Path
			if tt.AcceptEncoding == "" {
				req.Header.Del("Accept-Encoding")
			} else {
				req.Header.Set("Accept-Encoding", tt.AcceptEncoding)
			}
			resp, err := req.Send()
			require.Nil(t, err)
			require.Equal(t, http.StatusOK, resp.StatusCode())
			if tt.ContentEncoding {
				assert.True(t, strings.Contains(resp.Header().Get("Content-Encoding"), "gzip"))
			} else {
				assert.False(t, strings.Contains(resp.Header().Get("Content-Encoding"), "gzip"))
			}
		})
	}

}

func createMockHTMLHandler(ctrl *gomock.Controller) http.Handler {
	mockHandler := mocks.NewMockHandler(ctrl)

	mockHandler.EXPECT().ServeHTTP(gomock.Any(), gomock.Any()).DoAndReturn(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			io.WriteString(w, "<html><body>"+strings.Repeat("Hello, world<br>", 20)+"</body></html>")
		}).AnyTimes()
	return mockHandler
}

func createMockJSONHandler(ctrl *gomock.Controller) http.Handler {
	mockHandler := mocks.NewMockHandler(ctrl)

	mockHandler.EXPECT().ServeHTTP(gomock.Any(), gomock.Any()).DoAndReturn(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			io.WriteString(w, "{ "+strings.Repeat(`"msg":"Hello, world",`, 19)+`"msg":"Hello, world"`+"}")
		}).AnyTimes()
	return mockHandler
}

func createMockTextHandler(ctrl *gomock.Controller) http.Handler {
	mockHandler := mocks.NewMockHandler(ctrl)

	mockHandler.EXPECT().ServeHTTP(gomock.Any(), gomock.Any()).DoAndReturn(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			io.WriteString(w, strings.Repeat("Hello, world\n", 20))
		}).AnyTimes()
	return mockHandler
}
