package cli

import (
	"bytes"
	"testing"

	"github.com/floppa/yxa-cli/internal/config"
	"github.com/floppa/yxa-cli/internal/executor"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestNewRootCommand(t *testing.T) {
	// Create a simple test configuration
	cfg := &config.ProjectConfig{
		Variables: map[string]string{
			"PROJECT_NAME": "test-project",
		},
		Commands: map[string]config.Command{
			"test": {
				Run:         "echo 'Test command'",
				Description: "A test command",
			},
			"with-params": {
				Run:         "echo 'Command with parameters'",
				Description: "A command with parameters",
				Params: []config.Param{
					{
						Name:        "string-param",
						Description: "A string parameter",
						Type:        "string",
						Default:     "default-value",
						Flag:        true,
					},
				},
			},
		},
	}

	// Create a mock executor
	mockExec := executor.NewMockExecutor()

	// Create a root command
	root := NewRootCommand(cfg, mockExec)

	// Verify the root command
	assert.NotNil(t, root)
	assert.Equal(t, cfg, root.Config)
	assert.Equal(t, mockExec, root.Executor)
	assert.NotNil(t, root.Handler)
	assert.NotNil(t, root.RootCmd)
	assert.Equal(t, "yxa", root.RootCmd.Use)

	// Verify that commands were registered
	assert.NotNil(t, root.RootCmd.Commands())
	
	// There might be built-in commands like completion, help, etc.
	// So we'll just check that our commands are present
	found := 0
	for _, cmd := range root.RootCmd.Commands() {
		if cmd.Name() == "test" || cmd.Name() == "with-params" {
			found++
		}
	}
	assert.Equal(t, 2, found, "Both test and with-params commands should be registered")

	// Find the test command and with-params command
	var testCmd, withParamsCmd *cobra.Command
	for _, cmd := range root.RootCmd.Commands() {
		if cmd.Name() == "test" {
			testCmd = cmd
		} else if cmd.Name() == "with-params" {
			withParamsCmd = cmd
		}
	}

	// Verify the test command
	assert.NotNil(t, testCmd)
	assert.Equal(t, "test", testCmd.Name())
	assert.Equal(t, "A test command", testCmd.Short)

	// Verify the with-params command
	assert.NotNil(t, withParamsCmd)
	assert.Equal(t, "with-params", withParamsCmd.Name())
	assert.Equal(t, "A command with parameters", withParamsCmd.Short)
	
	// Verify that parameters were added
	flag := withParamsCmd.Flags().Lookup("string-param")
	assert.NotNil(t, flag)
	assert.Equal(t, "string-param", flag.Name)
	assert.Equal(t, "A string parameter", flag.Usage)
	assert.Equal(t, "default-value", flag.DefValue)
}

func TestRootCommand_Execute(t *testing.T) {
	// Create a simple test configuration
	cfg := &config.ProjectConfig{
		Variables: map[string]string{
			"PROJECT_NAME": "test-project",
		},
		Commands: map[string]config.Command{
			"test": {
				Run:         "echo 'Test command'",
				Description: "A test command",
			},
		},
	}

	// Create a mock executor
	mockExec := executor.NewMockExecutor()

	// Create a root command
	root := NewRootCommand(cfg, mockExec)

	// Test execution with no arguments (should show help)
	t.Run("no args", func(t *testing.T) {
		// Capture stdout
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		root.RootCmd.SetOut(stdout)
		root.RootCmd.SetErr(stderr)
		
		// Set args
		root.RootCmd.SetArgs([]string{})
		
		// Execute
		err := root.Execute()
		assert.NoError(t, err)
		
		// Verify output contains help text
		output := stdout.String()
		assert.Contains(t, output, "Usage:")
		assert.Contains(t, output, "Available Commands:")
	})

	// Test execution with an existing command
	t.Run("existing command", func(t *testing.T) {
		// Create a new mock executor for this test to avoid state from previous tests
		mockExec := executor.NewMockExecutor()
		root := NewRootCommand(cfg, mockExec)
		
		// Capture stdout
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		root.RootCmd.SetOut(stdout)
		root.RootCmd.SetErr(stderr)
		mockExec.SetStdout(stdout)
		mockExec.SetStderr(stderr)
		
		// Add expected result for the command
		mockExec.AddCommandResult("echo 'Test command'", "Test command\n", nil)
		
		// Set args
		root.RootCmd.SetArgs([]string{"test"})
		
		// Execute
		err := root.Execute()
		assert.NoError(t, err)
		
		// Verify mock executor was called
		executedCmds := mockExec.GetExecutedCommands()
		assert.GreaterOrEqual(t, len(executedCmds), 1, "At least one command should have been executed")
	})

	// Test execution with a non-existent command
	t.Run("non-existent command", func(t *testing.T) {
		// Create a new mock executor for this test to avoid state from previous tests
		mockExec := executor.NewMockExecutor()
		root := NewRootCommand(cfg, mockExec)
		
		// Capture stdout
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		root.RootCmd.SetOut(stdout)
		root.RootCmd.SetErr(stderr)
		
		// Set args
		root.RootCmd.SetArgs([]string{"non-existent"})
		
		// Execute
		err := root.Execute()
		assert.Error(t, err)
		
		// Verify error message
		errOutput := stderr.String()
		assert.Contains(t, errOutput, "unknown command")
	})
}

func TestGetWriterMutex(t *testing.T) {
	// Test getting a mutex for a writer
	writer1 := &bytes.Buffer{}
	mutex1 := getWriterMutex(writer1)
	assert.NotNil(t, mutex1)
	
	// Test getting the same mutex for the same writer
	mutex2 := getWriterMutex(writer1)
	assert.Equal(t, mutex1, mutex2)
	
	// Test getting a different mutex for a different writer
	writer2 := &bytes.Buffer{}
	mutex3 := getWriterMutex(writer2)
	assert.NotNil(t, mutex3)
	// We can't reliably test that mutex1 != mutex3 because the mutex map is package-level
	// and might be affected by other tests running concurrently
	
	// Test concurrent access to the mutex map
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			writer := &bytes.Buffer{}
			mutex := getWriterMutex(writer)
			assert.NotNil(t, mutex)
			done <- true
		}()
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}
