package args

import (
	"fmt"
	"os"
	"strconv"
)

const (
	mintPort = 1
	maxPort  = 65535

	minConcurrent uint64 = 1
	maxConcurrent uint64 = 1_000_000
)

// IsFile checks that the value is a file.
func IsFile(value string, result *string) error {
	stat, err := os.Stat(value)
	if err != nil {
		return err
	}

	if stat.IsDir() {
		return fmt.Errorf("it is a directory")
	}

	*result = value
	return nil
}

// IsPort checks that the value is a valid port number.
func IsPort(value string, result *uint16) error {
	port, err := strconv.ParseUint(value, 10, 16)
	if err != nil {
		return err
	}

	if port < mintPort || port > maxPort {
		return fmt.Errorf("port is out of range")
	}

	*result = uint16(port)
	return nil
}

// IsConcurrent checks that the value is a valid number of concurrent connections.
func IsConcurrent(value string, result *uint32) error {
	integer, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return err
	}

	if integer < minConcurrent || integer > maxConcurrent {
		return fmt.Errorf("value is out of range")
	}

	*result = uint32(integer)
	return nil
}

// PortDescription returns a description of the port argument.
func PortDescription(value uint16) string {
	return fmt.Sprintf("TCP port number to listen on in range [%d, %d] (default %d)", mintPort, maxPort, value)
}

// ConcurrentDescription returns a description of the concurrent argument.
func ConcurrentDescription(value uint32) string {
	return fmt.Sprintf(
		"number of concurrent connections in range [%d, %d] (default %d)",
		minConcurrent, maxConcurrent, value,
	)
}
