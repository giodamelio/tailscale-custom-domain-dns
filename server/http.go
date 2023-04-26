package server

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"tailscale.com/tsnet"
)

func SetupHttpServer(tsServer *tsnet.Server, readDevices chan ReadDevicesOp) {
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

	// Start an http server on the port
	err = http.Serve(listener, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello World")
	}))
	if err != nil {
		log.Fatal().Err(err).Msg("http server error")
	}
}
