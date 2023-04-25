package server

import (
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"tailscale.com/tsnet"

	"github.com/giodamelio/tailscale-custom-domain-dns/tsapi"
)

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
	tsServer.Hostname = viper.GetString("tailscale.hostname")
	tsServer.AuthKey = viper.GetString("tailscale.auth-key")
	if viper.IsSet("tailscale.state-directory") {
		tsServer.Dir = viper.GetString("tailscale.state-directory")
	}
	tsServer.Logf = func(format string, args ...any) {
		log.
			Trace().
			Str("library", "tsnet").
			Msgf(format, args...)
	}
	defer tsServer.Close()

	// Setup the tailscale api client
	ts := tsapi.NewTSClient(viper.GetString("tailscale.organization-name"))
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
