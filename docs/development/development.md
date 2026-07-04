# Development

[简体中文](development.zh-CN.md)

## Requirements

- Go 1.26 or newer.
- Docker only if you want to build/run the container image locally.
- `agently-cli` only for real integration smoke tests.

The local development machine does not need Docker for normal Go tests.

## Common commands

```bash
go test ./...
go vet ./...
make build
make run
make compose-build
```

Build without make:

```bash
go build -ldflags "-X main.version=$(git describe --tags --always --dirty 2>/dev/null || echo dev)" \
  -o bin/qq-agent-mail-mcp ./cmd/qq-agent-mail-mcp
```

## Tests

The server package includes a StreamableHTTP integration test using
`httptest`. It needs permission to listen on a temporary local port.

## Versioning

The binary version is injected through:

```text
-ldflags "-X main.version=<version>"
```

`make build` uses the latest git tag when available and falls back to `dev`.
Release/GHCR builds inject the release tag through the Docker build argument
`VERSION`. Do not manually maintain a release version in `main.go`.

Plain `go build ./cmd/qq-agent-mail-mcp` and plain `docker compose up -d --build`
produce a `dev` binary unless you inject a version. Use `make build`,
`make compose-build`, or the GHCR release workflow when you need traceable
versions.

For manual source-build compose deployments:

```bash
QQ_AGENT_MAIL_MCP_BUILD_VERSION="$(git describe --tags --always --dirty 2>/dev/null || echo dev)" \
  docker compose up -d --build
```

## Docker image

```bash
docker compose build
docker compose up -d
```

The runtime image uses:

- Debian slim / glibc, not Alpine;
- Node.js for the npm-distributed `agently-cli` wrapper;
- a non-root `app` user;
- `tini` as PID 1.

## Git commit requirements

Keep private application notes, local workflows, credentials, downloaded
attachments, and smoke-test artifacts out of git.

Use `private/` or another ignored local path for project-specific notes.
