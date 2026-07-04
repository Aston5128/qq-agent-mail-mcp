package agently

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
)

func TestRunnerReturnsParsedJSON(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "fake-agently")
	writeFakeExecutable(t, script, `{"ok":true,"data":{"message":"pong"}}`)

	r := Runner{Binary: script}
	got, err := r.Run(context.Background(), []string{"+me"})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if got["ok"] != true {
		t.Fatalf("ok = %#v, want true", got["ok"])
	}
}

func TestRunnerRejectsNonJSONOutput(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "fake-agently")
	writeFakeExecutable(t, script, `not json`)

	r := Runner{Binary: script}
	_, err := r.Run(context.Background(), []string{"+me"})
	if err == nil {
		t.Fatal("Run returned nil error for non-JSON output")
	}
}

func TestRunnerReturnsCLIErrorWithParsedStdoutForNonzeroExit(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "fake-agently")
	writeFakeExecutableWithExit(
		t,
		script,
		`{"ok":false,"error":{"type":"auth","message":"login required"}}`,
		"please login",
		7,
	)

	r := Runner{Binary: script}
	_, err := r.Run(context.Background(), []string{"+me"})
	if err == nil {
		t.Fatal("Run returned nil error for nonzero exit")
	}

	var cliErr *CLIError
	if !errors.As(err, &cliErr) {
		t.Fatalf("error = %T %[1]v, want CLIError", err)
	}
	if cliErr.ExitCode != 7 {
		t.Fatalf("ExitCode = %d, want 7", cliErr.ExitCode)
	}
	if cliErr.Stderr != "please login" {
		t.Fatalf("Stderr = %q, want please login", cliErr.Stderr)
	}
	gotError, ok := cliErr.Output["error"].(map[string]any)
	if !ok {
		t.Fatalf("Output[error] = %#v, want map[string]any", cliErr.Output["error"])
	}
	if gotError["type"] != "auth" {
		t.Fatalf("error type = %#v, want auth", gotError["type"])
	}
	if cliErr.Code() != "auth" {
		t.Fatalf("Code() = %q, want auth", cliErr.Code())
	}
}

// Code() resolves error.code ahead of error.type ahead of top-level code. When
// a payload carries both error.code and error.type (e.g. a richer future CLI),
// the explicit code wins and the auth discriminator stays scoped to stdout.
func TestRunnerCLIErrorCodeTakesPrecedenceOverType(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "fake-agently")
	writeFakeExecutableWithExit(
		t,
		script,
		`{"ok":false,"error":{"code":"UNAUTHORIZED","type":"auth","message":"login required"}}`,
		"please login",
		7,
	)

	r := Runner{Binary: script}
	_, err := r.Run(context.Background(), []string{"+me"})
	if err == nil {
		t.Fatal("Run returned nil error for nonzero exit")
	}

	var cliErr *CLIError
	if !errors.As(err, &cliErr) {
		t.Fatalf("error = %T %[1]v, want CLIError", err)
	}
	if got := cliErr.Code(); got != "UNAUTHORIZED" {
		t.Fatalf("Code() = %q, want UNAUTHORIZED (error.code must beat error.type)", got)
	}
}

func writeFakeExecutable(t *testing.T, path string, stdout string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("shell script fake is unix-only")
	}
	body := "#!/bin/sh\nprintf '%s' " + shellQuote(stdout) + "\n"
	if err := os.WriteFile(path, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
}

func writeFakeExecutableWithExit(t *testing.T, path string, stdout string, stderr string, exitCode int) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("shell script fake is unix-only")
	}
	body := "#!/bin/sh\n" +
		"printf '%s' " + shellQuote(stdout) + "\n" +
		"printf '%s' " + shellQuote(stderr) + " >&2\n" +
		"exit " + strconv.Itoa(exitCode) + "\n"
	if err := os.WriteFile(path, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
}
