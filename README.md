# QQ Agent Mail MCP

**English** | [简体中文](docs/readme/README.zh-CN.md)

A lightweight StreamableHTTP MCP server that exposes QQ Agent Mail capabilities
through [QQ Agent Mail](https://agent.qq.com)'s `agently-cli`.

It wraps existing `agently-cli` capabilities as structured MCP tools while
keeping the CLI installation and OAuth credential state outside the client agent
runtime.

## Features

- StreamableHTTP MCP server written in Go.
- Thin, structured pass-through to `agently-cli`.
- Parsed JSON stdout returned as MCP structured content.
- Nonzero CLI exits returned as structured MCP tool errors.
- Docker deployment with persisted `agently-cli` OAuth credentials.
- No raw shell or arbitrary CLI argument tool.

Current limitations:

- MCP bearer-token authentication is not implemented yet.
- stdio transport is not implemented yet.
- `agently-cli` OAuth management commands are not exposed by this server.
- Delete/trash operations are intentionally not exposed.

## Why this exists

This started from a Dockerized Hermes Agent deployment. Installing
`agently-cli` directly inside the Hermes container works, but every
`docker compose down` can lose that state. It is possible to patch the Hermes
`docker-compose.yaml`, but that felt less elegant, so this project provides a
separate MCP server that forwards mail operations.

This project separates those responsibilities:

```text
Agent runtime
  -> MCP tools
     -> QQ Agent Mail MCP
        -> agently-cli
           -> QQ Agent Mail
```

## Quick deployment with Docker Compose

Clone the latest code:

```bash
git clone https://github.com/Aston5128/qq-agent-mail-mcp.git
```

Build and start the server:

```bash
docker compose up -d --build
```

Authorize `agently-cli` inside the running container:

```bash
# This command prints a URL. Open it in any browser and authorize with WeChat.
docker compose exec qq-agent-mail-mcp agently-cli auth login
docker compose exec qq-agent-mail-mcp agently-cli +me
```

The default StreamableHTTP endpoint is:

```text
http://<host>:8765
```

The compose file persists both `agently-cli` config and OAuth keychain data, so
authorization survives `docker compose down` / `up`.

## MCP client configuration

Configure your MCP-capable agent to use a remote StreamableHTTP server at the
endpoint above.

Until MCP bearer-token authentication is implemented, do not expose this server
directly to an untrusted network. Bind it to localhost, place it on a private
network, or protect it with a reverse proxy / firewall.

## Tools

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

See [Tools](docs/tools/tools.md) for arguments and behavior notes.

## Local development

```bash
go test ./...
go vet ./...
make build
```

Run from source:

```bash
QQ_AGENTLY_CLI_BIN=agently-cli \
QQ_AGENT_MAIL_MCP_BIND=127.0.0.1:8765 \
go run ./cmd/qq-agent-mail-mcp
```

## Documentation

- [Deployment](docs/deployment/deployment.md)
- [Authentication and credential persistence](docs/authentication/authentication.md)
- [Tools](docs/tools/tools.md)
- [Error responses](docs/errors/errors.md)
- [Security](docs/security/security.md)
- [Design notes](docs/design/design.md)
- [Development](docs/development/development.md)
- [Changelog](CHANGELOG.md)
