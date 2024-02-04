package main

import (
	"context"
	"github.com/StasMerzlyakov/go-metrics/internal/server"
)

func main() {
	ctx := context.Background()             // TODO
	configuration := server.Configuration{} // TODO
	if err := server.CreateServer(ctx, configuration); err != nil {
		panic(err)
	}
}
