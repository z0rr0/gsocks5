package dns

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/armon/go-socks5"
)

// ErrHostIP is returned when the DNS host is invalid.
var ErrHostIP = errors.New("DNS host is not an IP address")

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
func New(dnsHost string, timeout time.Duration, loggerInfo, loggerDebug *log.Logger) (socks5.NameResolver, error) {
	const port = "53"

	if dnsHost == "" {
		loggerInfo.Printf("use default DNS name resolver")
		return socks5.DNSResolver{}, nil
	}

	ip := net.ParseIP(dnsHost)
	if ip == nil {
		return nil, errors.Join(ErrHostIP, fmt.Errorf("invalid DNS host: %s", dnsHost))
	}

	address := net.JoinHostPort(ip.String(), port)
	loggerInfo.Printf("using DNS server %q", address)

	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, _ string) (net.Conn, error) {
			var d = net.Dialer{Timeout: timeout}
			loggerDebug.Printf("dialing DNS server %s, network %s, timeout %v", address, network, timeout)

			return d.DialContext(ctx, network, address)
		},
	}

	return &nameResolver{r: resolver}, nil
}
