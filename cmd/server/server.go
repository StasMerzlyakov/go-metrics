package main

import (
	"flag"
	"fmt"
	"github.com/StasMerzlyakov/go-metrics/internal/server"
	"os"
)

func main() {
	configuration := new(server.Configuration)
	configuration.Set(":8080") // Значение по-умолчанию

	flag.Var(configuration, "a", "serverAddress")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	addr, isExists := os.LookupEnv("ADDRESS")
	if isExists {
		configuration.Set(addr)
	}

	if err := server.CreateServer(configuration); err != nil {
		panic(err)
	}
}
