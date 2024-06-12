package digest

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"net/http"

	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/http/middleware"
	"github.com/StasMerzlyakov/go-metrics/internal/server/domain"
)

// NewCheckHashDigestRequestMW Реализация без буфера. Хэш проверяется при возникновении EOF при чтении req.Body
// Требуется обязательное вычитывание req.Body в http.Handler и обработка ответа
//
// _, err := io.ReadAll(req.Body)
//
//	if err != nil && err != io.EOF {
//		http.Error(w, err.Error(), http.StatusBadRequest)
func NewCheckHashDigestRequestMW(key string) middleware.Middleware {
	return func(next http.Handler) http.Handler {
		cmprFn := func(w http.ResponseWriter, r *http.Request) {
			log := domain.GetCtxLogger(r.Context())

			if r.Method == http.MethodGet {
				next.ServeHTTP(w, r)
				return
			}

			hashSHA256Hex := r.Header.Get("HashSHA256")
			if hashSHA256Hex == "" {
				if _, err := io.ReadAll(r.Body); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				defer r.Body.Close()

				errMsg := fmt.Errorf("%w: HashSHA256 header is not specified", domain.ErrDataFormat)
				log.Infow("check_hash_digest_request_mw", "err", errMsg.Error())
				http.Error(w, errMsg.Error(), http.StatusNotFound)
				return
			}

			hashSHA256, err := hex.DecodeString(hashSHA256Hex)
			if err != nil {
				if _, err := io.ReadAll(r.Body); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				defer r.Body.Close()

				errMsg := fmt.Errorf("%w: decode HashSHA256 header err: %v", domain.ErrDataDigestMismath, err.Error())
				log.Infow("check_hash_digest_request_mw", "err", errMsg.Error())
				http.Error(w, errMsg.Error(), http.StatusBadRequest)
				return
			}

			hasher := hmac.New(sha256.New, []byte(key))
			reader := &hasherReader{
				hasher:     hasher,
				expected:   hashSHA256,
				ReadCloser: r.Body,
			}
			r.Body = reader
			next.ServeHTTP(w, r)

		}
		return http.HandlerFunc(cmprFn)
	}
}

type hasherReader struct {
	hasher   hash.Hash
	expected []byte
	io.ReadCloser
}

func (hr *hasherReader) Read(p []byte) (n int, err error) {
	n, err = hr.ReadCloser.Read(p)

	if err != nil && err != io.EOF {
		return
	}

	if err == nil {
		return hr.hasher.Write(p[:n])
	} else {
		// проверяем хэш при появлении EOF
		hr.hasher.Write(p[:n])
		value := hr.hasher.Sum(nil)
		if !bytes.Equal(hr.expected, value) {
			fullErr := fmt.Errorf("%w: expected %v, actual %v",
				domain.ErrDataDigestMismath,
				hex.EncodeToString(hr.expected),
				hex.EncodeToString(value))
			return 0, fullErr
		}
	}
	return
}

func (hr *hasherReader) Close() error {
	return hr.ReadCloser.Close()
}
