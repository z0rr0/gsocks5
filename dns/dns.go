package dns

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/armon/go-socks5"
)

// nameResolver is a nameResolver that uses a custom DNS server.
type nameResolver struct {
	r *net.Resolver
}

// Resolve resolves the given host name to an address.
func (nr *nameResolver) Resolve(ctx context.Context, name string) (context.Context, net.IP, error) {
	ips, err := nr.r.LookupIP(ctx, "ip", name)
	if err != nil {
		return ctx, nil, err
	}
	if len(ips) == 0 {
		return ctx, nil, nil
	}
	return ctx, ips[0], nil
}

// New returns a new name nameResolver.
func New(dnsHost string, timeout time.Duration, loggerInfo, loggerDebug *log.Logger) socks5.NameResolver {
	const port = "53"
	if dnsHost == "" {
		loggerInfo.Printf("using default DNS nameResolver")
		return socks5.DNSResolver{}
	}

	address := net.JoinHostPort(dnsHost, port)
	loggerInfo.Printf("using DNS server %s", address)

	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, _ string) (net.Conn, error) {
			loggerDebug.Printf("dialing DNS server %s, network %s", address, network)
			d := net.Dialer{Timeout: time.Second * 5}
			return d.DialContext(ctx, network, address)
		},
	}
	return &nameResolver{r: resolver}
}
