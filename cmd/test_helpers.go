package cmd

import (
	"testing"
)

// setupTestEnvironment sets up a test environment with a mock executor
// and returns a cleanup function to restore the original state
func setupTestEnvironment(t *testing.T) func() {
	// Store original values
	originalConfig := cfg
	originalExecutor := cmdExecutor
	originalExecutedCommands := executedCommands

	// Create a new mock executor
	mockExec := NewMockExecutor()
	
	// Set the mock executor as the global executor
	cmdExecutor = mockExec
	
	// Set the mock config as the global config
	cfg = mockExec.Config
	
	// Reset executed commands
	executedCommands = make(map[string]bool)

	// Return cleanup function
	return func() {
		cfg = originalConfig
		cmdExecutor = originalExecutor
		executedCommands = originalExecutedCommands
	}
}


