package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const textPlaint = "text/plain; charset=utf-8"

type mockHTTPAdapter struct {
	counterVal int64
	gaugeVal   float64
}

func (httpAdapter *mockHTTPAdapter) PostGauge(w http.ResponseWriter, req *http.Request) {}
func (httpAdapter *mockHTTPAdapter) GetGauge(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(fmt.Sprintf("%v", httpAdapter.gaugeVal)))
}
func (httpAdapter *mockHTTPAdapter) PostCounter(w http.ResponseWriter, req *http.Request) {}
func (httpAdapter *mockHTTPAdapter) GetCounter(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(fmt.Sprintf("%v", httpAdapter.counterVal)))
}

func (httpAdapter *mockHTTPAdapter) AllMetrics(w http.ResponseWriter, request *http.Request) {

}

func TestCounterValueHandler(t *testing.T) {
	testValue := 123
	testValueStr := fmt.Sprintf("%v", testValue)
	handler := createHTTPHandler(&mockHTTPAdapter{
		counterVal: int64(testValue),
	})

	srv := httptest.NewServer(handler)
	defer srv.Close()

	req := resty.New().R()
	req.Method = http.MethodPost

	req.URL = srv.URL + "/update/counter/TestCounter/" + testValueStr
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
	require.Equal(t, testValueStr, respBody)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
}

func TestGaugeValueHandler(t *testing.T) {

	testValue := 234.123
	testValueStr := fmt.Sprintf("%v", testValue)
	handler := createHTTPHandler(&mockHTTPAdapter{
		gaugeVal: float64(testValue),
	})
	srv := httptest.NewServer(handler)
	defer srv.Close()

	req := resty.New().R()
	req.Method = http.MethodPost
	req.URL = srv.URL + "/update/gauge/TestCounter/" + testValueStr
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
	require.Equal(t, testValueStr, respBody) // Не попасть бы на потерю точности string -> float64 -> string
	assert.Equal(t, http.StatusOK, resp.StatusCode())
}
