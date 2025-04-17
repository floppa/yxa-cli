package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/floppa/yxa-cli/config"
)

// TestExecuteSingleCommand tests the executeSingleCommand function
func TestExecuteSingleCommand(t *testing.T) {
	tests := []struct {
		name          string
		command       string
		timeout       time.Duration
		expectError   bool
		expectTimeout bool
	}{
		{
			name:        "Simple echo command",
			command:     "echo 'Hello, World!'",
			timeout:     0,
			expectError: false,
		},
		{
			name:        "Command with pipes",
			command:     "echo 'Hello' | grep 'Hello'",
			timeout:     0,
			expectError: false,
		},
		{
			name:        "Invalid command",
			command:     "invalid_command_that_does_not_exist",
			timeout:     0,
			expectError: true,
		},
		{
			name:          "Command with timeout",
			command:       "sleep 0.1",
			timeout:       5 * time.Millisecond,
			expectError:   true,
			expectTimeout: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Capture stdout and stderr
			oldStdout := os.Stdout
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stdout = w
			os.Stderr = w

			// Execute the command
			err := executeSingleCommand(tc.command, tc.timeout)

			// Close the pipe writer and restore stdout/stderr
			if err := w.Close(); err != nil {
				t.Errorf("Failed to close writer: %v", err)
			}
			os.Stdout = oldStdout
			os.Stderr = oldStderr

			// Read the output
			var buf bytes.Buffer
			if _, err := io.Copy(&buf, r); err != nil {
				t.Errorf("Failed to read output: %v", err)
			}

			// Check the result
			if tc.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if tc.expectTimeout && !strings.Contains(err.Error(), "timed out") {
				t.Errorf("Expected timeout error but got: %v", err)
			}
		})
	}
}

// TestExecuteSequentialCommands tests the executeSequentialCommands function
func TestExecuteSequentialCommands(t *testing.T) {
	// Set up test environment with mock executor
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Get the mock executor
	mockExec := cmdExecutor.(*MockExecutor)

	// Clear the output buffer
	mockExec.ClearOutput()

	// Create a test command
	cmd := config.Command{
		Commands: map[string]string{
			"step1": "echo Step 1",
			"step2": "echo Step 2",
			"step3": "echo Step 3",
		},
	}

	// Set up specific command results for the mock executor
	mockExec.AddCommandResult("echo Step 1", "Step 1\n", nil)
	mockExec.AddCommandResult("echo Step 2", "Step 2\n", nil)
	mockExec.AddCommandResult("echo Step 3", "Step 3\n", nil)

	// Execute the sequential commands
	err := executeSequentialCommands("test-sequential", cmd, 0)

	// Check the result
	if err != nil {
		t.Errorf("Expected no error but got: %v", err)
	}

	// Get the output from the mock executor
	output := mockExec.GetOutput()

	// Check that each step was executed
	expectedSteps := []string{"Step 1", "Step 2", "Step 3"}
	for _, step := range expectedSteps {
		if !strings.Contains(output, step) {
			t.Errorf("Output does not contain step '%s': %s", step, output)
		}
	}

	// Check that all commands were executed (we don't check the exact order because
	// map iteration order in Go is not guaranteed, which affects how commands are executed)
	expectedCmds := []string{"echo Step 1", "echo Step 2", "echo Step 3"}
	for _, expectedCmd := range expectedCmds {
		found := false
		for _, executedCmd := range mockExec.ExecutedCommands {
			if strings.Contains(executedCmd, expectedCmd) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected command '%s' to be executed, but it wasn't", expectedCmd)
		}
	}
	
	// Check that exactly 3 commands were executed
	if len(mockExec.ExecutedCommands) != 3 {
		t.Errorf("Expected 3 commands to be executed, got %d", len(mockExec.ExecutedCommands))
	}
}

// TestSafeWriter tests the safeWriter implementation
func TestSafeWriter(t *testing.T) {
	// Create a buffer to write to
	var buf bytes.Buffer

	// Create a safe writer
	prefix := "[TEST] "
	writer := newSafeWriter(&buf, prefix)

	// Test writing a single line
	if _, err := writer.Write([]byte("Hello, World!\n")); err != nil {
		t.Errorf("Failed to write to buffer: %v", err)
	}
	
	// Flush the buffer
	if err := writer.Flush(); err != nil {
		t.Errorf("Failed to flush buffer: %v", err)
	}
	
	expected := prefix + "Hello, World!\n"
	if buf.String() != expected {
		t.Errorf("Expected %q, got %q", expected, buf.String())
	}

	// Reset the buffer
	buf.Reset()

	// Test partial line followed by complete line
	if _, err := writer.Write([]byte("Partial")); err != nil {
		t.Errorf("Failed to write to buffer: %v", err)
	}
	if _, err := writer.Write([]byte(" Line\nComplete Line\n")); err != nil {
		t.Errorf("Failed to write to buffer: %v", err)
	}
	
	// Flush the buffer
	if err := writer.Flush(); err != nil {
		t.Errorf("Failed to flush buffer: %v", err)
	}
	
	expected = prefix + "Partial Line\n" + prefix + "Complete Line\n"
	if buf.String() != expected {
		t.Errorf("Expected %q, got %q", expected, buf.String())
	}

	// Reset the buffer
	buf.Reset()

	// Test writing multiple lines
	if _, err := writer.Write([]byte("Line 1\nLine 2\nLine 3\n")); err != nil {
		t.Errorf("Failed to write to buffer: %v", err)
	}
	
	// Flush the buffer
	if err := writer.Flush(); err != nil {
		t.Errorf("Failed to flush buffer: %v", err)
	}
	

	expected = prefix + "Line 1\n" + prefix + "Line 2\n" + prefix + "Line 3\n"
	if buf.String() != expected {
		t.Errorf("Expected %q, got %q", expected, buf.String())
	}
}

// TestExecuteParallelCommands tests the executeParallelCommands function
func TestExecuteParallelCommands(t *testing.T) {
	// Skip in CI environment
	if os.Getenv("CI") != "" {
		t.Skip("Skipping parallel command test in CI environment")
	}

	// Set up test environment with mock executor
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Get the mock executor
	mockExec := cmdExecutor.(*MockExecutor)

	// Clear the output buffer
	mockExec.ClearOutput()

	// Create a test command with shorter sleep times to speed up tests
	cmd := config.Command{
		Commands: map[string]string{
			"task1": "echo Task 1",
			"task2": "echo Task 2",
			"task3": "echo Task 3",
		},
		Parallel: true,
	}

	// Set up specific command results for the mock executor
	mockExec.AddCommandResult("echo Task 1", "Task 1\n", nil)
	mockExec.AddCommandResult("echo Task 2", "Task 2\n", nil)
	mockExec.AddCommandResult("echo Task 3", "Task 3\n", nil)

	// Execute the parallel commands
	err := executeParallelCommands("test-parallel", cmd, 0)

	// Check the result
	if err != nil {
		t.Errorf("Expected no error but got: %v", err)
	}

	// Get the output from the mock executor
	output := mockExec.GetOutput()

	// Check that all tasks were executed
	expectedTasks := []string{"Task 1", "Task 2", "Task 3"}
	for _, task := range expectedTasks {
		if !strings.Contains(output, task) {
			t.Errorf("Output does not contain task '%s': %s", task, output)
		}
	}

	// Test with a timeout - use a separate subtest to avoid race conditions
	t.Run("Timeout", func(t *testing.T) {
		// Set up a fresh test environment with a new mock executor
		cleanup := setupTestEnvironment(t)
		defer cleanup()
		
		// Get the mock executor
		mockExec := cmdExecutor.(*MockExecutor)
		
		// Create a test command with a slow task (using shorter duration)
		cmd := config.Command{
			Commands: map[string]string{
				"slow": "sleep 0.2",
			},
			Parallel: true,
		}
		
		// Add a timeout result for the slow command
		mockExec.AddCommandResult("sleep 0.2", "", fmt.Errorf("command timed out after 10ms"))
		
		// Execute the parallel commands with a short timeout
		err := executeParallelCommands("test-parallel", cmd, 10*time.Millisecond)
		
		// Check that we got a timeout error
		if err == nil {
			t.Errorf("Expected timeout error but got none")
		}
		if !strings.Contains(err.Error(), "timed out") {
			t.Errorf("Expected timeout error message, got: %v", err)
		}
	})
}
