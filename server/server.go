package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
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
	Connections uint32
	Done        chan struct{} // only for testing
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
	connections, semaphore, err := s.listen(ctx, p, done)
	if err != nil {
		return err
	}

	s.logDebug.Printf("listener started on %s", p.Addr)
	p.Ready()
	go s.start(p, connections, semaphore)

	return s.waitClose(p, done)
}

// listen starts goroutine to accept incoming connections and sends them to a returned channel.
func (s *Server) listen(ctx context.Context, p *Params, done chan<- struct{}) (<-chan net.Conn, <-chan struct{}, error) {
	var lc net.ListenConfig

	listener, err := lc.Listen(ctx, "tcp", p.Addr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to listen on %s: %w", p.Addr, err)
	}

	p.listener = listener // to close it later
	connections := make(chan net.Conn)
	semaphore := make(chan struct{}, p.Connections)

	go func() {
		for {
			semaphore <- struct{}{} // limit connections, Server.handle will release it
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
		close(semaphore)   // no new incoming connections
		close(done)
	}()

	return connections, semaphore, nil
}

// accept accepts a new connection.
func (s *Server) accept(listener net.Listener, p *Params) (net.Conn, error) {
	conn, err := listener.Accept()
	if err != nil {
		return nil, fmt.Errorf("failed to accept connection: %w", err)
	}

	if p.Timeout > 0 {
		if err = conn.SetReadDeadline(time.Now().Add(p.Timeout)); err != nil {
			return nil, fmt.Errorf("failed to set read deadline for connection: %w", err)
		}
	}

	s.logDebug.Printf("accepted connection from %s with timeout %v", conn.RemoteAddr().String(), p.Timeout)
	return conn, nil
}

// start starts workers to handle incoming connections.
func (s *Server) start(p *Params, connections <-chan net.Conn, semaphore <-chan struct{}) {
	for conn := range connections {
		go s.handle(p, conn, semaphore)
	}
	s.logInfo.Printf("finished connections handling cycle")
}

func (s *Server) handle(p *Params, conn net.Conn, semaphore <-chan struct{}) {
	const skipError = "i/o timeout"
	var (
		t      = time.Now()
		client = conn.RemoteAddr().String()
		err    error
	)
	p.wg.Add(1)
	s.logDebug.Printf("accepted connection from %s", client)

	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			if errors.Is(closeErr, net.ErrClosed) {
				s.logDebug.Printf("connection from %s is closed", client)
			} else {
				s.logInfo.Printf("failed to close connection from client %q: %v", client, closeErr)
			}
		}
		<-semaphore // release the limitation
		p.wg.Done()
	}()

	if err = s.S.ServeConn(conn); err != nil {
		if errMsg := err.Error(); strings.HasSuffix(errMsg, skipError) {
			s.logDebug.Printf("connection from %s is closed due to timeout: %v", client, err)
		} else {
			s.logInfo.Printf("failed to serve connection from client %q [%T]: %v", client, err, err)
		}
	} else {
		s.logDebug.Printf("connection served from %s during %v", client, time.Since(t))
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

	<-done
	s.logInfo.Println("listener is closed")

	p.wg.Wait()
	s.logInfo.Println("all connections are handled")

	return nil
}
