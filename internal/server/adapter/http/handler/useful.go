package handler

import (
	"errors"
	"fmt"
	"net/http"
	"runtime"

	"github.com/StasMerzlyakov/go-metrics/internal/server/domain"
	"go.uber.org/zap"
)

const (
	ApplicationJSON = "application/json"
	TextPlain       = "text/plain"
	TextHTML        = "text/html"
)

func BadRequestHandler(w http.ResponseWriter, req *http.Request) {
	http.Error(w, "BadRequest", http.StatusBadRequest)
}

func StatusMethodNotAllowedHandler(w http.ResponseWriter, req *http.Request) {
	http.Error(w, "StatusMethodNotAllowed", http.StatusMethodNotAllowed)
}
func StatusNotImplemented(w http.ResponseWriter, req *http.Request) {
	http.Error(w, "StatusMethodNotAllowed", http.StatusNotImplemented)
}

func StatusNotFound(w http.ResponseWriter, req *http.Request) {
	http.Error(w, "StatusNotFound", http.StatusNotFound)
}

func TodoResponse(res http.ResponseWriter, message string) {
	res.Header().Set("Content-Type", ApplicationJSON)
	res.WriteHeader(http.StatusNotImplemented)
	fmt.Fprintf(res, `
      {
        "response": {
          "text": "%v"
        },
        "version": "1.0"
      }
    `, message)
}

func handleAppError(w http.ResponseWriter, err error, logger *zap.SugaredLogger) {

	// используется для получения имени вызывющей функции
	// по мотивам https://stackoverflow.com/questions/25927660/how-to-get-the-current-function-name
	pc, _, _, _ := runtime.Caller(1)
	action := runtime.FuncForPC(pc).Name()

	if errors.Is(err, ErrMediaType) {
		logger.Infow(action, "status", "error", "msg", err.Error())
		http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
		return
	}

	if errors.Is(err, domain.ErrDataFormat) {
		logger.Infow(action, "status", "error", "msg", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if errors.Is(err, domain.ErrServerInternal) ||
		errors.Is(err, domain.ErrDBConnection) {
		logger.Infow(action, "status", "error", "msg", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if errors.Is(err, domain.ErrNotFound) {
		logger.Infow(action, "status", "error", "msg", err.Error())
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
}
