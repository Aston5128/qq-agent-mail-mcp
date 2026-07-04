// Package config loads runtime configuration for the QQ Agent Mail MCP server.
//
// Currently its only responsibility is loading environment variables from a .env
// file at startup (callers should invoke LoadEnvFile before reading any other
// configuration). Loading is non-overriding: variables already present in the
// real environment take precedence, so .env only supplies defaults. This keeps
// docker -e / mounted-secret environment values authoritative in deployments.
package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// EnvFileKey names the environment variable used to override the .env file path.
const EnvFileKey = "QQ_AGENT_MAIL_MCP_ENV_FILE"

// LoadEnvFile loads variables from a .env file into the process environment,
// without overriding variables that are already set.
//
// The path defaults to ".env" in the working directory; when EnvFileKey is set,
// that path is used instead. A missing file is treated as "not configured" and
// returns nil. A present-but-unreadable or malformed file returns an error.
func LoadEnvFile() error {
	path := os.Getenv(EnvFileKey)
	if path == "" {
		path = ".env"
	}
	if err := godotenv.Load(path); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("load env file %q: %w", path, err)
	}
	return nil
}
