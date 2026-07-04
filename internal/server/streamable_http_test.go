package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/Aston5128/qq-agent-mail-mcp/internal/agently"
)

func TestStreamableHTTPToolCallForwardsToAgentlyCLI(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell script fake is unix-only")
	}

	dir := t.TempDir()
	fakeCLI := filepath.Join(dir, "agently-cli")
	writeFakeAgentlyCLI(t, fakeCLI, `{"ok":true,"data":{"message":"pong"}}`)

	mcpServer := New(Config{Runner: agently.Runner{Binary: fakeCLI}})
	handler := mcp.NewStreamableHTTPHandler(
		func(*http.Request) *mcp.Server { return mcpServer },
		&mcp.StreamableHTTPOptions{JSONResponse: true},
	)
	httpServer := httptest.NewServer(handler)
	t.Cleanup(httpServer.Close)

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0.0.0"}, nil)
	session, err := client.Connect(
		context.Background(),
		&mcp.StreamableClientTransport{Endpoint: httpServer.URL, DisableStandaloneSSE: true},
		nil,
	)
	if err != nil {
		t.Fatalf("Connect returned error: %v", err)
	}
	t.Cleanup(func() {
		if err := session.Close(); err != nil {
			t.Fatalf("session.Close returned error: %v", err)
		}
	})

	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: "agently_me"})
	if err != nil {
		t.Fatalf("CallTool returned error: %v", err)
	}
	if result.IsError {
		t.Fatalf("CallTool returned tool error: %#v", result.Content)
	}

	structured, ok := result.StructuredContent.(map[string]any)
	if !ok {
		t.Fatalf("structured content = %#v, want map[string]any", result.StructuredContent)
	}
	if structured["ok"] != true {
		t.Fatalf("ok = %#v, want true", structured["ok"])
	}
}

func TestStreamableHTTPToolCallReturnsStructuredCLIError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell script fake is unix-only")
	}

	dir := t.TempDir()
	fakeCLI := filepath.Join(dir, "agently-cli")
	writeFailingAgentlyCLI(
		t,
		fakeCLI,
		`{"ok":false,"error":{"type":"auth","message":"login required"}}`,
		"please login",
		7,
	)

	mcpServer := New(Config{Runner: agently.Runner{Binary: fakeCLI}})
	handler := mcp.NewStreamableHTTPHandler(
		func(*http.Request) *mcp.Server { return mcpServer },
		&mcp.StreamableHTTPOptions{JSONResponse: true},
	)
	httpServer := httptest.NewServer(handler)
	t.Cleanup(httpServer.Close)

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0.0.0"}, nil)
	session, err := client.Connect(
		context.Background(),
		&mcp.StreamableClientTransport{Endpoint: httpServer.URL, DisableStandaloneSSE: true},
		nil,
	)
	if err != nil {
		t.Fatalf("Connect returned error: %v", err)
	}
	t.Cleanup(func() {
		if err := session.Close(); err != nil {
			t.Fatalf("session.Close returned error: %v", err)
		}
	})

	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: "agently_me"})
	if err != nil {
		t.Fatalf("CallTool returned protocol error: %v", err)
	}
	if !result.IsError {
		t.Fatal("CallTool IsError = false, want true")
	}

	structured, ok := result.StructuredContent.(map[string]any)
	if !ok {
		t.Fatalf("structured content = %#v, want map[string]any", result.StructuredContent)
	}
	errorInfo, ok := structured["error"].(map[string]any)
	if !ok {
		t.Fatalf("structured error = %#v, want map[string]any", structured["error"])
	}
	if errorInfo["type"] != "agently_cli_error" {
		t.Fatalf("error type = %#v, want agently_cli_error", errorInfo["type"])
	}
	if errorInfo["code"] != "auth" {
		t.Fatalf("error code = %#v, want auth", errorInfo["code"])
	}
	if errorInfo["message"] != "login required" {
		t.Fatalf("error message = %#v, want login required", errorInfo["message"])
	}
	if errorInfo["exit_code"] != float64(7) {
		t.Fatalf("exit_code = %#v, want 7", errorInfo["exit_code"])
	}
	stdout, ok := errorInfo["stdout"].(map[string]any)
	if !ok {
		t.Fatalf("stdout = %#v, want map[string]any", errorInfo["stdout"])
	}
	if stdout["ok"] != false {
		t.Fatalf("stdout ok = %#v, want false", stdout["ok"])
	}
}

func writeFakeAgentlyCLI(t *testing.T, path string, stdout string) {
	t.Helper()

	body := "#!/bin/sh\nprintf '%s' " + strconv.Quote(stdout) + "\n"
	if err := os.WriteFile(path, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
}

func writeFailingAgentlyCLI(t *testing.T, path string, stdout string, stderr string, exitCode int) {
	t.Helper()

	body := "#!/bin/sh\n" +
		"printf '%s' " + strconv.Quote(stdout) + "\n" +
		"printf '%s' " + strconv.Quote(stderr) + " >&2\n" +
		"exit " + strconv.Itoa(exitCode) + "\n"
	if err := os.WriteFile(path, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
}
