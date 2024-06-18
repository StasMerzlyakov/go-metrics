package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/caarlos0/env"
)

type serverFileConf struct {
	Address       string   `json:"address"`
	Restore       bool     `json:"restore"`
	StoreInterval Duration `json:"store_interval"`
	StoreFile     string   `json:"store_file"`
	DatabaseDSN   string   `json:"database_dsn"`
	CryptoKey     string   `json:"crypto_key"`
}

const (
	ServerDefaultAddr          = "localhost:8080"
	ServerDefaultRestore       = true
	ServerDefaultStoreInterval = 300
	ServerDefaultStoreFile     = "/tmp/metrics-db.json"
	ServerDefaultDatabaseDSN   = ""
	ServerDefaultCryptoKey     = ""
)

func LoadServerConfigFromFile(fileName string) *serverFileConf {

	f, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	var serverFileConf = &serverFileConf{
		Restore: ServerDefaultRestore, // значение по-умолчанию true
	}

	if err := json.NewDecoder(f).Decode(serverFileConf); err != nil {
		panic(err)
	}

	return serverFileConf
}

func UpdateServerDefaultValues(sFileConf *serverFileConf, sConf *ServerConfiguration, isRestoreSet bool) {

	if sConf.URL == ServerDefaultAddr && sFileConf.Address != "" {
		sConf.URL = sFileConf.Address
	}

	if !isRestoreSet {
		sConf.Restore = sFileConf.Restore
	}

	if sConf.StoreInterval == ServerDefaultStoreInterval && sFileConf.StoreInterval != 0 {
		seconds := sFileConf.StoreInterval
		dur := time.Duration(seconds)
		sConf.StoreInterval = uint(dur.Seconds())
	}

	if sConf.FileStoragePath == ServerDefaultStoreFile && sFileConf.StoreFile != "" {
		sConf.FileStoragePath = sFileConf.StoreFile
	}

	if sConf.DatabaseDSN == ServerDefaultDatabaseDSN && sFileConf.DatabaseDSN != "" {
		sConf.DatabaseDSN = sFileConf.DatabaseDSN
	}

	if sConf.CryptoKey == ServerDefaultCryptoKey && sFileConf.CryptoKey != "" {
		sConf.CryptoKey = sFileConf.CryptoKey
	}
}

type ServerConfiguration struct {
	URL             string `env:"ADDRESS"`
	Restore         bool   `env:"RESTORE"`
	StoreInterval   uint   `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
	Key             string `env:"KEY"`
	CryptoKey       string `env:"CRYPTO_KEY"`
}

type RestoreConfiguration struct {
	IsSet   bool // флаг, указывающий что переменная устанавливалась
	Restore bool
}

func (r *RestoreConfiguration) String() string {
	return fmt.Sprintf("%v", r.Restore)
}

func (r *RestoreConfiguration) Set(s string) (err error) {
	r.IsSet = true
	r.Restore, err = strconv.ParseBool(s)
	return
}

var _ flag.Value = (*RestoreConfiguration)(nil)

func LoadServerConfig() (*ServerConfiguration, error) {
	srvConf := &ServerConfiguration{}

	flag.StringVar(&srvConf.URL, "a", ServerDefaultAddr, "server address (format \":PORT\")")
	flag.UintVar(&srvConf.StoreInterval, "i", ServerDefaultStoreInterval, "Backup store interval in seconds")
	flag.StringVar(&srvConf.FileStoragePath, "f", ServerDefaultStoreFile, "Backup file path")
	flag.StringVar(&srvConf.DatabaseDSN, "d", ServerDefaultDatabaseDSN, "PostgreSQL URL like 'postgres://username:password@localhost:5432/database_name'")
	flag.StringVar(&srvConf.Key, "k", "", "hashSha256 key")
	flag.StringVar(&srvConf.CryptoKey, "crypto-key", ServerDefaultCryptoKey, "rsa public key file name")

	var configFileName string

	flag.StringVar(&configFileName, "c", "", "config file")      // config file short format
	flag.StringVar(&configFileName, "config", "", "config file") // config file long format

	// Шаманстрва из-за того, что Go хитро обрабатывает Bool-флаги(проверяет просто наличие флага в коммандной строке)
	restoreConf := &RestoreConfiguration{
		Restore: ServerDefaultRestore, // Значение по-умолчанию
	}
	flag.Var(restoreConf, "r", "is backup restore need")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if configFileName != "" {
		sFileConf := LoadServerConfigFromFile(configFileName)
		UpdateServerDefaultValues(sFileConf, srvConf, restoreConf.IsSet)
	}

	srvConf.Restore = restoreConf.Restore

	err := env.Parse(srvConf)
	if err != nil {
		return nil, err
	}

	return srvConf, nil
}
