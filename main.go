package main

import (
	"os"

	"github.com/BurntSushi/toml"
	"github.com/adrg/xdg"
	"github.com/omeid/uconfig"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/giodamelio/tailscale-custom-domain-dns/server"
)

func main() {
	// Setup logging
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})

	// Find the config files
	configPath, err := xdg.SearchConfigFile("tailscale-custom-domain-dns.toml")
	if err != nil {
		log.Fatal().Err(err).Msg("could not find config file")
	}

	// Load/parse the config file
	config := &server.Config{}
	c, err := uconfig.Classic(&config, uconfig.Files{
		{configPath, toml.Unmarshal},
	})
	if err != nil {
		log.Fatal().Err(err).Msg("could not parse config")
	}

	// Print usage if necessary
	for _, arg := range os.Args {
		if arg == "-h" || arg == "-help" || arg == "--help" || arg == "help" {
			c.Usage()
			os.Exit(0)
		}
	}

	// Set the log level
	level, err := zerolog.ParseLevel(config.LogLevel)
	if err != nil {
		log.Fatal().Err(err).Msg("invalid log level")
	}
	zerolog.SetGlobalLevel(level)

	// This has to be after the log level is set
	log.Trace().Any("config", config).Msg("Loaded Config")

	server.Start(config)
}
