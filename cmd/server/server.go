package main

import (
	"github.com/StasMerzlyakov/go-metrics/internal/server"
)

func main() {
	if err := server.CreateServer(); err != nil {
		panic(err)
	}
}
