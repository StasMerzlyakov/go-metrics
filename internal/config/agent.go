package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/caarlos0/env"
)

type agentFileConf struct {
	Address        *string        `json:"address"`
	ReportInterval *time.Duration `json:"report_interval"`
	PoolInterval   *time.Duration `json:"poll_interval"`
	CryptoKey      *string        `json:"crypto_key"`
}

type AgentConfiguration struct {
	ServerAddr     string `env:"ADDRESS"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	Key            string `env:"KEY"`
	BatchSize      int    `env:"BATCH_SIZE"`
	RateLimit      int    `env:"RATE_LIMIT"`
	CryptoKey      string `env:"CRYPTO_KEY"`
}

func LoadConfigFromFile(fileName string) *agentFileConf {

	f, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	var agentFileConf = &agentFileConf{}

	if err := json.NewDecoder(f).Decode(agentFileConf); err != nil {
		panic(err)
	}

	return agentFileConf
}

const (
	DefaultServerAddr     = "localhost:8080"
	DefautlPollInterval   = 2
	DefaultReportInterval = 10
	DefaultCryptoKey      = ""
)

func UpdateDefaultValues(aFileConf *agentFileConf, aConf *AgentConfiguration) {
	if aConf.ServerAddr == DefaultServerAddr && aFileConf.Address != nil {
		aConf.ServerAddr = *aFileConf.Address
	}

	if aConf.PollInterval == DefautlPollInterval && aFileConf.PoolInterval != nil {
		seconds := *aFileConf.PoolInterval
		aConf.PollInterval = int(seconds.Seconds())
	}

	if aConf.ReportInterval == DefaultReportInterval && aFileConf.ReportInterval != nil {
		seconds := *aFileConf.ReportInterval
		aConf.ReportInterval = int(seconds.Seconds())
	}

	if aConf.CryptoKey == DefaultCryptoKey && aFileConf.CryptoKey != nil {
		aConf.CryptoKey = *aFileConf.CryptoKey
	}
}

func LoadAgentConfig() (*AgentConfiguration, error) {

	agentCfg := &AgentConfiguration{}

	flag.StringVar(&agentCfg.ServerAddr, "a", DefaultServerAddr, "serverAddress")
	flag.IntVar(&agentCfg.PollInterval, "p", DefautlPollInterval, "poolInterval in seconds")
	flag.IntVar(&agentCfg.ReportInterval, "r", DefaultReportInterval, "reportInterval in seconds")
	flag.IntVar(&agentCfg.BatchSize, "b", 5, "metric count of metrics per update request")
	flag.IntVar(&agentCfg.RateLimit, "l", 1, "max update simultaneous request count")
	flag.StringVar(&agentCfg.CryptoKey, "crypto-key", DefaultCryptoKey, "rsa public key file name")

	var configFileName string

	flag.StringVar(&configFileName, "c", "", "config file")      // config file short format
	flag.StringVar(&configFileName, "config", "", "config file") // config file long format

	flag.StringVar(&agentCfg.Key, "k", "", "hashSha256 key")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if configFileName != "" {
		aFileConf := LoadConfigFromFile(configFileName)
		UpdateDefaultValues(aFileConf, agentCfg)
	}

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
