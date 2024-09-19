package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
	_ "time/tzdata"

	"github.com/armon/go-socks5"

	"github.com/z0rr0/gsocks5/args"
	"github.com/z0rr0/gsocks5/auth"
	"github.com/z0rr0/gsocks5/dns"
	"github.com/z0rr0/gsocks5/server"
)

const name = "GSocks5"

var (
	// Version is git version
	Version = ""
	// Revision is revision number
	Revision = ""
	// BuildDate is build date
	BuildDate = ""
	// GoVersion is runtime Go language version
	GoVersion = runtime.Version()

	logInfo  = log.New(os.Stdout, name+" [INFO]: ", log.LstdFlags)
	logDebug = log.New(io.Discard, name+" [DEBUG]: ", log.LstdFlags|log.Lshortfile|log.Lmicroseconds)
)

func main() {
	var (
		authFile    string
		customDNS   string
		host        string
		version     bool
		debugMode   bool
		connections uint32 = 1024
		port        uint16 = 1080
		timeout            = 3 * time.Minute
		timeoutDNS         = 5 * time.Second
	)
	defer func() {
		if r := recover(); r != nil {
			logInfo.Printf("abnormal termination [%v]: %v\n", Version, r)
		}
	}()

	flag.StringVar(&customDNS, "dns", customDNS, "custom DNS server")
	flag.BoolVar(&version, "version", false, "show version")
	flag.StringVar(&host, "host", "", "server host")
	flag.DurationVar(&timeout, "t", timeout, "connection timeout")
	flag.DurationVar(&timeoutDNS, "dt", timeoutDNS, "dns connection timeout")
	flag.BoolVar(&debugMode, "debug", false, "debug mode")
	flag.Func("port", args.PortDescription(port), func(s string) error { return args.IsPort(s, &port) })
	flag.Func("auth", "authentication file", func(s string) error { return args.IsFile(s, &authFile) })
	flag.Func("connections", args.ConcurrentDescription(connections), func(s string) error {
		return args.IsConcurrent(s, &connections)
	})

	flag.Parse()

	versionInfo := fmt.Sprintf("%v: %v %v %v %v", name, Version, Revision, GoVersion, BuildDate)
	if version {
		fmt.Println(versionInfo)
		return
	}
	if debugMode {
		logDebug.SetOutput(os.Stdout)
	}

	credentials, err := auth.New(authFile, logInfo)
	if err != nil {
		logInfo.Fatal(err)
	}

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
	logInfo.Println(versionInfo)
	logInfo.Printf(
		"starting server on %q, dns=%q, dns timeout=%v, cons timeout=%v, connections=%d, debug=%v, auth=%q\n",
		addr, customDNS, timeoutDNS, timeout, connections, debugMode, authFile,
	)

	params := &server.Params{Addr: addr, Connections: connections, Sigint: sigint, Timeout: timeout}
	if err = s.ListenAndServe(params); err != nil {
		logInfo.Printf("server listen error: %s", err)
	}

	logInfo.Println("server stopped")
}
