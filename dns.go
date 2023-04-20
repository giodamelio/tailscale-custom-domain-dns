package main

import (
	"fmt"
	"net/netip"
	"strconv"
	"strings"

	"github.com/giodamelio/tailscale-custom-domain-dns/tsapi"
	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
)

func buildRR(name string, question dns.Question, address netip.Addr) dns.RR {
	log.
		Trace().
		Str("name", name).
		Str("question_type", dns.TypeToString[question.Qtype]).
		IPAddr("ip", address.AsSlice()).
		Bool("is_ipv4", address.Is4()).
		Bool("is_ipv6", address.Is6()).
		Msg("Building resource record")

	var rr dns.RR
	var err error

	if (question.Qtype == dns.TypeA || question.Qtype == dns.TypeANY) && address.Is4() {
		rr, err = dns.NewRR(fmt.Sprintf("%s A %s", name, address.String()))
	}

	if (question.Qtype == dns.TypeAAAA || question.Qtype == dns.TypeANY) && address.Is6() {
		rr, err = dns.NewRR(fmt.Sprintf("%s AAAA %s", name, address.String()))
	}

	if err != nil {
		log.
			Debug().
			Err(err).
			IPAddr("ip", address.AsSlice()).
			Msg("could not create resource record")
		return nil
	}

	return rr
}

func constructResponses(name string, device tsapi.Device, question dns.Question) []dns.RR {
	var result []dns.RR

	for _, rawAddress := range device.Addresses {
		address, err := netip.ParseAddr(rawAddress)
		if err != nil {
			log.
				Error().
				Str("raw_address", rawAddress).
				Err(err).
				Msg("invalid IP from Tailscale")
			continue
		}

		if rr := buildRR(name, question, address); rr != nil {
			result = append(result, rr)
		}
	}

	return result
}

type DnsHandler func(dns.ResponseWriter, *dns.Msg)

func makeHandler(readDevices chan readDevicesOp, host string) DnsHandler {
	return func(w dns.ResponseWriter, r *dns.Msg) {
		m := new(dns.Msg)
		m.SetReply(r)

		// Only allow DNS Queries
		if r.Opcode != dns.OpcodeQuery {
			log.Debug().Str("opcode", dns.OpcodeToString[r.Opcode]).Msg("Invalid Opcode")
			w.WriteMsg(m)
			return
		}

		// Allow exactly one question. Most DNS servers can only handle one
		// even though the original spec allows more
		if len(m.Question) != 1 {
			log.
				Debug().
				Int("question_count", len(m.Question)).
				Msg("Invalid number of questions")
			w.WriteMsg(m)
			return
		}

		// Get the DeviceMap
		read := readDevicesOp{
			response: make(chan DeviceMap),
		}
		readDevices <- read
		deviceMap := <-read.response

		// Respond to the question
		question := m.Question[0]
		log.
			Debug().
			Str("hostname", question.Name).
			Str("type", dns.TypeToString[question.Qtype]).
			Msgf("%s query for %s", dns.TypeToString[question.Qtype], question.Name)

		// Get just the subdomain name from the request
		name := strings.ReplaceAll(question.Name, "."+host, "")

		// Respond if a device with the hostname exists
		if device, ok := deviceMap[name]; ok {
			rrs := constructResponses(name, device, question)
			log.Trace().Any("records", rrs).Msg("Sending records to client")
			m.Answer = rrs
		}

		w.WriteMsg(m)
	}
}

// Run the DNS server
func setupDnsServer(config *Config, readDevices chan readDevicesOp, host string) {
	// Create the server
	server := &dns.Server{Addr: ":" + strconv.Itoa(config.DNSServer.Port), Net: "udp"}

	// Listen at our domain
	dns.HandleFunc(host, makeHandler(readDevices, host))

	// Start the server
	log.
		Info().
		Int("port", config.DNSServer.Port).
		Msgf("Starting DNS server on port %d", config.DNSServer.Port)
	err := server.ListenAndServe()
	defer server.Shutdown()
	if err != nil {
		log.Fatal().Err(err).Send()
	}
}
