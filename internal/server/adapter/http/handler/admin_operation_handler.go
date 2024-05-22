package handler

import (
	"context"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
)

//go:generate mockgen -destination "../mocks/$GOFILE" -package mocks . AdminApp
type AdminApp interface {
	Ping(ctx context.Context) error
}

func AddAdminOperations(r *chi.Mux, adminApp AdminApp) {

	adapter := &adminOperationAdpater{
		adminApp: adminApp,
	}

	r.Get("/ping", adapter.Ping)
}

type adminOperationAdpater struct {
	adminApp AdminApp
}

// Ping Отвечает за проверку соединения с базой данных.
//
// GET /ping
// Возвращает [http.StatusOK] в случае успешной проверки.
func (h *adminOperationAdpater) Ping(w http.ResponseWriter, req *http.Request) {

	_, _ = io.ReadAll(req.Body)
	defer req.Body.Close()

	if err := h.adminApp.Ping(req.Context()); err != nil {
		handleAppError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
