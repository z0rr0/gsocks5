ARG GOLANG_VERSION="1.20.5"

FROM golang:${GOLANG_VERSION}-alpine as builder
ARG LDFLAGS
RUN apk --no-cache add tzdata git
WORKDIR /go/src/github.com/z0rr0/gsocks5
COPY . .
RUN echo "LDFLAGS = $LDFLAGS"
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "$LDFLAGS" -o ./gsocks5

FROM scratch
LABEL org.opencontainers.image.authors="me@axv.email" \
        org.opencontainers.image.url="https://hub.docker.com/r/z0rr0/gsocks5" \
        org.opencontainers.image.documentation="https://github.com/z0rr0/gsocks5" \
        org.opencontainers.image.source="https://github.com/z0rr0/gsocks5" \
        org.opencontainers.image.licenses="MIT" \
        org.opencontainers.image.title="GSocks5" \
        org.opencontainers.image.description="Simple socks5 server"
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/github.com/z0rr0/gsocks5/gsocks5 /bin/
ENTRYPOINT ["/bin/gsocks5"]
