package server

import (
	"log"
	"net"
	"os"
	"strconv"
	"testing"

	"github.com/armon/go-socks5"
	"golang.org/x/net/proxy"
)

var logger = log.New(os.Stdout, "[test] ", log.LstdFlags|log.Lshortfile)

func run(t *testing.T, s *Server, i, port int, isErr bool) (string, chan os.Signal) {
	sigint := make(chan os.Signal)
	start := make(chan struct{})
	addr := net.JoinHostPort("localhost", strconv.Itoa(port))

	go func() {
		if err := s.ListenAndServe(addr, start, sigint); err != nil {
			if !isErr {
				t.Errorf("run server case %d: %v", i, err)
			}
		}
	}()

	<-start
	return addr, sigint
}

func TestNew(t *testing.T) {
	cases := []struct {
		name string
		port int
		host string
		err  bool
	}{
		{name: "empty", port: 1080, host: "github.com:443"},
		{name: "badPort", port: 131072, err: true},
	}
	for i, c := range cases {
		t.Run(c.name, func(tt *testing.T) {
			cfg := &socks5.Config{Logger: logger}
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

			conn, err := dialer.Dial("tcp", c.host)
			if err != nil {
				tt.Fatalf("set connection, case [%d] %s: %v", i, c.name, err)
			}

			if err = conn.Close(); err != nil {
				tt.Errorf("close connection, case [%d] %s: %v", i, c.name, err)
			}

			//time.Sleep(100 * time.Millisecond) // wait for server handle connection
			sigint <- os.Interrupt
		})
	}
}
