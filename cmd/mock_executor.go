package cmd

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/floppa/yxa-cli/config"
)

// MockExecutor is a mock implementation of CommandExecutor for testing
type MockExecutor struct {
	// Stdout and Stderr for the executor
	Stdout io.Writer
	Stderr io.Writer
	
	// OutputBuffer captures all command outputs for verification in tests
	OutputBuffer *bytes.Buffer
	
	// ExecutedCommands tracks which commands were executed
	ExecutedCommands []string
	
	// CommandResults maps command strings to their expected results
	CommandResults map[string]struct {
		Output string
		Error  error
	}
	
	// Config is the test configuration to use for testing
	Config *config.ProjectConfig
	
	// mutex protects concurrent access to shared resources
	mutex sync.Mutex
}

// NewMockExecutor creates a new MockExecutor for testing
func NewMockExecutor() *MockExecutor {
	// Create a buffer to capture output
	outputBuffer := &bytes.Buffer{}
	
	return &MockExecutor{
		Stdout:           outputBuffer,
		Stderr:           outputBuffer,
		OutputBuffer:     outputBuffer,
		ExecutedCommands: []string{},
		CommandResults:   make(map[string]struct{Output string; Error error}),
		Config:           createDefaultTestConfig(),
	}
}

// createDefaultTestConfig creates a default test configuration
func createDefaultTestConfig() *config.ProjectConfig {
	return &config.ProjectConfig{
		Name: "test-project",
		Commands: map[string]config.Command{
			"simple": {
				Run:         "echo 'Simple command'",
				Description: "A simple command",
			},
			"with-pre": {
				Pre:         "echo 'Pre-hook'",
				Run:         "echo 'Main command with Pre-hook'",
				Description: "Command with pre-hook",
			},
			"with-post": {
				Run:         "echo 'Main command with Post-hook'",
				Post:        "echo 'Post-hook'",
				Description: "Command with post-hook",
			},
			"with-pre-post": {
				Pre:         "echo 'Pre-hook'",
				Run:         "echo 'Main command with Pre-hook and Post-hook'",
				Post:        "echo 'Post-hook'",
				Description: "Command with pre and post hooks",
			},
			"with-timeout": {
				Run:         "sleep 10",
				Timeout:     "50ms",
				Description: "Command with timeout",
			},
			"with-condition-true": {
				Run:         "echo 'This should run'",
				Condition:   "equal HOME /Users/magnuseriksson",
				Description: "Command with true condition",
			},
			"with-condition-false": {
				Run:         "echo 'This should not run'",
				Condition:   "equal NONEXISTENT_VAR some_value",
				Description: "Command with false condition",
			},
			"task-aggregator": {
				Depends:     []string{"simple"},
				Run:         "echo 'Task aggregator command'",
				Description: "Command that depends on other commands",
			},
		},
	}
}

// GetStdout returns the stdout writer
func (m *MockExecutor) GetStdout() io.Writer {
	return m.Stdout
}

// GetStderr returns the stderr writer
func (m *MockExecutor) GetStderr() io.Writer {
	return m.Stderr
}

// SetStdout sets the stdout writer
func (m *MockExecutor) SetStdout(w io.Writer) {
	m.Stdout = w
}

// SetStderr sets the stderr writer
func (m *MockExecutor) SetStderr(w io.Writer) {
	m.Stderr = w
}

// SetConfig sets the test configuration for the mock executor
func (m *MockExecutor) SetConfig(config *config.ProjectConfig) {
	m.Config = config
}

// GetOutput returns the contents of the output buffer
func (m *MockExecutor) GetOutput() string {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.OutputBuffer.String()
}

// ClearOutput clears the output buffer
func (m *MockExecutor) ClearOutput() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.OutputBuffer.Reset()
}

// AddCommandResult sets the expected result for a command
func (m *MockExecutor) AddCommandResult(cmdStr string, output string, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	if m.CommandResults == nil {
		m.CommandResults = make(map[string]struct {
			Output string
			Error  error
		})
	}
	m.CommandResults[cmdStr] = struct {
		Output string
		Error  error
	}{
		Output: output,
		Error:  err,
	}
}

// handleCommand is a helper function that handles common command processing logic
// for both Execute and ExecuteWithOutput methods
func (m *MockExecutor) handleCommand(cmdStr string, timeout time.Duration) (string, error) {
	// Record that this command was executed
	m.ExecutedCommands = append(m.ExecutedCommands, cmdStr)
	
	// Check if we have a predefined result for this command
	if result, ok := m.CommandResults[cmdStr]; ok {
		return result.Output, result.Error
	}
	
	// Fast path for timeout simulation with sleep commands
	if timeout > 0 && strings.Contains(cmdStr, "sleep") {
		return "", fmt.Errorf("command timed out after %s", timeout)
	}
	
	// Fast path for echo commands
	if strings.Contains(cmdStr, "echo") {
		// Extract the content being echoed
		content := extractEchoContent(cmdStr)
		if content != "" {
			return content + "\n", nil
		}
	}
	
	// Default behavior for unhandled commands
	return "", nil
}

// Execute mocks executing a command and returns the predefined result
func (m *MockExecutor) Execute(cmdStr string, timeout time.Duration) error {
	// Lock the mutex to protect concurrent access to shared resources
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Use the common handler to process the command
	output, err := m.handleCommand(cmdStr, timeout)
	
	// If there's output and no error, write it to stdout
	if output != "" && err == nil && m.Stdout != nil {
		_, writeErr := fmt.Fprint(m.Stdout, output)
		if writeErr != nil {
			return fmt.Errorf("failed to write output: %w", writeErr)
		}
	}
	
	return err
}

// ExecuteWithOutput mocks executing a command and returns the predefined output
func (m *MockExecutor) ExecuteWithOutput(cmdStr string, timeout time.Duration) (string, error) {
	// Lock the mutex to protect concurrent access to shared resources
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Use the common handler to process the command
	output, err := m.handleCommand(cmdStr, timeout)
	
	// If there's output and no error, also write to the output buffer
	if output != "" && err == nil && m.OutputBuffer != nil {
		_, _ = fmt.Fprint(m.OutputBuffer, output)
	}
	
	return output, err
}

// Helper function to extract echo content from a command
func extractEchoContent(cmdStr string) string {
	// Fast path for common echo patterns
	start := strings.Index(cmdStr, "echo")
	if start < 0 {
		return ""
	}
	
	// Skip 'echo' and any whitespace
	start += 4
	for start < len(cmdStr) && (cmdStr[start] == ' ' || cmdStr[start] == '\t') {
		start++
	}
	
	// Check for quoted content
	if start < len(cmdStr) {
		switch cmdStr[start] {
		case '\'': // Single quotes
			end := strings.Index(cmdStr[start+1:], "'")
			if end >= 0 {
				return cmdStr[start+1 : start+1+end]
			}
		case '"': // Double quotes
			end := strings.Index(cmdStr[start+1:], `"`)
			if end >= 0 {
				return cmdStr[start+1 : start+1+end]
			}
		default: // Unquoted content
			// Take the rest of the string
			return strings.TrimSpace(cmdStr[start:])
		}
	}
	
	return ""
}


