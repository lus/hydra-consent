package main

import (
	"github.com/lus/hydra-consent/internal/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if cfg.IsDevEnv() {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
		log.Warn().Msg("The service was started in development mode. Please change the 'ENVIRONMENT' variable to 'prod' in production!")
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		logLevel, err := zerolog.ParseLevel(cfg.LogLevel)
		if err != nil {
			log.Warn().Msg("An invalid log level was provided. Using the 'info' fallback value.")
			logLevel = zerolog.InfoLevel
		}
		zerolog.SetGlobalLevel(logLevel)
	}
}
