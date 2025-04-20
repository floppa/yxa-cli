package cli

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/floppa/yxa-cli/internal/config"
	"github.com/floppa/yxa-cli/internal/executor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSafeWriter tests the thread-safe writer implementation
func TestSafeWriter(t *testing.T) {
	t.Run("Basic Writing and Flushing", func(t *testing.T) {
		var buf bytes.Buffer
		writer := NewSafeWriter(&buf, "[prefix] ")

		// Write some data
		n, err := writer.Write([]byte("line1\nline2"))
		require.NoError(t, err)
		assert.Equal(t, 11, n)

		// Flush the buffer
		err = writer.Flush()
		require.NoError(t, err)

		// Check the output
		expected := "[prefix] line1\n[prefix] line2\n"
		assert.Equal(t, expected, buf.String())

		// Write more data
		n, err = writer.Write([]byte("line3"))
		require.NoError(t, err)
		assert.Equal(t, 5, n)

		// Flush again
		err = writer.Flush()
		require.NoError(t, err)

		// Check the updated output
		expected += "[prefix] line3\n"
		assert.Equal(t, expected, buf.String())
	})

	t.Run("Empty Buffer Flush", func(t *testing.T) {
		var buf bytes.Buffer
		writer := NewSafeWriter(&buf, "[prefix] ")

		// Flush an empty buffer
		err := writer.Flush()
		require.NoError(t, err)

		// Buffer should still be empty
		assert.Equal(t, "", buf.String())
	})

	t.Run("Concurrent Writing", func(t *testing.T) {
		var buf bytes.Buffer
		writer := NewSafeWriter(&buf, "[prefix] ")

		// Create a wait group to synchronize goroutines
		var wg sync.WaitGroup

		// Number of concurrent writers
		numWriters := 10

		// Launch multiple goroutines to write concurrently
		for i := 0; i < numWriters; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				// Write data specific to this goroutine
				data := fmt.Sprintf("data from goroutine %d", id)
				_, err := writer.Write([]byte(data))
				assert.NoError(t, err)
			}(i)
		}

		// Wait for all writers to complete
		wg.Wait()

		// Flush the buffer
		err := writer.Flush()
		require.NoError(t, err)

		// Check that all data was written
		output := buf.String()
		for i := 0; i < numWriters; i++ {
			assert.Contains(t, output, fmt.Sprintf("data from goroutine %d", i))
		}
	})
}

// We're using the MockExecutor from the executor package

func TestExecuteParallelCommands_Timeout(t *testing.T) {
	buf := &bytes.Buffer{}
	exec := executor.NewDefaultExecutor()
	exec.SetStdout(buf)
	exec.SetStderr(buf)

	cfg := &config.ProjectConfig{
		Name: "test-project",
		Commands: map[string]config.Command{
			"parallel-timeout": {
				Parallel: true,
				Timeout:  "100ms",
				Commands: map[string]string{
					"slow1": "sleep 1",
					"slow2": "sleep 1",
				},
			},
		},
	}
	handler := NewCommandHandler(cfg, exec)
	err := handler.ExecuteCommand("parallel-timeout", nil)
	if err == nil || !strings.Contains(err.Error(), "timed out") {
		t.Errorf("Expected timeout error, got %v", err)
	}
}

// TestExecuteParallelCommands tests the parallel command execution functionality
func TestExecuteParallelCommands(t *testing.T) {
	t.Run("Successful Parallel Execution", func(t *testing.T) {
		// Create a mock executor
		realExec := executor.NewDefaultExecutor()

		// Use a buffer for output
		buf := &bytes.Buffer{}
		realExec.SetStdout(buf)
		realExec.SetStderr(buf)

		// Create a command handler with the real executor
		handler := &CommandHandler{
			Config: &config.ProjectConfig{
				Variables: map[string]string{
					"VAR1": "value1",
					"VAR2": "value2",
				},
			},
			Executor:     realExec,
			executedCmds: make(map[string]bool),
		}

		// Create a command with parallel sub-commands
		cmd := config.Command{
			Commands: map[string]string{
				"cmd1": "echo $VAR1",
				"cmd2": "echo $VAR2",
				"cmd3": "echo test",
			},
			Parallel: true,
		}

		// Execute the parallel commands
		err := handler.executeParallelCommands("test-parallel", cmd, 0)
		assert.NoError(t, err)

		// Check the output contains the expected content
		output := buf.String()
		t.Logf("Output: %s", output)
		assert.Contains(t, output, "value1")
		assert.Contains(t, output, "value2")
		assert.Contains(t, output, "test")
		buf.Reset()

	})

	t.Run("Parallel Execution With Errors", func(t *testing.T) {
		// Use a buffer for output
		realExec := executor.NewDefaultExecutor()
		buf := &bytes.Buffer{}
		realExec.SetStdout(buf)
		realExec.SetStderr(buf)

		// Create a command handler with the real executor
		handler := &CommandHandler{
			Config: &config.ProjectConfig{
				Variables: map[string]string{},
			},
			Executor:     realExec,
			executedCmds: make(map[string]bool),
		}

		// Create a command with parallel sub-commands, one of which will fail
		cmd := config.Command{
			Commands: map[string]string{
				"cmd1": "echo success1",
				"cmd2": "false",
				"cmd3": "echo success2",
			},
			Parallel: true,
		}

		// Execute the parallel commands
		err := handler.executeParallelCommands("test-parallel-errors", cmd, 0)

		// Should return an error
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "one or more parallel commands failed")
		assert.Contains(t, err.Error(), "sub-command 'cmd2' for 'test-parallel-errors' failed")

		// Note: Output from parallel commands may not be in order or may be missing if a command fails early.
		output := buf.String()
		t.Logf("Output: %s", output)
		assert.Contains(t, output, "success1")
		assert.Contains(t, output, "success2")
		buf.Reset()

	})

	t.Run("Parallel Execution With Timeout", func(t *testing.T) {
		// Skip this test in CI environments as it's timing-dependent
		if testing.Short() {
			t.Skip("Skipping timeout test in short mode")
		}

		// Use a buffer for output
		realExec := executor.NewDefaultExecutor()
		buf := &bytes.Buffer{}
		realExec.SetStdout(buf)
		realExec.SetStderr(buf)

		// Create a command handler with the real executor
		handler := &CommandHandler{
			Config: &config.ProjectConfig{
				Variables: map[string]string{},
			},
			Executor:     realExec,
			executedCmds: make(map[string]bool),
		}

		// Create a command with parallel sub-commands, one of which is slow
		cmd := config.Command{
			Commands: map[string]string{
				"cmd1": "echo quick1",
				"cmd2": "sleep 2",
				"cmd3": "echo quick2",
			},
			Parallel: true,
		}

		// Execute the parallel commands with a short timeout
		err := handler.executeParallelCommands("test-parallel-timeout", cmd, 50*time.Millisecond)

		// Log the error for debugging
		if err != nil {
			t.Logf("Error: %v", err)
		}

		// Should return an error
		assert.Error(t, err, "Expected an error due to timeout or command failure")

		// Check the output for the quick commands
		output := buf.String()
		t.Logf("Output: %s", output)
		assert.Contains(t, output, "quick1")
		assert.Contains(t, output, "quick2")
		buf.Reset()

	})
}

// TestSafeWriterIntegration tests the SafeWriter with actual command execution
func TestSafeWriterIntegration(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("SafeWriter With Real Commands", func(t *testing.T) {
		// Create a buffer to capture output
		var buf bytes.Buffer

		// Create a SafeWriter with a prefix
		writer := NewSafeWriter(&buf, "[TEST] ")

		// Write some data
		n, err := writer.Write([]byte("line1\nline2\nline3"))
		require.NoError(t, err)
		assert.Equal(t, 17, n)

		// Flush the buffer
		err = writer.Flush()
		require.NoError(t, err)

		// Check the output
		expected := "[TEST] line1\n[TEST] line2\n[TEST] line3\n"
		assert.Equal(t, expected, buf.String())
	})
}
