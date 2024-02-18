package server

import (
	"net/http"

	"github.com/StasMerzlyakov/go-metrics/internal"
	"github.com/StasMerzlyakov/go-metrics/internal/storage"
	"github.com/go-chi/chi/v5"
)

func CreateServer(config *Configuration) error {
	serverHandler := CreateServerHandler()
	server := &http.Server{Addr: config.url, Handler: serverHandler, ReadTimeout: 0, IdleTimeout: 0}
	return server.ListenAndServe()
}

func CreateServerHandler() http.Handler {
	counterStorage := storage.NewMemoryInt64Storage()
	gaugeStorage := storage.NewMemoryFloat64Storage()
	metricController := NewMetricController(counterStorage, gaugeStorage)
	businessHandler := NewBusinessHandler(metricController)
	return CreateFullHttpHandler(businessHandler)
}

func CreateFullHttpHandler(businessHandler BusinessHandler) http.Handler {

	fullGaugeHandler := CreateFullPostGaugeHandler(businessHandler.PostGauge)
	fullCounterHandler := CreateFullPostCounterHandler(businessHandler.PostCounter)
	r := chi.NewRouter()
	r.Get("/", businessHandler.AllMetrics)
	r.Post("/update/gauge/{name}/{value}", fullGaugeHandler)

	r.Route("/update", func(r chi.Router) {
		r.Post("/gauge/{name}/{value}", fullGaugeHandler)
		r.Post("/gauge/{name}", internal.StatusNotFound)
		r.Post("/counter/{name}/{value}", fullCounterHandler)
		r.Post("/{type}/{name}/{value}", internal.StatusNotImplemented)
	})

	r.Route("/value", func(r chi.Router) {
		r.Get("/gauge/{name}", businessHandler.GetGauge)
		r.Get("/counter/{name}", businessHandler.GetCounter)
	})
	return r
}
