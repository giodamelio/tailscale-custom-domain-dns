package main

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/giodamelio/tailscale-custom-domain-dns/server"
)

func main() {
	// Setup logging
	log.Logger = log.Output(createFormatter())

	// Setup the config
	loadConfig()

	// Set the log level
	level, err := zerolog.ParseLevel(viper.GetString("log-level"))
	if err != nil {
		log.Fatal().Err(err).Msg("invalid log level")
	}
	zerolog.SetGlobalLevel(level)
	// If we are at trace level, show more info then our custom logger
	if level == zerolog.TraceLevel {
		log.Logger = log.Output(createTraceFormatter())
	}

	// This has to be after the log level is set
	log.Trace().Any("config", viper.AllSettings()).Msg("Loaded Config")

	server.Start()
}
