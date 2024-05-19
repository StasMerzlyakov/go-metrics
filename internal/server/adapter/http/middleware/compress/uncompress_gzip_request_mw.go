package compress

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/http/middleware"
	"github.com/StasMerzlyakov/go-metrics/internal/server/domain"
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
			log := domain.GetMainLogger()
			contentEncodingHeader := r.Header.Get("Content-Encoding")
			if strings.Contains(contentEncodingHeader, "gzip") {
				zr, err := gzip.NewReader(r.Body)
				if err != nil {
					log.Infow("uncompress request error:", err.Error())
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				r.Body = gzipreadCloser{zr, r.Body}
				log.Infow("readGZIP", "Content-Encoding", r.Header.Get("Content-Encoding"), "msq", "the request will be unzip")
			}
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(uncmprFn)
	}
}
