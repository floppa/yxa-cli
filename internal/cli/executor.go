package cli

import (
	"io"
	"time"

	"github.com/floppa/yxa-cli/internal/executor"
)

// DefaultExecutor creates a new default executor
// This is a convenience function to create an executor from the CLI package
func NewDefaultExecutor() executor.CommandExecutor {
	return executor.NewDefaultExecutor()
}


// CommandExecutor is an interface that wraps the executor.CommandExecutor interface
// This allows the CLI package to use the executor package without direct imports in client code
type CommandExecutor interface {
	// Execute runs a shell command with optional timeout
	Execute(cmdStr string, timeout time.Duration) error
	
	// ExecuteWithOutput runs a shell command and returns its output
	ExecuteWithOutput(cmdStr string, timeout time.Duration) (string, error)
	
	// GetStdout returns the stdout writer
	GetStdout() io.Writer
	
	// GetStderr returns the stderr writer
	GetStderr() io.Writer
	
	// SetStdout sets the stdout writer
	SetStdout(w io.Writer)
	
	// SetStderr sets the stderr writer
	SetStderr(w io.Writer)
}
