package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/armon/go-socks5"
)

// Server is a socks5 server struct.
type Server struct {
	S        *socks5.Server
	logInfo  *log.Logger
	logDebug *log.Logger
}

// Params is a start parameters for the server.
type Params struct {
	Addr        string
	Connections int
	Done        chan struct{}
	Sigint      chan os.Signal
	Timeout     time.Duration
	setReady    sync.Once
	wg          sync.WaitGroup
	listener    net.Listener
}

// Ready closes Done channel if it is not closed yet.
func (p *Params) Ready() {
	p.setReady.Do(func() {
		if p.Done != nil {
			close(p.Done)
		}
	})
}

// New creates a new socks5 server.
func New(cfg *socks5.Config, logInfo, logDebug *log.Logger) (*Server, error) {
	server, err := socks5.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create socks5 server: %w", err)
	}
	return &Server{S: server, logInfo: logInfo, logDebug: logDebug}, nil
}

// ListenAndServe starts the socks5 server.
func (s *Server) ListenAndServe(p *Params) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		p.Ready()
		cancel()
	}()

	done := make(chan struct{})
	connections, err := s.listen(ctx, p, done)
	if err != nil {
		return err
	}

	s.logDebug.Printf("listener started on %s", p.Addr)
	p.Ready()
	s.startWorkers(p, connections)

	return s.waitClose(p, done)
}

// accept accepts a new connection.
func (s *Server) accept(listener net.Listener, p *Params) (net.Conn, error) {
	conn, err := listener.Accept()
	if err != nil {
		return nil, fmt.Errorf("failed to accept connection: %w", err)
	}

	if p.Timeout > 0 {
		if err = conn.SetReadDeadline(time.Now().Add(p.Timeout)); err != nil {
			return nil, fmt.Errorf("failed to set deadline for connection: %w", err)
		}
	}

	s.logDebug.Printf("accepted connection from %s with timeout %v", conn.RemoteAddr().String(), p.Timeout)
	return conn, nil
}

// listen starts goroutine to accept incoming connections and sends them to a returned channel.
func (s *Server) listen(ctx context.Context, p *Params, done chan<- struct{}) (<-chan net.Conn, error) {
	var lc net.ListenConfig

	listener, err := lc.Listen(ctx, "tcp", p.Addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s: %w", p.Addr, err)
	}

	p.listener = listener // to close it later
	connections := make(chan net.Conn)

	go func() {
		for {
			if conn, e := s.accept(listener, p); e != nil {
				if errors.Is(e, net.ErrClosed) {
					break
				}
				s.logInfo.Printf("failed to accept connection [%T]: %v", e, e)
			} else {
				connections <- conn
			}
		}

		s.logDebug.Printf("listener stopped")
		close(connections) // finish workers
		close(done)
	}()

	return connections, nil
}

// startWorkers starts workers to handle incoming connections.
func (s *Server) startWorkers(p *Params, connections <-chan net.Conn) {
	for i := 0; i < p.Connections; i++ {
		go func() {
			var (
				client string
				err    error
				t      time.Time
			)

			for conn := range connections {
				p.wg.Add(1)

				t = time.Now()
				client = conn.RemoteAddr().String()
				s.logDebug.Printf("accepted connection from %s", client)

				if err = s.S.ServeConn(conn); err != nil {
					s.logInfo.Printf("failed to serve connection from client %q: %v", client, err)
				} else {
					s.logDebug.Printf("connection served from %s during %v", client, time.Since(t))
				}

				p.wg.Done()
			}
		}()
	}
}

// waitClose waits for a signal to close the listener.
// It's a blocking function that returns when the listener is closed and all connections are handled.
func (s *Server) waitClose(p *Params, done <-chan struct{}) error {
	signal := <-p.Sigint
	s.logInfo.Printf("taken signal %v", signal)

	if err := p.listener.Close(); err != nil {
		return fmt.Errorf("failed to close listener: %w", err)
	}

	<-done      // wait for listener accept was stopped
	p.wg.Wait() // wait for all connections to be handled

	s.logInfo.Printf("all connections are handled")
	return nil
}
