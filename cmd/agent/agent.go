package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/StasMerzlyakov/go-metrics/internal/agent"
	"github.com/StasMerzlyakov/go-metrics/internal/common/errors/retriable"
	"github.com/StasMerzlyakov/go-metrics/internal/config"
)

type Agent interface {
	Start(ctx context.Context)
	Wait()
}

func main() {
	agentCfg, err := config.LoadAgentConfig()
	if err != nil {
		log.Fatal(err)
	}

	metricStorage := agent.NewMemStatsStorage()
	resultSender := agent.NewHTTPResultSender(agentCfg.ServerAddr)
	retryCfg := retriable.DefaultConf(syscall.ECONNREFUSED)
	retryableResultSender := agent.NewHTTPRetryableResultSender(*retryCfg, resultSender)

	var agnt Agent = agent.Create(agentCfg,
		retryableResultSender,
		metricStorage,
	)

	// Взято отсюда: "Реализация Graceful Shutdown в Go"(https://habr.com/ru/articles/771626/)
	// Сейчас выглядит избыточным - оставил как задел на будущее для сервера
	ctx, cancelFn := context.WithCancel(context.Background())
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)

	agnt.Start(ctx)
	defer func() {
		cancelFn()
		agnt.Wait()
	}()
	<-exit
}
