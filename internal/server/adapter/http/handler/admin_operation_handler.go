package handler

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
)

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

	if _, err := io.ReadAll(req.Body); err != nil {
		handleAppError(req.Context(), w, err)
		return
	}

	defer req.Body.Close()

	if err := h.adminApp.Ping(req.Context()); err != nil {
		handleAppError(req.Context(), w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
