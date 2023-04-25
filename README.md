# Tailscale Custom Domain DNS

[![Latest Release](https://flat.badgen.net/github/release/giodamelio/tailscale-custom-domain-dns)](https://github.com/giodamelio/tailscale-custom-domain-dns/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/giodamelio/tailscale-custom-domain-dns)](https://goreportcard.com/report/github.com/giodamelio/tailscale-custom-domain-dns)
![Dependabot Status](https://flat.badgen.net/github/dependabot/giodamelio/tailscale-custom-domain-dns)
![Open Issues](https://flat.badgen.net/github/open-issues/giodamelio/tailscale-custom-domain-dns)
[![License](https://flat.badgen.net/github/license/giodamelio/tailscale-custom-domain-dns)](https://github.com/giodamelio/tailscale-custom-domain-dns/blob/master/LICENSE)

A tiny DNS server that fetches your list of Tailscale machines and serves records for them on any domain you want.

# Why

I love using [Tailscale](https://tailscale.com/) for all my devices, but I am paranoid about configuring my services to use the `*.ts.net` domain given to me by Tailscale in case I ever need to migrate away from Tailscale.

This small DNS server reads the list of all your Tailscale devices and returns `A` and `AAAA` records as subdomains on an arbitrary domain you specify.

# Install

 - Download the [latest release](https://github.com/giodamelio/tailscale-custom-domain-dns/releases) from Github
 - Install via Docker from GHCR: `$ docker pull ghcr.io/giodamelio/tailscale-custom-domain-dns:0.1.0`
 - Install with Golang cli: `go install github.com/giodamelio/tailscale-custom-domain-dns`
 - Clone and build from repo: `git clone https://github.com/giodamelio/tailscale-custom-domain-dns.git`

# Configuration

For docs on all the config optons, see the [example config file](examples/tailscale-custom-domain-dns.toml)

## Environment variables

The config file can be overridden with environment variables. They all have the prefix `TSDNS`. Nested options are seperated by underscores and dashes are removed. For example:

```
[dns-server]
port = 2222

# becomes

$ export TSDNS_DNSSERVER_PORT=2222
```

# Possible Future Enhancements

 - Webhook endpoint allowing automatic refreshing of devices when a new device is added
 - LetsEncrypt DNS-01 Challenge integration
 - Config based static records/aliases
 - Simple web ui listing status
 - Status url for Prometheus or monitoring
