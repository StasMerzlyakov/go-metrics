package digest

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"

	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/http/middleware"
	"github.com/StasMerzlyakov/go-metrics/internal/server/domain"
)

// NewCheckHashDigestRequestBufferedMW Реализация с буфером. Хэш проверяется прямо в мидле. Для этого читается req.Body
func NewCheckHashDigestRequestBufferedMW(key string) middleware.Middleware {
	return func(next http.Handler) http.Handler {
		cmprFn := func(w http.ResponseWriter, req *http.Request) {
			log := domain.GetCtxLogger(req.Context())
			action := domain.GetAction(1)
			hashSHA256Hex := req.Header.Get("HashSHA256")
			if hashSHA256Hex == "" {
				errMsg := fmt.Errorf("%w: HashSHA256 header is not specified", domain.ErrDataFormat)
				log.Errorw(action, "error", errMsg.Error())
				http.Error(w, "", http.StatusNotFound)
				return
			}

			hashSHA256, err := hex.DecodeString(hashSHA256Hex)
			if err != nil {
				errMsg := fmt.Errorf("%w: decode HashSHA256 header err: %v", domain.ErrDataDigestMismath, err.Error())
				log.Infow(action, "error", errMsg.Error())
				http.Error(w, "", http.StatusBadRequest)
				return
			}

			hasher := hmac.New(sha256.New, []byte(key))

			reqBytes, _ := io.ReadAll(req.Body)
			defer req.Body.Close()

			hasher.Write(reqBytes)
			hashValue := hasher.Sum(nil)

			if !bytes.Equal(hashSHA256, hashValue) {
				err := fmt.Errorf("%w: expected %v, actual %v",
					domain.ErrDataDigestMismath, hex.EncodeToString(hashSHA256), hex.EncodeToString(hashValue))
				log.Infow(action, "error", err.Error())
				http.Error(w, "", http.StatusBadRequest)
				return

			}

			reader := bytes.NewReader(reqBytes)
			req.Body = io.NopCloser(reader)
			next.ServeHTTP(w, req)
		}
		return http.HandlerFunc(cmprFn)
	}
}
