package executor

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
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

// GetExecutedCommands returns the list of executed commands
func (m *MockExecutor) GetExecutedCommands() []string {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Return a copy to avoid race conditions
	commands := make([]string, len(m.ExecutedCommands))
	copy(commands, m.ExecutedCommands)
	return commands
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
	m.mutex.Lock()
	m.ExecutedCommands = append(m.ExecutedCommands, cmdStr)
	m.mutex.Unlock()
	
	// Check if we have a predefined result for this command
	m.mutex.Lock()
	result, ok := m.CommandResults[cmdStr]
	m.mutex.Unlock()
	if ok {
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
	// Use the common handler to process the command
	output, err := m.handleCommand(cmdStr, timeout)
	
	// If there's output and no error, write it to stdout
	if output != "" && err == nil {
		m.mutex.Lock()
		if m.Stdout != nil {
			_, _ = fmt.Fprint(m.Stdout, output)
		}
		if m.OutputBuffer != nil {
			_, _ = fmt.Fprint(m.OutputBuffer, output)
		}
		m.mutex.Unlock()
	}
	
	return err
}

// ExecuteWithOutput mocks executing a command and returns the predefined output
func (m *MockExecutor) ExecuteWithOutput(cmdStr string, timeout time.Duration) (string, error) {
	// Use the common handler to process the command
	output, err := m.handleCommand(cmdStr, timeout)
	
	// Record the output in the buffer for inspection
	if output != "" {
		m.mutex.Lock()
		if m.OutputBuffer != nil {
			_, _ = fmt.Fprint(m.OutputBuffer, output)
		}
		m.mutex.Unlock()
	}
	
	return output, err
}

// Helper function to extract echo content from a command
func extractEchoContent(cmdStr string) string {
	// Simple parsing for echo commands
	// This is a basic implementation and may not handle all cases correctly
	
	// Check if it's an echo command
	if !strings.HasPrefix(cmdStr, "echo") {
		return ""
	}
	
	// Remove the echo command
	content := strings.TrimPrefix(cmdStr, "echo")
	
	// Handle quoted content
	content = strings.TrimSpace(content)
	
	// Handle single quotes
	if strings.HasPrefix(content, "'") && strings.HasSuffix(content, "'") {
		return content[1 : len(content)-1]
	}
	
	// Handle double quotes
	if strings.HasPrefix(content, "\"") && strings.HasSuffix(content, "\"") {
		return content[1 : len(content)-1]
	}
	
	// Handle unquoted content
	return content
}
