package executor

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewDefaultExecutor(t *testing.T) {
	// Test that the constructor creates an executor with stdout and stderr set to os.Stdout and os.Stderr
	executor := NewDefaultExecutor()
	
	assert.Equal(t, os.Stdout, executor.GetStdout(), "Default stdout should be os.Stdout")
	assert.Equal(t, os.Stderr, executor.GetStderr(), "Default stderr should be os.Stderr")
}

func TestDefaultExecutor_GettersAndSetters(t *testing.T) {
	// Create a default executor
	executor := NewDefaultExecutor()
	
	// Test the default values
	assert.Equal(t, os.Stdout, executor.GetStdout(), "Default stdout should be os.Stdout")
	assert.Equal(t, os.Stderr, executor.GetStderr(), "Default stderr should be os.Stderr")
	
	// Create custom writers
	customStdout := &bytes.Buffer{}
	customStderr := &bytes.Buffer{}
	
	// Test SetStdout
	executor.SetStdout(customStdout)
	assert.Equal(t, customStdout, executor.GetStdout(), "Stdout should be set to custom writer")
	
	// Test SetStderr
	executor.SetStderr(customStderr)
	assert.Equal(t, customStderr, executor.GetStderr(), "Stderr should be set to custom writer")
	
	// Test concurrent access to ensure mutex works properly
	wg := sync.WaitGroup{}
	wg.Add(2)
	
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			executor.SetStdout(customStdout)
			_ = executor.GetStdout()
		}
	}()
	
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			executor.SetStderr(customStderr)
			_ = executor.GetStderr()
		}
	}()
	
	wg.Wait()
}

func TestDefaultExecutor_Execute(t *testing.T) {
	tests := []struct {
		name       string
		cmdStr     string
		timeout    time.Duration
		wantErr    bool
		wantOutput string
	}{
		{
			name:       "echo command",
			cmdStr:     "echo 'Hello, World!'",
			timeout:    0,
			wantErr:    false,
			wantOutput: "Hello, World!\n",
		},
		{
			name:       "command with error",
			cmdStr:     "exit 1",
			timeout:    0,
			wantErr:    true,
			wantOutput: "",
		},
		{
			name:       "command with timeout",
			cmdStr:     "sleep 2",
			timeout:    50 * time.Millisecond,
			wantErr:    true,
			wantOutput: "",
		},
		{
			name:       "invalid command",
			cmdStr:     "command_that_does_not_exist",
			timeout:    0,
			wantErr:    true,
			wantOutput: "",
		},
		{
			name:       "command with stderr output",
			cmdStr:     "echo 'Error message' >&2",
			timeout:    0,
			wantErr:    false,
			wantOutput: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a buffer to capture output
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			// Create executor with buffers
			e := &DefaultExecutor{
				Stdout: &stdout,
				Stderr: &stderr,
			}

			// Execute the command
			err := e.Execute(tt.cmdStr, tt.timeout)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("DefaultExecutor.Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// For timeout errors, check the error message
			if tt.timeout > 0 && err != nil {
				if !strings.Contains(err.Error(), "timed out") && !strings.Contains(err.Error(), "signal: killed") {
					t.Errorf("Expected timeout error, got: %v", err)
				}
			}

			// Check output for echo commands
			if strings.HasPrefix(tt.cmdStr, "echo") && !strings.Contains(tt.cmdStr, ">&2") && !tt.wantErr {
				if stdout.String() != tt.wantOutput {
					t.Errorf("DefaultExecutor.Execute() output = %q, want %q", stdout.String(), tt.wantOutput)
				}
			}
			
			// Check stderr output for commands redirecting to stderr
			if strings.Contains(tt.cmdStr, ">&2") && !tt.wantErr {
				if !strings.Contains(stderr.String(), "Error message") {
					t.Errorf("Expected stderr to contain 'Error message', got: %q", stderr.String())
				}
			}
		})
	}
}

func TestDefaultExecutor_ExecuteWithOutput(t *testing.T) {
	tests := []struct {
		name       string
		cmdStr     string
		timeout    time.Duration
		wantErr    bool
		wantOutput string
	}{
		{
			name:       "echo command",
			cmdStr:     "echo 'Hello, World!'",
			timeout:    0,
			wantErr:    false,
			wantOutput: "Hello, World!\n",
		},
		{
			name:       "command with error",
			cmdStr:     "exit 1",
			timeout:    0,
			wantErr:    true,
			wantOutput: "",
		},
		{
			name:       "command with timeout",
			cmdStr:     "sleep 2",
			timeout:    50 * time.Millisecond,
			wantErr:    true,
			wantOutput: "",
		},
		{
			name:       "command with stderr output",
			cmdStr:     "echo 'Error message' >&2 && echo 'Standard output'",
			timeout:    0,
			wantErr:    false,
			wantOutput: "Standard output\n", // Only stdout should be returned
		},
		{
			name:       "multiline output",
			cmdStr:     "echo 'Line 1' && echo 'Line 2'",
			timeout:    0,
			wantErr:    false,
			wantOutput: "Line 1\nLine 2\n",
		},
		{
			name:       "invalid command",
			cmdStr:     "command_that_does_not_exist",
			timeout:    0,
			wantErr:    true,
			wantOutput: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a buffer to capture stderr (for debugging)
			var stderr bytes.Buffer

			// Create executor
			e := &DefaultExecutor{
				Stdout: io.Discard, // Discard normal output since we're capturing it directly
				Stderr: &stderr,
			}

			// Execute the command
			output, err := e.ExecuteWithOutput(tt.cmdStr, tt.timeout)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("DefaultExecutor.ExecuteWithOutput() error = %v, wantErr %v, stderr: %s", 
					err, tt.wantErr, stderr.String())
				return
			}

			// For timeout errors, check the error message
			if tt.timeout > 0 && err != nil {
				if !strings.Contains(err.Error(), "timed out") && !strings.Contains(err.Error(), "signal: killed") {
					t.Errorf("Expected timeout error, got: %v", err)
				}
			}

			// Check output for successful commands
			if !tt.wantErr && tt.wantOutput != "" {
				if output != tt.wantOutput {
					t.Errorf("DefaultExecutor.ExecuteWithOutput() output = %q, want %q", output, tt.wantOutput)
				}
			}
		})
	}

	// Test with custom stdout writer
	t.Run("with custom stdout", func(t *testing.T) {
		// Create a buffer to capture stdout
		var stdout bytes.Buffer
		var stderr bytes.Buffer

		// Create executor with custom stdout
		e := &DefaultExecutor{
			Stdout: &stdout,
			Stderr: &stderr,
		}

		// Execute a command
		output, err := e.ExecuteWithOutput("echo 'Test with custom stdout'", 0)
		assert.NoError(t, err)
		assert.Equal(t, "Test with custom stdout\n", output)

		// The output should also be written to our custom stdout
		assert.Equal(t, "Test with custom stdout\n", stdout.String())
	})
}

// Test timeout behavior more thoroughly
func TestExecuteWithTimeout(t *testing.T) {
	tests := []struct {
		name    string
		cmdStr  string
		timeout time.Duration
		wantErr bool
	}{
		{
			name:    "command completes before timeout",
			cmdStr:  "echo 'Quick command'",
			timeout: 1 * time.Second,
			wantErr: false,
		},
		{
			name:    "command times out",
			cmdStr:  "sleep 2",
			timeout: 100 * time.Millisecond,
			wantErr: true,
		},
		{
			name:    "zero timeout means no timeout",
			cmdStr:  "sleep 0.5",
			timeout: 0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create executor
			e := &DefaultExecutor{
				Stdout: io.Discard,
				Stderr: io.Discard,
			}

			// Test with Execute
			err := e.Execute(tt.cmdStr, tt.timeout)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() with timeout error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && err != nil {
				if !strings.Contains(err.Error(), "timed out") && !strings.Contains(err.Error(), "signal: killed") {
					t.Errorf("Expected timeout error, got: %v", err)
				}
			}

			// Test with ExecuteWithOutput
			_, err = e.ExecuteWithOutput(tt.cmdStr, tt.timeout)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteWithOutput() with timeout error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && err != nil {
				if !strings.Contains(err.Error(), "timed out") && !strings.Contains(err.Error(), "signal: killed") {
					t.Errorf("Expected timeout error, got: %v", err)
				}
			}
		})
	}
}

func TestMockExecutor(t *testing.T) {
	// Create a mock executor
	m := NewMockExecutor()

	// Add some command results
	m.AddCommandResult("echo 'test'", "test output", nil)
	m.AddCommandResult("failing command", "", fmt.Errorf("command failed"))

	// Test successful command
	err := m.Execute("echo 'test'", 0)
	if err != nil {
		t.Errorf("MockExecutor.Execute() error = %v, wantErr %v", err, false)
	}
	if !strings.Contains(m.GetOutput(), "test output") {
		t.Errorf("MockExecutor.Execute() output = %q, want to contain %q", m.GetOutput(), "test output")
	}

	// Clear output
	m.ClearOutput()

	// Test failing command
	err = m.Execute("failing command", 0)
	if err == nil {
		t.Error("MockExecutor.Execute() error = nil, wantErr true")
	}

	// Test command with output
	output, err := m.ExecuteWithOutput("echo 'test'", 0)
	if err != nil {
		t.Errorf("MockExecutor.ExecuteWithOutput() error = %v, wantErr %v", err, false)
	}
	if output != "test output" {
		t.Errorf("MockExecutor.ExecuteWithOutput() output = %q, want %q", output, "test output")
	}

	// Test echo command extraction
	if extractEchoContent("echo 'Hello'") != "Hello" {
		t.Errorf("extractEchoContent() failed to extract content")
	}
	if extractEchoContent("echo \"Hello\"") != "Hello" {
		t.Errorf("extractEchoContent() failed to extract content with double quotes")
	}
	if extractEchoContent("echo Hello") != "Hello" {
		t.Errorf("extractEchoContent() failed to extract unquoted content")
	}
}
