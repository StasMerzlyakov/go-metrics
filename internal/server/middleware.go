package server

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

func CheckMethodPostMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(w, "only post methods", http.StatusMethodNotAllowed)
			return
		}
		next.ServeHTTP(w, req)
	}
}

func CheckContentTypeMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		contentType := req.Header.Get("Content-Type")
		if contentType != "" && !strings.HasPrefix(contentType, "text/plain") {
			http.Error(w, "only 'text/plain' supported", http.StatusUnsupportedMediaType)
			return
		}
		next.ServeHTTP(w, req)
	}
}

func CheckDigitalMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		valueStr := chi.URLParam(req, "value")
		if !CheckDecimal(valueStr) {
			http.Error(w, "wrong decimal value", http.StatusBadRequest)
			return
		}
		next.ServeHTTP(w, req)
	}
}

func CheckIntegerMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		valueStr := chi.URLParam(req, "value")
		if !CheckInteger(valueStr) {
			http.Error(w, "wrong integer value", http.StatusBadRequest)
			return
		}
		next.ServeHTTP(w, req)
	}
}

func CheckMetricNameMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		name := chi.URLParam(req, "name")
		if !CheckName(name) {
			http.Error(w, "wrong name value", http.StatusBadRequest)
			return
		}
		next.ServeHTTP(w, req)
	}
}
