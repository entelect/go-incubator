package config

import (
	"fmt"
	"os"
	"strconv"
)

type Configuration struct {
	HttpPort int
	ApiKey   string
}

func ReadConfig(prefix string) (Configuration, error) {
	var cfg Configuration

	port, err := strconv.ParseInt(os.Getenv(prefix+"HTTPPORT"), 10, 64)
	if err != nil {
		return cfg, fmt.Errorf("unable to parse value for %sHTTPPORT (%s)", prefix, os.Getenv(prefix+"HTTPPORT"))
	}

	cfg.HttpPort = int(port)
	cfg.ApiKey = os.Getenv(prefix + "APIKEY")

	return cfg, nil
}
