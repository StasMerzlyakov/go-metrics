package compress

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/http/middleware"
	"github.com/StasMerzlyakov/go-metrics/internal/server/domain"
)

type bufferWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w bufferWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// Вариант мидлы через буфер. Можно оценить ответ.
func NewCompressGZIPBufferResponseMW() middleware.Middleware {
	return func(next http.Handler) http.Handler {
		cmprFn := func(w http.ResponseWriter, r *http.Request) {
			log := domain.GetMainLogger()
			acceptEncodingReqHeader := r.Header.Get("Accept-Encoding")
			if !strings.Contains(acceptEncodingReqHeader, "gzip") {
				next.ServeHTTP(w, r)
			} else {
				var buff bytes.Buffer // Пишем в буфер
				next.ServeHTTP(bufferWriter{ResponseWriter: w, Writer: &buff}, r)

				contentTypeRespHeader := w.Header().Get("Content-Type")
				if strings.Contains(contentTypeRespHeader, "application/json") ||
					strings.Contains(contentTypeRespHeader, "text/html") {
					w.Header().Add("Content-Encoding", "gzip") // добавляем заголовок
					gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
					if err != nil {
						log.Errorln("writeGZIP", "desc", "can't initialize gzip", "err", err.Error())
						http.Error(w, "can't initialize gzip", http.StatusInternalServerError)
						return
					}
					defer gz.Close()
					_, err = gz.Write(buff.Bytes())
					if err != nil {
						log.Errorln("writeGZIP", "status", "err", "msg", fmt.Sprintf("can't initialize gzip: %v", err.Error()))
						http.Error(w, "can't write gzip", http.StatusInternalServerError)
						return
					}

					log.Infow("writeGZIP", "header", "Content-Type", "value", contentTypeRespHeader, "msg", "response will be zipped")
				} else {
					w.Write(buff.Bytes())
				}

			}
		}
		return http.HandlerFunc(cmprFn)
	}
}
