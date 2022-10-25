package config

import (
	"fmt"
	"os"
	"strconv"
)

type Configuration struct {
	HttpPort int
	Address  string
	GrpcPort int
	ApiKey   string
	Database DBConfig
}

type DBConfig struct {
	DBMS      string
	ConString string
}

func ReadConfig(prefix string) (Configuration, error) {
	var cfg Configuration

	p := os.Getenv(prefix + "HTTPPORT")
	// Set HTTP Port to default value of 80 if no value is provided
	if p == "" {
		p = "80"
	}
	port, err := strconv.ParseInt(p, 10, 64)
	if err != nil {
		return Configuration{}, fmt.Errorf("unable to parse value for %sHTTPPORT (%s)", prefix, os.Getenv(prefix+"HTTPPORT"))
	}
	cfg.HttpPort = int(port)

	p = os.Getenv(prefix + "GRPCPORT")
	// Set gRPC Port to default value of 81 if no value is provided
	if p == "" {
		p = "80"
	}
	port, err = strconv.ParseInt(p, 10, 64)
	if err != nil {
		return Configuration{}, fmt.Errorf("unable to parse value for %sGRPCPORT (%s)", prefix, os.Getenv(prefix+"GRPCPORT"))
	}
	cfg.GrpcPort = int(port)

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
