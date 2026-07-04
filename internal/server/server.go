package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/Aston5128/qq-agent-mail-mcp/internal/agently"
)

const DefaultBind = "127.0.0.1:8765"

type Config struct {
	Name    string
	Version string
	Runner  agently.Runner
}

type argsBuilder interface {
	Args() []string
}

func New(cfg Config) *mcp.Server {
	name := cfg.Name
	if name == "" {
		name = "mcp-server"
	}
	version := cfg.Version
	if version == "" {
		version = "dev"
	}

	s := mcp.NewServer(&mcp.Implementation{Name: name, Version: version}, nil)
	addForwardTool[agently.EmptyInput](s, cfg.Runner, "agently_me", "Show current user info and alias list.")
	addForwardTool[agently.MessageListInput](s, cfg.Runner, "agently_message_list", "List messages in a folder with pagination support.")
	addForwardTool[agently.MessageReadInput](s, cfg.Runner, "agently_message_read", "Read a single message in full.")
	addForwardTool[agently.MessageSearchInput](s, cfg.Runner, "agently_message_search", "Search messages with keyword and filters.")
	addForwardTool[agently.MessageSendInput](s, cfg.Runner, "agently_message_send", "Send a new email using agently-cli confirmation flow.")
	addForwardTool[agently.MessageReplyInput](s, cfg.Runner, "agently_message_reply", "Reply to an existing message using agently-cli confirmation flow.")
	addForwardTool[agently.MessageForwardInput](s, cfg.Runner, "agently_message_forward", "Forward a message using agently-cli confirmation flow.")
	addForwardTool[agently.AttachmentUploadInput](s, cfg.Runner, "agently_attachment_upload", "Upload a file for later attachment.")
	addForwardTool[agently.AttachmentDownloadInput](s, cfg.Runner, "agently_attachment_download", "Download a message attachment.")
	return s
}

func RunStreamableHTTP(ctx context.Context, bind string, cfg Config) error {
	if bind == "" {
		bind = DefaultBind
	}
	mcpServer := New(cfg)
	handler := mcp.NewStreamableHTTPHandler(
		func(*http.Request) *mcp.Server { return mcpServer },
		&mcp.StreamableHTTPOptions{JSONResponse: true},
	)
	httpServer := &http.Server{Addr: bind, Handler: handler}

	errc := make(chan error, 1)
	go func() {
		errc <- httpServer.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		if err := httpServer.Shutdown(context.Background()); err != nil {
			return err
		}
		return ctx.Err()
	case err := <-errc:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	}
}

func addForwardTool[In argsBuilder](s *mcp.Server, runner agently.Runner, name string, description string) {
	mcp.AddTool(s, &mcp.Tool{Name: name, Description: description},
		func(ctx context.Context, _ *mcp.CallToolRequest, input In) (*mcp.CallToolResult, any, error) {
			out, err := runner.Run(ctx, input.Args())
			if err != nil {
				var cliErr *agently.CLIError
				if errors.As(err, &cliErr) {
					return cliErrorResult(name, cliErr), nil, nil
				}
				return nil, nil, fmt.Errorf("%s failed: %w", name, err)
			}
			return nil, out, nil
		})
}

func cliErrorResult(toolName string, cliErr *agently.CLIError) *mcp.CallToolResult {
	errorInfo := map[string]any{
		"type":      "agently_cli_error",
		"tool":      toolName,
		"message":   cliErr.Message(),
		"exit_code": cliErr.ExitCode,
	}
	if code := cliErr.Code(); code != "" {
		errorInfo["code"] = code
	}
	if cliErr.Stderr != "" {
		errorInfo["stderr"] = cliErr.Stderr
	}
	if cliErr.Output != nil {
		errorInfo["stdout"] = cliErr.Output
	} else if cliErr.RawStdout != "" {
		errorInfo["stdout_text"] = cliErr.RawStdout
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: cliErr.Error()},
		},
		StructuredContent: map[string]any{
			"ok":    false,
			"error": errorInfo,
		},
		IsError: true,
	}
}
