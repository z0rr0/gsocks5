package dns

import (
	"context"
	"io"
	"log"
	"testing"
	"time"

	"github.com/armon/go-socks5"
)

var (
	logger  = log.New(io.Discard, "test", log.LstdFlags)
	timeout = 5 * time.Second
)

func TestNew(t *testing.T) {
	cases := []struct {
		name    string
		dnsHost string
		host    string
		err     bool
	}{
		{name: "default"},
		{name: "google", dnsHost: "8.8.8.8"},
		{name: "badDefault", host: "bad.bad.github.bad", err: true},
		{name: "badCustom", host: "bad.bad.github.bad", dnsHost: "8.8.8.8", err: true},
	}
	for _, c := range cases {
		t.Run(c.name, func(tt *testing.T) {
			nr := New(c.dnsHost, timeout, logger, logger)
			if len(c.dnsHost) == 0 {
				// default nameResolver
				if _, ok := nr.(socks5.DNSResolver); !ok {
					tt.Errorf("expected default nameResolver, gotten: %T", nr)
				}
			} else {
				// custom nameResolver
				if _, ok := nr.(*nameResolver); !ok {
					tt.Fatal("expected custom nameResolver")
				}
			}
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			_, ip, err := nr.Resolve(ctx, c.host)
			if err != nil {
				if !c.err {
					tt.Errorf("unexpected error: %v", err)
				}
			} else {
				if c.err {
					tt.Errorf("expected error, got: %v", ip)
				} else {
					if ipString := ip.String(); len(ipString) == 0 {
						tt.Errorf("expected ip")
					}
				}
			}
		})
	}
}
