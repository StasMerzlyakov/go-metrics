// Package retry client request retry middleware
package retry

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"

	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/http/middleware"
	"github.com/StasMerzlyakov/go-metrics/internal/server/domain"
	"go.uber.org/zap/buffer"
)

func NewRetriableRequestMW() middleware.Middleware {
	return NewRetriableRequestMWConf(time.Duration(time.Second), time.Duration(2*time.Second), 4, nil)
}

func NewRetriableRequestMWConf(firstRetryDelay time.Duration, delayIncrement time.Duration, retryCount int, preProccFn domain.ErrPreProcessFn) middleware.Middleware {
	rConf := &domain.RetriableInvokerConf{
		RetriableErr:    domain.ErrServerInternal,
		FirstRetryDelay: firstRetryDelay,
		DelayIncrement:  delayIncrement,
		RetryCount:      retryCount,
		PreProccFn:      nil,
	}

	invoker := domain.CreateRetriableInvokerByConf(rConf)

	return func(next http.Handler) http.Handler {
		infokeFn := func(w http.ResponseWriter, req *http.Request) {

			log := domain.GetCtxLogger(req.Context())

			// Кешируем данные запроса
			body, err := io.ReadAll(req.Body)
			if err != nil {
				errMsg := "can't read request content"
				log.Errorw("RetriableRequestMW", "err", errMsg)
				http.Error(w, errMsg, http.StatusBadRequest)
				return
			}
			defer req.Body.Close()

			respWriter := &responseWriter{
				header: make(map[string][]string),
			}

			invokableFn := func(ctx context.Context) error {
				respWriter.Clear()                             // Нужно очистить данные перед каждым вызовом
				req.Body = io.NopCloser(bytes.NewReader(body)) // Устанавливаем тело запроса
				next.ServeHTTP(respWriter, req)
				if respWriter.status == http.StatusInternalServerError {
					// По конфигурации rConf на ошибку domain.ErrServerInternal invoker будет повторять операцию
					return domain.ErrServerInternal
				}
				return nil
			}

			err = invoker.Invoke(req.Context(), invokableFn)
			if err != nil {
				log.Errorw("RetriableRequestMW", "error", err.Error())
				w.WriteHeader(domain.MapDomainErrorToHTTPStatusErr(err))
				return
			}

			// Заполняем данные для ответа в порядке Header, StatusCode, Body
			for k, vs := range respWriter.Header() {
				for _, v := range vs {
					w.Header().Add(k, v)
				}
			}

			if respWriter.status != 0 {
				w.WriteHeader(respWriter.status)
			}

			_, err = w.Write(respWriter.buf.Bytes())
			if err != nil {
				log.Errorw("RetriableRequestMW", "err", err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		return http.HandlerFunc(infokeFn)
	}
}

type responseWriter struct {
	buf    buffer.Buffer
	status int
	header http.Header
}

var _ http.ResponseWriter = (*responseWriter)(nil)

func (rw *responseWriter) Clear() {
	rw.buf.Reset()
	rw.status = 0
	for k := range rw.header {
		delete(rw.header, k)
	}
}

func (rw *responseWriter) Header() http.Header {
	return rw.header
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	size, err := rw.buf.Write(data)
	return size, err
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.status = statusCode
}
