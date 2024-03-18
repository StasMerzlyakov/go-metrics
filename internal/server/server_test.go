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

func TestCounterValueHandler(t *testing.T) {
	// TODO использовать mock-storage
	serverHandler := CreateServerHandler()
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
	serverHandler := CreateServerHandler()

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
