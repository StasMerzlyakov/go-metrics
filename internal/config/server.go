package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/caarlos0/env"
)

type ServerConfiguration struct {
	URL             string `env:"ADDRESS"`
	StoreInterval   uint   `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
}

type RestoreConfiguration struct {
	Restore bool
}

func (r *RestoreConfiguration) String() string {
	return fmt.Sprintf("%v", r.Restore)
}

func (r *RestoreConfiguration) Set(s string) (err error) {
	r.Restore, err = strconv.ParseBool(s)
	return
}

var _ flag.Value = (*RestoreConfiguration)(nil)

func LoadServerConfig() (*ServerConfiguration, error) {
	srvConf := &ServerConfiguration{}

	flag.StringVar(&srvConf.URL, "a", ":8080", "server address (format \":PORT\")")
	flag.UintVar(&srvConf.StoreInterval, "i", 300, "Backup store interval in seconds")
	flag.StringVar(&srvConf.FileStoragePath, "f", "/tmp/metrics-db.json", "Backup file path")
	flag.StringVar(&srvConf.DatabaseDSN, "d", "", "PostgreSQL URL like 'postgres://username:password@localhost:5432/database_name'")

	// Шаманстрва из-за того, что Go хитро обрабатывает Bool-флаги(проверяет просто наличие флага в коммандной строке)
	restoreConf := &RestoreConfiguration{
		Restore: true, // Значение по-умолчанию
	}
	flag.Var(restoreConf, "r", "is backup restore need")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	srvConf.Restore = restoreConf.Restore

	err := env.Parse(srvConf)
	if err != nil {
		return nil, err
	}
	return srvConf, nil
}
