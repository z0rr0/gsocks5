package auth

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/armon/go-socks5"
)

// New returns a new credential store.
func New(fileName string, logger *log.Logger) (socks5.CredentialStore, error) {
	auth, err := parse(fileName)
	if err != nil {
		return nil, err
	}
	if len(auth) == 0 {
		logger.Println("no authentication required")
		return nil, nil
	}
	for user := range auth {
		logger.Printf("adding credential for user %s", user)
	}
	return socks5.StaticCredentials(auth), nil
}

// parse reads the given file and returns a map of username/password pairs.
func parse(fileName string) (map[string]string, error) {
	if fileName == "" {
		return nil, nil
	}
	f, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to open auth file: %w", err)
	}

	result := make(map[string]string)
	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		values := strings.SplitN(scanner.Text(), " ", 2)
		if len(values) == 2 {
			user, pass := strings.Trim(values[0], " "), strings.Trim(values[1], " ")
			if user != "" && pass != "" {
				result[user] = pass
			}
		}
	}
	if err = scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read auth file: %w", err)
	}
	if err = f.Close(); err != nil {
		return nil, fmt.Errorf("failed to close auth file: %w", err)
	}
	return result, nil
}
