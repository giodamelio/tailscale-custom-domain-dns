package main

import (
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/adrg/xdg"
	"github.com/omeid/uconfig"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"tailscale.com/tsnet"

	"github.com/giodamelio/tailscale-custom-domain-dns/tsapi"
)

type DeviceMap map[string]tsapi.Device

type writeDevicesOp struct {
	deviceMap DeviceMap
	response  chan bool
}

type readDevicesOp struct {
	response chan DeviceMap
}

type DNSConfig struct {
	Port int `default:"5353"`
}

type FetcherConfig struct {
	Interval string `default:"1h"`
}

type Config struct {
	Domain      string `default:""`
	TailnetName string `default:""`
	LogLevel    string `default:"info"`
	DNSServer   DNSConfig
	Fetcher     FetcherConfig
}

func main() {
	// Setup logging
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})

	// Find the config files
	configPath, err := xdg.SearchConfigFile("tailscale-custom-domain-dns.toml")
	if err != nil {
		log.Fatal().Err(err).Msg("could not find config file")
	}

	// Load/parse the config file
	config := &Config{}
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

	// Startup tsnet
	tsServer := new(tsnet.Server)
	// TODO: allow this to be configured
	tsServer.Hostname = "tailscale-custom-domain-dns"
	defer tsServer.Close()

	// Setup the tailscale api client
	ts := tsapi.NewTSClient(config.TailnetName)

	// Channels for reads and writes
	reads := make(chan readDevicesOp)
	writes := make(chan writeDevicesOp)

	// Fetch the Devices on a regular basis
	duration, err := time.ParseDuration(config.Fetcher.Interval)
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot parse config item fetchinterval")
	}
	go setupDeviceFetcher(writes, ts, duration)

	// Setup the DNS server
	go setupDnsServer(config, tsServer, reads, config.Domain)

	// Keep track of all the devices
	var state = make(DeviceMap)
	for {
		select {
		case read := <-reads:
			log.Trace().Int("count", len(state)).Msg("Devices read")
			read.response <- state
		case write := <-writes:
			log.Trace().Int("count", len(write.deviceMap)).Msg("Devices written")
			state = write.deviceMap
			write.response <- true
		}
	}
}
