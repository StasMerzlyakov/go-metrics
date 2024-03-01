package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/caarlos0/env"
)

type AgentConfiguration struct {
	ServerAddr        string `env:"ADDRESS"`
	PollIntervalSec   int    `env:"POLL_INTERVAL"`
	ReportIntervalSec int    `env:"REPORT_INTERVAL"`
}

func LoadAgentConfig() (*AgentConfiguration, error) {

	agentCfg := &AgentConfiguration{}

	flag.StringVar(&agentCfg.ServerAddr, "a", "localhost:8080", "serverAddress")
	flag.IntVar(&agentCfg.PollIntervalSec, "p", 2, "poolInterval in seconds")
	flag.IntVar(&agentCfg.ReportIntervalSec, "r", 10, "reportInterval in seconds")
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
