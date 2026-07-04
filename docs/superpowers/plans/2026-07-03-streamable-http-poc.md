# StreamableHTTP PoC 实现规划

> **给 agentic worker：** 必备子技能：使用 superpowers:executing-plans 按任务逐步实现本规划。步骤使用复选框（`- [ ]`）语法跟踪进度。

**目标：** 构建一个 0.1 概念验证（PoC），通过远程 StreamableHTTP MCP Server 暴露 QQ Agent Mail 能力，并把支持的调用转发给 `agently-cli`。

**架构：** 使用官方 Go MCP SDK 来实现 MCP Server。保持封装层很薄：类型化的 MCP 工具输入会被转换成安全的 `agently-cli` argv 列表，CLI 的 stdout JSON 作为结构化工具输出返回，CLI 的 stderr 会做脱敏处理。本 PoC 不实现 OAuth 存储、Docker 部署和 git 初始化。

**技术栈：** Go 1.26、`github.com/modelcontextprotocol/go-sdk`、`agently-cli@latest`、标准库 `os/exec`、`encoding/json`、`net/http`、`httptest`。

---

## 0.1 工具集

暴露以下 MCP 工具：

| MCP 工具 | CLI 命令 |
| --- | --- |
| `agently_me` | `agently-cli +me` |
| `agently_message_list` | `agently-cli message +list` |
| `agently_message_read` | `agently-cli message +read` |
| `agently_message_search` | `agently-cli message +search` |
| `agently_message_send` | `agently-cli message +send` |
| `agently_message_reply` | `agently-cli message +reply` |
| `agently_message_forward` | `agently-cli message +forward` |
| `agently_attachment_upload` | `agently-cli attachment +upload` |
| `agently_attachment_download` | `agently-cli attachment +download` |

不暴露以下命令为 MCP 工具：

| 命令 | 原因 |
| --- | --- |
| `agently-cli auth login` | 管理操作；不应允许 Agent 调用。 |
| `agently-cli auth refresh` | 管理操作；不应允许 Agent 调用。 |
| `agently-cli help` | 作为 MCP 工具没有意义。 |

## 文件清单

- 新建：`go.mod`
- 新建：`cmd/qq-agent-mail-mcp/main.go`
- 新建：`internal/agently/runner.go`
- 新建：`internal/agently/runner_test.go`
- 新建：`internal/agently/tools.go`
- 新建：`internal/agently/tools_test.go`
- 新建：`internal/server/server.go`
- 新建：`internal/server/server_test.go`

## 任务 1：Go module 和 CLI 执行器

**文件：**
- 新建：`go.mod`
- 新建：`internal/agently/runner.go`
- 新建：`internal/agently/runner_test.go`

- [x] **步骤 1：初始化 Go module**

运行：

```bash
go mod init github.com/Aston5128/qq-agent-mail-mcp
go get github.com/modelcontextprotocol/go-sdk/mcp@latest
```

预期：

```text
go.mod 存在
go.sum 存在
```

- [x] **步骤 2：为命令执行写失败测试**

创建 `internal/agently/runner_test.go`：

```go
package agently

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
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

func writeFakeExecutable(t *testing.T, path string, stdout string) {
	t.Helper()
	body := "#!/bin/sh\nprintf '%s' " + shellQuote(stdout) + "\n"
	if runtime.GOOS == "windows" {
		t.Skip("shell script fake is unix-only")
	}
	if err := os.WriteFile(path, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
}
```

- [x] **步骤 3：运行测试并确认失败**

运行：

```bash
go test ./internal/agently
```

预期失败包含未定义符号，例如：

```text
undefined: Runner
undefined: shellQuote
```

- [x] **步骤 4：实现执行器**

创建 `internal/agently/runner.go`：

```go
package agently

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"time"
)

type Runner struct {
	Binary  string
	Timeout time.Duration
}

func (r Runner) Run(ctx context.Context, args []string) (map[string]any, error) {
	binary := r.Binary
	if binary == "" {
		binary = "agently-cli"
	}
	timeout := r.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, binary, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, fmt.Errorf("agently-cli failed: %w: %s", err, sanitize(stderr.String()))
	}

	var out map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		return nil, fmt.Errorf("agently-cli returned non-JSON output: %w", err)
	}
	return out, nil
}

func sanitize(value string) string {
	if len(value) > 512 {
		value = value[:512] + "..."
	}
	return value
}

func shellQuote(value string) string {
	return strconv.Quote(value)
}

var ErrInvalidArguments = errors.New("invalid arguments")
```

- [x] **步骤 5：运行测试并确认通过**

运行：

```bash
go test ./internal/agently
```

预期：

```text
ok  	github.com/Aston5128/qq-agent-mail-mcp/internal/agently
```

## 任务 2：1:1 工具的参数构造器

**文件：**
- 新建：`internal/agently/tools.go`
- 新建：`internal/agently/tools_test.go`

- [x] **步骤 1：为代表性的 argv 构造器写失败测试**

创建 `internal/agently/tools_test.go`：

```go
package agently

import (
	"reflect"
	"testing"
)

func TestMessageListArgs(t *testing.T) {
	got := MessageListInput{Limit: 10, Dir: "inbox", HasAttachments: true}.Args()
	want := []string{"message", "+list", "--limit", "10", "--dir", "inbox", "--has-attachments"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("args = %#v, want %#v", got, want)
	}
}

func TestMessageSendArgs(t *testing.T) {
	got := MessageSendInput{
		To: []string{"a@example.com", "b@example.com"},
		Subject: "Hi",
		Body: "Hello",
		ConfirmationToken: "tok_123",
	}.Args()
	want := []string{
		"message", "+send",
		"--to", "a@example.com",
		"--to", "b@example.com",
		"--subject", "Hi",
		"--body", "Hello",
		"--confirmation-token", "tok_123",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("args = %#v, want %#v", got, want)
	}
}

func TestAttachmentDownloadArgs(t *testing.T) {
	got := AttachmentDownloadInput{MessageID: "msg_1", AttachmentID: "att_1", Output: "downloads"}.Args()
	want := []string{"attachment", "+download", "--msg", "msg_1", "--att", "att_1", "--output", "downloads"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("args = %#v, want %#v", got, want)
	}
}
```

- [x] **步骤 2：运行测试并确认失败**

运行：

```bash
go test ./internal/agently
```

预期失败包含未定义的输入类型。

- [x] **步骤 3：实现输入结构体和 Args 方法**

创建 `internal/agently/tools.go`，为所有 0.1 工具定义结构体：

```go
package agently

import "strconv"

type MessageListInput struct {
	Limit          int    `json:"limit,omitempty" jsonschema:"messages per page, max 50"`
	Cursor         string `json:"cursor,omitempty" jsonschema:"pagination cursor"`
	Dir            string `json:"dir,omitempty" jsonschema:"folder: inbox, sent, trash, spam"`
	After          string `json:"after,omitempty" jsonschema:"only messages after this ISO 8601 time"`
	Before         string `json:"before,omitempty" jsonschema:"only messages before this ISO 8601 time"`
	IsUnread       bool   `json:"is_unread,omitempty" jsonschema:"only show unread messages"`
	HasAttachments bool   `json:"has_attachments,omitempty" jsonschema:"only show messages with attachments"`
}

func (in MessageListInput) Args() []string {
	args := []string{"message", "+list"}
	addCommonListFlags(&args, in.Limit, in.Cursor, in.Dir, in.After, in.Before, in.IsUnread, in.HasAttachments)
	return args
}

type MessageReadInput struct {
	ID string `json:"id" jsonschema:"message_id with msg_ prefix"`
}

func (in MessageReadInput) Args() []string {
	return []string{"message", "+read", "--id", in.ID}
}

type MessageSearchInput struct {
	Query          string `json:"q,omitempty" jsonschema:"search keyword or phrase"`
	SearchIn       string `json:"search_in,omitempty" jsonschema:"SEARCH_IN_ALL, SEARCH_IN_SUBJECT, or SEARCH_IN_CONTENT"`
	From           string `json:"from,omitempty" jsonschema:"sender email filter"`
	To             string `json:"to,omitempty" jsonschema:"recipient email filter"`
	Limit          int    `json:"limit,omitempty" jsonschema:"results per page, max 50"`
	Cursor         string `json:"cursor,omitempty" jsonschema:"pagination cursor"`
	Dir            string `json:"dir,omitempty" jsonschema:"folder: inbox, sent, trash, spam"`
	After          string `json:"after,omitempty" jsonschema:"only messages after this ISO 8601 time"`
	Before         string `json:"before,omitempty" jsonschema:"only messages before this ISO 8601 time"`
	IsUnread       bool   `json:"is_unread,omitempty" jsonschema:"only show unread messages"`
	HasAttachments bool   `json:"has_attachments,omitempty" jsonschema:"only show messages with attachments"`
}

func (in MessageSearchInput) Args() []string {
	args := []string{"message", "+search"}
	addString(&args, "--q", in.Query)
	addString(&args, "--search-in", in.SearchIn)
	addString(&args, "--from", in.From)
	addString(&args, "--to", in.To)
	addCommonListFlags(&args, in.Limit, in.Cursor, in.Dir, in.After, in.Before, in.IsUnread, in.HasAttachments)
	return args
}

type MessageSendInput struct {
	To                []string `json:"to" jsonschema:"recipient email addresses"`
	Cc                []string `json:"cc,omitempty" jsonschema:"CC recipient email addresses"`
	Bcc               []string `json:"bcc,omitempty" jsonschema:"BCC recipient email addresses"`
	Subject           string   `json:"subject" jsonschema:"email subject"`
	Body              string   `json:"body,omitempty" jsonschema:"email body"`
	BodyFile          string   `json:"body_file,omitempty" jsonschema:"relative UTF-8 body file path"`
	BodyFormat        string   `json:"body_format,omitempty" jsonschema:"body format, e.g. plain"`
	Attachment         []string `json:"attachment,omitempty" jsonschema:"relative attachment file paths"`
	ConfirmationToken string   `json:"confirmation_token,omitempty" jsonschema:"token returned by first send call"`
}

func (in MessageSendInput) Args() []string {
	args := []string{"message", "+send"}
	addRepeated(&args, "--to", in.To)
	addRepeated(&args, "--cc", in.Cc)
	addRepeated(&args, "--bcc", in.Bcc)
	addString(&args, "--subject", in.Subject)
	addString(&args, "--body", in.Body)
	addString(&args, "--body-file", in.BodyFile)
	addString(&args, "--body-format", in.BodyFormat)
	addRepeated(&args, "--attachment", in.Attachment)
	addString(&args, "--confirmation-token", in.ConfirmationToken)
	return args
}

type MessageReplyInput struct {
	ID                string   `json:"id" jsonschema:"message_id to reply to"`
	Body              string   `json:"body,omitempty" jsonschema:"reply body"`
	BodyFile          string   `json:"body_file,omitempty" jsonschema:"relative UTF-8 body file path"`
	BodyFormat        string   `json:"body_format,omitempty" jsonschema:"body format, e.g. plain"`
	Cc                []string `json:"cc,omitempty" jsonschema:"additional CC recipients"`
	Bcc               []string `json:"bcc,omitempty" jsonschema:"additional BCC recipients"`
	Attachment         []string `json:"attachment,omitempty" jsonschema:"relative attachment file paths"`
	ConfirmationToken string   `json:"confirmation_token,omitempty" jsonschema:"token returned by first reply call"`
	ReplyAll          bool     `json:"reply_all,omitempty" jsonschema:"reply to all recipients"`
}

func (in MessageReplyInput) Args() []string {
	args := []string{"message", "+reply", "--id", in.ID}
	addString(&args, "--body", in.Body)
	addString(&args, "--body-file", in.BodyFile)
	addString(&args, "--body-format", in.BodyFormat)
	addRepeated(&args, "--cc", in.Cc)
	addRepeated(&args, "--bcc", in.Bcc)
	addRepeated(&args, "--attachment", in.Attachment)
	addString(&args, "--confirmation-token", in.ConfirmationToken)
	addBool(&args, "--reply-all", in.ReplyAll)
	return args
}

type MessageForwardInput struct {
	ID                 string   `json:"id" jsonschema:"message_id to forward"`
	To                 []string `json:"to" jsonschema:"forward recipient email addresses"`
	Cc                 []string `json:"cc,omitempty" jsonschema:"CC recipient email addresses"`
	Bcc                []string `json:"bcc,omitempty" jsonschema:"BCC recipient email addresses"`
	Body               string   `json:"body,omitempty" jsonschema:"forwarding note"`
	BodyFile           string   `json:"body_file,omitempty" jsonschema:"relative UTF-8 body file path"`
	BodyFormat         string   `json:"body_format,omitempty" jsonschema:"body format, e.g. plain"`
	Attachment          []string `json:"attachment,omitempty" jsonschema:"additional relative attachment paths"`
	ConfirmationToken  string   `json:"confirmation_token,omitempty" jsonschema:"token returned by first forward call"`
	IncludeAttachments bool     `json:"include_attachments,omitempty" jsonschema:"include original message attachments"`
}

func (in MessageForwardInput) Args() []string {
	args := []string{"message", "+forward", "--id", in.ID}
	addRepeated(&args, "--to", in.To)
	addRepeated(&args, "--cc", in.Cc)
	addRepeated(&args, "--bcc", in.Bcc)
	addString(&args, "--body", in.Body)
	addString(&args, "--body-file", in.BodyFile)
	addString(&args, "--body-format", in.BodyFormat)
	addRepeated(&args, "--attachment", in.Attachment)
	addString(&args, "--confirmation-token", in.ConfirmationToken)
	addBool(&args, "--include-attachments", in.IncludeAttachments)
	return args
}

type AttachmentUploadInput struct {
	File string `json:"file" jsonschema:"relative local file path to upload"`
}

func (in AttachmentUploadInput) Args() []string {
	return []string{"attachment", "+upload", "--file", in.File}
}

type AttachmentDownloadInput struct {
	MessageID    string `json:"message_id" jsonschema:"message_id with msg_ prefix"`
	AttachmentID string `json:"attachment_id" jsonschema:"attachment_id with att_ prefix"`
	Output       string `json:"output,omitempty" jsonschema:"relative output directory"`
}

func (in AttachmentDownloadInput) Args() []string {
	args := []string{"attachment", "+download", "--msg", in.MessageID, "--att", in.AttachmentID}
	addString(&args, "--output", in.Output)
	return args
}

type EmptyInput struct{}

func addCommonListFlags(args *[]string, limit int, cursor, dir, after, before string, isUnread, hasAttachments bool) {
	if limit > 0 {
		*args = append(*args, "--limit", strconv.Itoa(limit))
	}
	addString(args, "--cursor", cursor)
	addString(args, "--dir", dir)
	addString(args, "--after", after)
	addString(args, "--before", before)
	addBool(args, "--is-unread", isUnread)
	addBool(args, "--has-attachments", hasAttachments)
}

func addString(args *[]string, flag string, value string) {
	if value != "" {
		*args = append(*args, flag, value)
	}
}

func addRepeated(args *[]string, flag string, values []string) {
	for _, value := range values {
		if value != "" {
			*args = append(*args, flag, value)
		}
	}
}

func addBool(args *[]string, flag string, value bool) {
	if value {
		*args = append(*args, flag)
	}
}
```

- [x] **步骤 4：运行测试并确认通过**

运行：

```bash
go test ./internal/agently
```

预期：

```text
ok  	github.com/Aston5128/qq-agent-mail-mcp/internal/agently
```

## 任务 3：MCP Server 注册

**文件：**
- 新建：`internal/server/server.go`
- 新建：`internal/server/server_test.go`

- [x] **步骤 1：写一个测试，验证所有预期的工具名都已注册**

创建 `internal/server/server_test.go`：

```go
package server

import "testing"

func TestToolNames(t *testing.T) {
	got := ToolNames()
	want := []string{
		"agently_me",
		"agently_message_list",
		"agently_message_read",
		"agently_message_search",
		"agently_message_send",
		"agently_message_reply",
		"agently_message_forward",
		"agently_attachment_upload",
		"agently_attachment_download",
	}
	if len(got) != len(want) {
		t.Fatalf("len(ToolNames()) = %d, want %d: %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ToolNames()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
```

- [x] **步骤 2：运行测试并确认失败**

运行：

```bash
go test ./internal/server
```

预期失败包含未定义的 `ToolNames`。

- [x] **步骤 3：实现 Server 构造和工具名**

创建 `internal/server/server.go`。先用 `go doc github.com/modelcontextprotocol/go-sdk/mcp` 查阅官方 Go SDK 的实际 StreamableHTTP API，再使用它。这个文件只负责 MCP Server 的构造和工具注册。

最小导出接口：

```go
package server

import "github.com/Aston5128/qq-agent-mail-mcp/internal/agently"

func ToolNames() []string {
	return []string{
		"agently_me",
		"agently_message_list",
		"agently_message_read",
		"agently_message_search",
		"agently_message_send",
		"agently_message_reply",
		"agently_message_forward",
		"agently_attachment_upload",
		"agently_attachment_download",
	}
}

type Config struct {
	Name    string
	Version string
	Runner  agently.Runner
}
```

- [x] **步骤 4：运行包测试**

运行：

```bash
go test ./...
```

预期所有测试通过。

## 任务 4：主命令和 StreamableHTTP PoC

**文件：**
- 新建：`cmd/qq-agent-mail-mcp/main.go`

- [x] **步骤 1：创建主命令**

创建 `cmd/qq-agent-mail-mcp/main.go`：

```go
package main

import (
	"context"
	"log"
	"os"

	"github.com/Aston5128/qq-agent-mail-mcp/internal/agently"
	"github.com/Aston5128/qq-agent-mail-mcp/internal/server"
)

func main() {
	bind := getenv("QQ_AGENT_MAIL_MCP_BIND", "127.0.0.1:8765")
	binary := getenv("QQ_AGENTLY_CLI_BIN", "agently-cli")

	cfg := server.Config{
		Name:    "qq-agent-mail-mcp",
		Version: "0.1.0",
		Runner:  agently.Runner{Binary: binary},
	}

	if err := server.RunStreamableHTTP(context.Background(), bind, cfg); err != nil {
		log.Fatal(err)
	}
}

func getenv(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
```

- [x] **步骤 2：用官方 SDK 实现 `server.RunStreamableHTTP`**

使用官方 SDK API，不要自己实现协议。如果 SDK API 名称与本规划不一致，按 SDK 的名称调整，但保持对外行为不变：

```text
监听 QQ_AGENT_MAIL_MCP_BIND。
暴露所有 ToolNames()。
每个工具调用 Runner.Run(ctx, typedInput.Args())。
把解析后的 JSON 作为结构化工具输出返回。
```

- [x] **步骤 3：运行构建**

运行：

```bash
go build ./cmd/qq-agent-mail-mcp
```

预期：

```text
构建成功
```

## 任务 5：无真实 QQ 授权的本地 PoC 验证

**文件：**
- 不需要新文件。

状态说明：已通过 `internal/server/streamable_http_test.go` 中的自动化 `httptest` StreamableHTTP 集成测试完成。该测试会创建临时的 fake `agently-cli`，启动 MCP handler，用官方 Go SDK 客户端连接，并调用 `agently_me`。

- [x] **步骤 1：在 `/private/tmp` 下创建 fake `agently-cli`**

运行：

```bash
mkdir -p /private/tmp/qq-agent-mail-mcp-poc
```

创建 `/private/tmp/qq-agent-mail-mcp-poc/agently-cli`：

```sh
#!/bin/sh
case "$*" in
  "+me")
    printf '{"ok":true,"data":{"email":"fake@example.com"}}'
    ;;
  "message +list"*)
    printf '{"ok":true,"data":{"data":[],"pagination":{"has_more":false}}}'
    ;;
  *)
    printf '{"ok":true,"argv":['
    first=1
    for arg in "$@"; do
      if [ "$first" = 0 ]; then printf ','; fi
      first=0
      printf '"%s"' "$arg"
    done
    printf ']}'
    ;;
esac
```

加上可执行权限：

```bash
chmod +x /private/tmp/qq-agent-mail-mcp-poc/agently-cli
```

- [x] **步骤 2：用 fake CLI 启动 Server**

运行：

```bash
QQ_AGENTLY_CLI_BIN=/private/tmp/qq-agent-mail-mcp-poc/agently-cli \
QQ_AGENT_MAIL_MCP_BIND=127.0.0.1:8765 \
go run ./cmd/qq-agent-mail-mcp
```

预期：

```text
server 监听 127.0.0.1:8765
```

- [x] **步骤 3：用 MCP 客户端连接**

用目标运行时里可用的任意 MCP 客户端调用：

```text
agently_me
agently_message_list
```

预期：

```text
工具调用返回来自 fake agently-cli 的假 JSON。
```

## 任务 6：部署宿主机烟测

**文件：**
- 不需要仓库文件。

- [ ] **步骤 1：把源码或构建出的二进制拷到部署宿主机**

从本机用 SSH/scp 传到部署宿主机。

- [ ] **步骤 2：在部署宿主机上安装或校验 `agently-cli@latest`**

在部署宿主机上运行：

```bash
npm install -g @tencent-qqmail/agently-cli@latest
agently-cli --version
```

- [ ] **步骤 3：用真实 `agently-cli` 启动 PoC Server**

在部署宿主机上运行：

```bash
QQ_AGENT_MAIL_MCP_BIND=127.0.0.1:8765 ./qq-agent-mail-mcp
```

- [ ] **步骤 4：从目标 Agent 运行时用 StreamableHTTP 连接**

调用：

```text
agently_me
agently_message_list
```

预期：

```text
目标 Agent 可以调用远程 MCP Server，并收到真实的 agently-cli JSON 输出。
```

## 自检

- 规格覆盖：覆盖了 Go、StreamableHTTP、官方 SDK、最新 agently-cli、无本地 Docker、无 git 初始化、1:1 业务命令透传，以及部署宿主机测试。
- 占位符扫描：没有遗留 `TBD` 或未指定的工具集；只有 SDK API 名称推迟到 `go doc` 阶段确定，因为它们必须匹配已安装的官方 SDK。
- 范围检查：OAuth 持久化和生产认证刻意排除在 0.1 PoC 之外。
