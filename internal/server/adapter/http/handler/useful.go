package handler

import (
	"errors"
	"fmt"
	"net/http"
	"github.com/StasMerzlyakov/go-metrics/internal/server/domain"
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

func handleAppError(w http.ResponseWriter, err error) {
	logger := domain.GetMainLogger()

	action := domain.GetAction(2) //  интересует имя метода, из которого взывался handleAppError

	if errors.Is(err, ErrMediaType) {
		logger.Infow(action, "status", "error", "msg", err.Error())
		http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
		return
	}

	if errors.Is(err, domain.ErrDataFormat) ||
		errors.Is(err, domain.ErrDataDigestMismath) {
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
