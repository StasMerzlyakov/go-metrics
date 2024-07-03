package retry

import "net/http"

//go:generate mockgen -destination "./generated_mocks_test.go" -package ${GOPACKAGE}_test . Handler
type Handler interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
}
