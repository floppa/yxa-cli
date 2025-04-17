package cmd

import (
	"bytes"
	"strings"
	"testing"
)

// TestExecuteCommand tests the executeCommand function
func TestExecuteCommand(t *testing.T) {
	// Set up test environment with mock executor
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Test cases
	tests := []struct {
		name        string
		command     string
		expectError bool
		checkLog    bool
		logContains []string
	}{
		{
			name:        "Simple command",
			command:     "simple",
			expectError: false,
			checkLog:    true,
			logContains: []string{"Simple command"},
		},
		{
			name:        "Command with pre-hook",
			command:     "with-pre",
			expectError: false,
			checkLog:    true,
			logContains: []string{"Pre-hook", "Main command with Pre-hook"},
		},
		{
			name:        "Command with post-hook",
			command:     "with-post",
			expectError: false,
			checkLog:    true,
			logContains: []string{"Main command with Post-hook", "Post-hook"},
		},
		{
			name:        "Command with pre and post hooks",
			command:     "with-pre-post",
			expectError: false,
			checkLog:    true,
			logContains: []string{"Pre-hook", "Main command with Pre-hook and Post-hook", "Post-hook"},
		},
		{
			name:        "Command with timeout",
			command:     "with-timeout",
			expectError: true,
			checkLog:    false,
		},
		{
			name:        "Command with true condition",
			command:     "with-condition-true",
			expectError: false,
			checkLog:    false,
		},
		{
			name:        "Command with false condition",
			command:     "with-condition-false",
			expectError: false,
			checkLog:    false,
		},
		{
			name:        "Task aggregator command",
			command:     "task-aggregator",
			expectError: false,
			checkLog:    true,
			logContains: []string{"Simple command", "Task aggregator command"},
		},
		{
			name:        "Command with no run or depends",
			command:     "non-existent",
			expectError: true,
			checkLog:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Reset the output buffer
			mockExec := cmdExecutor.(*MockExecutor)
			mockExec.ClearOutput()

			// Reset executed commands
			executedCommands = make(map[string]bool)

			// Execute command
			err := executeCommand(tc.command)

			// Check error
			if tc.expectError && err == nil {
				t.Errorf("Expected error but got none")
			} else if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Check output buffer contents if needed
			if tc.checkLog {
				output := mockExec.GetOutput()
				outputBytes := []byte(output)

				for _, expected := range tc.logContains {
					if !bytes.Contains(outputBytes, []byte(expected)) {
						t.Errorf("Expected output to contain '%s', but got: %s", expected, output)
					}
				}
			}
		})
	}
}

// TestExecute tests the Execute function
func TestExecute(t *testing.T) {
	// Set up test environment
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Store original args
	origArgs := rootCmd.Args
	defer func() {
		rootCmd.Args = origArgs
	}()

	// Test with no args
	rootCmd.SetArgs([]string{})
	// Execute doesn't return an error, so we just call it
	Execute()

	// Test with valid command
	rootCmd.SetArgs([]string{"simple"})
	Execute()

	// Test with invalid command - we can't check for errors directly
	// since Execute() doesn't return an error, but we can check that
	// the command was executed by checking the mock executor's output
	rootCmd.SetArgs([]string{"invalid-command"})
	Execute()
	
	// Verify that the mock executor was called
	mockExec := cmdExecutor.(*MockExecutor)
	found := false
	for _, cmd := range mockExec.ExecutedCommands {
		if strings.Contains(cmd, "invalid-command") {
			found = true
			break
		}
	}
	
	// Since invalid-command doesn't exist, it shouldn't be executed
	if found {
		t.Errorf("Execute() with invalid command should not have executed the command")
	}
}

// TestInitFunction tests the init function indirectly by checking rootCmd setup
func TestInitFunction(t *testing.T) {
	// Check that rootCmd is properly initialized
	if rootCmd == nil {
		t.Fatal("rootCmd is nil")
	}

	// Check that rootCmd has the expected properties
	if rootCmd.Use != "yxa" {
		t.Errorf("Expected rootCmd.Use to be 'yxa', got '%s'", rootCmd.Use)
	}

	// Check that rootCmd has the expected short description
	if !strings.Contains(rootCmd.Short, "cli") {
		t.Errorf("Expected rootCmd.Short to contain 'cli', got '%s'", rootCmd.Short)
	}
}
