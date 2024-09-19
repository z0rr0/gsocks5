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
)

const timeout = 2 * time.Second

var logger = log.New(os.Stdout, "[test] ", log.LstdFlags|log.Lshortfile|log.Lmicroseconds)

func run(t *testing.T, s *Server, i, port int, isErr bool) (string, chan os.Signal) {
	params := &Params{
		Addr:        net.JoinHostPort("localhost", strconv.Itoa(port)),
		Connections: 1,
		Done:        make(chan struct{}),
		Sigint:      make(chan os.Signal),
		Timeout:     timeout,
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
				conn net.Conn
				cfg  = &socks5.Config{Logger: logger}
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
				conn, err = dialer.Dial("tcp", net.JoinHostPort(h.host, strconv.Itoa(h.port)))
				if err != nil {
					tt.Errorf("case [%d] %s: %v", i, c.name, err)
				}

				if h.close {
					if err = conn.Close(); err != nil {
						tt.Errorf("case [%d] %s: %v", i, c.name, err)
					}
				}
			}
			sigint <- os.Interrupt
		})
	}
}
