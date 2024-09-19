package args

import (
	"os"
	"testing"
)

func TestIsFile(t *testing.T) {
	tempFile, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatalf("failed to create temporary file: %v", err)
	}

	defer func() {
		var e error

		if e = tempFile.Close(); e != nil {
			t.Fatalf("failed to close temporary file: %v", e)
		}

		if e = os.Remove(tempFile.Name()); e != nil {
			t.Fatalf("failed to remove temporary file: %v", e)
		}
	}()

	testCases := []struct {
		name    string
		value   string
		want    string
		wantErr bool
	}{
		{name: "ValidFile", value: tempFile.Name(), want: tempFile.Name()},
		{name: "InvalidFile", value: "nonExistentFile.txt", wantErr: true},
		{name: "Directory", value: os.TempDir(), wantErr: true},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			var result string
			err = IsFile(tc.value, &result)

			if (err != nil) != tc.wantErr {
				t.Errorf("IsFile() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if result != tc.want {
				t.Errorf("IsFile() = %v, want %v", result, tc.want)
			}
		})
	}
}

func TestIsPort(t *testing.T) {
	testCases := []struct {
		name    string
		value   string
		want    uint16
		wantErr bool
	}{
		{name: "ValidPort", value: "8080", want: 8080},
		{name: "InvalidPort", value: "70000", wantErr: true},
		{name: "NonNumericPort", value: "abc", wantErr: true},
		{name: "ZeroPort", value: "0", wantErr: true},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			var result uint16

			err := IsPort(tc.value, &result)
			if (err != nil) != tc.wantErr {
				t.Errorf("IsPort() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if result != tc.want {
				t.Errorf("IsPort() = %v, want %v", result, tc.want)
			}
		})
	}
}

func TestIsConcurrent(t *testing.T) {
	testCases := []struct {
		name    string
		value   string
		want    uint64
		wantErr bool
	}{
		{name: "ValidConcurrent", value: "100", want: 100},
		{name: "TooLowConcurrent", value: "0", wantErr: true},
		{name: "TooHighConcurrent", value: "4294967296", wantErr: true},
		{name: "NonNumericConcurrent", value: "abc", wantErr: true},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			var result uint32
			err := IsConcurrent(tc.value, &result)

			if (err != nil) != tc.wantErr {
				t.Errorf("IsConcurrent() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if uint64(result) != tc.want {
				t.Errorf("IsConcurrent() = %v, want %v", result, tc.want)
			}
		})
	}
}

func TestPortDescription(t *testing.T) {
	result := PortDescription(8080)
	expected := "TCP port number to listen on in range [1, 65535] (default 8080)"

	if result != expected {
		t.Errorf("PortDescription() = %v, want %v", result, expected)
	}
}

func TestConcurrentDescription(t *testing.T) {
	result := ConcurrentDescription(10000)
	expected := "number of concurrent connections in range [1, 1000000] (default 10000)"

	if result != expected {
		t.Errorf("ConcurrentDescription() = %v, want %v", result, expected)
	}
}
