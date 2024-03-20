package handler

import (
	"context"
	"errors"
	"net/http"
	"runtime"
	"strings"

	"github.com/StasMerzlyakov/go-metrics/internal/server/domain"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

//go:generate mockgen -destination "../mocks/$GOFILE" -package mocks . AdminApp
type AdminApp interface {
	Ping(ctx context.Context) error
}

func AddAdminOperations(r *chi.Mux, adminApp AdminApp, log *zap.SugaredLogger) {

	adapter := &adminOperationAdpater{
		adminApp: adminApp,
		logger:   log,
	}

	r.Get("/ping", adapter.Ping)
}

type adminOperationAdpater struct {
	adminApp AdminApp
	logger   *zap.SugaredLogger
}

func (h *adminOperationAdpater) Ping(w http.ResponseWriter, req *http.Request) {
	action := "Ping"
	if err := h.adminApp.Ping(req.Context()); err != nil {
		h.handlerAppError(err, w)
		return
	}

	w.WriteHeader(http.StatusOK)
	h.logger.Infow(action, "status", "ok")
}

func (h *adminOperationAdpater) handlerAppError(err error, w http.ResponseWriter) {
	pc, _, _, _ := runtime.Caller(1)
	action := runtime.FuncForPC(pc).Name()

	// Получаем строку вида "github.com/StasMerzlyakov/go-metrics/internal/server/adapter/http/handler.(*adminOperationAdpater).Ping"
	lst := strings.Split(action, ".")
	if len(lst) > 1 {
		action = lst[len(lst)-1]
	}

	if errors.Is(err, domain.ErrDataFormat) {
		h.logger.Infow(action, "status", "error", "msg", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	} else {
		h.logger.Infow(action, "status", "error", "msg", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
