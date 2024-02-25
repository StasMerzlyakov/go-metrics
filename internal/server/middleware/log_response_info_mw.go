package middleware

import (
	"net/http"

	"go.uber.org/zap"
)

type responseData struct {
	status int
	size   int
}

type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

var _ http.ResponseWriter = (*loggingResponseWriter)(nil)

func (lw *loggingResponseWriter) Header() http.Header {
	return lw.ResponseWriter.Header()
}

func (lw *loggingResponseWriter) Write(data []byte) (int, error) {
	size, err := lw.ResponseWriter.Write(data)
	lw.responseData.size += size
	return size, err
}

func (lw *loggingResponseWriter) WriteHeader(statusCode int) {
	lw.ResponseWriter.WriteHeader(statusCode)
	lw.responseData.status = statusCode
}

func NewLogResponseMW(log *zap.SugaredLogger) MWHandlerFn {
	return func(next http.Handler) http.Handler {
		lrw := func(w http.ResponseWriter, r *http.Request) {
			lw := &loggingResponseWriter{
				responseData: &responseData{
					status: 0,
					size:   0,
				},
				ResponseWriter: w,
			}

			next.ServeHTTP(lw, r)
			log.Infoln("statusCode", lw.responseData.status, "size", lw.responseData.size)
		}
		return http.HandlerFunc(lrw)
	}
}
