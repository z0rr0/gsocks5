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
	"github.com/z0rr0/gsocks5/conn"
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
		authFile         string
		customDNS        string
		host             string
		version          bool
		debugMode        bool
		connections      uint32 = 1024
		port             uint16 = 1080
		timeoutIdle             = 15 * time.Second
		timeoutDNS              = 5 * time.Second
		timeoutKeepAlive        = 30 * time.Second
		timeoutConn             = 5 * time.Second
	)
	defer func() {
		if r := recover(); r != nil {
			logInfo.Printf("abnormal termination [%v]: %v\n", Version, r)
		}
	}()

	flag.StringVar(&customDNS, "dns", customDNS, "custom DNS server")
	flag.BoolVar(&version, "version", false, "show version")
	flag.StringVar(&host, "host", "", "server host")
	flag.DurationVar(&timeoutIdle, "ti", timeoutIdle, "idle timeout")
	flag.DurationVar(&timeoutDNS, "td", timeoutDNS, "dns timeout")
	flag.DurationVar(&timeoutKeepAlive, "tk", timeoutKeepAlive, "keepalive timeout")
	flag.DurationVar(&timeoutConn, "tc", timeoutConn, "connection timeout")
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

	dialer := &net.Dialer{Timeout: timeoutConn, KeepAlive: timeoutKeepAlive}
	cfg := &socks5.Config{
		Logger:      logInfo,
		Credentials: credentials,
		Resolver:    resolver,
		Dial:        conn.Dial(dialer, timeoutIdle, logInfo),
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
		"starting server on %q, dns=%q, dns timeoutIdle=%v, cons timeoutIdle=%v, connections=%d, debug=%v, auth=%q\n",
		addr, customDNS, timeoutDNS, timeoutIdle, connections, debugMode, authFile,
	)

	params := &server.Params{Addr: addr, Connections: connections, Sigint: sigint}
	if err = s.ListenAndServe(params); err != nil {
		logInfo.Printf("server listen error: %s", err)
	}

	logInfo.Println("server stopped")
}
