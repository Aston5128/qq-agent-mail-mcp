# QQ Agent Mail MCP

**English** | [简体中文](docs/readme/README.zh-CN.md)

A lightweight MCP server for QQ Agent Mail, backed by Tencent QQ Mail's
`agently-cli`.

This project is designed as a thin, structured pass-through layer. It exposes
the useful `agently-cli` mail operations as MCP tools while keeping CLI
installation and OAuth state outside the client agent process.

## Status

0.0.3 PoC implemented.

Implemented:

- Go StreamableHTTP MCP server;
- structured 1:1 MCP tool pass-through to `agently-cli`;
- CLI stdout JSON returned as MCP structured content;
- nonzero `agently-cli` exits return stdout JSON as structured tool errors;
- local fake `agently-cli` integration test;
- container deployment (`Dockerfile` + `docker-compose.yml`) with `agently-cli`
  OAuth credential persistence verified on a real host (token survives
  `docker compose down` / `up`).

Not implemented yet:

- MCP bearer token authentication;
- `agently-cli` OAuth management commands;
- stdio transport.

## Why this project exists

Installing `agently-cli` directly inside every agent container is fragile. A
container rebuild can remove both the CLI and its authorization state.

This project separates those concerns:

```text
Agent runtime
  -> MCP tools
     -> QQ Agent Mail MCP
        -> agently-cli
           -> QQ Agent Mail
```

The agent runtime uses MCP. This server owns the CLI binary, OAuth state,
message operations, and attachment downloads.

## Design principles

- Thin wrapper: map MCP tools to `agently-cli` commands with minimal business
  logic.
- Structured pass-through: preserve the CLI's JSON output shape where possible.
- No arbitrary command execution: expose named tools, not a raw `args` runner.
- Small memory footprint: prefer a simple implementation and avoid background
  workers unless required.
- Explicit send confirmation: keep `agently-cli`'s two-phase send flow visible.
- Portable deployment: support container sidecar first, with stdio fallback if
  required by the client runtime.

## Implemented tools

The first version mirrors stable CLI operations:

| MCP tool | Backing command |
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

The tool names intentionally stay close to the CLI. This keeps the server easy
to audit and avoids embedding application-specific workflows into a general
mail bridge.

## Quick start

Run local tests:

```bash
go test ./...
```

Build a binary. Version is injected from the latest git tag — `make build` does this automatically:

```bash
make build
# or without make:
go build -ldflags "-X main.version=$(git describe --tags --always --dirty 2>/dev/null || echo 0.0.3)" \
  -o bin/qq-agent-mail-mcp ./cmd/qq-agent-mail-mcp
```

Start the StreamableHTTP server:

```bash
QQ_AGENTLY_CLI_BIN=agently-cli \
QQ_AGENT_MAIL_MCP_BIND=127.0.0.1:8765 \
go run ./cmd/qq-agent-mail-mcp
```

Common environment variables:

| Variable | Default | Description |
| --- | --- | --- |
| `QQ_AGENTLY_CLI_BIN` | `agently-cli` | Path to the `agently-cli` executable. |
| `QQ_AGENT_MAIL_MCP_BIND` | `127.0.0.1:8765` | StreamableHTTP bind address. |
| `QQ_AGENT_MAIL_MCP_ENV_FILE` | `./.env` | Path to a `.env` file loaded at startup (see below). |

Environment variables may also be supplied via a `.env` file. At startup the server loads `.env` from the working directory (or the path in `QQ_AGENT_MAIL_MCP_ENV_FILE`), without overriding variables already present in the environment — so real environment values always win. A missing file is silently ignored. See [`.env.example`](.env.example) for a template; copy it to `.env` for local runs (`.env` is gitignored).

## Pass-through boundary

The server should pass through supported command parameters, not arbitrary shell
arguments.

Good:

```json
{
  "tool": "agently_message_list",
  "arguments": {
    "limit": 10,
    "dir": "inbox",
    "has_attachments": true
  }
}
```

Avoid:

```json
{
  "tool": "agently_cli",
  "arguments": {
    "args": ["message", "+list", "--limit", "10"]
  }
}
```

The structured form keeps the MCP interface safe, discoverable, and compatible
with agent tool schemas.

## Error responses

If `agently-cli` exits nonzero, the server distinguishes two cases:

- JSON stdout is parsed and returned at `structuredContent.error.stdout`;
- non-JSON stdout is returned as sanitized `stdout_text` together with `stderr`.

CLI-level errors such as unauthorized, not found, or invalid arguments are
returned as MCP tool results with `isError=true`, not as MCP protocol errors.
This lets the caller inspect structured fields:

```json
{
  "ok": false,
  "error": {
    "type": "agently_cli_error",
    "code": "auth",
    "message": "login required",
    "exit_code": 7,
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

## Send behavior

`agently-cli message +send` uses a two-step confirmation flow:

1. The first call returns a confirmation token and message summary.
2. The caller sends again with `--confirmation-token` to complete delivery.

The MCP server should preserve this behavior:

- if no confirmation token is supplied, return the confirmation response;
- if a confirmation token is supplied, complete the send request;
- do not add hidden auto-confirm behavior in the generic tool.

Application-specific allowlists can be built by the caller, or by future
specialized tools layered on top of the generic pass-through tools.

## Authentication model

`agently-cli auth login` performs OAuth authorization and stores credentials in
the system credential store.

The MCP server should not implement QQ Mail OAuth itself. Account binding is a
local administration step:

```bash
qq-agent-mail-mcp auth login
qq-agent-mail-mcp auth status
qq-agent-mail-mcp auth refresh
```

Those commands are expected to call the corresponding `agently-cli` operations
under the hood. They are management commands, not MCP tools exposed to agents.

The recommended model is one MCP server instance bound to one QQ Agent Mail
account. Run multiple instances if multiple accounts are needed.

On startup, the server should call `agently-cli +me` and verify that the CLI is
authorized. If `QQ_AGENT_MAIL_EXPECTED_ACCOUNT` is configured, the server should
also verify that the authorized account matches it and fail fast if it does not.

Credential persistence was validated on a real Linux host with the included
`docker-compose.yml`.

`agently-cli` stores its real OAuth token in a **file-backed keychain**, not in
the system keychain (the slim runtime image has none) and not in
`AGENTLY_CLI_CONFIG_DIR` (that holds only `config.json`). The token lives in:

```text
$HOME/.local/share/agently-cli/master.key          # encryption key
$HOME/.local/share/agently-cli/bootstrap_token.enc # encrypted OAuth token
```

Persistence therefore requires mounting **that directory** as a volume. The
compose file mounts `agently-keyring` at `/home/app/.local/share/agently-cli`;
with it, `agently-cli +me` still succeeds after `docker compose down` / `up`.

To re-validate persistence after any deployment change:

```bash
agently-cli auth login
docker compose down && docker compose up -d
agently-cli +me            # still authorized without re-login => persisted
```

## MCP access control

There are two separate trust boundaries:

```text
agently-cli -> QQ Agent Mail
MCP client  -> QQ Agent Mail MCP
```

`agently-cli` handles the first boundary. This server should handle the second
boundary when it listens on a network socket.

Recommended modes:

| Mode | Use case | Behavior |
| --- | --- | --- |
| `stdio` | Agent launches the MCP command locally | No network listener; no extra MCP auth. |
| `none` | Localhost-only development or isolated test network | Accept requests without an MCP auth token. |
| `bearer` | HTTP/SSE or any remote MCP transport | Require a bearer token on every request. |

Suggested defaults:

- stdio transport: `QQ_AGENT_MAIL_MCP_AUTH_MODE=stdio`;
- HTTP/SSE transport: `QQ_AGENT_MAIL_MCP_AUTH_MODE=bearer`;
- `none` should be explicit and used only for local testing.

Example:

```env
QQ_AGENT_MAIL_MCP_TRANSPORT=http
QQ_AGENT_MAIL_MCP_BIND=127.0.0.1:8765
QQ_AGENT_MAIL_MCP_AUTH_MODE=bearer
QQ_AGENT_MAIL_MCP_AUTH_TOKEN_FILE=/run/secrets/qq-agent-mail-mcp-token
QQ_AGENT_MAIL_EXPECTED_ACCOUNT=example@agent.qq.com
```

## Deployment shapes

### Sidecar MCP server

Preferred when the agent runtime can connect to a remote MCP server:

```text
agent-container
  -> remote MCP

qq-agent-mail-mcp
  -> agently-cli
  -> persistent credential volume
```

### Local stdio MCP command

Fallback when the agent runtime only supports launching local MCP commands:

```text
agent-container
  -> stdio MCP command
     -> agently-cli
     -> mounted credential volume
```

The sidecar shape keeps the agent container smaller. The stdio shape is more
compatible with runtimes that do not support remote MCP connections.

## Recommended host layout

Keep the MCP server separate from any application-specific agent directory.

Recommended layout:

```text
/opt/qq-agent-mail-mcp/        # application files, binary, compose files
/etc/qq-agent-mail-mcp/        # environment files and local config
/var/lib/qq-agent-mail-mcp/    # persisted auth state and runtime data
/var/cache/qq-agent-mail-mcp/  # temporary attachment downloads and cache
```

If another agent is installed under `/opt/<agent-name>`, this project should
usually live next to it rather than inside it:

```text
/opt/<agent-name>
/opt/qq-agent-mail-mcp
```

That keeps the mail bridge reusable and avoids coupling its OAuth state or
deployment lifecycle to a specific agent application.

## Security notes

- Do not expose a raw shell or arbitrary CLI argument runner.
- Do not log OAuth tokens, refresh tokens, message bodies, or attachment
  contents.
- Download attachments into a controlled working directory.
- Reject path traversal in output paths.
- Return structured errors, but avoid dumping raw CLI output if it may contain
  sensitive data.
- Keep application-specific mail policies outside the generic tools unless they
  are added as explicitly named specialized tools.

## Implementation language

The server will be implemented in Go.

Reasons for choosing Go:

- small memory footprint;
- simple static binary deployment;
- good fit for a thin CLI wrapper;
- easy to run as a small sidecar service.

TypeScript was considered because MCP's TypeScript SDK is mature and
`agently-cli` is distributed through npm, but this project prioritizes a small
runtime footprint and a simple deployable binary.

## Development environment

Required local tools:

- Go, current stable release recommended;
- `agently-cli`, installed only for smoke tests and runtime integration;
- no local Docker requirement.

On macOS with Homebrew:

```bash
brew install go
go version
```

Keep the development machine clean. Docker-based or Linux-container credential
persistence tests should run on the target server or another disposable Linux
environment, not on the local development machine.

The first implementation should include a deployment-host smoke test for
`agently-cli` credential persistence before relying on sidecar deployment.

## Repository hygiene

Public repository files should describe only the generic MCP bridge.

Keep local application notes, private workflow details, credentials, downloaded
attachments, and smoke-test artifacts out of git.

See `.gitignore` for ignored local paths.
