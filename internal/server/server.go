package server

import (
	"context"
	"net/http"
	"sync"

	"github.com/StasMerzlyakov/go-metrics/internal/config"
	"go.uber.org/zap"
)

type Backuper interface {
	RestoreBackUp() error
	Run(ctx context.Context) error
	WaitDone()
}

func NewMetricsServer(config *config.ServerConfiguration,
	sugar *zap.SugaredLogger,
	httpHandler http.Handler,
	backUper Backuper,
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
	backuper     Backuper
	startContext context.Context
}

func (s *meterServer) Start(startContext context.Context) {
	s.startContext = startContext
	go func() {
		s.wg.Add(1)
		defer s.wg.Done()
		if err := s.srv.ListenAndServe(); err != http.ErrServerClosed {
			s.sugar.Fatalw("srv.ListenAndServe", "msg", err.Error())
		}
	}()
	if s.backuper != nil {
		go func() {
			s.wg.Add(1)
			defer s.wg.Done()
			if err := s.backuper.Run(startContext); err != nil {
				s.sugar.Fatalw("backuper.Run", "msg", err.Error())
			}
		}()
	}
	s.sugar.Infof("Server started")
}

func (s *meterServer) WaitDone() {
	s.srv.Shutdown(s.startContext)
	s.wg.Wait()
	s.sugar.Infof("WaitDone")
}
