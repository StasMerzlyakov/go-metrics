// Package middleware contains go-metrics middleware
package middleware

import (
	"net/http"
)

//go:generate mockgen -destination "../mocks/$GOFILE" -package mocks . Handler
type Handler interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
}
