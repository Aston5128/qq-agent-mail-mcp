# Authentication and credential persistence

[简体中文](authentication.zh-CN.md)

QQ Agent Mail MCP does not implement QQ Mail OAuth itself. OAuth authorization
is owned by `agently-cli`.

## Bind an account

Run the login flow in the same runtime environment where the MCP server runs:

```bash
agently-cli auth login
agently-cli +me
```

With Docker Compose:

```bash
docker compose exec qq-agent-mail-mcp agently-cli auth login
docker compose exec qq-agent-mail-mcp agently-cli +me
```

The recommended model is one MCP server instance per QQ Agent Mail account. Run
multiple server instances if you need multiple accounts.

## What is persisted

In the verified container runtime, `agently-cli` persists two separate things:

```text
/var/lib/agently                         # config.json (AGENTLY_CLI_CONFIG_DIR)
/home/app/.local/share/agently-cli       # master.key + bootstrap_token.enc
```

The real OAuth token is stored in a file-backed keychain under
`$HOME/.local/share/agently-cli/`, not in `AGENTLY_CLI_CONFIG_DIR`.

The included compose file therefore mounts two named volumes:

```yaml
volumes:
  - agently-data:/var/lib/agently
  - agently-keyring:/home/app/.local/share/agently-cli
```

Both are required. If the keychain volume is missing, `agently-cli +me` may stop
working after a container recreate.

## Validate persistence

After deployment or volume changes, re-check persistence:

```bash
docker compose exec qq-agent-mail-mcp agently-cli auth login
docker compose exec qq-agent-mail-mcp agently-cli +me
docker compose down
docker compose up -d
docker compose exec qq-agent-mail-mcp agently-cli +me
```

If the final `+me` succeeds without logging in again, credentials are persisted.

## MCP authentication

There are two separate trust boundaries:

```text
agently-cli -> QQ Agent Mail
MCP client  -> QQ Agent Mail MCP
```

`agently-cli` handles the first boundary. MCP request authentication is not
implemented in the current server version. Until bearer-token authentication is
implemented, keep the server on localhost, a private network, or behind a
trusted reverse proxy / firewall.

## Not implemented yet

- `qq-agent-mail-mcp auth login`
- `qq-agent-mail-mcp auth status`
- `qq-agent-mail-mcp auth refresh`
- startup fail-fast account verification
- MCP bearer-token authentication
