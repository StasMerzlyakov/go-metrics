package compress

import (
	"io"
	"net/http"
	"strings"

	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/http/middleware"
	gPool "github.com/ungerik/go-pool"
)

// NewCompressGZIPResponseMW Вариант мидлы без дополнительного буфера при обработке ответа.
func NewCompressGZIPResponseMW() middleware.Middleware {
	return func(next http.Handler) http.Handler {
		cmprFn := func(w http.ResponseWriter, r *http.Request) {
			acceptEncodingReqHeader := r.Header.Get("Accept-Encoding")
			if !strings.Contains(acceptEncodingReqHeader, "gzip") {
				next.ServeHTTP(w, r)
			} else {
				gz := gPool.Gzip.GetWriter(w)
				defer gPool.Gzip.PutWriter(gz)
				next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz, IsGZIPHeaderRecorded: false, IsContentTypeVerified: false}, r)
			}
		}
		return http.HandlerFunc(cmprFn)
	}
}

type gzipWriter struct {
	http.ResponseWriter
	Writer                io.Writer
	IsGZIPHeaderRecorded  bool
	IsContentTypeVerified bool
}

func (w gzipWriter) Write(b []byte) (int, error) {
	// С помощью двух флагов определяем нужно ли записать заголовок и сжать
	if !w.IsContentTypeVerified {
		w.IsContentTypeVerified = true
		contentTypeRespHeader := w.Header().Get("Content-Type")
		if strings.Contains(contentTypeRespHeader, "application/json") ||
			strings.Contains(contentTypeRespHeader, "text/html") {
			w.IsGZIPHeaderRecorded = true
			w.Header().Add("Content-Encoding", "gzip") // добавляем заголовок
		}
	}

	if w.IsGZIPHeaderRecorded {
		return w.Writer.Write(b) // пишем в gzip
	} else {
		return w.ResponseWriter.Write(b) // пишем в обычный поток
	}
}
