package digest

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"net/http"

	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/http/middleware"
	"go.uber.org/zap"
)

type hashDigestWriter struct {
	http.ResponseWriter
	hasher hash.Hash
}

var _ http.ResponseWriter = (*hashDigestWriter)(nil)

func (hdw *hashDigestWriter) Header() http.Header {
	return hdw.ResponseWriter.Header()
}

func (hdw *hashDigestWriter) Write(data []byte) (int, error) {
	n, err := hdw.hasher.Write(data)
	if err != nil {
		return n, err
	}

	hashValue := hdw.hasher.Sum(nil)
	hdw.ResponseWriter.Header().Set("HashSHA256", hex.EncodeToString(hashValue))

	size, err := hdw.ResponseWriter.Write(data)
	return size, err
}

func (lw *hashDigestWriter) WriteHeader(statusCode int) {
	hashValue := lw.hasher.Sum(nil)
	lw.ResponseWriter.Header().Set("HashSHA256", hex.EncodeToString(hashValue))
	lw.ResponseWriter.WriteHeader(statusCode)
}

func NewWriteHashDigestResponseHeaderMW(log *zap.SugaredLogger, key string) middleware.Middleware {
	return func(next http.Handler) http.Handler {
		lrw := func(w http.ResponseWriter, r *http.Request) {
			lw := &hashDigestWriter{
				hasher:         hmac.New(sha256.New, []byte(key)),
				ResponseWriter: w,
			}

			next.ServeHTTP(lw, r)
		}
		return http.HandlerFunc(lrw)
	}
}
