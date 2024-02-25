package config

import (
	"flag"
	"fmt"
	"os"

	"go.uber.org/zap"
)

type ServerConfiguration struct {
	Url string
	Log *zap.SugaredLogger
}

func (c *ServerConfiguration) String() string {
	return c.Url
}

func (c *ServerConfiguration) Set(s string) error {
	c.Url = s
	return nil
}

var _ flag.Value = (*ServerConfiguration)(nil)

func LoadServerConfig() *ServerConfiguration {
	logger, err := zap.NewDevelopment()
	if err != nil {
		// вызываем панику, если ошибка
		panic("cannot initialize zap")
	}
	defer logger.Sync()

	srvConf := &ServerConfiguration{
		Log: logger.Sugar(),
	}
	srvConf.Set(":8080") // Значение по-умолчанию

	flag.Var(srvConf, "a", "serverAddress")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	addr, isExists := os.LookupEnv("ADDRESS")
	if isExists {
		srvConf.Set(addr)
	}

	return srvConf
}
