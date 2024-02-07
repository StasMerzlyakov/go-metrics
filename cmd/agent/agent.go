package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/StasMerzlyakov/go-metrics/internal/agent"
	"github.com/caarlos0/env"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	agentCfg := agent.Configuration{}

	flag.StringVar(&agentCfg.ServerAddr, "a", "localhost:8080", "serverAddress")
	flag.IntVar(&agentCfg.PollInterval, "p", 2, "poolInterval in seconds")
	flag.IntVar(&agentCfg.ReportInterval, "r", 10, "poolInterval in seconds")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	err := env.Parse(&agentCfg)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("agentCfg: %v\n", agentCfg)

	// Взято отсюда: "Реализация Graceful Shutdown в Go"(https://habr.com/ru/articles/771626/)
	// Сейчас выглядит избыточным - оставил как задел на будущее для сервера
	ctx, cancel := context.WithCancel(context.Background())
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)

	if agent, err := agent.CreateAgent(ctx, &agentCfg); err != nil {
		panic(err)
	} else {
		<-exit
		cancel()
		agent.Wait() // ожидаение завершения go-рутин в агенте
	}
}
