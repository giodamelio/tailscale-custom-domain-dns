package main

import (
	"os"

	"github.com/adrg/xdg"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

var configName = "tailscale-custom-domain-dns.toml"

func loadConfig() {
	// Find the config file
	configPath, err := xdg.SearchConfigFile(configName)
	if err != nil {
		log.Error().Msg("Could not find config file")
		log.Error().Msgf(`No config file "%s" found in directories:`, configName)
		log.Error().Msgf("  %s", xdg.ConfigHome)
		for _, configDir := range xdg.ConfigDirs {
			log.Error().Msgf("  %s", configDir)
		}
		log.Error().Msg("You can generate config with `tailscale-custom-domain-dns --generate-config`")
		log.Fatal().Err(err).Msg("Exiting")
	}

	// Setup viper
	viper.SetConfigName("tailscale-custom-domain-dns")
	viper.SetConfigType("toml")

	// Set some default values
	viper.SetDefault("log-level", "info")
	viper.SetDefault("fetcher.interval", "1h")
	viper.SetDefault("dns-server.port", 53)
	viper.SetDefault("tailscale.hostname", "tailscale-custom-domain-dns")

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
