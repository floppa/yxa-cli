package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"
)

// TestDefaultExecutor tests the DefaultExecutor implementation
func TestDefaultExecutor(t *testing.T) {
	// Create a buffer to capture output
	var stdout, stderr bytes.Buffer
	
	// Create a default executor with the buffer as stdout/stderr
	executor := &DefaultExecutor{
		Stdout: &stdout,
		Stderr: &stderr,
	}
	
	// Test simple command execution
	err := executor.Execute("echo 'Hello, World!'", 0)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	// Check output
	output := stdout.String()
	if !strings.Contains(output, "Hello, World!") {
		t.Errorf("Expected output to contain 'Hello, World!', got: %s", output)
	}
	
	// Reset buffers
	stdout.Reset()
	stderr.Reset()
	
	// Test command with error
	err = executor.Execute("invalid_command_that_does_not_exist", 0)
	if err == nil {
		t.Errorf("Expected error for invalid command, got none")
	}
	
	// Test command with timeout (using shorter duration)
	start := time.Now()
	err = executor.Execute("sleep 0.2", 10*time.Millisecond)
	duration := time.Since(start)
	
	// Check that the command timed out
	if err == nil {
		t.Errorf("Expected timeout error, got none")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Errorf("Expected timeout error message, got: %v", err)
	}
	if duration >= 100*time.Millisecond {
		t.Errorf("Expected command to timeout after 10ms, but it took %v", duration)
	}
}

// TestMockExecutor tests the MockExecutor implementation
func TestMockExecutor(t *testing.T) {
	// Create a mock executor
	mockExec := NewMockExecutor()
	
	// Test default behavior
	err := mockExec.Execute("echo 'Hello, World!'", 0)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	// Check output
	output := mockExec.GetOutput()
	if !strings.Contains(output, "Hello, World!") {
		t.Errorf("Expected output to contain 'Hello, World!', got: %s", output)
	}
	
	// Clear output
	mockExec.ClearOutput()
	
	// Test custom result
	mockExec.AddCommandResult("custom-command", "Custom output", nil)
	err = mockExec.Execute("custom-command", 0)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	// Check output
	output = mockExec.GetOutput()
	if !strings.Contains(output, "Custom output") {
		t.Errorf("Expected output to contain 'Custom output', got: %s", output)
	}
	
	// Test error result
	expectedErr := fmt.Errorf("custom error")
	mockExec.AddCommandResult("error-command", "", expectedErr)
	err = mockExec.Execute("error-command", 0)
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
	
	// Test timeout (using shorter duration)
	err = mockExec.Execute("sleep 0.2", 10*time.Millisecond)
	if err == nil {
		t.Errorf("Expected timeout error, got none")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Errorf("Expected timeout error message, got: %v", err)
	}
	
	// Test ExecuteWithOutput
	output, err = mockExec.ExecuteWithOutput("echo 'Test output'", 0)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !strings.Contains(output, "Test output") {
		t.Errorf("Expected output to contain 'Test output', got: %s", output)
	}
}

// TestExecutorInterface tests that both implementations satisfy the CommandExecutor interface
func TestExecutorInterface(t *testing.T) {
	// This test doesn't actually run any code, it just verifies at compile time
	// that both implementations satisfy the CommandExecutor interface
	var _ CommandExecutor = &DefaultExecutor{}
	var _ CommandExecutor = &MockExecutor{}
}

// TestDefaultExecutorGettersSetters tests the getter and setter methods of DefaultExecutor
func TestDefaultExecutorGettersSetters(t *testing.T) {
	// Create a default executor
	executor := NewDefaultExecutor()
	
	// Create test writers
	stdoutWriter := &bytes.Buffer{}
	stderrWriter := &bytes.Buffer{}
	
	// Test SetStdout
	executor.SetStdout(stdoutWriter)
	gotStdout := executor.GetStdout()
	if gotStdout != stdoutWriter {
		t.Errorf("Expected stdout writer %v, got %v", stdoutWriter, gotStdout)
	}
	
	// Test SetStderr
	executor.SetStderr(stderrWriter)
	gotStderr := executor.GetStderr()
	if gotStderr != stderrWriter {
		t.Errorf("Expected stderr writer %v, got %v", stderrWriter, gotStderr)
	}
	
	// Test concurrent access to getters/setters
	done := make(chan bool)
	go func() {
		for i := 0; i < 100; i++ {
			executor.SetStdout(&bytes.Buffer{})
			_ = executor.GetStdout()
		}
		done <- true
	}()
	
	go func() {
		for i := 0; i < 100; i++ {
			executor.SetStderr(&bytes.Buffer{})
			_ = executor.GetStderr()
		}
		done <- true
	}()
	
	<-done
	<-done
}

// TestMockExecutorSetters tests the setter methods of MockExecutor
func TestMockExecutorSetters(t *testing.T) {
	// Create a mock executor
	mockExec := NewMockExecutor()
	
	// Test SetStdout
	stdoutWriter := &bytes.Buffer{}
	mockExec.SetStdout(stdoutWriter)
	gotStdout := mockExec.GetStdout()
	if gotStdout != stdoutWriter {
		t.Errorf("Expected stdout writer %v, got %v", stdoutWriter, gotStdout)
	}
	
	// Test SetStderr
	stderrWriter := &bytes.Buffer{}
	mockExec.SetStderr(stderrWriter)
	gotStderr := mockExec.GetStderr()
	if gotStderr != stderrWriter {
		t.Errorf("Expected stderr writer %v, got %v", stderrWriter, gotStderr)
	}
	
	// Test SetConfig
	cfg := createDefaultTestConfig()
	mockExec.SetConfig(cfg)
	if mockExec.Config != cfg {
		t.Errorf("Expected config %v, got %v", cfg, mockExec.Config)
	}
	
	// Test AddCommandResult with nil CommandResults
	// Reset the mock executor to ensure CommandResults is nil
	mockExec = NewMockExecutor()
	mockExec.CommandResults = nil // Explicitly set to nil
	
	// This should initialize the map
	mockExec.AddCommandResult("test-cmd", "test-output", nil)
	
	// Check that the command was added
	result, ok := mockExec.CommandResults["test-cmd"]
	if !ok {
		t.Errorf("Expected command 'test-cmd' to be added to CommandResults")
	}
	if result.Output != "test-output" || result.Error != nil {
		t.Errorf("Expected output 'test-output' and nil error, got '%s' and %v", result.Output, result.Error)
	}
}

// TestExecuteWithOutput tests the ExecuteWithOutput method
func TestExecuteWithOutput(t *testing.T) {
	// Create a default executor
	executor := NewDefaultExecutor()
	
	// Test successful command
	output, err := executor.ExecuteWithOutput("echo 'Test Output'", 0)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !strings.Contains(output, "Test Output") {
		t.Errorf("Expected output to contain 'Test Output', got: %s", output)
	}
	
	// Test command with error
	_, err = executor.ExecuteWithOutput("invalid_command_that_does_not_exist", 0)
	if err == nil {
		t.Errorf("Expected error for invalid command, got none")
	}
	
	// Test command with timeout
	start := time.Now()
	_, err = executor.ExecuteWithOutput("sleep 0.2", 10*time.Millisecond)
	duration := time.Since(start)
	
	// Check that the command timed out
	if err == nil {
		t.Errorf("Expected timeout error, got none")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Errorf("Expected timeout error message, got: %v", err)
	}
	if duration >= 100*time.Millisecond {
		t.Errorf("Expected command to timeout after 10ms, but it took %v", duration)
	}
	
	// Verify that stdout is properly restored after ExecuteWithOutput
	var buf bytes.Buffer
	executor.SetStdout(&buf)
	
	// Execute a command that should not affect our buffer
	_, _ = executor.ExecuteWithOutput("echo 'Should not appear in buffer'", 0)
	
	// Now execute a regular command that should write to our buffer
	_ = executor.Execute("echo 'Should appear in buffer'", 0)
	
	// Check that the buffer contains only the output from the second command
	if !strings.Contains(buf.String(), "Should appear in buffer") {
		t.Errorf("Expected buffer to contain 'Should appear in buffer', got: %s", buf.String())
	}
	if strings.Contains(buf.String(), "Should not appear in buffer") {
		t.Errorf("Buffer should not contain output from ExecuteWithOutput, got: %s", buf.String())
	}
}

// TestMockExecutorExecuteWithOutput tests the ExecuteWithOutput method of MockExecutor
func TestMockExecutorExecuteWithOutput(t *testing.T) {
	// Create a mock executor
	mockExec := NewMockExecutor()
	
	// Test with echo command
	output, err := mockExec.ExecuteWithOutput("echo 'Test Output'", 0)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !strings.Contains(output, "Test Output") {
		t.Errorf("Expected output to contain 'Test Output', got: %s", output)
	}
	
	// Test with custom result
	mockExec.AddCommandResult("custom-command", "Custom output", nil)
	output, err = mockExec.ExecuteWithOutput("custom-command", 0)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if output != "Custom output" {
		t.Errorf("Expected output 'Custom output', got: %s", output)
	}
	
	// Test with error result
	expectedErr := fmt.Errorf("custom error")
	mockExec.AddCommandResult("error-command", "", expectedErr)
	output, err = mockExec.ExecuteWithOutput("error-command", 0)
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
	if output != "" {
		t.Errorf("Expected empty output, got: %s", output)
	}
	
	// Test with sleep command to ensure it's properly handled
	_, err = mockExec.ExecuteWithOutput("sleep 0.1", 0)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

// TestExtractEchoContent tests the extractEchoContent function
func TestExtractEchoContent(t *testing.T) {
	tests := []struct {
		name     string
		cmdStr   string
		expected string
	}{
		{
			name:     "No echo command",
			cmdStr:   "ls -la",
			expected: "",
		},
		{
			name:     "Echo with single quotes",
			cmdStr:   "echo 'Hello, World!'",
			expected: "Hello, World!",
		},
		{
			name:     "Echo with double quotes",
			cmdStr:   "echo \"Hello, World!\"",
			expected: "Hello, World!",
		},
		{
			name:     "Echo without quotes",
			cmdStr:   "echo Hello, World!",
			expected: "Hello, World!",
		},
		{
			name:     "Echo with unclosed single quote",
			cmdStr:   "echo 'Hello, World!",
			expected: "",
		},
		{
			name:     "Echo with unclosed double quote",
			cmdStr:   "echo \"Hello, World!",
			expected: "",
		},
		{
			name:     "Echo with extra whitespace",
			cmdStr:   "echo    Hello, World!",
			expected: "Hello, World!",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractEchoContent(tt.cmdStr)
			if got != tt.expected {
				t.Errorf("extractEchoContent(%q) = %q, want %q", tt.cmdStr, got, tt.expected)
			}
		})
	}
}
