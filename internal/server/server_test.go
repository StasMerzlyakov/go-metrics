package server

import (
	"fmt"
	"github.com/StasMerzlyakov/go-metrics/internal/storage"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

const textPlaint = "text/plain; charset=utf-8"

func TestCreateServerHandler(t *testing.T) {
	serverHandler := Builder.NewConfigurationBuilder().
		CounterPostHandler(mockSuccessHandler).
		GaugePostHandler(mockSuccessHandler).
		Build()
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

func TestCounterValueHandler(t *testing.T) {
	// TODO использовать mock-storage
	counterStorage := storage.NewMemoryInt64Storage()
	counterPostHandler := CounterPostHandlerCreator(counterStorage)
	counterGetHandler := CounterGetHandlerCreator(counterStorage)
	serverHandler := Builder.NewConfigurationBuilder().
		CounterGetHandler(counterGetHandler).
		CounterPostHandler(counterPostHandler).
		Build()

	srv := httptest.NewServer(serverHandler)
	defer srv.Close()

	req := resty.New().R()
	req.Method = http.MethodPost
	value1 := 123
	value1Str := fmt.Sprintf("%v", value1)
	req.URL = srv.URL + "/update/counter/TestCounter/" + value1Str
	req.Header.Add("Content-Type", textPlaint)
	_, err := req.Send()
	require.Nil(t, err)

	req = resty.New().R()
	req.Method = http.MethodGet
	req.URL = srv.URL + "/value/counter/TestCounter"
	req.Header.Add("Content-Type", textPlaint)

	resp, err := req.Send()
	require.Nil(t, err)
	respBody := string(resp.Body())
	require.Equal(t, value1Str, respBody)

	req = resty.New().R()
	req.Method = http.MethodPost
	value2 := 234
	req.URL = srv.URL + "/update/counter/TestCounter/" + fmt.Sprintf("%v", value2)
	req.Header.Add("Content-Type", textPlaint)
	_, err = req.Send()
	require.Nil(t, err)

	req = resty.New().R()
	req.Method = http.MethodGet
	req.URL = srv.URL + "/value/counter/TestCounter"
	req.Header.Add("Content-Type", textPlaint)

	resp, err = req.Send()
	require.Nil(t, err)
	respBody = string(resp.Body())
	value3 := fmt.Sprintf("%v", value1+value2)
	assert.Equal(t, value3, respBody)

}

func TestGaugeValueHandler(t *testing.T) {
	// TODO использовать mock-storage
	gaugeStorage := storage.NewMemoryFloat64Storage()
	gaugePostHandler := GaugePostHandlerCreator(gaugeStorage)
	gaugeGetHandler := GaugeGetHandlerCreator(gaugeStorage)

	serverHandler := Builder.NewConfigurationBuilder().
		GaugePostHandler(gaugePostHandler).
		GaugeGetHandler(gaugeGetHandler).
		Build()

	srv := httptest.NewServer(serverHandler)
	defer srv.Close()

	req := resty.New().R()
	req.Method = http.MethodPost
	value1 := 234.123
	value1Str := fmt.Sprintf("%v", value1)
	req.URL = srv.URL + "/update/gauge/TestCounter/" + value1Str
	req.Header.Add("Content-Type", textPlaint)
	_, err := req.Send()
	require.Nil(t, err)

	req = resty.New().R()
	req.Method = http.MethodGet
	req.URL = srv.URL + "/value/gauge/TestCounter"
	req.Header.Add("Content-Type", textPlaint)
	resp, err := req.Send()
	require.Nil(t, err)
	respBody := string(resp.Body())
	require.Equal(t, value1Str, respBody) // Не попасть бы на потерю точности string -> float64 -> string

	req = resty.New().R()
	req.Method = http.MethodPost
	value2 := 534.123
	value2Str := fmt.Sprintf("%v", value2)
	req.URL = srv.URL + "/update/gauge/TestCounter/" + value2Str
	req.Header.Add("Content-Type", textPlaint)
	_, err = req.Send()
	require.Nil(t, err)

	req = resty.New().R()
	req.Method = http.MethodGet
	req.URL = srv.URL + "/value/gauge/TestCounter"
	req.Header.Add("Content-Type", textPlaint)
	resp, err = req.Send()
	require.Nil(t, err)
	respBody = string(resp.Body())
	require.Equal(t, value2Str, respBody) // Не попасть бы на потерю точности string -> float64 -> string

}

var mockSuccessHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
})

type want struct {
	code        int
	contentType string
}
