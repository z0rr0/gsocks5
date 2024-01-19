package dns

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/armon/go-socks5"
)

var (
	logger  = log.New(os.Stdout, "[test] ", log.LstdFlags|log.Lshortfile)
	timeout = 5 * time.Second
)

func TestNew(t *testing.T) {
	testCases := []struct {
		name    string
		dnsHost string
		host    string
		err     bool
	}{
		{name: "default"},
		{name: "google", dnsHost: "8.8.8.8", host: "github.com"},
		{name: "badDNS", dnsHost: "bad", err: true},
		{name: "badDefault", host: "bad.bad.github.bad", err: true},
		{name: "badCustom", host: "bad.bad.github.bad", dnsHost: "8.8.8.8", err: true},
	}
	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			nr, err := New(tc.dnsHost, timeout, logger, logger)
			if err != nil {
				if !tc.err {
					t.Errorf("unexpected error: %v", err)
				}
				return
			}

			if tc.dnsHost == "" {
				// default nameResolver
				if _, ok := nr.(socks5.DNSResolver); !ok {
					t.Errorf("expected default nameResolver, gotten: %T", nr)
					return
				}
			} else {
				// custom nameResolver
				if _, ok := nr.(*nameResolver); !ok {
					t.Errorf("expected custom nameResolver")
					return
				}
			}

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			_, ip, err := nr.Resolve(ctx, tc.host)
			if err != nil {
				if !tc.err {
					t.Errorf("unexpected error: %v", err)
				}
			} else {
				if tc.err {
					t.Errorf("expected error, got: %v", ip)
				} else {
					if ipString := ip.String(); ipString == "" {
						t.Errorf("expected ip")
					}
				}
			}
		})
	}
}
