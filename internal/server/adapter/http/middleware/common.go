// Package middleware содержит мидлы приложения
package middleware

import "net/http"

type Middleware func(http.Handler) http.Handler

func Conveyor(h http.Handler, middlewares ...Middleware) http.Handler {
	for _, middleware := range middlewares {
		h = middleware(h)
	}
	return h
}

func ConveyorFunc(h http.Handler, middlewares ...Middleware) http.HandlerFunc {
	handler := Conveyor(h, middlewares...)
	return handler.ServeHTTP
}
