package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/StasMerzlyakov/go-metrics/internal/agent"
)

func main() {
	agentCfg, err := agent.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	// Взято отсюда: "Реализация Graceful Shutdown в Go"(https://habr.com/ru/articles/771626/)
	// Сейчас выглядит избыточным - оставил как задел на будущее для сервера
	ctx, cancel := context.WithCancel(context.Background())
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)

	if agent, err := agent.CreateAgent(ctx, agentCfg); err != nil {
		panic(err)
	} else {
		<-exit
		cancel()
		agent.Wait() // ожидаение завершения go-рутин в агенте
	}
}
