package server

import (
	"log"
	"net"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/armon/go-socks5"
	"golang.org/x/net/proxy"

	"github.com/z0rr0/gsocks5/conn"
)

const timeout = 200 * time.Millisecond

var logger = log.New(os.Stdout, "[test] ", log.LstdFlags|log.Lshortfile|log.Lmicroseconds)

func run(t *testing.T, s *Server, i, port int, isErr bool) (string, chan os.Signal) {
	params := &Params{
		Addr:        net.JoinHostPort("localhost", strconv.Itoa(port)),
		Connections: 1,
		Done:        make(chan struct{}),
		Sigint:      make(chan os.Signal),
	}

	go func() {
		if err := s.ListenAndServe(params); err != nil {
			if !isErr {
				t.Errorf("run server case %d: %v", i, err)
			}
		}
	}()

	<-params.Done
	return params.Addr, params.Sigint
}

type testHost struct {
	host  string
	port  int
	close bool
}

func TestNew(t *testing.T) {
	cases := []struct {
		name  string
		port  int
		hosts []testHost
		err   bool
	}{
		{name: "one", port: 1080, hosts: []testHost{{host: "github.com", port: 443}}},
		{
			name: "two",
			port: 1080,
			hosts: []testHost{
				{host: "github.com", port: 443},
				{host: "leetcode.com", port: 443, close: true},
			},
		},
		{
			name: "three",
			port: 1080,
			hosts: []testHost{
				{host: "github.com", port: 443, close: true},
				{host: "leetcode.com", port: 443, close: true},
				{host: "leetcode.com", port: 80},
			},
		},
		{
			name: "many",
			port: 1080,
			hosts: []testHost{
				{host: "github.com", port: 443, close: true},
				{host: "github.com", port: 80},
				{host: "leetcode.com", port: 443, close: true},
				{host: "leetcode.com", port: 80},
			},
		},
		{name: "badPort", port: 131072, err: true},
	}
	for i, c := range cases {
		t.Run(c.name, func(tt *testing.T) {
			var (
				connection net.Conn
				cfgDialer  = &net.Dialer{Timeout: timeout * 15}
				cfg        = &socks5.Config{Logger: logger, Dial: conn.Dial(cfgDialer, timeout, logger)}
			)
			s, err := New(cfg, logger, logger)
			if err != nil {
				tt.Errorf("case [%d] %s: unexpected error: %v", i, c.name, err)
			}

			addr, sigint := run(tt, s, i, c.port, c.err)
			defer close(sigint)

			if c.err {
				return // processing in run()
			}

			dialer, err := proxy.SOCKS5("tcp", addr, nil, proxy.Direct)
			if err != nil {
				tt.Errorf("case [%d] %s: unexpected error: %v", i, c.name, err)
			}

			for _, h := range c.hosts {
				connection, err = dialer.Dial("tcp", net.JoinHostPort(h.host, strconv.Itoa(h.port)))
				if err != nil {
					tt.Errorf("case [%d] %s: %v", i, c.name, err)
				}

				if h.close {
					if err = connection.Close(); err != nil {
						tt.Errorf("case [%d] %s: %v", i, c.name, err)
					}
				}
			}
			sigint <- os.Interrupt
		})
	}
}
