package server

import (
	"context"
	"net/http"
	"sync"

	"github.com/StasMerzlyakov/go-metrics/internal/config"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type HttpAdapter interface {
	PostGauge(w http.ResponseWriter, req *http.Request)
	GetGauge(w http.ResponseWriter, req *http.Request)
	PostCounter(w http.ResponseWriter, req *http.Request)
	GetCounter(w http.ResponseWriter, req *http.Request)
	AllMetrics(w http.ResponseWriter, request *http.Request)
}

func CreateMeterServer(config *config.ServerConfiguration,
	logger *zap.Logger,
	httpAdapter HttpAdapter,
) *meterServer {
	return &meterServer{
		sugar: logger.Sugar(),
		srv: &http.Server{
			Addr:        config.Url,
			Handler:     createHTTPHandler(httpAdapter),
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
}

func (s *meterServer) WaitDone() {
	s.srv.Shutdown(s.startContext)
	s.wg.Wait()
	s.sugar.Infof("server complete")
}

func createHTTPHandler(httpAdapter HttpAdapter) http.Handler {

	fullGaugeHandler := createFullPostGaugeHandler(httpAdapter.PostGauge)
	fullCounterHandler := createFullPostCounterHandler(httpAdapter.PostCounter)
	r := chi.NewRouter()
	r.Get("/", httpAdapter.AllMetrics)
	r.Post("/update/gauge/{name}/{value}", fullGaugeHandler)

	r.Route("/update", func(r chi.Router) {
		r.Post("/gauge/{name}/{value}", fullGaugeHandler)
		r.Post("/gauge/{name}", StatusNotFound)
		r.Post("/counter/{name}/{value}", fullCounterHandler)
		r.Post("/{type}/{name}/{value}", StatusNotImplemented)
	})

	r.Route("/value", func(r chi.Router) {
		r.Get("/gauge/{name}", httpAdapter.GetGauge)
		r.Get("/counter/{name}", httpAdapter.GetCounter)
	})
	return r
}

func createFullPostCounterHandler(counterHandler http.HandlerFunc) http.HandlerFunc {
	return Conveyor(
		counterHandler,
		CheckIntegerMiddleware,
		CheckMetricNameMiddleware,
		CheckContentTypeMiddleware,
		CheckMethodPostMiddleware,
	)
}

func createFullPostGaugeHandler(gaugeHandler http.HandlerFunc) http.HandlerFunc {
	return Conveyor(
		gaugeHandler,
		CheckDigitalMiddleware,
		CheckMetricNameMiddleware,
		CheckContentTypeMiddleware,
		CheckMethodPostMiddleware,
	)
}
