package cli

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIExitCodes(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "cli_test.wal")

	var stdout, stderr bytes.Buffer

	// 1. Missing / invalid args -> ExitUsage (2)
	code := Run([]string{}, &stdout, &stderr)
	if code != ExitUsage {
		t.Errorf("expected ExitUsage (2), got %d", code)
	}

	code = Run([]string{"put", dbPath}, &stdout, &stderr)
	if code != ExitUsage {
		t.Errorf("expected ExitUsage (2), got %d", code)
	}

	// 2. Put valid key-value -> ExitOK (0)
	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"put", dbPath, "user", "alice"}, &stdout, &stderr)
	if code != ExitOK {
		t.Errorf("expected ExitOK (0) for put, got %d, stderr: %s", code, stderr.String())
	}

	// 3. Get existing key -> ExitOK (0) and stdout contains value
	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"get", dbPath, "user"}, &stdout, &stderr)
	if code != ExitOK {
		t.Errorf("expected ExitOK (0) for get, got %d, stderr: %s", code, stderr.String())
	}
	if strings.TrimSpace(stdout.String()) != "alice" {
		t.Errorf("expected stdout 'alice', got %q", stdout.String())
	}

	// 4. Get missing key -> ExitNotFound (3)
	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"get", dbPath, "unknown_key"}, &stdout, &stderr)
	if code != ExitNotFound {
		t.Errorf("expected ExitNotFound (3) for missing get, got %d", code)
	}

	// 5. Del missing key -> ExitNotFound (3)
	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"del", dbPath, "unknown_key"}, &stdout, &stderr)
	if code != ExitNotFound {
		t.Errorf("expected ExitNotFound (3) for missing del, got %d", code)
	}

	// 6. Del existing key -> ExitOK (0)
	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"del", dbPath, "user"}, &stdout, &stderr)
	if code != ExitOK {
		t.Errorf("expected ExitOK (0) for del, got %d, stderr: %s", code, stderr.String())
	}

	// 7. Get deleted key -> ExitNotFound (3)
	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"get", dbPath, "user"}, &stdout, &stderr)
	if code != ExitNotFound {
		t.Errorf("expected ExitNotFound (3) for get deleted key, got %d", code)
	}
}
