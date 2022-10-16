package config

import (
	"fmt"
	"os"
	"strconv"
)

type Configuration struct {
	HttpPort int
	Address  string
	ApiKey   string
	Database DBConfig
}

type DBConfig struct {
	DBMS      string
	ConString string
}

func ReadConfig(prefix string) (Configuration, error) {
	var cfg Configuration

	port, err := strconv.ParseInt(os.Getenv(prefix+"HTTPPORT"), 10, 64)
	if err != nil {
		return cfg, fmt.Errorf("unable to parse value for %sHTTPPORT (%s)", prefix, os.Getenv(prefix+"HTTPPORT"))
	}

	cfg.HttpPort = int(port)

	cfg.Address = os.Getenv(prefix + "ADDRESS")
	if cfg.Address == "" {
		cfg.Address = "127.0.0.1"
	}

	cfg.ApiKey = os.Getenv(prefix + "APIKEY")

	cfg.Database = DBConfig{
		DBMS:      os.Getenv(prefix + "DBMS"),
		ConString: os.Getenv(prefix + "CONSTRING"),
	}

	return cfg, nil
}
