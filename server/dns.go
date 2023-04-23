package server

import (
	"fmt"
	"net/netip"
	"strconv"
	"strings"

	"github.com/giodamelio/tailscale-custom-domain-dns/tsapi"
	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"tailscale.com/tsnet"
	"tailscale.com/types/nettype"
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

func makeHandler(readDevices chan ReadDevicesOp, host string) DnsHandler {
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
		read := ReadDevicesOp{
			Response: make(chan DeviceMap),
		}
		readDevices <- read
		deviceMap := <-read.Response

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

func serveDNSConn(conn nettype.ConnPacketConn, readDevices chan ReadDevicesOp, host string) {
	server := &dns.Server{
		PacketConn: conn,
		Net:        "udp",
	}
	defer server.Shutdown()

	// TODO we should probably make our own ServeMux here instead of using the default one
	dns.HandleFunc(host, makeHandler(readDevices, host))

	err := server.ActivateAndServe()
	if err != nil {
		log.Fatal().Err(err).Msg("dns server error")
	}
}

// Run the DNS server
func SetupDnsServer(tsServer *tsnet.Server, readDevices chan ReadDevicesOp, host string) {
	// Create the Tailscale listener
	listener, err := tsServer.Listen("udp", ":"+strconv.Itoa(viper.GetInt("dns-server.port")))
	if err != nil {
		log.Fatal().Err(err).Msg("could not listen on tailnet")
	}

	// Handle connections
	log.
		Info().
		Str("host", tsServer.Hostname).
		Int("port", viper.GetInt("dns-server.port")).
		Msgf(
			`DNS started. Host: %s, Port: %d, Tailnet: %s`,
			viper.GetString("tailscale.hostname"),
			viper.GetInt("dns-server.port"),
			viper.GetString("tailscale.tailnet"),
		)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Error().Err(err).Msg("could not accept connection")
		}
		go serveDNSConn(conn.(nettype.ConnPacketConn), readDevices, host)
	}
}
