package server

import (
	"context"
	"github.com/StasMerzlyakov/go-metrics/internal"
	"github.com/StasMerzlyakov/go-metrics/internal/storage"
	"github.com/go-chi/chi/v5"
	"net/http"
)

type Configuration struct {
}

// TODO сделать более удобную конфигурацию (напрашивается Builder).

func CreateServer(ctx context.Context, config Configuration) error {
	counterStorage := storage.NewMemoryInt64Storage()
	gaugeStorage := storage.NewMemoryFloat64Storage()
	gaugePostHandler := GaugePostHandlerCreator(gaugeStorage)
	counterPostHandler := CounterPostHandlerCreator(counterStorage)
	allMetricsHandler := AllMetricsViewHandlerCreator(counterStorage, gaugeStorage)
	gaugeGetHandler := GaugeGetHandlerCreator(gaugeStorage)
	counterGetHandler := CounterGetHandlerCreator(counterStorage)

	serverHandler := CreateServerHandler(
		gaugePostHandler,
		gaugeGetHandler,
		counterPostHandler,
		counterGetHandler,
		allMetricsHandler)
	return http.ListenAndServe(":8080", serverHandler)
}

func CreateServerHandler(
	gaugePostHandler http.HandlerFunc,
	gaugeGetHandler http.HandlerFunc,
	counterPostHandler http.HandlerFunc,
	counterGetHandler http.HandlerFunc,
	allMetricsHandler http.HandlerFunc,

) http.Handler {

	fullGaugeHandler := CreateFullPostGaugeHandler(gaugePostHandler)
	fullCounterHandler := CreateFullPostCounterHandler(counterPostHandler)
	r := chi.NewRouter()
	r.Get("/", allMetricsHandler)
	r.Post("/update/gauge/{name}/{value}", fullGaugeHandler)

	r.Route("/update", func(r chi.Router) {
		r.Post("/gauge/{name}/{value}", fullGaugeHandler)
		r.Post("/gauge/{name}", internal.StatusNotFound)
		r.Post("/counter/{name}/{value}", fullCounterHandler)
		r.Post("/{type}/{name}/{value}", internal.StatusNotImplemented)
	})

	r.Route("/value", func(r chi.Router) {
		r.Get("/gauge/{name}", gaugeGetHandler)
		r.Get("/counter/{name}", counterGetHandler)
	})
	return r
}
