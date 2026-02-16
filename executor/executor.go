package executor

import "context"

// Executor defines the interface for executing external commands
type Executor interface {
	// Execute runs a command and returns its stdout output
	Execute(ctx context.Context, command string, args ...string) ([]byte, error)
}
