package config

import (
	"github.com/caarlos0/env/v8"
)

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:":8080"`
	PostgresConn    string `env:"POSTGRES_CONN"`
	PostgresJDBCUrl string `env:"POSTGRES_JDBC_URL"`
	PostgresUser    string `env:"POSTGRES_USERNAME"`
	PostgresPass    string `env:"POSTGRES_PASSWORD"`
	PostgresHost    string `env:"POSTGRES_HOST"`
	PostgresPort    string `env:"POSTGRES_PORT" envDefault:"5432"`
	PostgresDB      string `env:"POSTGRES_DATABASE"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
