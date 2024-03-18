package server

import (
	"flag"
	"fmt"
	"os"
)

type Configuration struct {
	url string
}

func (c *Configuration) String() string {
	return c.url
}

func (c *Configuration) Set(s string) error {
	c.url = s
	return nil
}

var _ flag.Value = (*Configuration)(nil)

func LoadConfig() (*Configuration, error) {
	srvConf := &Configuration{}
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

	return srvConf, nil
}
