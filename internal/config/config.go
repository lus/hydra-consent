package config

import (
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"strings"
)

type Config struct {
	Environment   string `default:"dev"`
	LogLevel      string `default:"info" split_words:"true"`
	ListenAddress string `default:":8080" split_words:"true"`
	HydraAdminAPI string `required:"true" split_words:"true"`
}

func (cfg *Config) IsDevEnv() bool {
	return strings.ToLower(cfg.Environment) == "dev"
}

func Load() (*Config, error) {
	_ = godotenv.Overload()
	cfg := new(Config)
	if err := envconfig.Process("", cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
