package cryptomw

import (
	"bytes"
	"crypto/rsa"
	"io"
	"net/http"

	"github.com/StasMerzlyakov/go-metrics/internal/keygen"
	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/http/middleware"
	"github.com/StasMerzlyakov/go-metrics/internal/server/domain"
)

// NewDecrytpMw дешифрует входящие данные
func NewDecrytpMw(privKey *rsa.PrivateKey) middleware.Middleware {
	return func(next http.Handler) http.Handler {
		lrw := func(w http.ResponseWriter, req *http.Request) {
			log := domain.GetCtxLogger(req.Context())
			action := domain.GetAction(1)
			encryptedData, err := io.ReadAll(req.Body)

			if err != nil {
				log.Infow(action, "error", err.Error())
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			decryptedBytes, err := keygen.DecryptWithPrivateKey(encryptedData, privKey)

			if err != nil {
				log.Infow(action, "error", err.Error())
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			reader := bytes.NewReader(decryptedBytes)
			req.Body = io.NopCloser(reader)
			next.ServeHTTP(w, req)
		}
		return http.HandlerFunc(lrw)
	}
}
