package main

import (
	"os"

	"github.com/miekg/dns"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/giodamelio/tailscale-custom-domain-dns/tsapi"
)

func parseQuery(m *dns.Msg) {
	for _, q := range m.Question {
		switch q.Qtype {
		case dns.TypeA:
			log.Debug().Str("hostname", q.Name).Msgf("Query for %s", q.Name)
			rr, err := dns.NewRR("test A 1.1.1.1")
			if err == nil {
				m.Answer = append(m.Answer, rr)
			}
		}
	}
}

func handleDnsRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)

	switch r.Opcode {
	case dns.OpcodeQuery:
		parseQuery(m)
	}

	w.WriteMsg(m)
}

func main() {
	// Setup logging
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})

	// Setup the tailscale api
	tsapi := tsapi.NewTSClient("giodamelio.github")
	devices, err := tsapi.Devices()
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	log.Debug().Any("devices", devices).Msgf("There are %d devices", len(devices))

	// Create the server
	// port := 5353
	// server := &dns.Server{Addr: ":" + strconv.Itoa(port), Net: "udp"}
	//
	// // Listen at our domain
	// dns.HandleFunc("home.gio.ninja.", handleDnsRequest)
	//
	// // Start the server
	// log.Info().Int("port", port).Msgf("Starting DNS server on port %d", port)
	// err := server.ListenAndServe()
	// defer server.Shutdown()
	// if err != nil {
	// 	log.Fatal().Err(err).Send()
	// }
}
