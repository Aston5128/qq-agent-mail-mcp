package main

import (
	"context"
	"log"
	"os"

	"github.com/Aston5128/qq-agent-mail-mcp/internal/agently"
	"github.com/Aston5128/qq-agent-mail-mcp/internal/config"
	"github.com/Aston5128/qq-agent-mail-mcp/internal/server"
)

// version 是 MCP server 自报的二进制版本。
//
// 不要在源码里手动维护发布版本号；GitHub Release / GHCR 构建会通过
// -ldflags "-X main.version=<release-tag>" 注入真实版本。
// 本地 go run 或未注入版本的构建使用 "dev"，表示这是开发构建。
var version = "dev"

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
