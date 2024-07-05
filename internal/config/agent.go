package config

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/caarlos0/env"
)

type Duration time.Duration

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case string:
		tmp, err := time.ParseDuration(value)
		if err != nil {
			return err
		}
		*d = Duration(tmp)
		return nil
	default:
		return errors.New("invalid duration")
	}
}

type agentFileConf struct {
	Address        string   `json:"address"`
	ReportInterval Duration `json:"report_interval"`
	PoolInterval   Duration `json:"poll_interval"`
	CryptoKey      string   `json:"crypto_key"`
	UseGRPC        bool     `json:"use_grpc"`
}

type AgentConfiguration struct {
	ServerAddr     string `env:"ADDRESS"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	Key            string `env:"KEY"`
	BatchSize      int    `env:"BATCH_SIZE"`
	RateLimit      int    `env:"RATE_LIMIT"`
	CryptoKey      string `env:"CRYPTO_KEY"`
	UseGRPC        bool   `env:"USE_GRPC"`
}

func LoadAgentConfigFromFile(fileName string) *agentFileConf {

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
	AgentDefaultServerAddr     = "localhost:8080"
	AgentDefautlPollInterval   = 2
	AgentDefaultReportInterval = 10
	AgentDefaultCryptoKey      = ""
	AgentDefaultUseGRPCValue   = false
)

func UpdateAgentDefaultValues(aFileConf *agentFileConf, aConf *AgentConfiguration) {
	if aConf.ServerAddr == AgentDefaultServerAddr && aFileConf.Address != "" {
		aConf.ServerAddr = aFileConf.Address
	}

	if aConf.PollInterval == AgentDefautlPollInterval && aFileConf.PoolInterval != 0 {
		seconds := aFileConf.PoolInterval
		dur := time.Duration(seconds)
		aConf.PollInterval = int(dur.Seconds())
	}

	if aConf.ReportInterval == AgentDefaultReportInterval && aFileConf.ReportInterval != 0 {
		seconds := aFileConf.ReportInterval
		dur := time.Duration(seconds)
		aConf.ReportInterval = int(dur.Seconds())
	}

	if aConf.CryptoKey == AgentDefaultCryptoKey && aFileConf.CryptoKey != "" {
		aConf.CryptoKey = aFileConf.CryptoKey
	}

	if !aConf.UseGRPC && aFileConf.UseGRPC {
		aConf.UseGRPC = aFileConf.UseGRPC
	}
}

func LoadAgentConfig() (*AgentConfiguration, error) {

	agentCfg := &AgentConfiguration{}

	flag.StringVar(&agentCfg.ServerAddr, "a", AgentDefaultServerAddr, "serverAddress")
	flag.IntVar(&agentCfg.PollInterval, "p", AgentDefautlPollInterval, "poolInterval in seconds")
	flag.IntVar(&agentCfg.ReportInterval, "r", AgentDefaultReportInterval, "reportInterval in seconds")
	flag.IntVar(&agentCfg.BatchSize, "b", 5, "metric count of metrics per update request")
	flag.IntVar(&agentCfg.RateLimit, "l", 1, "max update simultaneous request count")
	flag.StringVar(&agentCfg.CryptoKey, "crypto-key", AgentDefaultCryptoKey, "rsa public key file name")
	flag.BoolVar(&agentCfg.UseGRPC, "grpc", false, "use grpc")

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
		aFileConf := LoadAgentConfigFromFile(configFileName)
		UpdateAgentDefaultValues(aFileConf, agentCfg)
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
