package server

import (
	"net/http"

	"github.com/StasMerzlyakov/go-metrics/internal"
	"github.com/StasMerzlyakov/go-metrics/internal/storage"
	"github.com/go-chi/chi/v5"
)

func CreateServer(config *Configuration) error {
	counterStorage := storage.NewMemoryInt64Storage()
	gaugeStorage := storage.NewMemoryFloat64Storage()
	gaugePostHandler := GaugePostHandlerCreator(gaugeStorage)
	counterPostHandler := CounterPostHandlerCreator(counterStorage)
	allMetricsHandler := AllMetricsViewHandlerCreator(counterStorage, gaugeStorage)
	gaugeGetHandler := GaugeGetHandlerCreator(gaugeStorage)
	counterGetHandler := CounterGetHandlerCreator(counterStorage)
	serverHandler := Builder.NewConfigurationBuilder().
		GaugePostHandler(gaugePostHandler).
		GaugeGetHandler(gaugeGetHandler).
		CounterPostHandler(counterPostHandler).
		CounterGetHandler(counterGetHandler).
		AllMetricsHandler(allMetricsHandler).Build()
	server := &http.Server{Addr: config.url, Handler: serverHandler, ReadTimeout: 0, IdleTimeout: 0}
	return server.ListenAndServe()
}

var Builder = &handlerConfigurationBuilder{}

type handlerConfigurationBuilder struct{}

func (b *handlerConfigurationBuilder) NewConfigurationBuilder() *handlerConfiguration {
	return &handlerConfiguration{
		allMetricsHandler:  DefaultHandler,
		gaugeGetHandler:    DefaultHandler,
		gaugePostHandler:   DefaultHandler,
		counterGetHandler:  DefaultHandler,
		counterPostHandler: DefaultHandler,
	}
}

type handlerConfiguration struct {
	allMetricsHandler  http.HandlerFunc
	gaugePostHandler   http.HandlerFunc
	counterPostHandler http.HandlerFunc
	gaugeGetHandler    http.HandlerFunc
	counterGetHandler  http.HandlerFunc
}

func (hCfg *handlerConfiguration) AllMetricsHandler(handler http.HandlerFunc) *handlerConfiguration {
	hCfg.allMetricsHandler = handler
	return hCfg
}

func (hCfg *handlerConfiguration) GaugePostHandler(handler http.HandlerFunc) *handlerConfiguration {
	hCfg.gaugePostHandler = handler
	return hCfg
}

func (hCfg *handlerConfiguration) CounterPostHandler(handler http.HandlerFunc) *handlerConfiguration {
	hCfg.counterPostHandler = handler
	return hCfg
}

func (hCfg *handlerConfiguration) GaugeGetHandler(handler http.HandlerFunc) *handlerConfiguration {
	hCfg.gaugeGetHandler = handler
	return hCfg
}

func (hCfg *handlerConfiguration) CounterGetHandler(handler http.HandlerFunc) *handlerConfiguration {
	hCfg.counterGetHandler = handler
	return hCfg
}

func (hCfg *handlerConfiguration) Build() http.Handler {
	fullGaugeHandler := CreateFullPostGaugeHandler(hCfg.gaugePostHandler)
	fullCounterHandler := CreateFullPostCounterHandler(hCfg.counterPostHandler)
	r := chi.NewRouter()
	r.Get("/", hCfg.allMetricsHandler)
	r.Post("/update/gauge/{name}/{value}", fullGaugeHandler)

	r.Route("/update", func(r chi.Router) {
		r.Post("/gauge/{name}/{value}", fullGaugeHandler)
		r.Post("/gauge/{name}", internal.StatusNotFound)
		r.Post("/counter/{name}/{value}", fullCounterHandler)
		r.Post("/{type}/{name}/{value}", internal.StatusNotImplemented)
	})

	r.Route("/value", func(r chi.Router) {
		r.Get("/gauge/{name}", hCfg.gaugeGetHandler)
		r.Get("/counter/{name}", hCfg.counterGetHandler)
	})
	return r
}
