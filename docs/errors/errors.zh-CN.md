# 错误返回

[English](errors.md)

server 会区分 MCP 协议级错误和 CLI 级 tool error。

## CLI 级错误

如果 `agently-cli` 以非零状态退出，MCP 调用会返回：

```json
{
  "isError": true
}
```

这是刻意设计。未授权、未找到、参数错误这类 CLI 级失败应该被 Agent 看见，方便它自我修正。

## 结构化错误形状

如果 stdout 是 JSON，server 会解析并保留到 `structuredContent.error.stdout`。

鉴权失败示例：

```json
{
  "ok": false,
  "error": {
    "type": "agently_cli_error",
    "tool": "agently_me",
    "code": "auth",
    "message": "login required",
    "exit_code": 7,
    "stderr": "please login",
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

## code 归一化

server 会从这些 CLI 输出形状归一化顶层 `error.code`：

1. `stdout.error.code`
2. `stdout.error.type`
3. 顶层 `stdout.code`

真实未登录的 `agently-cli` 形状使用 `stdout.error.type:"auth"`；server 会把它提升成顶层 `error.code:"auth"`。

## 非 JSON stdout

如果 stdout 不是 JSON，server 会返回脱敏后的文本：

```json
{
  "ok": false,
  "error": {
    "type": "agently_cli_error",
    "message": "...",
    "exit_code": 1,
    "stderr": "...",
    "stdout_text": "..."
  }
}
```

## MCP 协议级错误

server 内部异常、transport 错误或 MCP 层无效调用仍可能作为协议级错误返回。结构化 tool error 只用于 `agently-cli` 报告的错误。
