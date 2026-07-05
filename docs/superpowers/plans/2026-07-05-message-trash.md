# Message Trash Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `agently_message_trash`, a typed MCP tool that forwards to `agently-cli message +trash`.

**Architecture:** Follow the existing typed pass-through pattern. Add one input struct in `internal/agently`, register one MCP tool in `internal/server`, and update docs to describe recoverable trash behavior instead of permanent deletion.

**Tech Stack:** Go 1.26.4, `github.com/modelcontextprotocol/go-sdk`, standard Go tests, Markdown docs.

---

### Task 1: Add MessageTrashInput With TDD

**Files:**
- Modify: `internal/agently/tools_test.go`
- Modify: `internal/agently/tools.go`

- [ ] **Step 1: Write the failing tests**

Add these tests to `internal/agently/tools_test.go`:

```go
func TestMessageTrashArgs(t *testing.T) {
	got := MessageTrashInput{ID: "msg_1"}.Args()
	want := []string{"message", "+trash", "--id", "msg_1"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("args = %#v, want %#v", got, want)
	}
}

func TestMessageTrashArgsWithConfirmationToken(t *testing.T) {
	got := MessageTrashInput{ID: "msg_1", ConfirmationToken: "tok_123"}.Args()
	want := []string{"message", "+trash", "--id", "msg_1", "--confirmation-token", "tok_123"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("args = %#v, want %#v", got, want)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
go test ./internal/agently
```

Expected: FAIL because `MessageTrashInput` is undefined.

- [ ] **Step 3: Write minimal implementation**

Add this type and method to `internal/agently/tools.go` near the other message input types:

```go
type MessageTrashInput struct {
	ID                string `json:"id" jsonschema:"message_id to move to trash"`
	ConfirmationToken string `json:"confirmation_token,omitempty" jsonschema:"token returned by first trash call"`
}

func (in MessageTrashInput) Args() []string {
	args := []string{"message", "+trash", "--id", in.ID}
	addString(&args, "--confirmation-token", in.ConfirmationToken)
	return args
}
```

- [ ] **Step 4: Run test to verify it passes**

Run:

```bash
go test ./internal/agently
```

Expected: PASS.

### Task 2: Register agently_message_trash With TDD

**Files:**
- Modify: `internal/server/streamable_http_test.go`
- Modify: `internal/server/server.go`

- [ ] **Step 1: Write the failing StreamableHTTP test**

Add this test to `internal/server/streamable_http_test.go`:

```go
func TestStreamableHTTPTrashToolForwardsToAgentlyCLI(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell script fake is unix-only")
	}

	dir := t.TempDir()
	fakeCLI := filepath.Join(dir, "agently-cli")
	writeFakeAgentlyCLI(t, fakeCLI, `{"ok":true,"data":{"trashed":true}}`)

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

	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "agently_message_trash",
		Arguments: map[string]any{"id": "msg_1"},
	})
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
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
go test ./internal/server
```

Expected: FAIL because `agently_message_trash` is not registered.

- [ ] **Step 3: Register the tool**

Add this line in `internal/server/server.go` after `agently_message_forward`:

```go
addForwardTool[agently.MessageTrashInput](s, cfg.Runner, "agently_message_trash", "Move a message to trash using agently-cli confirmation flow.")
```

- [ ] **Step 4: Run test to verify it passes**

Run:

```bash
go test ./internal/server
```

Expected: PASS.

### Task 3: Update Documentation

**Files:**
- Modify: `README.md`
- Modify: `docs/readme/README.zh-CN.md`
- Modify: `docs/tools/tools.md`
- Modify: `docs/tools/tools.zh-CN.md`
- Modify: `docs/security/security.md`
- Modify: `docs/security/security.zh-CN.md`
- Modify: `docs/design/design.md`
- Modify: `docs/design/design.zh-CN.md`
- Modify: `CHANGELOG.md`
- Modify: `docs/changelogs/CHANGELOG.zh-CN.md`

- [ ] **Step 1: Update tool lists**

Add `agently_message_trash` to the README and tools tables with backing command `agently-cli message +trash`.

- [ ] **Step 2: Replace old non-exposure wording**

Replace statements saying delete/trash is not exposed with wording that says `agently_message_trash` moves messages to trash, not permanent deletion, and QQ Agent Mail retains trash messages for 30 days.

- [ ] **Step 3: Update confirmation notes**

Mention that trash follows the CLI confirmation-token flow and the MCP server does not auto-confirm it.

- [ ] **Step 4: Add changelog entries**

Add an unreleased entry to English and Chinese changelogs describing `agently_message_trash`.

### Task 4: Verify And Review

**Files:**
- All modified files.

- [ ] **Step 1: Run full test suite**

Run:

```bash
go test ./...
```

Expected: PASS.

- [ ] **Step 2: Run vet**

Run:

```bash
go vet ./...
```

Expected: no output and exit code 0.

- [ ] **Step 3: Inspect final diff**

Run:

```bash
git diff --stat
git diff --check
git status --short
```

Expected: only intended code, docs, and plan changes; no whitespace errors.
