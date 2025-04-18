package cli

import (
	"github.com/floppa/yxa-cli/internal/executor"
)

// NewMockExecutor creates a new mock executor for testing
// This is a convenience function to create a mock executor from the CLI package
func NewMockExecutor() *executor.MockExecutor {
	return executor.NewMockExecutor()
}
