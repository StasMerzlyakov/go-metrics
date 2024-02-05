package server

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCheckInputMiddleware(t *testing.T) {
	checkInputMiddleware := CheckInputMiddleware(mockSuccessHandler)
	srv := httptest.NewServer(checkInputMiddleware)
	defer srv.Close()
	testCases := []struct {
		name        string
		url         string
		method      string
		contentType string
		want        want
	}{
		{
			"wrong method",
			"/m1/123",
			http.MethodGet,
			"text/plain",
			want{
				http.StatusMethodNotAllowed,
				"text/plain; charset=utf-8",
			},
		},
		{
			"wrong content type",
			"/m1/123",
			http.MethodPost,
			"application/json",
			want{
				http.StatusUnsupportedMediaType,
				"text/plain; charset=utf-8",
			},
		},
		{
			"wong metric value",
			"/m1/123_",
			http.MethodPost,
			"text/plain; charset=utf-8",
			want{
				http.StatusBadRequest,
				"text/plain; charset=utf-8",
			},
		},
		{
			"float value",
			"/m1/123.05",
			http.MethodPost,
			"text/plain; charset=utf-8",
			want{
				http.StatusOK,
				"text/plain; charset=utf-8",
			},
		},
		{
			"negative value",
			"/m1/-123.05",
			http.MethodPost,
			"text/plain; charset=utf-8",
			want{
				http.StatusOK,
				"text/plain; charset=utf-8",
			},
		},
		{
			"float zero value",
			"/m1/-0.0",
			http.MethodPost,
			"text/plain; charset=utf-8",
			want{
				http.StatusOK,
				"text/plain; charset=utf-8",
			},
		},
		{
			"int value",
			"/m1/123",
			http.MethodPost,
			"text/plain; charset=utf-8",
			want{
				http.StatusOK,
				"text/plain; charset=utf-8",
			},
		},
		{
			"negative int value",
			"/m1/-123",
			http.MethodPost,
			"text/plain; charset=utf-8",
			want{
				http.StatusOK,
				"text/plain; charset=utf-8",
			},
		},
		{
			"zero value",
			"/m1/0",
			http.MethodPost,
			"text/plain; charset=utf-8",
			want{
				http.StatusOK,
				"text/plain; charset=utf-8",
			},
		},
		{
			"request without metrics name",
			"/123",
			http.MethodPost,
			"text/plain; charset=utf-8",
			want{
				http.StatusNotFound,
				"text/plain; charset=utf-8",
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

func TestGetURLRegexp(t *testing.T) {
	reg := getURLRegexp()
	testCases := []struct {
		input  string
		result bool
	}{
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

	for id, test := range testCases {
		t.Run(fmt.Sprintf("TestGetURLRegexp_%v", id), func(t *testing.T) {
			res := reg.MatchString(test.input)
			assert.Equal(t, test.result, res)
		})
	}
}

var mockSuccessHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
})

type want struct {
	code        int
	contentType string
}

type extractFloat64Result struct {
	name              string
	value             float64
	isSuccessExpected bool
}

func Test_extractFloat64(t *testing.T) {
	tests := []struct {
		name   string
		url    string
		result extractFloat64Result
	}{
		{
			"good value",
			"/m1/123.5",
			extractFloat64Result{
				"m1",
				123.5,
				true,
			},
		},
		{
			"bad value",
			"/m1/123.F",
			extractFloat64Result{
				"m1",
				-1,
				false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, tt.url, nil)
			name, value, err := extractFloat64(req)
			assert.Equal(t, tt.result.name, name)
			assert.Equal(t, tt.result.value, value)
			assert.Equal(t, tt.result.isSuccessExpected, err == nil)
		})
	}
}

type extractInt64Result struct {
	name              string
	value             int64
	isSuccessExpected bool
}

func Test_extractInt64(t *testing.T) {
	tests := []struct {
		name   string
		url    string
		result extractInt64Result
	}{
		{
			"good value",
			"/m1/123",
			extractInt64Result{
				"m1",
				123,
				true,
			},
		},
		{
			"bad value",
			"/m1/123F",
			extractInt64Result{
				"m1",
				-1,
				false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, tt.url, nil)
			name, value, err := extractInt64(req)
			assert.Equal(t, tt.result.name, name)
			assert.Equal(t, tt.result.value, value)
			assert.Equal(t, tt.result.isSuccessExpected, err == nil)
		})
	}
}
