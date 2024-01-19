package auth

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/armon/go-socks5"
)

const dataDir = "/data"

var (
	// ErrAuthFile is returned when the auth file path or content is invalid.
	ErrAuthFile = fmt.Errorf("invalid auth file path")
)

type usersData map[string]string

func (u usersData) String() string {
	const sep = ", "
	var b strings.Builder

	for user := range u {
		b.WriteString(user)
		b.WriteString(sep)
	}

	return strings.TrimRight(b.String(), sep)
}

// New returns a new credential store.
func New(fileName string, logger *log.Logger) (socks5.CredentialStore, error) {
	users, err := parseFile(fileName)
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		logger.Println("no credentials found")
		return nil, nil
	}

	logger.Printf("found credentials for users: %s", users.String())
	return socks5.StaticCredentials(users), nil
}

// parseFile reads the given file and returns a map of username/password pairs.
func parseFile(fileName string) (usersData, error) {
	if fileName == "" {
		return nil, nil
	}

	fileName = filepath.Clean(fileName)
	if !(strings.HasPrefix(fileName, dataDir) || strings.HasPrefix(fileName, os.TempDir())) {
		return nil, errors.Join(ErrAuthFile, fmt.Errorf("invalid auth file path: %s", fileName))
	}

	f, err := os.Open(fileName)
	if err != nil {
		return nil, errors.Join(ErrAuthFile, fmt.Errorf("failed to open file: %w", err))
	}

	users, scanErr := readFile(f)
	if err = errors.Join(scanErr, f.Close()); err != nil {
		return nil, errors.Join(ErrAuthFile, fmt.Errorf("failed to scan or close file: %w", err))
	}

	return users, nil
}

func readFile(f *os.File) (usersData, error) {
	var (
		users   = make(map[string]string)
		scanner = bufio.NewScanner(f)
	)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		values := strings.Fields(scanner.Text())

		if len(values) == 2 {
			user := strings.TrimSpace(values[0])
			password := strings.TrimSpace(values[1])

			if user != "" && password != "" {
				users[user] = password
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
