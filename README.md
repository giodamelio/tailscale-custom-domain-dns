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
