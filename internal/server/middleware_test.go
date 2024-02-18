package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

const textPlaint = "text/plain; charset=utf-8"

type mockBusinessHandler struct{}

func (*mockBusinessHandler) PostGauge(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", textPlaint)

}
func (*mockBusinessHandler) GetGauge(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", textPlaint)
}
func (*mockBusinessHandler) PostCounter(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", textPlaint)
}
func (*mockBusinessHandler) GetCounter(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", textPlaint)
}
func (*mockBusinessHandler) AllMetrics(w http.ResponseWriter, request *http.Request) {
	w.Header().Set("Content-Type", textPlaint)
}

func TestServerMiddlewareChain(t *testing.T) {
	mockBusinessHandler := &mockBusinessHandler{}
	serverHandler := CreateFullHttpHandler(mockBusinessHandler)

	srv := httptest.NewServer(serverHandler)
	defer srv.Close()
	testCases := []struct {
		name        string
		url         string
		method      string
		contentType string
		want        want
	}{
		{
			"unknown type",
			"/update/unknown/testCounter/100",
			http.MethodPost,
			"text/plain",
			want{
				http.StatusNotImplemented,
				textPlaint,
			},
		},
		{
			"wrong content type",
			"/update/gauge/m1/123",
			http.MethodPost,
			"application/json",
			want{
				http.StatusUnsupportedMediaType,
				textPlaint,
			},
		},
		{
			"wong metric value",
			"/update/gauge/m1/123_",
			http.MethodPost,
			textPlaint,
			want{
				http.StatusBadRequest,
				textPlaint,
			},
		},
		{
			"float value",
			"/update/gauge/m1/123.05",
			http.MethodPost,
			textPlaint,
			want{
				http.StatusOK,
				textPlaint,
			},
		},
		{
			"negative value",
			"/update/gauge/m1/-123.05",
			http.MethodPost,
			textPlaint,
			want{
				http.StatusOK,
				textPlaint,
			},
		},
		{
			"float zero value",
			"/update/gauge/m1/-0.0",
			http.MethodPost,
			textPlaint,
			want{
				http.StatusOK,
				textPlaint,
			},
		},
		{
			"int value",
			"/update/counter/m1/123",
			http.MethodPost,
			textPlaint,
			want{
				http.StatusOK,
				textPlaint,
			},
		},
		{
			"negative int value",
			"/update/counter/m1/-123",
			http.MethodPost,
			textPlaint,
			want{
				http.StatusOK,
				textPlaint,
			},
		},
		{
			"zero value",
			"/update/counter/m1/0",
			http.MethodPost,
			textPlaint,
			want{
				http.StatusOK,
				textPlaint,
			},
		},
		{
			"request without metrics name",
			"/update/counter/123",
			http.MethodPost,
			textPlaint,
			want{
				http.StatusNotFound,
				textPlaint,
			},
		},
		{
			"request without metrics name 2",
			"/update/gauge/123",
			http.MethodPost,
			textPlaint,
			want{
				http.StatusNotFound,
				textPlaint,
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			req := resty.New().R()
			req.Method = test.method
			req.URL = srv.URL + test.url
			req.Header.Add("Content-Type", test.contentType)
			resp, err := req.Send()
			assert.Nil(t, err)
			assert.Equal(t, test.want.code, resp.StatusCode())
			assert.Equal(t, test.want.contentType, resp.Header().Get("Content-Type"))
		})
	}
}
