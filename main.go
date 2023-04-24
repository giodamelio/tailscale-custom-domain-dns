package main

import (
	_ "embed"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/giodamelio/tailscale-custom-domain-dns/server"
)

func printHelp() {
	log.Info().Msg("Small DNS server that serves records of your Tailnet as a subdomain of any domain")
	log.Info().Msg("Instructions:")
	log.Info().Msg("  Generate config with `tailscale-custom-domain-dns --generate-config`")
	log.Info().Msg("  Fill in config")
	log.Info().Msgf("  place %s somewhere in an XDG config directory", configName)
	log.Info().Msg("  Keep this server running")
}

//go:embed examples/tailscale-custom-domain-dns.toml
var exampleConfig []byte

func generateConfig() {
	configFile, err := os.OpenFile(configName, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		if os.IsExist(err) {
			log.Fatal().Err(err).Msg("Example configuration already exists")
		}
		log.Fatal().Err(err).Msg("could not open config file")
	}
	defer configFile.Close()

	// Write the example config
	_, err = configFile.Write(exampleConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("could not write to config file")
	}

	log.Info().Msgf(`Wrote example config to "./%s"`, configName)
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

	if argsContain([]string{"-generate-config", "--generate-config"}) {
		generateConfig()
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
