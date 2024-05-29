package handler

import (
	"context"
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

func handleAppError(ctx context.Context, w http.ResponseWriter, err error) {
	logger := domain.GetCtxLogger(ctx)

	action := domain.GetAction(2) //  интересует имя метода, из которого взывался handleAppError
	logger.Infow(action, "error", err.Error())
	http.Error(w, err.Error(), domain.MapDomainErrorToHTTPStatusErr(err))
}
