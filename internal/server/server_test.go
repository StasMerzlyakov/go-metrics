package server

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

type TestCounterCases struct {
	url           string
	method        string
	contentType   string
	expectedCode  int
	expectedKey   string
	expectedValue int64
}

func TestServerCounter(t *testing.T) {
	counterStorage := NewInt64Storage()
	handler := CreateCounterConveyor(counterStorage)
	testCases := []TestCounterCases{
		{
			"/m1/123",
			http.MethodGet,
			"text/plain",
			http.StatusMethodNotAllowed,
			"",
			-1,
		},
		{
			"/m1/123",
			http.MethodPost,
			"application/json",
			http.StatusUnsupportedMediaType,
			"",
			-1,
		},
		{
			"/m1/123_",
			http.MethodPost,
			"text/plain; charset=utf-8",
			http.StatusBadRequest,
			"",
			-1,
		},
		{
			"/m1/123.05",
			http.MethodPost,
			"text/plain; charset=utf-8",
			http.StatusBadRequest,
			"",
			-1,
		},
		{
			"/123_",
			http.MethodPost,
			"text/plain; charset=utf-8",
			http.StatusNotFound,
			"",
			-1,
		},
		{
			"/m1/123",
			http.MethodPost,
			"text/plain; charset=utf-8",
			http.StatusOK,
			"m1",
			123,
		},
		{
			"/m1/-123",
			http.MethodPost,
			"text/plain; charset=utf-8",
			http.StatusOK,
			"m1",
			0,
		},
	}

	for _, test := range testCases {
		req, _ := http.NewRequest(test.method, test.url, nil)
		req.Header.Add("Content-Type", test.contentType)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		res := w.Result()
		assert.Equal(t, test.expectedCode, res.StatusCode)
		if res.StatusCode == http.StatusOK {
			v, ok := counterStorage.Get(test.expectedKey)
			assert.True(t, ok)
			assert.Equal(t, test.expectedValue, v)
		}
	}
}

type TestGaugeCases struct {
	url           string
	method        string
	contentType   string
	expectedCode  int
	expectedKey   string
	expectedValue float64
}

func TestGaugeCounter(t *testing.T) {
	gaugeStorage := NewFloat64Storage()
	handler := CreateGaugeConveyor(gaugeStorage)
	testCases := []TestGaugeCases{
		{
			"/m1/123",
			http.MethodGet,
			"text/plain",
			http.StatusMethodNotAllowed,
			"",
			-1,
		},
		{
			"/m1/123",
			http.MethodPost,
			"application/json",
			http.StatusUnsupportedMediaType,
			"",
			-1,
		},
		{
			"/m1/123_",
			http.MethodPost,
			"text/plain; charset=utf-8",
			http.StatusBadRequest,
			"",
			-1,
		},
		{
			"/m1/123.05",
			http.MethodPost,
			"text/plain; charset=utf-8",
			http.StatusOK,
			"m1",
			123.05,
		},
		{
			"/123_",
			http.MethodPost,
			"text/plain; charset=utf-8",
			http.StatusNotFound,
			"",
			-1,
		},
		{
			"/m1/123",
			http.MethodPost,
			"text/plain; charset=utf-8",
			http.StatusOK,
			"m1",
			123,
		},
		{
			"/m1/-123.5",
			http.MethodPost,
			"text/plain; charset=utf-8",
			http.StatusOK,
			"m1",
			-123.5,
		},
		{
			"/m1/-1",
			http.MethodPost,
			"",
			http.StatusOK,
			"m1",
			-1,
		},
	}

	for _, test := range testCases {
		req, _ := http.NewRequest(test.method, test.url, nil)
		if test.contentType != "" {
			req.Header.Add("Content-Type", test.contentType)
		}

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		res := w.Result()
		assert.Equal(t, test.expectedCode, res.StatusCode)
		if res.StatusCode == http.StatusOK {
			v, ok := gaugeStorage.Get(test.expectedKey)
			assert.True(t, ok)
			assert.Equal(t, test.expectedValue, v)
		}
	}
}

type UrlTestCases struct {
	input  string
	result bool
}

func TestGetUrlRegexp(t *testing.T) {
	reg := GetUrlRegexp()
	testCases := []UrlTestCases{
		{
			"/a1s_asd1_1",
			false,
		},
		{
			"/a1s_asd1_1/00.123",
			false,
		},
		{
			"/a1s_asd1_1/0.123",
			true,
		},
		{
			"/a1s_asd1_1/-0.123",
			true,
		},
		{
			"/a1s_asd1_1/0.123?channel=fs&client=ubuntu",
			false,
		},
		{
			"/1_/0.123",
			false,
		},
		{
			"/_m1_/0.123",
			false,
		},
		{
			"/m1_/123",
			true,
		},
		{
			"/m1_/-123",
			true,
		},
		{
			"//m1_/123",
			false,
		},
		{
			"/m1_/123/",
			false,
		},
		{
			"/m1_//123",
			false,
		},
		{
			"/m1_/123.",
			false,
		},
		{
			"/m1_/123.123.1",
			false,
		},
		{
			"/m1_/123.123.",
			false,
		},
		{
			"/m1_/123.123",
			true,
		},
		{
			"/a1s_asd1_1/0.123/asdsavb",
			false,
		},
	}

	for _, test := range testCases {
		res := reg.MatchString(test.input)
		assert.Equal(t, test.result, res)
	}

}
