# Error responses

[简体中文](errors.zh-CN.md)

The server distinguishes MCP protocol errors from CLI-level tool errors.

## CLI-level errors

If `agently-cli` exits with a nonzero status, the MCP call returns a tool result
with:

```json
{
  "isError": true
}
```

This is intentional. CLI-level failures such as unauthorized, not found, or
invalid arguments should be visible to the agent so it can self-correct.

## Structured error shape

When stdout is JSON, it is parsed and preserved under
`structuredContent.error.stdout`.

Example auth failure:

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

## Code normalization

The server normalizes the top-level `error.code` from these CLI output shapes:

1. `stdout.error.code`
2. `stdout.error.type`
3. top-level `stdout.code`

The real unauthenticated `agently-cli` shape uses `stdout.error.type:"auth"`;
the server promotes that to top-level `error.code:"auth"`.

## Non-JSON stdout

If stdout is not JSON, the server returns sanitized text:

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

## MCP protocol errors

Unexpected server failures, transport errors, or invalid MCP-level calls may
still be returned as protocol errors. The structured tool-error path is for
errors reported by `agently-cli`.
