package auth

import (
	"log"
	"os"
	"testing"

	"github.com/armon/go-socks5"
)

var logger = log.New(os.Stdout, "[test] ", log.LstdFlags|log.Lshortfile)

func userFile(rows []string) (string, error) {
	f, err := os.CreateTemp("", "users_gsocks5_test")
	if err != nil {
		return "", err
	}
	for _, row := range rows {
		if _, err = f.WriteString(row); err != nil {
			return "", err
		}
	}
	if err = f.Sync(); err != nil {
		return "", err
	}
	return f.Name(), nil
}

func TestNew(t *testing.T) {
	var cases = []struct {
		name     string
		badName  string
		empty    bool
		rows     []string
		expected map[string]string
	}{
		{name: "empty", empty: true, rows: []string{}},
		{name: "emptyFile", rows: []string{}},
		{name: "skipped", rows: []string{"bad"}},
		{name: "bad", badName: "/tmp/bad", rows: []string{}},
		{
			name:     "one",
			rows:     []string{"user1 password1\n"},
			expected: map[string]string{"user1": "password1"},
		},
		{
			name:     "two",
			rows:     []string{"user1 password1\n", "u2 pwd2\n"},
			expected: map[string]string{"user1": "password1", "u2": "pwd2"},
		},
		{
			name:     "multilines",
			rows:     []string{"\nuser1 password1\n\n\n", "u2 pwd2\n\nbad\n\n"},
			expected: map[string]string{"user1": "password1", "u2": "pwd2"},
		},
	}
	for i, c := range cases {
		t.Run(c.name, func(tt *testing.T) {
			var (
				fileName string
				err      error
				failed   bool
			)
			if len(c.badName) == 0 {
				if !c.empty {
					fileName, err = userFile(c.rows)
					if err != nil {
						tt.Fatal(err)
					}
					defer func() {
						if err = os.Remove(fileName); err != nil {
							tt.Fatal(err)
						}
					}()
				}
			} else {
				fileName = c.badName
				failed = true
			}
			credentials, err := New(fileName, logger)
			if err != nil {
				if failed {
					return // expected error
				}
				tt.Fatal(err)
			}
			if failed {
				tt.Fatal("expected error")
			}
			// check empty credentials
			if n := len(c.expected); n == 0 {
				if credentials != nil {
					tt.Fatal("expected nil credentials")
				}
				return
			}
			// no error, values is to be non-nil
			values, ok := credentials.(socks5.StaticCredentials)
			if !ok {
				tt.Fatalf("unexpected credentials type: %T", credentials)
			}
			if n, m := len(values), len(c.expected); n != m {
				t.Errorf("case %d: expected %d rows, got %d", i, m, n)
			}
			for k, v := range c.expected {
				if actual := values[k]; actual != v {
					t.Errorf("case [%d] %s: expected %s, got %s", i, c.name, v, actual)
				}
			}
		})
	}
}
