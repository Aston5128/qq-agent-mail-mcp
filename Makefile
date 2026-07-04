VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -X main.version=$(VERSION)

.PHONY: build run compose-build test vet clean

build:
	go build -ldflags "$(LDFLAGS)" -o bin/qq-agent-mail-mcp ./cmd/qq-agent-mail-mcp

run:
	go run -ldflags "$(LDFLAGS)" ./cmd/qq-agent-mail-mcp

compose-build:
	QQ_AGENT_MAIL_MCP_BUILD_VERSION=$(VERSION) docker compose up -d --build

test:
	go test ./...

vet:
	go vet ./...

clean:
	rm -f bin/qq-agent-mail-mcp
