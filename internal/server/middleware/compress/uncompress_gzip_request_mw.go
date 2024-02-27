package compress

import (
	"io"
	"net/http"
	"strings"

	"github.com/StasMerzlyakov/go-metrics/internal/server/middleware"
	gPool "github.com/ungerik/go-pool"
	"go.uber.org/zap"
)

func NewUncompressGZIPRequestMW(log *zap.SugaredLogger) middleware.Middleware {

	return func(next http.Handler) http.Handler {

		uncmprFn := func(w http.ResponseWriter, r *http.Request) {
			var reader io.ReadCloser

			if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
				gz := gPool.Gzip.GetReader(r.Body)
				defer gPool.Gzip.PutReader(gz)
				reader = gz
				log.Info("content", "compresed")
			} else {
				reader = r.Body
				log.Info("content", "uncompresed")
			}

			r.Body = reader
			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(uncmprFn)
	}
}
