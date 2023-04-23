package main

import (
	"os"

	"github.com/adrg/xdg"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func loadConfig() {
	// Find the config file
	configPath, err := xdg.SearchConfigFile("tailscale-custom-domain-dns.toml")
	if err != nil {
		log.Fatal().Err(err).Msg("could not find config file")
	}

	// Setup viper
	viper.SetConfigName("tailscale-custom-domain-dns")
	viper.SetConfigType("toml")

	// Set some default values
	viper.SetDefault("log-level", "info")
	viper.SetDefault("fetcher.interval", "1h")
	viper.SetDefault("dns-server.port", 53)

	// Read the config
	configFile, err := os.Open(configPath)
	if err != nil {
		log.Fatal().Err(err).Str("path", configPath).Msg("could not open conf file")
	}
	err = viper.ReadConfig(configFile)
	if err != nil {
		log.Fatal().Err(err).Str("path", configPath).Msg("could not parse conf file")
	}
}
