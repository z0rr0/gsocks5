package conn

import (
	"context"
	"log"
	"net"
	"os"
	"testing"
	"time"
)

const timeout = 3 * time.Second

var logger = log.New(os.Stdout, "[test] ", log.LstdFlags|log.Lshortfile|log.Lmicroseconds)

// testConn is a test net.Conn implementation.
type testConn struct {
	readDeadline  time.Time
	writeDeadline time.Time
}

func (c *testConn) Read([]byte) (int, error)           { return 0, nil }
func (c *testConn) Write([]byte) (int, error)          { return 0, nil }
func (c *testConn) Close() error                       { return nil }
func (c *testConn) LocalAddr() net.Addr                { return nil }
func (c *testConn) RemoteAddr() net.Addr               { return nil }
func (c *testConn) SetDeadline(time.Time) error        { return nil }
func (c *testConn) SetReadDeadline(t time.Time) error  { c.readDeadline = t; return nil }
func (c *testConn) SetWriteDeadline(t time.Time) error { c.writeDeadline = t; return nil }

func TestIdleTimeoutConn_Read(t *testing.T) {
	var connection = &testConn{}
	idleConn := newIdleTimeoutConn(connection, timeout, logger)

	if idleConn == nil {
		t.Fatal("unexpected nil idleConn")
	}

	if idleConn.timeout != timeout {
		t.Errorf("unexpected timeout: %v", idleConn.timeout)
	}
	if idleConn.logger == nil {
		t.Error("unexpected nil logger")
	}

	if _, err := idleConn.Read(nil); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if connection.readDeadline.IsZero() {
		t.Error("unexpected read deadline")
	}
}

func TestIdleTimeoutConn_Write(t *testing.T) {
	var connection = &testConn{}
	idleConn := newIdleTimeoutConn(connection, timeout, logger)

	if idleConn == nil {
		t.Fatal("unexpected nil idleConn")
	}

	if idleConn.timeout != timeout {
		t.Errorf("unexpected timeout: %v", idleConn.timeout)
	}
	if idleConn.logger == nil {
		t.Error("unexpected nil logger")
	}

	if _, err := idleConn.Write(nil); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if connection.writeDeadline.IsZero() {
		t.Error("unexpected write deadline")
	}
}

func TestDial(t *testing.T) {
	connFunc := Dial(&net.Dialer{}, timeout, logger)
	if connFunc == nil {
		t.Fatal("unexpected nil connFunc")
	}

	// fail connection
	_, err := connFunc(context.Background(), "tcp", "localhost:1080")
	if err == nil {
		t.Errorf("unexpected nil error")
	}

	// success connection
	connection, err := connFunc(context.Background(), "tcp", "github.com:443")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if connection == nil {
		t.Error("unexpected nil connFunc")
	}

	if _, ok := connection.(*idleTimeoutConn); !ok {
		t.Errorf("unexpected connection type [%T]", connection)
	}
}
