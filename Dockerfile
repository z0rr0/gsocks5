ARG GOLANG_VERSION="1.20.5"

FROM golang:${GOLANG_VERSION}-alpine as builder
RUN apk --no-cache add tzdata git
WORKDIR /go/src/github.com/z0rr0/gsocks5
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-X main.Tag=`git tag --sort=version:refname | tail -1`" -o ./gsocks5

FROM scratch
MAINTAINER Alexander Zaitsev "me@axv.email"
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/github.com/z0rr0/gsocks5/gsocks5 /bin/
ENTRYPOINT ["/bin/gsocks5"]
