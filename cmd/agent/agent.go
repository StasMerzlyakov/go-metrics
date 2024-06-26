package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/StasMerzlyakov/go-metrics/internal/agent"
	"github.com/StasMerzlyakov/go-metrics/internal/agent/retriable"
	"github.com/StasMerzlyakov/go-metrics/internal/config"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func printVersion() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}

type Agent interface {
	Start(ctx context.Context)
	Wait()
}

func main() {

	printVersion()

	agentCfg, err := config.LoadAgentConfig()
	if err != nil {
		log.Fatal(err)
	}

	// Отвечает за сбор метрик
	metricStorage := agent.NewMemStatsStorage()

	// Отвечает за отправку по http
	resultSender := agent.NewHTTPResultSender(agentCfg)

	// Отвечает за повтор отправки
	retryCfg := retriable.DefaultConf(syscall.ECONNREFUSED)
	retryableResultSender := agent.NewHTTPRetryableResultSender(*retryCfg, resultSender)

	// Отвечает за пулы отправки
	limitedResultSender := agent.NewPoolResultSender(agentCfg, retryableResultSender)

	var agnt Agent = agent.Create(agentCfg,
		limitedResultSender,
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
