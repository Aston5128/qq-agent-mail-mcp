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
```

Build without make:

```bash
go build -ldflags "-X main.version=$(git describe --tags --always --dirty 2>/dev/null || echo 0.0.4)" \
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

`make build` uses the latest git tag when available and falls back to `0.0.4`.

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
