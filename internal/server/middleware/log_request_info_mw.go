package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

func NewLogRequestMW(log *zap.SugaredLogger) MWHandlerFn {
	return func(next http.Handler) http.Handler {
		logReqFn := func(w http.ResponseWriter, req *http.Request) {
			start := time.Now()
			uri := req.RequestURI
			method := req.Method

			next.ServeHTTP(w, req)

			duration := time.Since(start)

			log.Infoln(
				"uri", uri,
				"method", method,
				"duration", duration,
			)
		}
		return http.HandlerFunc(logReqFn)
	}
}
