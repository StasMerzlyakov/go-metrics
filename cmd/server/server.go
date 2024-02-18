package main

import (
	"log"

	"github.com/StasMerzlyakov/go-metrics/internal/server"
)

func main() {

	srvConf, err := server.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	if err := server.CreateServer(srvConf); err != nil {
		panic(err)
	}
}
