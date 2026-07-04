package agently

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"time"
)

const (
	DefaultBinary  = "agently-cli"
	DefaultTimeout = 30 * time.Second
)

var ErrInvalidArguments = errors.New("invalid arguments")

type Runner struct {
	Binary  string
	Timeout time.Duration
}

type CLIError struct {
	Err       error
	ExitCode  int
	Stderr    string
	RawStdout string
	Output    map[string]any
}

func (e *CLIError) Error() string {
	msg := "agently-cli failed"
	if e.ExitCode >= 0 {
		msg += fmt.Sprintf(" with exit code %d", e.ExitCode)
	}
	if code := e.Code(); code != "" {
		msg += ": " + code
	}
	if message, ok := nestedString(e.Output, "error", "message"); ok {
		msg += ": " + message
	} else if e.Stderr != "" {
		msg += ": " + e.Stderr
	}
	return msg
}

func (e *CLIError) Unwrap() error {
	return e.Err
}

func (e *CLIError) Code() string {
	if code, ok := nestedString(e.Output, "error", "code"); ok {
		return code
	}
	if code, ok := nestedString(e.Output, "error", "type"); ok {
		return code
	}
	if code, ok := e.Output["code"].(string); ok {
		return code
	}
	return ""
}

func (e *CLIError) Message() string {
	if message, ok := nestedString(e.Output, "error", "message"); ok {
		return message
	}
	if message, ok := e.Output["message"].(string); ok {
		return message
	}
	if e.Stderr != "" {
		return e.Stderr
	}
	return e.Error()
}

func (r Runner) Run(ctx context.Context, args []string) (map[string]any, error) {
	binary := r.Binary
	if binary == "" {
		binary = DefaultBinary
	}

	timeout := r.Timeout
	if timeout == 0 {
		timeout = DefaultTimeout
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, binary, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, newCLIError(err, stdout.Bytes(), stderr.String())
	}

	var out map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		return nil, fmt.Errorf("agently-cli returned non-JSON output: %w", err)
	}
	return out, nil
}

func newCLIError(err error, stdout []byte, stderr string) *CLIError {
	cliErr := &CLIError{
		Err:       err,
		ExitCode:  -1,
		Stderr:    sanitize(stderr),
		RawStdout: sanitize(string(stdout)),
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		cliErr.ExitCode = exitErr.ExitCode()
	}

	var out map[string]any
	if json.Unmarshal(stdout, &out) == nil {
		cliErr.Output = out
	}
	return cliErr
}

func nestedString(value map[string]any, firstKey string, secondKey string) (string, bool) {
	first, ok := value[firstKey].(map[string]any)
	if !ok {
		return "", false
	}
	second, ok := first[secondKey].(string)
	return second, ok
}

func sanitize(value string) string {
	if len(value) > 512 {
		return value[:512] + "..."
	}
	return value
}

func shellQuote(value string) string {
	return strconv.Quote(value)
}
