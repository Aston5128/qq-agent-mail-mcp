# syntax=docker/dockerfile:1

# ---- build stage: compile the MCP server binary ----
FROM golang:1.26-bookworm AS builder

WORKDIR /src

# Cache module downloads independently of source changes.
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

# version is injected via ldflags (same as Makefile); override with --build-arg.
ARG VERSION=0.0.3
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath \
        -ldflags "-X main.version=${VERSION}" \
        -o /out/qq-agent-mail-mcp ./cmd/qq-agent-mail-mcp

# ---- runtime stage: MCP server + agently-cli ----
# debian-slim (glibc), NOT alpine: the agently-cli native binary is a Go/glibc
# build, so musl would break it. node is required because agently-cli is an npm
# wrapper that dispatches to the platform native binary.
FROM node:22-slim AS runtime

# agently-cli release verified end-to-end; override to upgrade.
ARG AGENTLY_CLI_VERSION=1.0.6

RUN apt-get update \
 && apt-get install -y --no-install-recommends ca-certificates tini \
 && rm -rf /var/lib/apt/lists/* \
 && npm install -g "@tencent-qqmail/agently-cli@${AGENTLY_CLI_VERSION}" \
 && AGENTLY_CLI_NO_UPDATE_NOTIFIER=1 agently-cli --version

# Non-root runtime user.
RUN useradd --create-home --shell /usr/sbin/nologin app

COPY --from=builder /out/qq-agent-mail-mcp /usr/local/bin/qq-agent-mail-mcp

# agently-cli credential/config dir. A named volume mounted here inherits this
# ownership, so login survives container restarts.
RUN install -d -o app -g app /var/lib/agently

ENV QQ_AGENT_MAIL_MCP_BIND=0.0.0.0:8765 \
    AGENTLY_CLI_CONFIG_DIR=/var/lib/agently \
    AGENTLY_CLI_NO_UPDATE_NOTIFIER=1 \
    QQ_AGENTLY_CLI_BIN=agently-cli

USER app
WORKDIR /home/app

VOLUME ["/var/lib/agently"]
EXPOSE 8765

# tini reaps the agently-cli child processes the server exec's and forwards signals.
ENTRYPOINT ["/usr/bin/tini", "--", "/usr/local/bin/qq-agent-mail-mcp"]
