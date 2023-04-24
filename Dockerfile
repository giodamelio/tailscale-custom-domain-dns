FROM alpine:latest

ARG RELEASE_TAG
ARG RELEASE_FILE=tailscale-custom-domain-dns_${RELEASE_TAG}_linux_amd64.tar.gz

WORKDIR /srv

ADD https://github.com/giodamelio/tailscale-custom-domain-dns/releases/download/v0.1.0/${RELEASE_FILE} .

RUN tar xzvf ${RELEASE_FILE} && rm ${RELEASE_FILE}

ENTRYPOINT [ "/srv/tailscale-custom-domain-dns" ]
