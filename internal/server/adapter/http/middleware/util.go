package middleware

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func Add(r *chi.Mux, middlewares ...func(http.Handler) http.Handler) {
	r.Use(middlewares...)
}
