package server

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/giodamelio/tailscale-custom-domain-dns/tsapi"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"tailscale.com/tsnet"
)

func SetupHttpServer(
	tsServer *tsnet.Server,
	ts *tsapi.TSApi,
	readDevices chan ReadDevicesOp,
	writeDevices chan WriteDevicesOp,
) {
	port := viper.GetInt("http-server.port")

	log.
		Info().
		Int("port", port).
		Msgf("Starting http server on port %d", port)

	// Create the Tailscale listener
	listener, err := tsServer.Listen("tcp", ":"+strconv.Itoa(viper.GetInt("http-server.port")))
	if err != nil {
		log.Fatal().Err(err).Msg("could not listen on tailnet")
	}
	defer listener.Close()

	// Setup our router
	router := http.NewServeMux()
	router.Handle(
		"/api/devices/refresh",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			updateDevices(writeDevices, ts)
			fmt.Fprintf(w, "Devices refreshed")
		}),
	)
	router.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))

	// Start an http server on the port
	err = http.Serve(listener, router)
	if err != nil {
		log.Fatal().Err(err).Msg("http server error")
	}
}
