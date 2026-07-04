# Deployment

[简体中文](deployment.zh-CN.md)

Docker Compose is the recommended deployment path for now. The preferred path is
to pull the published GitHub Container Registry image with
`docker-compose.ghcr.yml`. If you want the deployment host to build from source,
use the default `docker-compose.yml`.

Because the current server has no built-in authentication, it is recommended to
deploy it on the same host as Hermes or inside another trusted private
environment.

The image packages the Go MCP server and `agently-cli` into one small runtime
container.

## Start from the published image

```bash
docker compose -f docker-compose.ghcr.yml pull
docker compose -f docker-compose.ghcr.yml up -d
```

Then authorize the CLI:

```bash
docker compose -f docker-compose.ghcr.yml exec qq-agent-mail-mcp agently-cli auth login
docker compose -f docker-compose.ghcr.yml exec qq-agent-mail-mcp agently-cli +me
```

To pin a specific image tag, set:

```env
QQ_AGENT_MAIL_MCP_VERSION=v0.0.5
```

## Build from source

For a traceable local image tag and binary version, use:

```bash
make compose-build
```

Or inject the version manually:

```bash
QQ_AGENT_MAIL_MCP_BUILD_VERSION="$(git describe --tags --always --dirty 2>/dev/null || echo dev)" \
  docker compose up -d --build
```

If you run plain `docker compose up -d --build`, the local image tag and binary
version default to `dev`.

Then authorize the CLI:

```bash
docker compose exec qq-agent-mail-mcp agently-cli auth login
docker compose exec qq-agent-mail-mcp agently-cli +me
```

Default endpoint:

```text
http://<host>:8765
```

## Ports

The compose file publishes:

```yaml
ports:
  - "8765:8765"
```

To make the service reachable only from the host, change it to:

```yaml
ports:
  - "127.0.0.1:8765:8765"
```

## Volumes

Two volumes are mounted:

```yaml
volumes:
  - agently-data:/var/lib/agently
  - agently-keyring:/home/app/.local/share/agently-cli
```

`agently-data` stores `config.json`. `agently-keyring` stores the encrypted OAuth
token and key material. Both are required for reliable credential persistence.

See [Authentication](../authentication/authentication.md) for details.

## Environment

The server reads normal environment variables and may also load a `.env` file.
Real environment variables take precedence over `.env` values.

Common variables:

| Variable | Default | Description |
| --- | --- | --- |
| `QQ_AGENT_MAIL_MCP_BIND` | `127.0.0.1:8765` | StreamableHTTP bind address. |
| `QQ_AGENTLY_CLI_BIN` | `agently-cli` | Path to `agently-cli`. |
| `QQ_AGENT_MAIL_MCP_ENV_FILE` | `./.env` | Optional env file path. |

Inside Docker, `QQ_AGENT_MAIL_MCP_BIND` is set to `0.0.0.0:8765`.

## Recommended host layout

For a manual host deployment:

```text
/opt/qq-agent-mail-mcp/        # app files, binary, compose file
/etc/qq-agent-mail-mcp/        # environment files and local config
/var/lib/qq-agent-mail-mcp/    # persisted auth/runtime data
/var/cache/qq-agent-mail-mcp/  # temporary downloads/cache
```

If another agent is installed under `/opt/<agent-name>`, keep this service as a
sibling directory rather than putting it inside that agent's application tree.

## Upgrade

For the published image:

```bash
docker compose -f docker-compose.ghcr.yml pull
docker compose -f docker-compose.ghcr.yml up -d
docker compose -f docker-compose.ghcr.yml exec qq-agent-mail-mcp agently-cli +me
```

For a local source build:

```bash
docker compose build --pull
docker compose up -d
docker compose exec qq-agent-mail-mcp agently-cli +me
```

If `+me` still succeeds, credential persistence survived the upgrade.
