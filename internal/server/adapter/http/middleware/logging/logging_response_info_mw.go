package logging

import (
	"net/http"

	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/http/middleware"
	"github.com/StasMerzlyakov/go-metrics/internal/server/domain"
)

type responseData struct {
	status int
	size   int
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCodeFixed bool
	responseData    *responseData
}

var _ http.ResponseWriter = (*loggingResponseWriter)(nil)

func (lw *loggingResponseWriter) Header() http.Header {
	return lw.ResponseWriter.Header()
}

func (lw *loggingResponseWriter) Write(data []byte) (int, error) {
	if !lw.statusCodeFixed {
		lw.statusCodeFixed = true
	}
	size, err := lw.ResponseWriter.Write(data)
	lw.responseData.size += size
	return size, err
}

func (lw *loggingResponseWriter) WriteHeader(statusCode int) {
	lw.ResponseWriter.WriteHeader(statusCode)
	if !lw.statusCodeFixed {
		lw.responseData.status = statusCode
		lw.statusCodeFixed = true
	}
}

func NewLoggingResponseMW() middleware.Middleware {
	return func(next http.Handler) http.Handler {
		log := domain.GetMainLogger()
		lrw := func(w http.ResponseWriter, r *http.Request) {
			lw := &loggingResponseWriter{
				responseData: &responseData{
					status: http.StatusOK,
					size:   0,
				},
				statusCodeFixed: false,
				ResponseWriter:  w,
			}

			next.ServeHTTP(lw, r)
			log.Infow("requestResult", "statusCode", lw.responseData.status, "size", lw.responseData.size)
		}
		return http.HandlerFunc(lrw)
	}
}
