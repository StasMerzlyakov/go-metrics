package handler

import (
	"context"
	"io"
	"net/http"

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

	_, _ = io.ReadAll(req.Body)
	defer req.Body.Close()

	if err := h.adminApp.Ping(req.Context()); err != nil {
		handleAppError(w, err, h.logger)
		return
	}

	w.WriteHeader(http.StatusOK)
}
