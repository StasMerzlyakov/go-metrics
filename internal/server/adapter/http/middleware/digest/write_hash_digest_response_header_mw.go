package digest

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"net/http"

	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/http/middleware"
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

func (hdw *hashDigestWriter) WriteHeader(statusCode int) {
	hashValue := hdw.hasher.Sum(nil)
	hdw.ResponseWriter.Header().Set("HashSHA256", hex.EncodeToString(hashValue))
	hdw.ResponseWriter.WriteHeader(statusCode)
}

func NewWriteHashDigestResponseHeaderMW(key string) middleware.Middleware {
	return func(next http.Handler) http.Handler {
		lrw := func(w http.ResponseWriter, r *http.Request) {
			hdw := &hashDigestWriter{
				hasher:         hmac.New(sha256.New, []byte(key)),
				ResponseWriter: w,
			}

			next.ServeHTTP(hdw, r)
		}
		return http.HandlerFunc(lrw)
	}
}
