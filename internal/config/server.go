package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/caarlos0/env"
)

type ServerConfiguration struct {
	URL                    string `env:"ADDRESS"`
	BackupStoreIntervalSec uint   `env:"STORE_INTERVAL"`
	FileStoragePath        string `env:"FILE_STORAGE_PATH"`
	Restore                bool   `env:"RESTORE"`
}

func LoadServerConfig() (*ServerConfiguration, error) {
	srvConf := &ServerConfiguration{}

	flag.StringVar(&srvConf.URL, "a", ":8080", "server address (format \":PORT\")")
	flag.UintVar(&srvConf.BackupStoreIntervalSec, "i", 300, "Backup store interval in seconds")
	flag.StringVar(&srvConf.FileStoragePath, "f", "/tmp/metrics-db.json", "Backup file path")
	flag.BoolVar(&srvConf.Restore, "r", true, "Is restore need")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	err := env.Parse(srvConf)
	if err != nil {
		return nil, err
	}
	return srvConf, nil
}
