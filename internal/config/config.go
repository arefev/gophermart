package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env"
)

const (
	address       string = "localhost:8080"
	logLevel      string = "info"
	databaseDSN   string = ""
	tokenSecret   string = ""
	tokenDuration int    = 60
)

type Config struct {
	TokenSecret   string `env:"TOKEN_SECRET"`
	Address       string `env:"ADDRESS"`
	LogLevel      string `env:"LOG_LEVEL"`
	DatabaseDSN   string `env:"DATABASE_DSN"`
	TokenDuration int    `env:"TOKEN_DURATION"`
}

func NewConfig(params []string) (Config, error) {
	cnf := Config{}

	if err := cnf.initFlags(params); err != nil {
		return Config{}, err
	}

	if err := cnf.initEnvs(); err != nil {
		return Config{}, err
	}

	return cnf, nil
}

func (cnf *Config) initFlags(params []string) error {
	f := flag.NewFlagSet("main", flag.ExitOnError)
	f.StringVar(&cnf.Address, "a", address, "address and port to run server")
	f.StringVar(&cnf.LogLevel, "l", logLevel, "log level")
	f.StringVar(&cnf.DatabaseDSN, "d", databaseDSN, "db connection string")
	f.StringVar(&cnf.TokenSecret, "s", tokenSecret, "token secret")
	f.IntVar(&cnf.TokenDuration, "t", tokenDuration, "token lifetime duration in minutes")
	if err := f.Parse(params); err != nil {
		return fmt.Errorf("InitFlags: parse flags fail: %w", err)
	}

	return nil
}

func (cnf *Config) initEnvs() error {
	if err := env.Parse(cnf); err != nil {
		return fmt.Errorf("InitEnvs: parse envs fail: %w", err)
	}

	return nil
}
