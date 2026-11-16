package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestShouldRetry(t *testing.T) {
	for _, message := range []string{
		"fatal: couldn't find remote ref main",
		"fatal: requested URL returned error: 429",
		"fatal: requested URL returned error: 500",
		"fatal: requested URL returned error: 502\nfatal: could not fetch object from promisor remote",
		"fatal: requested URL returned error: 503",
		"fatal: requested URL returned error: 504",
		"fatal: Could not resolve host: example.org",
		"fatal: Failed to connect to example.org port 443",
		"fatal: Recv failure: Connection reset by peer",
		"fatal: Operation timed out",
		"fatal: remote end hung up unexpectedly",
		"fatal: TLS connection was non-properly terminated",
	} {
		t.Run(message, func(t *testing.T) {
			if !shouldRetry(message) {
				t.Errorf("expected retry for %q", message)
			}
		})
	}
	for _, message := range []string{
		"fatal: requested URL returned error: 401",
		"fatal: requested URL returned error: 403",
		"fatal: requested URL returned error: 404",
		"fatal: requested URL returned error: 5020",
		"fatal: requested URL returned error: ",
		"fatal: Authentication failed",
		"fatal: could not fetch object from promisor remote",
		"fatal: repository not found",
		"fatal: SSL certificate problem: certificate has expired",
		"",
	} {
		t.Run(message, func(t *testing.T) {
			if shouldRetry(message) {
				t.Errorf("unexpected retry for %q", message)
			}
		})
	}
}

func TestRetryExecStopsAfterPermanentFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("requires a POSIX shell")
	}
	path := t.TempDir()
	cmd := exec.Command("sh", "-c", `printf 'attempt\n' >> attempts; printf 'fatal: requested URL returned error: 401\n' >&2; exit 1`)
	cmd.Dir = path
	if err := retryExec(cmd, 0, 3); err == nil {
		t.Fatal("expected authentication error")
	}
	attempts, err := os.ReadFile(filepath.Join(path, "attempts"))
	if err != nil {
		t.Fatal(err)
	}
	if string(attempts) != "attempt\n" {
		t.Fatalf("permanent failure retried: %q", attempts)
	}
}

func TestRetryExecBoundsTransientFailures(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("requires a POSIX shell")
	}
	path := t.TempDir()
	cmd := exec.Command("sh", "-c", `printf 'attempt\n' >> attempts; printf 'fatal: requested URL returned error: 503\n' >&2; exit 1`)
	cmd.Dir = path
	if err := retryExec(cmd, 0, 2); err == nil {
		t.Fatal("expected server error")
	}
	attempts, err := os.ReadFile(filepath.Join(path, "attempts"))
	if err != nil {
		t.Fatal(err)
	}
	if string(attempts) != "attempt\nattempt\n" {
		t.Fatalf("unexpected retry count: %q", attempts)
	}
}
