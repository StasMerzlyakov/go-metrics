package main

import (
	"context"
	"github.com/StasMerzlyakov/go-metrics/internal/agent"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	configuration := agent.Configuration{
		ServerAddr:        "http://localhost:8080",
		ContentType:       "text/plain",
		PollIntervalSec:   2,
		ReportIntervalSec: 10,
	}

	// Взято отсюда: "Реализация Graceful Shutdown в Go"(https://habr.com/ru/articles/771626/)
	// скорее на будущее для сервера
	ctx, cancel := context.WithCancel(context.Background())
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)

	if agent, err := agent.CreateAgent(ctx, configuration); err != nil {
		panic(err)
	} else {
		<-exit
		cancel()
		agent.Wait() // ожидаение завершения go-рутин в агенте
	}
}
