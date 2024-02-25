package server

import (
	"context"
	"net/http"
	"sync"

	"github.com/StasMerzlyakov/go-metrics/internal/config"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type HTTPAdapter interface {
	PostGauge(w http.ResponseWriter, req *http.Request)
	GetGauge(w http.ResponseWriter, req *http.Request)
	PostCounter(w http.ResponseWriter, req *http.Request)
	GetCounter(w http.ResponseWriter, req *http.Request)
	AllMetrics(w http.ResponseWriter, request *http.Request)
}

func CreateMeterServer(config *config.ServerConfiguration,
	httpAdapter HTTPAdapter,
	middlewares ...func(http.Handler) http.Handler,
) *meterServer {
	return &meterServer{
		sugar: config.Log,
		srv: &http.Server{
			Addr:        config.URL,
			Handler:     createHTTPHandler(httpAdapter, middlewares...),
			ReadTimeout: 0,
			IdleTimeout: 0,
		},
	}
}

type meterServer struct {
	sugar        *zap.SugaredLogger
	srv          *http.Server
	wg           sync.WaitGroup
	startContext context.Context
}

func (s *meterServer) Start(startContext context.Context) {
	s.startContext = startContext
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := s.srv.ListenAndServe(); err != http.ErrServerClosed {
			s.sugar.Fatalf("ListenAndServe(): %v", err)
		}
	}()
	s.sugar.Infof("Server started")
}

func (s *meterServer) WaitDone() {
	s.srv.Shutdown(s.startContext)
	s.wg.Wait()
	s.sugar.Infof("WaitDone")
}

func createHTTPHandler(httpAdapter HTTPAdapter, middlewares ...func(http.Handler) http.Handler) http.Handler {
	r := chi.NewRouter()

	r.Use(middlewares...)

	r.Get("/", httpAdapter.AllMetrics)

	r.Route("/update", func(r chi.Router) {
		r.Post("/gauge/{name}/{value}", httpAdapter.PostGauge)
		r.Post("/gauge/{name}", StatusNotFound)
		r.Post("/counter/{name}/{value}", httpAdapter.PostCounter)
		r.Post("/counter/{name}", StatusNotFound)
		r.Post("/{type}/{name}/{value}", StatusNotImplemented)
	})

	r.Route("/value", func(r chi.Router) {
		r.Get("/gauge/{name}", httpAdapter.GetGauge)
		r.Get("/counter/{name}", httpAdapter.GetCounter)
	})
	return r
}
