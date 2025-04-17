package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"
)

// CommandExecutor defines an interface for executing shell commands
// This allows for easy mocking in tests
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

// DefaultExecutor is the standard implementation of CommandExecutor
// that actually runs commands on the system
type DefaultExecutor struct {
	Stdout io.Writer
	Stderr io.Writer
	mutex  sync.Mutex // Protects concurrent access to Stdout/Stderr
}

// NewDefaultExecutor creates a new DefaultExecutor with standard output/error
func NewDefaultExecutor() *DefaultExecutor {
	return &DefaultExecutor{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
}

// GetStdout returns the stdout writer
func (e *DefaultExecutor) GetStdout() io.Writer {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	return e.Stdout
}

// GetStderr returns the stderr writer
func (e *DefaultExecutor) GetStderr() io.Writer {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	return e.Stderr
}

// SetStdout sets the stdout writer
func (e *DefaultExecutor) SetStdout(w io.Writer) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	e.Stdout = w
}

// SetStderr sets the stderr writer
func (e *DefaultExecutor) SetStderr(w io.Writer) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	e.Stderr = w
}

// executeWithContext is a helper function that executes a command with timeout handling
// It's used internally by both Execute and ExecuteWithOutput to avoid code duplication
func executeWithContext(cmd *exec.Cmd, timeout time.Duration) error {
	// If no timeout is specified, just run the command
	if timeout == 0 {
		return cmd.Run()
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Start the command
	err := cmd.Start()
	if err != nil {
		return err
	}

	// Create a channel for the command completion
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	// Wait for either command completion or timeout
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		// Command timed out, kill it
		if err := cmd.Process.Kill(); err != nil {
			return fmt.Errorf("command timed out after %s and failed to kill process: %v", timeout, err)
		}
		return fmt.Errorf("command timed out after %s", timeout)
	}
}

// Execute runs a shell command with optional timeout
func (e *DefaultExecutor) Execute(cmdStr string, timeout time.Duration) error {
	// Lock to safely access stdout/stderr
	e.mutex.Lock()
	
	// Create a command
	cmdExec := exec.Command("sh", "-c", cmdStr) // #nosec G204
	cmdExec.Stdout = e.Stdout
	cmdExec.Stderr = e.Stderr
	
	// Unlock after setting up the command
	e.mutex.Unlock()

	// Execute the command with timeout handling
	return executeWithContext(cmdExec, timeout)
}

// ExecuteWithOutput runs a shell command and returns its output
func (e *DefaultExecutor) ExecuteWithOutput(cmdStr string, timeout time.Duration) (string, error) {
	// For thread safety, we need to use a different approach than Execute
	// We'll create a separate command and buffer for this operation
	
	// Create a buffer to capture output
	var outputBuffer bytes.Buffer
	
	// For no timeout case, use a simpler approach to avoid race conditions
	if timeout == 0 {
		// Create and configure the command
		cmdExec := exec.Command("sh", "-c", cmdStr) // #nosec G204
		cmdExec.Stdout = &outputBuffer
		cmdExec.Stderr = &outputBuffer
		
		// Run the command and wait for it to complete
		err := cmdExec.Run()
		return outputBuffer.String(), err
	}
	
	// For timeout case, we need to handle it carefully
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	// Create and configure the command with context
	cmdExec := exec.CommandContext(ctx, "sh", "-c", cmdStr) // #nosec G204
	cmdExec.Stdout = &outputBuffer
	cmdExec.Stderr = &outputBuffer
	
	// Run the command and wait for it to complete
	err := cmdExec.Run()
	
	// Check if the context was canceled (timeout)
	if ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("command timed out after %s", timeout)
	}
	
	return outputBuffer.String(), err
}
