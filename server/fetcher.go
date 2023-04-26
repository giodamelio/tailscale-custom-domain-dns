package server

import (
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/giodamelio/tailscale-custom-domain-dns/tsapi"
)

func fetchDevices(tsapi *tsapi.TSApi) ([]tsapi.Device, error) {
	devices, err := tsapi.Devices()
	if err != nil {
		return nil, err
	}
	log.Debug().Msgf("Fetched %d devices from Tailscale", len(devices))
	return devices, nil
}

// Get the device name by chopping off the last three subdomain parts
// example computer.tsnet00000.ts.net -> compute
func getDeviceName(rawDeviceName string) string {
	domainParts := strings.Split(rawDeviceName, ".")
	return strings.Join(domainParts[:len(domainParts)-3], ".")
}

// Fetch the devices on a regular basis
func SetupDeviceFetcher(
	writeDevices chan WriteDevicesOp,
	ts *tsapi.TSApi,
) {
	duration, err := time.ParseDuration(viper.GetString("fetcher.interval"))
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot parse config item fetchinterval")
	}

	log.
		Info().
		Dur("duration", duration).
		Msgf("Fetching tailnet devices every %s", duration)

	for {
		devices, err := fetchDevices(ts)
		if err != nil {
			log.Warn().Err(err).Msg("Cannot fetch devices")
		}

		// Build a DeviceMap
		var deviceMap = make(DeviceMap)
		for _, device := range devices {
			name := getDeviceName(device.Name)
			deviceMap[name] = device
		}

		// Write the device map to the central store
		write := WriteDevicesOp{
			DeviceMap: deviceMap,
			Response:  make(chan bool),
		}
		writeDevices <- write
		<-write.Response

		// Take a nap
		time.Sleep(duration)
	}
}
