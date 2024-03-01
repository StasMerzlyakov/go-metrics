package logging

import (
	"net/http"
	"time"

	"github.com/StasMerzlyakov/go-metrics/internal/server/controller/http/middleware"
	"go.uber.org/zap"
)

func NewLoggingRequestMW(log *zap.SugaredLogger) middleware.Middleware {
	return func(next http.Handler) http.Handler {
		logReqFn := func(w http.ResponseWriter, req *http.Request) {
			start := time.Now()
			uri := req.RequestURI
			method := req.Method

			next.ServeHTTP(w, req)

			duration := time.Since(start)

			log.Infow("requestStatus",
				"uri", uri,
				"method", method,
				"duration", duration,
			)
		}
		return http.HandlerFunc(logReqFn)
	}
}
