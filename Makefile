VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo 0.0.4)
LDFLAGS := -X main.version=$(VERSION)

.PHONY: build run test vet clean

build:
	go build -ldflags "$(LDFLAGS)" -o bin/qq-agent-mail-mcp ./cmd/qq-agent-mail-mcp

run:
	go run -ldflags "$(LDFLAGS)" ./cmd/qq-agent-mail-mcp

test:
	go test ./...

vet:
	go vet ./...

clean:
	rm -f bin/qq-agent-mail-mcp
