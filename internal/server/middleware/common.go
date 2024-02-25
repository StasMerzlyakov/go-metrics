package middleware

import "net/http"

type MWHandlerFn = func(next http.Handler) http.Handler
