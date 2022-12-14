package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"runtime/debug"

	"github.com/armon/go-socks5"
)

// Server is a socks5 server struct.
type Server struct {
	S        *socks5.Server
	logInfo  *log.Logger
	logDebug *log.Logger
}

// New creates a new socks5 server.
func New(cfg *socks5.Config, logInfo, logDebug *log.Logger) (*Server, error) {
	server, err := socks5.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create socks5 server: %w", err)
	}
	return &Server{S: server, logInfo: logInfo, logDebug: logDebug}, nil
}

// listen accepts for incoming connections and sends them to a returned channel.
func (s *Server) listen(listener net.Listener, done chan<- struct{}) <-chan net.Conn {
	connections := make(chan net.Conn)
	go func() {
		for {
			conn, e := listener.Accept()
			if e != nil {
				if errors.Is(e, net.ErrClosed) {
					break
				}
				s.logInfo.Printf("failed to accept connection: %T %#v", e, e)
			}
			connections <- conn
		}
		close(connections)
		close(done)
		s.logDebug.Printf("listener stopped")
	}()
	return connections
}

// ListenAndServe starts the socks5 server.
func (s *Server) ListenAndServe(addr string, sigint <-chan os.Signal) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	lc := &net.ListenConfig{}
	listener, err := lc.Listen(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	done := make(chan struct{})
	connections := s.listen(listener, done)
	s.logDebug.Printf("listener started on %s", addr)
	for {
		select {
		case signal := <-sigint:
			s.logInfo.Printf("taken signal %v", signal)
			if err = listener.Close(); err != nil {
				return fmt.Errorf("failed to close listener: %w", err)
			}
			<-done
			return nil
		case conn := <-connections:
			s.logDebug.Printf("accepted connection from %s", conn.RemoteAddr())
			go s.handleConnection(conn)
		}
	}
}

// handleConnection handles a single connection.
func (s *Server) handleConnection(conn net.Conn) {
	err := s.S.ServeConn(conn)
	if err != nil {
		s.logInfo.Printf("failed to serve connection: %v", err)
	}
}

// Version prints the version of the program.
func Version(name, tag string) string {
	var keys = map[string]string{
		"vcs":          "",
		"vcs.revision": "",
		"vcs.time":     "",
	}
	if bi, ok := debug.ReadBuildInfo(); ok {
		for _, bs := range bi.Settings {
			if _, exists := keys[bs.Key]; exists {
				keys[bs.Key] = bs.Value
			}
		}
	}
	return fmt.Sprintf(
		"%s %s\n%s:%s\nbuild: %s",
		name, tag, keys["vcs"], keys["vcs.revision"], keys["vcs.time"],
	)
}
