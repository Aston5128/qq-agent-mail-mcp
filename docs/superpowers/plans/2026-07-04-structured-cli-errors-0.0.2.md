# 0.0.2 Structured CLI Error Iteration

## 背景

`agently-cli` 在未登录、未找到、参数错误等场景下可能会以非零状态退出，同时把结构化错误 JSON 输出到 stdout。0.0.1 的 runner 在非零退出路径只保留 stderr，导致上层 Agent 无法可靠区分错误类型。

真实未登录场景的错误鉴别字段是 `stdout.error.type:"auth"`，不是 `stdout.error.code:"UNAUTHORIZED"`。因此 0.0.2 的回归测试必须使用真实 `error.type` 形状，且归一化后的 MCP tool error 顶层 `error.code` 应该提升为 `"auth"`。

## 目标

- [x] 非零退出时解析 stdout JSON。
- [x] 暴露 `agently.CLIError`，保留 exit code、stderr、raw stdout 和 parsed stdout。
- [x] MCP 层把 CLI 级错误返回为 `isError=true` 的 tool result。
- [x] structured content 中包含标准化 `type/code/message/exit_code/stdout`。
- [x] `CLIError.Code()` 兼容 `stdout.error.code` 和真实未登录形状 `stdout.error.type`。
- [x] runner 单元测试和 StreamableHTTP 集成测试使用 `error.type:"auth"` 覆盖真实未登录场景。
- [x] 保持非 CLI 异常为协议级错误。
- [x] 增加 runner 单元测试和 StreamableHTTP 集成测试。

## 真实 auth 错误形状

输入示例：

```json
{
  "ok": false,
  "error": {
    "type": "auth",
    "message": "login required"
  }
}
```

MCP tool error 归一化后应该包含：

```json
{
  "ok": false,
  "error": {
    "type": "agently_cli_error",
    "code": "auth",
    "message": "login required",
    "stdout": {
      "ok": false,
      "error": {
        "type": "auth",
        "message": "login required"
      }
    }
  }
}
```

## 验证

```bash
go test ./internal/agently
go test ./internal/server
go test ./...
go vet ./...
go build -o bin/qq-agent-mail-mcp ./cmd/qq-agent-mail-mcp
```

其中 StreamableHTTP 测试需要允许本机临时监听端口。
