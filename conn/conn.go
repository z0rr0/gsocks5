package conn

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"time"
)

// DialType is a dial function type alias.
type DialType = func(ctx context.Context, network, addr string) (net.Conn, error)

// idleTimeoutConn is a net.Conn wrapper with idle timeout.
type idleTimeoutConn struct {
	net.Conn
	timeout time.Duration
	logger  *log.Logger
}

// newIdleTimeoutConn creates a new idleTimeoutConn.
func newIdleTimeoutConn(conn net.Conn, timeout time.Duration, logger *log.Logger) *idleTimeoutConn {
	return &idleTimeoutConn{Conn: conn, timeout: timeout, logger: logger}
}

// Read reads data from the connection.
func (c *idleTimeoutConn) Read(b []byte) (int, error) {
	var netErr net.Error
	n, err := c.Conn.Read(b)

	if err != nil {
		if errors.As(err, &netErr) && netErr.Timeout() {
			c.logger.Printf("idleTimeoutConn read timeout from %s to %s", c.Conn.RemoteAddr(), c.Conn.LocalAddr())
		}
	} else if c.timeout > 0 {
		if deadlineErr := c.Conn.SetReadDeadline(time.Now().Add(c.timeout)); deadlineErr != nil {
			c.logger.Printf("idleTimeoutConn failed to set read deadline: %v", deadlineErr)
		}
	}

	return n, err
}

// Write writes data to the connection.
func (c *idleTimeoutConn) Write(b []byte) (int, error) {
	var netErr net.Error
	n, err := c.Conn.Write(b)

	if err != nil {
		if errors.As(err, &netErr) && netErr.Timeout() {
			c.logger.Printf("idleTimeoutConn write timeout from %s to %s", c.Conn.RemoteAddr(), c.Conn.LocalAddr())
		}
	} else if c.timeout > 0 {
		if deadlineErr := c.Conn.SetWriteDeadline(time.Now().Add(c.timeout)); deadlineErr != nil {
			c.logger.Printf("idleTimeoutConn failed to set write deadline: %v", deadlineErr)
		}
	}

	return n, err
}

// Dial creates a new DialType.
func Dial(dialer *net.Dialer, timeout time.Duration, logger *log.Logger) DialType {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		connection, err := dialer.DialContext(ctx, network, addr)
		if err != nil {
			return nil, fmt.Errorf("failed to dial %s: %w", addr, err)
		}

		return newIdleTimeoutConn(connection, timeout, logger), nil
	}
}
