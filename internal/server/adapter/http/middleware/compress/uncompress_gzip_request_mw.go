package compress

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/http/middleware"
)

type gzipreadCloser struct {
	*gzip.Reader
	io.Closer
}

func (gz gzipreadCloser) Close() error {
	return gz.Closer.Close()
}

func NewUncompressGZIPRequestMW() middleware.Middleware {

	return func(next http.Handler) http.Handler {
		uncmprFn := func(w http.ResponseWriter, r *http.Request) {
			contentEncodingHeader := r.Header.Get("Content-Encoding")
			if strings.Contains(contentEncodingHeader, "gzip") {
				zr, err := gzip.NewReader(r.Body)
				if err != nil {
					http.Error(w, fmt.Sprintf("uncompress request error - %s", err.Error()), http.StatusInternalServerError)
					return
				}
				r.Body = gzipreadCloser{zr, r.Body}
			}
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(uncmprFn)
	}
}
