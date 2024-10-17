ARG GOLANG_VERSION="1.23.2"

FROM golang:${GOLANG_VERSION}-alpine AS builder
ARG LDFLAGS
WORKDIR /go/src/github.com/z0rr0/gsocks5
COPY . .
RUN echo "LDFLAGS = $LDFLAGS"
RUN GOOS=linux go build -ldflags "$LDFLAGS" -o ./gsocks5

FROM alpine:3.20
LABEL org.opencontainers.image.authors="me@axv.email" \
        org.opencontainers.image.url="https://hub.docker.com/r/z0rr0/gsocks5" \
        org.opencontainers.image.documentation="https://github.com/z0rr0/gsocks5" \
        org.opencontainers.image.source="https://github.com/z0rr0/gsocks5" \
        org.opencontainers.image.licenses="MIT" \
        org.opencontainers.image.title="GSocks5" \
        org.opencontainers.image.description="Simple SOCKS5 proxy server."

COPY --from=builder /go/src/github.com/z0rr0/gsocks5/gsocks5 /bin/
RUN chmod 0755 /bin/gsocks5

VOLUME ["/data/"]
EXPOSE 1080
ENTRYPOINT ["/bin/gsocks5"]
