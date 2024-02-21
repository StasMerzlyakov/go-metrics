package agent

import (
	"flag"
	"fmt"
	"os"

	"github.com/caarlos0/env"
)

type Configuration struct {
	ServerAddr     string `env:"ADDRESS"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
}

func LoadConfig() (*Configuration, error) {

	agentCfg := &Configuration{}

	flag.StringVar(&agentCfg.ServerAddr, "a", "localhost:8080", "serverAddress")
	flag.IntVar(&agentCfg.PollInterval, "p", 2, "poolInterval in seconds")
	flag.IntVar(&agentCfg.ReportInterval, "r", 10, "reportInterval in seconds")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	err := env.Parse(agentCfg)
	if err != nil {
		return nil, err
	}
	return agentCfg, nil
}
