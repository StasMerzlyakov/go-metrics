package internal

import (
	"fmt"
	"net/http"
)

func BadRequestHandler(w http.ResponseWriter, req *http.Request) {
	http.Error(w, "BadRequest", http.StatusBadRequest)
}

func StatusMethodNotAllowedHandler(w http.ResponseWriter, req *http.Request) {
	http.Error(w, "StatusMethodNotAllowed", http.StatusMethodNotAllowed)
}
func StatusNotImplemented(w http.ResponseWriter, req *http.Request) {
	http.Error(w, "StatusMethodNotAllowed", http.StatusNotImplemented)
}

func StatusNotFound(w http.ResponseWriter, req *http.Request) {
	http.Error(w, "StatusNotFound", http.StatusNotFound)
}

func TodoResponse(res http.ResponseWriter, message string) {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusNotImplemented)
	fmt.Fprintf(res, `
      {
        "response": {
          "text": "%v"
        },
        "version": "1.0"
      }
    `, message)
}

type Middleware func(http.HandlerFunc) http.HandlerFunc

func Conveyor(h http.HandlerFunc, middlewares ...Middleware) http.HandlerFunc {
	for _, middleware := range middlewares {
		h = middleware(h)
	}
	return h
}
