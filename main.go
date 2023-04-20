package main

import (
	"os"
	"time"

	"github.com/adrg/xdg"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

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

func main() {
	// Setup logging
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})

	// Find the config files
	configPath, err := xdg.SearchConfigFile("tailscale-custom-domain-dns.toml")
	if err != nil {
		log.Fatal().Err(err).Msg("could not find config file")
	}
	log.Warn().Str("path", configPath).Msg("Found config file")

	// Setup the tailscale api client
	ts := tsapi.NewTSClient("giodamelio.github")

	// Channels for reads and writes
	reads := make(chan readDevicesOp)
	writes := make(chan writeDevicesOp)

	// Fetch the Devices on a regular basis
	go setupDeviceFetcher(writes, ts, time.Minute)

	// Setup the DNS server
	go setupDnsServer(reads, "home.gio.ninja.")

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
