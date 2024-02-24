package config

import (
	"flag"
	"fmt"
	"os"
)

type MemMeterConfiguration struct {
	url string
}

func (c *MemMeterConfiguration) String() string {
	return c.url
}

func (c *MemMeterConfiguration) Set(s string) error {
	c.url = s
	return nil
}

var _ flag.Value = (*MemMeterConfiguration)(nil)

func LoadMemMeterConfig() (*MemMeterConfiguration, error) {
	srvConf := &MemMeterConfiguration{}
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
