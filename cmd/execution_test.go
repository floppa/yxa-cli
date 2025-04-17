package cmd

import (
	"bytes"
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
			command:       "sleep 2",
			timeout:       500 * time.Millisecond,
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
	// Create a test command
	cmd := config.Command{
		Commands: map[string]string{
			"step1": "echo 'Step 1'",
			"step2": "echo 'Step 2'",
			"step3": "echo 'Step 3'",
		},
	}

	// Set up global config for variable replacement
	cfg = &config.ProjectConfig{
		Variables: map[string]string{},
	}

	// Capture stdout and stderr
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	// Execute the sequential commands
	err := executeSequentialCommands("test-sequential", cmd, 0)

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
	output := buf.String()

	// Check the result
	if err != nil {
		t.Errorf("Expected no error but got: %v", err)
	}
	if !strings.Contains(output, "Step 1") || !strings.Contains(output, "Step 2") || !strings.Contains(output, "Step 3") {
		t.Errorf("Output does not contain all steps: %s", output)
	}
}

// TestPrefixedWriter tests the prefixedWriter implementation
func TestPrefixedWriter(t *testing.T) {
	// Create a buffer to capture output
	var buf bytes.Buffer

	// Create a prefixed writer
	prefix := "[TEST] "
	writer := newPrefixedWriter(&buf, prefix)

	// Test writing a single line
	if _, err := writer.Write([]byte("Hello, World!\n")); err != nil {
		t.Errorf("Failed to write to buffer: %v", err)
	}
	expected := prefix + "Hello, World!\n"
	if buf.String() != expected {
		t.Errorf("Expected %q, got %q", expected, buf.String())
	}

	// Reset the buffer
	buf.Reset()

	// Test writing multiple lines
	if _, err := writer.Write([]byte("Line 1\nLine 2\nLine 3")); err != nil {
		t.Errorf("Failed to write to buffer: %v", err)
	}
	expected = prefix + "Line 1\n" + prefix + "Line 2\n"
	if !strings.HasPrefix(buf.String(), expected) {
		t.Errorf("Expected output to start with %q, got %q", expected, buf.String())
	}
	// The last line without a newline should be buffered
	if strings.Contains(buf.String(), "Line 3") {
		t.Errorf("Line without newline should be buffered, but found in output: %q", buf.String())
	}

	// Write a newline to flush the buffer
	if _, err := writer.Write([]byte("\n")); err != nil {
		t.Errorf("Failed to write to buffer: %v", err)
	}
	expected = prefix + "Line 1\n" + prefix + "Line 2\n" + prefix + "Line 3\n"
	if buf.String() != expected {
		t.Errorf("Expected %q, got %q", expected, buf.String())
	}
}

// TestExecuteParallelCommands tests the executeParallelCommands function
func TestExecuteParallelCommands(t *testing.T) {
	// Skip if running in CI environment
	if os.Getenv("CI") != "" {
		t.Skip("Skipping parallel command test in CI environment")
	}

	// Create a test command
	cmd := config.Command{
		Commands: map[string]string{
			"task1": "echo 'Task 1' && sleep 0.1",
			"task2": "echo 'Task 2' && sleep 0.1",
			"task3": "echo 'Task 3' && sleep 0.1",
		},
		Parallel: true,
	}

	// Set up global config for variable replacement
	cfg = &config.ProjectConfig{
		Variables: map[string]string{},
	}

	// Capture stdout and stderr
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	// Execute the parallel commands
	err := executeParallelCommands("test-parallel", cmd, 0)

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
	output := buf.String()

	// Check the result
	if err != nil {
		t.Errorf("Expected no error but got: %v", err)
	}
	if !strings.Contains(output, "Task 1") || !strings.Contains(output, "Task 2") || !strings.Contains(output, "Task 3") {
		t.Errorf("Output does not contain all tasks: %s", output)
	}

	// Test with a timeout
	cmd.Commands["slow"] = "sleep 2"
	
	// Capture stdout and stderr again
	r, w, _ = os.Pipe()
	os.Stdout = w
	os.Stderr = w

	// Execute with a short timeout
	err = executeParallelCommands("test-parallel-timeout", cmd, 100*time.Millisecond)

	// Close the pipe writer and restore stdout/stderr
	if err := w.Close(); err != nil {
		t.Errorf("Failed to close writer: %v", err)
	}

	// Read the output
	buf.Reset()
	if _, err := io.Copy(&buf, r); err != nil {
		t.Errorf("Failed to read output: %v", err)
	}

	// Check for timeout error
	if err == nil {
		t.Errorf("Expected timeout error but got none")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Errorf("Expected timeout error but got: %v", err)
	}
}
