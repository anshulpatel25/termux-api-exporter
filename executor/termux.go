package executor

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

// TermuxExecutor implements the Executor interface for Termux commands
type TermuxExecutor struct {
	timeout time.Duration
}

// NewTermuxExecutor creates a new TermuxExecutor with the specified timeout
func NewTermuxExecutor(timeout time.Duration) *TermuxExecutor {
	return &TermuxExecutor{
		timeout: timeout,
	}
}

// Execute runs a termux command and returns its stdout output
func (e *TermuxExecutor) Execute(ctx context.Context, command string, args ...string) ([]byte, error) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	// Create the command
	cmd := exec.CommandContext(ctx, command, args...)

	// Execute and capture output
	output, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("command execution timed out after %v: %w", e.timeout, err)
		}
		// Check if it's an ExitError to get stderr
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("command failed with exit code %d: %s", exitErr.ExitCode(), string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("failed to execute command: %w", err)
	}

	return output, nil
}
