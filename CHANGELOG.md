# Changelog

> 中文版：[CHANGELOG.zh-CN.md](docs/changelogs/CHANGELOG.zh-CN.md)

## Unreleased

- Fix credential loss across container restarts: `agently-cli` stores its OAuth token in a file-backed keychain at `$HOME/.local/share/agently-cli/` (`master.key` + `bootstrap_token.enc`), not in `AGENTLY_CLI_CONFIG_DIR` (which holds only `config.json`). Add a second named volume (`agently-keyring`) at that path so the token survives `docker compose down` / `up`. Verified on a real host.

## 0.0.3

- Add container deployment: multi-stage Dockerfile (MCP server build + agently-cli runtime) and docker-compose.
- Mount `AGENTLY_CLI_CONFIG_DIR` as a named volume for `config.json`.
- Run as a non-root user with `tini` as PID 1 for clean signal forwarding and child-process reaping.
- Pin agently-cli to the verified 1.0.6 release (overridable via build arg).
- Add `.dockerignore`.

## 0.0.2

- Return nonzero `agently-cli` stdout JSON as structured MCP tool errors.
- Preserve CLI exit code, stderr, normalized code/message, and original stdout JSON.
- Normalize both `error.code` and the real auth shape `error.type` into top-level `error.code`.
- Add regression coverage for unauthorized-style CLI failures over StreamableHTTP.

## 0.0.1

- Initial StreamableHTTP PoC.
- Expose named MCP tools for QQ Agent Mail operations.
- Forward typed MCP inputs to `agently-cli` and return parsed stdout JSON.
