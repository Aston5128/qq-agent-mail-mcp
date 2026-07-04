package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeEnvFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", p, err)
	}
	return p
}

func TestLoadEnvFile_MissingFileIsNoop(t *testing.T) {
	t.Setenv(EnvFileKey, filepath.Join(t.TempDir(), "does-not-exist.env"))
	if err := LoadEnvFile(); err != nil {
		t.Fatalf("missing file should be a no-op, got %v", err)
	}
}

func TestLoadEnvFile_LoadsValuesFromCustomPath(t *testing.T) {
	const key = "QQ_TEST_LOADENV_PROBE"
	dir := t.TempDir()
	writeEnvFile(t, dir, ".env", key+"=from-file\n")
	t.Setenv(EnvFileKey, filepath.Join(dir, ".env"))
	os.Unsetenv(key)
	t.Cleanup(func() { os.Unsetenv(key) })

	if err := LoadEnvFile(); err != nil {
		t.Fatalf("LoadEnvFile: %v", err)
	}
	if got := os.Getenv(key); got != "from-file" {
		t.Fatalf("got %q, want %q", got, "from-file")
	}
}

func TestLoadEnvFile_DoesNotOverrideExistingEnv(t *testing.T) {
	const key = "QQ_TEST_LOADENV_OVERRIDE_PROBE"
	dir := t.TempDir()
	writeEnvFile(t, dir, ".env", key+"=from-file\n")
	t.Setenv(EnvFileKey, filepath.Join(dir, ".env"))
	t.Setenv(key, "from-real-env") // real environment must win

	if err := LoadEnvFile(); err != nil {
		t.Fatalf("LoadEnvFile: %v", err)
	}
	if got := os.Getenv(key); got != "from-real-env" {
		t.Fatalf("real env should win: got %q, want %q", got, "from-real-env")
	}
}

func TestLoadEnvFile_DefaultsToCwdDotEnv(t *testing.T) {
	const key = "QQ_TEST_LOADENV_CWD_PROBE"
	dir := t.TempDir()
	writeEnvFile(t, dir, ".env", key+"=cwd-value\n")
	t.Setenv(EnvFileKey, "") // fall back to default ./.env
	t.Chdir(dir)             // cwd → temp dir, so ".env" resolves there
	os.Unsetenv(key)
	t.Cleanup(func() { os.Unsetenv(key) })

	if err := LoadEnvFile(); err != nil {
		t.Fatalf("LoadEnvFile: %v", err)
	}
	if got := os.Getenv(key); got != "cwd-value" {
		t.Fatalf("got %q, want %q", got, "cwd-value")
	}
}
