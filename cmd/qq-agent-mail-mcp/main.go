package main

import (
	"context"
	"log"
	"os"

	"github.com/Aston5128/qq-agent-mail-mcp/internal/agently"
	"github.com/Aston5128/qq-agent-mail-mcp/internal/config"
	"github.com/Aston5128/qq-agent-mail-mcp/internal/server"
)

// version 可在构建时通过 ldflags 注入；未注入时使用当前源码版本。
// 构建示例: go build -ldflags "-X main.version=0.0.4"
var version = "0.0.4"

const name = "qq-agent-mail-mcp"

func main() {
	if err := config.LoadEnvFile(); err != nil {
		log.Printf("warning: %v", err)
	}
	bind := getenv("QQ_AGENT_MAIL_MCP_BIND", server.DefaultBind)
	binary := getenv("QQ_AGENTLY_CLI_BIN", agently.DefaultBinary)

	cfg := server.Config{
		Name:    name,
		Version: version,
		Runner:  agently.Runner{Binary: binary},
	}

	log.Printf("%s listening on %s", cfg.Name, bind)
	if err := server.RunStreamableHTTP(context.Background(), bind, cfg); err != nil {
		log.Fatal(err)
	}
}

func getenv(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
