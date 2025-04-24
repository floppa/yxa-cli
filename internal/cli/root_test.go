package cli

import (
	"bytes"
	"fmt"
	"io"
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

// createTestConfig creates a test configuration with various command types for testing
func createTestConfig() *config.ProjectConfig {
	return &config.ProjectConfig{
		Name: "test-project",
		Commands: map[string]config.Command{
			"cmd1": {
				Run:         "echo cmd1",
				Description: "Command 1",
			},
			"cmd2": {
				Description: "Command 2 with subcommands",
				Commands: map[string]config.Command{
					"sub1": {
						Run:         "echo sub1",
						Description: "Subcommand 1",
					},
					"sub2": {
						Run:         "echo sub2",
						Description: "Subcommand 2",
					},
				},
			},
			"cmd-with-params": {
				Run:         "echo $PARAM1 $PARAM2",
				Description: "Command with parameters",
				Params: []config.Param{
					{Name: "param1", Type: "string", Default: "default1", Description: "Parameter 1"},
					{Name: "param2", Type: "int", Default: "42", Description: "Parameter 2"},
				},
			},
			"cmd-with-condition": {
				Run:         "echo condition-met",
				Description: "Command with condition",
				Condition:   "$PROJECT_NAME == test-project",
			},
			"cmd-with-timeout": {
				Run:         "echo timeout-command",
				Description: "Command with timeout",
				Timeout:     "5s",
			},
			"cmd-with-hooks": {
				Run:         "echo main-command",
				Description: "Command with hooks",
				Pre:         "echo pre-hook",
				Post:        "echo post-hook",
			},
			"cmd-with-tasks": {
				Tasks:       []string{"echo task1", "echo task2"},
				Description: "Command with tasks",
				Parallel:    true,
			},
		},
		Variables: map[string]string{
			"PROJECT_NAME": "test-project",
		},
	}
}

// setupTestRoot creates a RootCommand with the test configuration for testing
func setupTestRoot(cfg *config.ProjectConfig) *RootCommand {
	buf := &bytes.Buffer{}
	exec := executor.NewDefaultExecutor()
	exec.SetStdout(buf)
	root := NewRootCommand(cfg, exec)
	
	// Clear any existing commands
	root.clearUserCommands()
	
	// Register commands
	root.registerCommands()
	
	return root
}

// TestRegisterCommands is the main test for command registration
// It's now broken down into smaller, focused sub-tests
func TestRegisterCommands(t *testing.T) {
	// Use t.Run to organize tests into logical groups
	t.Run("BasicCommandRegistration", testBasicCommandRegistration)
	t.Run("SubcommandRegistration", testSubcommandRegistration)
	t.Run("ParameterizedCommandRegistration", testParameterizedCommandRegistration)
	t.Run("SpecialCommandsRegistration", testSpecialCommandsRegistration)
	t.Run("DryRunMode", testDryRunMode)
}

// testBasicCommandRegistration tests registration of a simple command
func testBasicCommandRegistration(t *testing.T) {
	cfg := createTestConfig()
	root := setupTestRoot(cfg)
	
	// Find the basic command and verify its properties
	cmd := findCommand(t, root.RootCmd, "cmd1")
	if cmd == nil {
		return // findCommand already reported the error
	}
	
	// Verify the command properties
	if cmd.Short != "Command 1" {
		t.Errorf("Expected description 'Command 1', got '%s'", cmd.Short)
	}
	if cmd.Long != "Command 1" {
		t.Errorf("Expected long description 'Command 1', got '%s'", cmd.Long)
	}
}

// testSubcommandRegistration tests registration of commands with subcommands
func testSubcommandRegistration(t *testing.T) {
	cfg := createTestConfig()
	root := setupTestRoot(cfg)
	
	// Find the command with subcommands
	cmd := findCommand(t, root.RootCmd, "cmd2")
	if cmd == nil {
		return // findCommand already reported the error
	}
	
	// Verify the command properties
	if cmd.Short != "Command 2 with subcommands" {
		t.Errorf("Expected description 'Command 2 with subcommands', got '%s'", cmd.Short)
	}
	
	// Test subcommand handling by checking if subcommands are registered
	subCmds := cmd.Commands()
	
	// Find and verify sub1
	sub1 := findCommandInList(subCmds, "sub1")
	if sub1 == nil {
		t.Error("Expected sub1 subcommand to be registered")
	} else if sub1.Short != "Subcommand 1" {
		t.Errorf("Expected description 'Subcommand 1', got '%s'", sub1.Short)
	}
	
	// Find and verify sub2
	sub2 := findCommandInList(subCmds, "sub2")
	if sub2 == nil {
		t.Error("Expected sub2 subcommand to be registered")
	} else if sub2.Short != "Subcommand 2" {
		t.Errorf("Expected description 'Subcommand 2', got '%s'", sub2.Short)
	}
}

// testParameterizedCommandRegistration tests registration of commands with parameters
func testParameterizedCommandRegistration(t *testing.T) {
	cfg := createTestConfig()
	root := setupTestRoot(cfg)
	
	// Find the command with parameters
	cmd := findCommand(t, root.RootCmd, "cmd-with-params")
	if cmd == nil {
		return // findCommand already reported the error
	}
	
	// Verify the command properties
	if cmd.Short != "Command with parameters" {
		t.Errorf("Expected description 'Command with parameters', got '%s'", cmd.Short)
	}
	
	// Verify that flags are registered
	flags := cmd.Flags()
	if flags.Lookup("param1") == nil {
		t.Error("Expected param1 flag to be registered")
	}
	if flags.Lookup("param2") == nil {
		t.Error("Expected param2 flag to be registered")
	}
}

// testSpecialCommandsRegistration tests registration of commands with special properties
func testSpecialCommandsRegistration(t *testing.T) {
	cfg := createTestConfig()
	root := setupTestRoot(cfg)
	
	// Verify that all special commands are registered
	specialCommands := []string{
		"cmd-with-condition",
		"cmd-with-timeout",
		"cmd-with-hooks",
		"cmd-with-tasks",
	}
	
	for _, cmdName := range specialCommands {
		cmd := findCommand(t, root.RootCmd, cmdName)
		if cmd == nil {
			t.Errorf("%s was not registered", cmdName)
		}
	}
}

// testDryRunMode tests command registration in dry run mode
func testDryRunMode(t *testing.T) {
	cfg := createTestConfig()
	root := setupTestRoot(cfg)
	
	// Set dry run mode and re-register commands
	root.DryRun = true
	root.clearUserCommands()
	root.registerCommands()
	
	// Verify that commands are still registered in dry run mode
	cmd := findCommand(t, root.RootCmd, "cmd1")
	if cmd == nil {
		t.Error("cmd1 was not registered in dry run mode")
	}
}

// Helper function to find a command by name in a cobra.Command
func findCommand(t *testing.T, parent *cobra.Command, name string) *cobra.Command {
	commands := parent.Commands()
	cmd := findCommandInList(commands, name)
	if cmd == nil {
		t.Errorf("%s was not registered", name)
	}
	return cmd
}

// Helper function to find a command by name in a list of commands
func findCommandInList(commands []*cobra.Command, name string) *cobra.Command {
	for _, cmd := range commands {
		if cmd.Name() == name {
			return cmd
		}
	}
	return nil
}

// setupConfigTestEnvironment creates a temporary directory with a test config file
func setupConfigTestEnvironment(t *testing.T) (string, string, func()) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "yxa-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create a test config file
	configContent := `
name: test-project
variables:
  PROJECT_NAME: test-project
commands:
  test-cmd:
    run: echo "test command"
    description: Test command
`
	configPath := filepath.Join(tmpDir, "yxa.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Return cleanup function
	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return tmpDir, configPath, cleanup
}

// verifyConfigLoaded checks if the config was properly loaded
func verifyConfigLoaded(t *testing.T, root *RootCommand, expectedName string) {
	if root.Config == nil {
		t.Error("Expected config to be loaded, got nil")
		return
	}
	
	if root.Config.Name != expectedName {
		t.Errorf("Expected project name '%s', got '%s'", expectedName, root.Config.Name)
	}
}

// verifyCommandRegistered checks if a specific command was registered
func verifyCommandRegistered(t *testing.T, root *RootCommand, cmdName string) {
	cmdFound := false
	for _, cmd := range root.RootCmd.Commands() {
		if cmd.Name() == cmdName {
			cmdFound = true
			break
		}
	}
	
	if !cmdFound {
		t.Errorf("Expected '%s' to be registered, but it wasn't found", cmdName)
	}
}

// TestRootCommand_LoadConfigAndRegisterCommands is the main test for config loading
// It's now broken down into smaller, focused sub-tests
func TestRootCommand_LoadConfigAndRegisterCommands(t *testing.T) {
	// Run sub-tests
	t.Run("LoadFromSpecificPath", testLoadConfigFromSpecificPath)
	t.Run("LoadFromCurrentDirectory", testLoadConfigFromCurrentDirectory)
}

// testLoadConfigFromSpecificPath tests loading config from a specific file path
func testLoadConfigFromSpecificPath(t *testing.T) {
	// Setup test environment
	_, configPath, cleanup := setupConfigTestEnvironment(t)
	defer cleanup()

	// Create a root command with nil config
	exec := executor.NewDefaultExecutor()
	root := NewRootCommand(nil, exec)

	// Test loading config from the specified path
	err := root.loadConfigAndRegisterCommands(configPath)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify that the config was loaded correctly
	verifyConfigLoaded(t, root, "test-project")
	
	// Verify that commands were registered
	verifyCommandRegistered(t, root, "test-cmd")
}

// testLoadConfigFromCurrentDirectory tests loading config from the current directory
func testLoadConfigFromCurrentDirectory(t *testing.T) {
	// Setup test environment
	tmpDir, _, cleanup := setupConfigTestEnvironment(t)
	defer cleanup()

	// Save current directory
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Change to the temp directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	defer os.Chdir(currentDir) // Restore original directory

	// Create a new root command with nil config
	exec := executor.NewDefaultExecutor()
	root := NewRootCommand(nil, exec)

	// Test loading config from the current directory (empty config flag)
	err = root.loadConfigAndRegisterCommands("")
	if err != nil {
		t.Errorf("Expected no error when loading from current dir, got: %v", err)
	}

	// Verify that the config was loaded
	verifyConfigLoaded(t, root, "test-project")
}

// setupCompletionTest creates a root command with completion removed
func setupCompletionTest(t *testing.T) *RootCommand {
	cfg := &config.ProjectConfig{
		Name:     "test-project",
		Commands: map[string]config.Command{},
	}
	root := NewRootCommand(cfg, executor.NewDefaultExecutor())

	// Remove completion if already present
	removeCompletionCommand(root)

	// Verify completion command is not present
	if hasCompletionCommand(root) {
		t.Fatal("Completion command should not be present before setupCompletion")
	}

	return root
}

// removeCompletionCommand removes the completion command if it exists
func removeCompletionCommand(root *RootCommand) {
	for _, cmd := range root.RootCmd.Commands() {
		if cmd.Name() == "completion" {
			root.RootCmd.RemoveCommand(cmd)
			break
		}
	}
}

// hasCompletionCommand checks if the completion command exists
func hasCompletionCommand(root *RootCommand) bool {
	for _, cmd := range root.RootCmd.Commands() {
		if cmd.Name() == "completion" {
			return true
		}
	}
	return false
}

// findCompletionCommand finds and returns the completion command
func findCompletionCommand(root *RootCommand) *cobra.Command {
	for _, cmd := range root.RootCmd.Commands() {
		if cmd.Name() == "completion" {
			return cmd
		}
	}
	return nil
}

// countCompletionCommands counts how many completion commands are present
func countCompletionCommands(root *RootCommand) int {
	count := 0
	for _, cmd := range root.RootCmd.Commands() {
		if cmd.Name() == "completion" {
			count++
		}
	}
	return count
}

// captureCommandOutput executes a command and captures its output
func captureCommandOutput(t *testing.T, cmd *cobra.Command, args []string) string {
	// Save original stdout
	oldStdout := os.Stdout
	
	// Create a pipe to capture output
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	// Execute the command
	cmd.SetArgs(args)
	err := cmd.Execute()
	assert.NoError(t, err)
	
	// Restore stdout and read captured output
	w.Close()
	os.Stdout = oldStdout
	outputBytes, _ := io.ReadAll(r)
	
	return string(outputBytes)
}

// TestRootCommand_SetupCompletion is the main test for completion setup
// It's now broken down into smaller, focused sub-tests
func TestRootCommand_SetupCompletion(t *testing.T) {
	// Run sub-tests
	t.Run("CompletionCommandCreation", testCompletionCommandCreation)
	t.Run("CompletionShells", testCompletionShells)
	t.Run("CompletionIdempotency", testCompletionIdempotency)
}

// testCompletionCommandCreation tests that the completion command is created
func testCompletionCommandCreation(t *testing.T) {
	// Setup test
	root := setupCompletionTest(t)
	
	// Setup completion
	root.setupCompletion()
	
	// Verify completion command is now present
	if !hasCompletionCommand(root) {
		t.Fatal("Completion command should be present after setupCompletion")
	}
}

// testCompletionShells tests the completion command for different shells
func testCompletionShells(t *testing.T) {
	// Setup test
	root := setupCompletionTest(t)
	root.setupCompletion()
	
	// Find the completion command
	completionCmd := findCompletionCommand(root)
	if completionCmd == nil {
		t.Fatal("Completion command not found")
	}
	
	// Test different shell completions
	shells := []string{"bash", "zsh", "fish", "powershell"}
	
	for _, shell := range shells {
		t.Run(shell, func(t *testing.T) {
			output := captureCommandOutput(t, completionCmd, []string{shell})
			assert.NotEmpty(t, output, fmt.Sprintf("%s completion output should not be empty", shell))
		})
	}
}

// testCompletionIdempotency tests that setupCompletion is idempotent
func testCompletionIdempotency(t *testing.T) {
	// Setup test
	root := setupCompletionTest(t)
	
	// Setup completion twice
	root.setupCompletion()
	root.setupCompletion()
	
	// Count how many completion commands are present
	completionCount := countCompletionCommands(root)
	
	// Verify only one completion command exists
	if completionCount != 1 {
		t.Fatalf("Expected exactly 1 completion command, got %d", completionCount)
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
			mtx1 := GetWriterMutex(tt.writer1)
			mtx2 := GetWriterMutex(tt.writer2)
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
	mutex1 := GetWriterMutex(writer1)
	assert.NotNil(t, mutex1)

	// Test getting the same mutex for the same writer
	mutex2 := GetWriterMutex(writer1)
	assert.Equal(t, mutex1, mutex2)

	// Test getting a different mutex for a different writer
	writer2 := &bytes.Buffer{}
	mutex3 := GetWriterMutex(writer2)
	assert.NotNil(t, mutex3)
	// We can't reliably test that mutex1 != mutex3 because the mutex map is package-level
	// and might be affected by other tests running concurrently

	// Test concurrent access to the mutex map
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			writer := &bytes.Buffer{}
			mutex := GetWriterMutex(writer)
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
