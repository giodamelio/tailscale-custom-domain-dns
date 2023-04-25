FROM alpine:latest

ARG RELEASE_TAG
ARG RELEASE_FILE=tailscale-custom-domain-dns_${RELEASE_TAG}_linux_amd64.tar.gz

WORKDIR /srv

ADD https://github.com/giodamelio/tailscale-custom-domain-dns/releases/download/v${RELEASE_TAG}/${RELEASE_FILE} .

RUN tar xzvf ${RELEASE_FILE} && rm ${RELEASE_FILE}

# Create empty config file so the cli doesn't complain
RUN mkdir /root/.config/ && touch /root/.config/tailscale-custom-domain-dns.toml

# Save the state directory in /data so it can be added to a volume easily
ENV TSDNS_TAILSCALE_STATEDIRECTORY=/data

ENTRYPOINT [ "/srv/tailscale-custom-domain-dns" ]
