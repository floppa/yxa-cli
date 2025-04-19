package cli

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/floppa/yxa-cli/internal/config"
	"github.com/floppa/yxa-cli/internal/executor"
)

// SafeWriter is a thread-safe writer implementation
type SafeWriter struct {
	writer io.Writer
	prefix string
	buffer strings.Builder
	mutex  sync.Mutex
}

// NewSafeWriter creates a new safe writer with a prefix
func NewSafeWriter(writer io.Writer, prefix string) *SafeWriter {
	return &SafeWriter{
		writer: writer,
		prefix: prefix,
	}
}

// Write appends data to the buffer in a thread-safe manner
func (w *SafeWriter) Write(p []byte) (n int, err error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	// Write to the buffer
	return w.buffer.Write(p)
}

// flushLocked is an internal method that flushes the buffer without locking
// It must be called with the mutex already locked
func (w *SafeWriter) flushLocked() error {
	// Get the buffered content
	content := w.buffer.String()
	if content == "" {
		return nil
	}

	// Reset the buffer
	w.buffer.Reset()

	// Split the content by newlines
	lines := strings.Split(content, "\n")

	// Process each line
	for i, line := range lines {
		// Skip the last empty line that results from a trailing newline
		if i == len(lines)-1 && line == "" {
			continue
		}

		// Write the line with the prefix
		_, err := fmt.Fprintf(w.writer, "%s%s\n", w.prefix, line)
		if err != nil {
			return err
		}
	}

	return nil
}

// Flush writes the buffered data to the underlying writer with the prefix
func (w *SafeWriter) Flush() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	return w.flushLocked()
}

// syncWrite writes to the output in a thread-safe manner
func syncWrite(writer io.Writer, format string, args ...interface{}) {
	// Use a mutex to protect access to the shared writer
	outputMutex.Lock()
	defer outputMutex.Unlock()

	// Write the formatted string to the writer
	_, err := fmt.Fprintf(writer, format, args...)
	if err != nil {
		// Log the error but don't fail the command
		fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
	}
}

// outputMutex protects access to the shared output writer
var outputMutex sync.Mutex

// executeParallelCommands executes multiple commands in parallel
func (h *CommandHandler) executeParallelCommands(cmdName string, cmd config.Command, timeout time.Duration) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(cmd.Commands))

	// We'll use a mutex to protect access to the shared output writer

	// Create a context with timeout if specified
	var ctx context.Context
	var cancel context.CancelFunc
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
		defer cancel()
	} else {
		ctx, cancel = context.WithCancel(context.Background())
		defer cancel()
	}

	// Start all commands in parallel
	for name, cmdStr := range cmd.Commands {
		wg.Add(1)
		go func(name, cmdStr string) {
			defer wg.Done()

			// Replace variables in the command
			cmdStr = h.replaceVariablesInString(cmdStr, nil)
			// Log the command execution to stdout so it's visible in the main output
			syncWrite(h.Executor.GetStdout(), "Executing parallel sub-command '%s' for '%s'...\n", name, cmdName)

			// Create a dedicated buffer for each command
			cmdOutputBuffer := &bytes.Buffer{}

			// Create a local executor with prefixed output
			localExecutor := executor.NewDefaultExecutor()
			localExecutor.SetStdout(cmdOutputBuffer)
			localExecutor.SetStderr(cmdOutputBuffer)

			// Use the syncWrite helper for thread-safe output
			syncWrite(h.Executor.GetStdout(), "[%s] Starting execution...\n", name)

			// Create a channel for command completion
			done := make(chan error, 1)
			go func() {
				// Execute the command and capture its output
				_, err := localExecutor.ExecuteWithOutput(cmdStr, timeout)

				// Get the buffered output
				output := cmdOutputBuffer.String()

				// Use the syncWrite helper for thread-safe output
				if output != "" {
					syncWrite(h.Executor.GetStdout(), "[%s] %s\n", name, output)
				}

				// Send the error (if any) to the done channel
				done <- err
			}()

			// Wait for command completion or timeout
			select {
			case err := <-done:
				if err != nil {
					errChan <- fmt.Errorf("sub-command '%s' for '%s' failed: %v", name, cmdName, err)
				}
			case <-ctx.Done():

				// Command timed out or context was canceled
				errChan <- fmt.Errorf("sub-command '%s' for '%s' timed out after %s", name, cmdName, timeout)
			}
		}(name, cmdStr)
	}

	// Wait for all commands to finish
	wg.Wait()
	close(errChan)

	// No need for any final output handling

	// Collect errors
	var errors []string
	for err := range errChan {
		errors = append(errors, err.Error())
	}

	// Return combined errors if any
	if len(errors) > 0 {
		return fmt.Errorf("one or more parallel commands failed:\n%s", strings.Join(errors, "\n"))
	}

	return nil
}
