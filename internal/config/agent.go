package config

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/caarlos0/env"
)

type AgentConfiguration struct {
	ServerAddr     string `env:"ADDRESS"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	Key            string `env:"KEY"`
	BatchSize      int    `env:"BATCH_SIZE"`
	RateLimit      int    `env:"RATE_LIMIT"`
	CryptoKey      string `env:"CRYPTO_KEY"`
}

func LoadAgentConfig() (*AgentConfiguration, error) {

	agentCfg := &AgentConfiguration{}

	flag.StringVar(&agentCfg.ServerAddr, "a", "localhost:8080", "serverAddress")
	flag.IntVar(&agentCfg.PollInterval, "p", 2, "poolInterval in seconds")
	flag.IntVar(&agentCfg.ReportInterval, "r", 10, "reportInterval in seconds")
	flag.IntVar(&agentCfg.BatchSize, "b", 5, "metric count of metrics per update request")
	flag.IntVar(&agentCfg.RateLimit, "l", 1, "max update simultaneous request count")
	flag.StringVar(&agentCfg.CryptoKey, "crypto-key", "", "rsa public key file name")

	flag.StringVar(&agentCfg.Key, "k", "", "hashSha256 key")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	err := env.Parse(agentCfg)
	if err != nil {
		return nil, err
	}

	// Доп. обработка переданных данных
	if !strings.HasPrefix(agentCfg.ServerAddr, "http") {
		agentCfg.ServerAddr = "http://" + agentCfg.ServerAddr
	}
	agentCfg.ServerAddr = strings.TrimSuffix(agentCfg.ServerAddr, "/")

	if agentCfg.PollInterval < 0 {
		agentCfg.PollInterval = 2
	}

	if agentCfg.ReportInterval < 0 {
		agentCfg.ReportInterval = 10
	}

	if agentCfg.BatchSize < 0 {
		agentCfg.BatchSize = 5
	}

	if agentCfg.RateLimit < 0 {
		agentCfg.RateLimit = 1
	}

	return agentCfg, nil
}
