package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/floppa/yxa-cli/internal/config"
	"github.com/floppa/yxa-cli/internal/executor"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func countUserCommands(cmd *cobra.Command) int {
	count := 0
	for _, c := range cmd.Commands() {
		if c.Name() != "help" && c.Name() != "completion" {
			count++
		}
	}
	return count
}

func TestNewRootCommand_NilConfig(t *testing.T) {
	root := NewRootCommand(nil, executor.NewDefaultExecutor())
	if root == nil || root.RootCmd == nil {
		t.Fatal("Expected RootCommand and RootCmd to be non-nil")
	}
	if countUserCommands(root.RootCmd) != 0 {
		t.Errorf("Expected no user-defined commands, got %d", countUserCommands(root.RootCmd))
	}
}

func TestNewRootCommand_EmptyCommands(t *testing.T) {
	cfg := &config.ProjectConfig{Name: "empty", Commands: map[string]config.Command{}}
	root := NewRootCommand(cfg, executor.NewDefaultExecutor())
	if countUserCommands(root.RootCmd) != 0 {
		t.Errorf("Expected no user-defined commands, got %d", countUserCommands(root.RootCmd))
	}
}

func TestRootCommand_SetupCompletion(t *testing.T) {
	cfg := &config.ProjectConfig{
		Name:     "test-project",
		Commands: map[string]config.Command{},
	}
	root := NewRootCommand(cfg, executor.NewDefaultExecutor())

	// Remove completion if already present (to test idempotency)
	for _, cmd := range root.RootCmd.Commands() {
		if cmd.Name() == "completion" {
			root.RootCmd.RemoveCommand(cmd)
		}
	}
	root.setupCompletion()

	found := false
	for _, cmd := range root.RootCmd.Commands() {
		if cmd.Name() == "completion" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'completion' command to be registered")
	}
}

func TestNewRootCommand_WithParams(t *testing.T) {
	cfg := &config.ProjectConfig{
		Commands: map[string]config.Command{
			"with-param": {
				Run: "echo ok",
				Params: []config.Param{
					{
						Name:        "flag",
						Type:        "string",
						Default:     "default",
						Description: "A test flag",
						Flag:        true,
					},
				},
			},
		},
	}
	exec := executor.NewDefaultExecutor()
	root := NewRootCommand(nil, exec)
	// Clear any existing commands to ensure a clean state for testing
	root.clearUserCommands()
	// Set the config and register commands directly
	root.Config = cfg
	root.Handler = NewCommandHandler(cfg, exec)
	root.registerCommands()

	// Find the command
	cmd, _, err := root.RootCmd.Find([]string{"with-param"})
	if err != nil {
		t.Fatalf("Find failed: %v", err)
	}
	if cmd == nil {
		t.Fatal("Command not found")
	}
	flag := cmd.Flags().Lookup("flag")
	if flag == nil {
		t.Fatalf("Expected flag 'flag' to be registered")
	}
	if flag.DefValue != "default" {
		t.Errorf("Expected default value 'default', got '%s'", flag.DefValue)
	}
}

func TestGetWriterMutex_NewAndExisting(t *testing.T) {
	b := &bytes.Buffer{}

	tests := []struct {
		name      string
		writer1   interface{}
		writer2   interface{}
		sameMutex bool
	}{
		{
			name:      "same writer",
			writer1:   b,
			writer2:   b,
			sameMutex: true,
		},
		{"different writers", &bytes.Buffer{}, &bytes.Buffer{}, false},
		{"nil writer", nil, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mtx1 := getWriterMutex(tt.writer1)
			mtx2 := getWriterMutex(tt.writer2)
			if tt.sameMutex && mtx1 != mtx2 {
				t.Errorf("Expected same mutex for same writer")
			}
			if !tt.sameMutex && mtx1 == mtx2 {
				t.Errorf("Expected different mutexes for different writers")
			}
		})
	}
}

func TestNewRootCommand(t *testing.T) {
	// Create a test config
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
	realExec := executor.NewDefaultExecutor()

	// Create a root command with a fresh config
	root := NewRootCommand(nil, realExec)
	// Clear any existing commands to ensure a clean state for testing
	root.clearUserCommands()
	// Set the config and register commands
	root.Config = cfg
	root.Handler = NewCommandHandler(cfg, realExec)
	root.registerCommands()

	// Verify the root command
	assert.NotNil(t, root)
	assert.Equal(t, cfg, root.Config)
	assert.Equal(t, realExec, root.Executor)
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
	tests := []struct {
		name        string
		args        []string
		expectError bool
		checkStdout func(t *testing.T, out string)
		checkStderr func(t *testing.T, errOut string)
	}{
		{
			name:        "no args (help)",
			args:        []string{},
			expectError: false,
			checkStdout: func(t *testing.T, out string) {
				assert.Contains(t, out, "Usage:")
				assert.Contains(t, out, "Available Commands:")
			},
		},
		{
			name:        "existing command",
			args:        []string{"test"},
			expectError: false,
			checkStdout: func(t *testing.T, out string) {
				assert.Contains(t, out, "Test command")
			},
		},
		{
			name:        "non-existent command",
			args:        []string{"non-existent"},
			expectError: true,
			checkStderr: func(t *testing.T, errOut string) {
				assert.Contains(t, errOut, "unknown command")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			realExec := executor.NewDefaultExecutor()
			realExec.SetStdout(stdout) // Set on executor instance
			realExec.SetStderr(stderr) // Set on executor instance

			root := NewRootCommand(nil, realExec)

			// Clear any existing commands to ensure a clean state for testing
			root.clearUserCommands()
			
			// Set the config and register commands directly
			root.Config = cfg
			root.Handler = NewCommandHandler(cfg, realExec)
			root.registerCommands()

			root.RootCmd.SetOut(stdout) // Set output on the command itself
			root.RootCmd.SetErr(stderr)
			root.RootCmd.SetArgs(tt.args) // Set args directly on the command

			err := root.Execute() // Execute should now find the command
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if tt.checkStdout != nil {
				out := stdout.String()
				tt.checkStdout(t, out)
			}
			if tt.checkStderr != nil {
				errOut := stderr.String()
				tt.checkStderr(t, errOut)
			}
		})
	}
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

// TestSetupCompletion tests the setupCompletion function
func TestSetupCompletion(t *testing.T) {
	// Create a root command
	exec := executor.NewDefaultExecutor()
	root := NewRootCommand(nil, exec)

	// Remove all completion commands if they exist
	var completionCmds []*cobra.Command
	for _, cmd := range root.RootCmd.Commands() {
		if cmd.Name() == "completion" {
			completionCmds = append(completionCmds, cmd)
		}
	}

	// Remove all completion commands
	for _, cmd := range completionCmds {
		root.RootCmd.RemoveCommand(cmd)
	}

	// Verify completion command was removed or didn't exist
	for _, cmd := range root.RootCmd.Commands() {
		if cmd.Name() == "completion" {
			t.Fatal("Completion command should not exist before setupCompletion")
		}
	}

	// Setup completion
	root.setupCompletion()

	// Find the completion command
	var completionCmd *cobra.Command
	for _, cmd := range root.RootCmd.Commands() {
		if cmd.Name() == "completion" {
			completionCmd = cmd
			break
		}
	}

	// Verify completion command exists
	assert.NotNil(t, completionCmd, "Completion command should exist after setupCompletion")

	// Verify completion command properties
	assert.Equal(t, "completion [bash|zsh|fish|powershell]", completionCmd.Use)
	assert.Equal(t, "Generate completion script", completionCmd.Short)
	assert.True(t, completionCmd.DisableFlagsInUseLine)
	assert.Equal(t, []string{"bash", "zsh", "fish", "powershell"}, completionCmd.ValidArgs)
}

func TestLoadConfigAndRegisterCommandsInRoot(t *testing.T) {
	// Save the original working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	defer func() {
		// Restore the original working directory
		if err := os.Chdir(originalWd); err != nil {
			t.Logf("Warning: Failed to restore original working directory: %v", err)
		}
	}()

	// Create a temporary directory for the test
	tempDir := t.TempDir()

	// Test with config flag value provided
	t.Run("with config flag value", func(t *testing.T) {
		// Create a test config file in the temp directory
		configPath := filepath.Join(tempDir, "custom-config.yml")
		configContent := `
name: test-project
commands:
  test-cmd:
    run: echo "Test command"
    description: A test command
`
		err := os.WriteFile(configPath, []byte(configContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create test config file: %v", err)
		}

		// Create a root command with nil config
		exec := executor.NewDefaultExecutor()
		root := NewRootCommand(nil, exec)

		// Load the config using the flag value
		err = root.loadConfigAndRegisterCommands(configPath)
		assert.NoError(t, err)
		assert.NotNil(t, root.Config)
		assert.Equal(t, "test-project", root.Config.Name)

		// Verify the command was registered
		cmd, _, _ := root.RootCmd.Find([]string{"test-cmd"})
		assert.NotNil(t, cmd)
	})

	// Test with local config file
	t.Run("with local config file", func(t *testing.T) {
		// Create a temporary directory and change to it
		localDir := filepath.Join(tempDir, "local")
		err := os.Mkdir(localDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create local directory: %v", err)
		}
		err = os.Chdir(localDir)
		if err != nil {
			t.Fatalf("Failed to change to local directory: %v", err)
		}

		// Create a local yxa.yml file
		localConfig := `
name: local-project
commands:
  local-cmd:
    run: echo "Local command"
    description: A local command
`
		err = os.WriteFile("yxa.yml", []byte(localConfig), 0644)
		if err != nil {
			t.Fatalf("Failed to create local config file: %v", err)
		}

		// Create a root command with nil config
		exec := executor.NewDefaultExecutor()
		root := NewRootCommand(nil, exec)

		// Load the config using empty flag value (should find local yxa.yml)
		err = root.loadConfigAndRegisterCommands("")
		assert.NoError(t, err)
		assert.NotNil(t, root.Config)
		assert.Equal(t, "local-project", root.Config.Name)

		// Verify the command was registered
		cmd, _, _ := root.RootCmd.Find([]string{"local-cmd"})
		assert.NotNil(t, cmd)
	})

	// Test with config already provided
	t.Run("with config already provided", func(t *testing.T) {
		// Create a config
		cfg := &config.ProjectConfig{
			Name: "provided-project",
			Commands: map[string]config.Command{
				"provided-cmd": {
					Run:         "echo \"Provided command\"",
					Description: "A provided command",
				},
			},
		}

		// Create a root command with the config
		exec := executor.NewDefaultExecutor()
		root := NewRootCommand(cfg, exec)

		// Load the config (should use the provided config)
		err = root.loadConfigAndRegisterCommands("")
		assert.NoError(t, err)
		assert.NotNil(t, root.Config)
		assert.Equal(t, "provided-project", root.Config.Name)

		// Verify the command was registered
		cmd, _, _ := root.RootCmd.Find([]string{"provided-cmd"})
		assert.NotNil(t, cmd)
	})

	// Test with invalid config path
	t.Run("with invalid config path", func(t *testing.T) {
		// Create a root command with nil config
		exec := executor.NewDefaultExecutor()
		root := NewRootCommand(nil, exec)

		// Load the config using a non-existent path
		err = root.loadConfigAndRegisterCommands("/non/existent/path.yml")
		assert.Error(t, err)
	})

	// Test with circular dependencies
	t.Run("with circular dependencies", func(t *testing.T) {
		// Create a temporary directory and change to it
		circularDir := filepath.Join(tempDir, "circular")
		err := os.Mkdir(circularDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create circular directory: %v", err)
		}
		err = os.Chdir(circularDir)
		if err != nil {
			t.Fatalf("Failed to change to circular directory: %v", err)
		}

		// Create a config with circular dependencies
		circularConfig := `
name: circular-project
commands:
  cmd1:
    depends: [cmd2]
    run: echo "Command 1"
  cmd2:
    depends: [cmd1]
    run: echo "Command 2"
`
		err = os.WriteFile("yxa.yml", []byte(circularConfig), 0644)
		if err != nil {
			t.Fatalf("Failed to create circular config file: %v", err)
		}

		// Create a root command with nil config
		exec := executor.NewDefaultExecutor()
		root := NewRootCommand(nil, exec)

		// Load the config (should fail due to circular dependencies)
		err = root.loadConfigAndRegisterCommands("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "circular dependency detected")
	})
}
