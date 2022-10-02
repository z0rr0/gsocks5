package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/armon/go-socks5"

	"github.com/z0rr0/gsocks5/auth"
	"github.com/z0rr0/gsocks5/dns"
	"github.com/z0rr0/gsocks5/server"
)

const name = "GSocks5"

var (
	// Tag is the git tag of the build.
	Tag string

	logInfo  = log.New(os.Stdout, name+" [INFO]: ", log.LstdFlags)
	logDebug = log.New(io.Discard, name+" [DEBUG]: ", log.LstdFlags|log.Lshortfile)
)

func main() {
	var (
		authFile  string
		customDNS string
		host      string
		version   bool
		port      uint
		timeout   uint
		debugMode bool
	)
	defer func() {
		if r := recover(); r != nil {
			logInfo.Printf("abnormal termination [%v]: %v\n%v", Tag, r, string(debug.Stack()))
		}
	}()
	flag.StringVar(&authFile, "auth", "", "authentication file")
	flag.StringVar(&customDNS, "dns", "", "custom DNS server")
	flag.BoolVar(&version, "version", false, "show version")
	flag.StringVar(&host, "host", "", "server host")
	flag.UintVar(&port, "port", 1080, "port to listen on")
	flag.UintVar(&timeout, "timeout", 5, "context timeout (seconds)")
	flag.BoolVar(&debugMode, "debug", false, "debug mode")

	flag.Parse()
	if version {
		fmt.Println(server.Version(name, Tag))
		return
	}
	if debugMode {
		logDebug.SetOutput(os.Stdout)
	}

	credentials, err := auth.New(authFile, logInfo)
	if err != nil {
		logInfo.Fatal(err)
	}

	timeoutDNS := time.Duration(timeout) * time.Second
	resolver, err := dns.New(customDNS, timeoutDNS, logInfo, logDebug)
	if err != nil {
		logInfo.Fatal(err)
	}

	cfg := &socks5.Config{
		Logger:      logInfo,
		Credentials: credentials,
		Resolver:    resolver,
	}

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt, os.Signal(syscall.SIGTERM), os.Signal(syscall.SIGQUIT))
	defer close(sigint)

	s, err := server.New(cfg, logInfo, logDebug)
	if err != nil {
		logInfo.Fatal(err)
	}

	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	logInfo.Printf("starting server on %s, debugMode=%v", addr, debugMode)

	if err = s.ListenAndServe(addr, sigint); err != nil {
		logInfo.Printf("server listen error: %s", err)
	}
	logInfo.Println("server stopped")
}
