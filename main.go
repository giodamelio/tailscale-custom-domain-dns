package main

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/giodamelio/tailscale-custom-domain-dns/server"
)

func printHelp() {
	log.Info().Msg("Small DNS server that serves records of your Tailnet as a subdomain of any domain")
	log.Info().Msg("Instructions:")
	log.Info().Msg("  Generate config with TODO")
	log.Info().Msg("  Fill in config")
	log.Info().Msgf("  place %s somewhere in an XDG config directory", configName)
	log.Info().Msg("  Keep this server running")
}

func argsContain(flags []string) bool {
	for _, arg := range os.Args {
		for _, flag := range flags {
			if arg == flag {
				return true
			}
		}
	}
	return false
}

func main() {
	// Setup logging
	log.Logger = log.Output(createFormatter())

	// Print help if necessary
	if argsContain([]string{"-h", "-help", "--help", "help"}) {
		printHelp()
		os.Exit(0)
	}

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
