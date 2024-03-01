package server

import (
	"context"
	"net/http"
	"sync"

	"github.com/StasMerzlyakov/go-metrics/internal/config"
	"go.uber.org/zap"
)

func NewMetricsServer(config *config.ServerConfiguration,
	sugar *zap.SugaredLogger,
	httpHandler http.Handler,
) *meterServer {
	return &meterServer{
		srv: &http.Server{
			Addr:        config.URL,
			Handler:     httpHandler,
			ReadTimeout: 0,
			IdleTimeout: 0,
		},
		sugar: sugar,
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
