package server

import (
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"tailscale.com/tsnet"

	"github.com/giodamelio/tailscale-custom-domain-dns/tsapi"
)

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

type DeviceMap map[string]tsapi.Device

type WriteDevicesOp struct {
	DeviceMap DeviceMap
	Response  chan bool
}

type ReadDevicesOp struct {
	Response chan DeviceMap
}

func Start() {
	// Startup tsnet
	tsServer := new(tsnet.Server)
	// TODO: allow this to be configured
	tsServer.Hostname = "tailscale-custom-domain-dns"
	tsServer.Logf = func(format string, args ...any) {
		log.
			Trace().
			Str("library", "tsnet").
			Msgf(format, args...)
	}
	defer tsServer.Close()

	// Setup the tailscale api client
	ts := tsapi.NewTSClient(viper.GetString("tailnet-name"))
	// Channels for reads and writes
	reads := make(chan ReadDevicesOp)
	writes := make(chan WriteDevicesOp)

	// Fetch the Devices on a regular basis
	duration, err := time.ParseDuration(viper.GetString("fetcher.interval"))
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot parse config item fetchinterval")
	}
	go SetupDeviceFetcher(writes, ts, duration)

	// Setup the DNS server
	go SetupDnsServer(tsServer, reads, viper.GetString("domain"))

	// Keep track of all the devices
	var state = make(DeviceMap)
	for {
		select {
		case read := <-reads:
			log.Trace().Int("count", len(state)).Msg("Devices read")
			read.Response <- state
		case write := <-writes:
			log.Trace().Int("count", len(write.DeviceMap)).Msg("Devices written")
			state = write.DeviceMap
			write.Response <- true
		}
	}
}
