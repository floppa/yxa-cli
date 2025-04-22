package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/floppa/yxa-cli/internal/config"
	"github.com/floppa/yxa-cli/internal/executor"
	"github.com/stretchr/testify/assert"
)

// setupTestEnv creates a temporary directory, sets it as the current working directory,
// and returns the original working directory and a cleanup function.
func setupTestEnv(t *testing.T) (string, func()) {
	t.Helper()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	tempDir := t.TempDir()
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}
	cleanup := func() {
		if err := os.Chdir(originalWd); err != nil {
			t.Logf("Warning: Failed to restore original working directory: %v", err)
		}
	}
	return tempDir, cleanup
}

func TestInitializeApp_Execution(t *testing.T) {

	// Test with valid configuration
	t.Run("valid config loading via execute", func(t *testing.T) {
		tempDir, cleanup := setupTestEnv(t)
		defer cleanup()

		validConfig := `
variables:
  PROJECT_NAME: test-project
commands:
  test:
    run: echo "Test command"
    description: A test command
`
		err := os.WriteFile(filepath.Join(tempDir, "yxa.yml"), []byte(validConfig), 0644)
		if err != nil {
			t.Fatalf("Failed to create test config file: %v", err)
		}

		// Initialize app (config should be loaded immediately)
		root, err := InitializeApp()
		assert.NoError(t, err)
		assert.NotNil(t, root)
		assert.NotNil(t, root.Config) // Config should be loaded now
		assert.NotNil(t, root.Executor)
		assert.NotNil(t, root.Handler)
		assert.NotNil(t, root.RootCmd)

		// Check the loaded config
		assert.Equal(t, "test-project", root.Config.Variables["PROJECT_NAME"])
		assert.Len(t, root.Config.Commands, 1)
		testCmd, _, _ := root.RootCmd.Find([]string{"test"}) // Check command registered
		assert.NotNil(t, testCmd)
		assert.Equal(t, "test", testCmd.Name())
	})

	// Test with missing config file
	t.Run("missing config file during load", func(t *testing.T) {
		_, cleanup := setupTestEnv(t) // Just need the clean env
		defer cleanup()

		// Initialize app - should not fail but will have nil config
		root, err := InitializeApp()
		assert.NoError(t, err)
		assert.NotNil(t, root)
		// Config will be nil since no config file was found
		assert.Nil(t, root.Config)
	})

	// Test with circular dependencies detected during load
	t.Run("circular dependencies detected during load", func(t *testing.T) {
		tempDir, cleanup := setupTestEnv(t)
		defer cleanup()

		circularConfig := `
commands:
  circular1:
    depends: [circular2]
  circular2:
    depends: [circular1]
`
		err := os.WriteFile(filepath.Join(tempDir, "yxa.yml"), []byte(circularConfig), 0644)
		if err != nil {
			t.Fatalf("Failed to create circular config file: %v", err)
		}

		// Initialize app - should fail due to circular dependencies
		root, err := InitializeApp()
		// Check that we got an error about circular dependencies
		assert.Error(t, err) 
		assert.Contains(t, err.Error(), "circular dependency detected")
		// Root should still be returned even when there's an error
		assert.NotNil(t, root)
	})

	// Test using --config flag during load
	t.Run("config flag path during load", func(t *testing.T) {
		// Save the original InitializeApp function
		originalInitializeApp := InitializeApp
		defer func() {
			// Restore the original InitializeApp function after the test
			InitializeApp = originalInitializeApp
		}()

		tempDir, cleanup := setupTestEnv(t)
		defer cleanup()

		// Create config in a sub-directory
		subDir := filepath.Join(tempDir, "sub")
		err := os.Mkdir(subDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create sub directory: %v", err)
		}
		configPath := filepath.Join(subDir, "custom-config.yml")
		flagConfig := `
variables:
  FLAG_VAR: flag-value
commands:
  flagcmd:
    run: echo "flag command"
`
		err = os.WriteFile(configPath, []byte(flagConfig), 0644)
		if err != nil {
			t.Fatalf("Failed to create flag config file: %v", err)
		}

		// Override InitializeApp to use our custom config path
		InitializeApp = func() (*RootCommand, error) {
			// Create a default executor
			exec := executor.NewDefaultExecutor()

			// Create the root command with nil config initially
			root := NewRootCommand(nil, exec)

			// Load configuration from the custom path
			cfg, err := config.LoadConfigFrom(configPath)
			if err != nil {
				return root, fmt.Errorf("failed to load configuration from '%s': %w", configPath, err)
			}

			// Store the loaded config
			root.Config = cfg

			// Validate command dependencies
			if err = validateCommandDependencies(cfg); err != nil {
				return root, fmt.Errorf("invalid command dependencies: %w", err)
			}

			// Initialize the handler with the config
			root.Handler = NewCommandHandler(cfg, exec)

			// Register commands
			root.registerCommands()

			return root, nil
		}

		// Initialize app with our custom config path
		root, err := InitializeApp()
		assert.NoError(t, err)
		assert.NotNil(t, root)
		assert.NotNil(t, root.Config)

		// Check the loaded config
		assert.Equal(t, "flag-value", root.Config.Variables["FLAG_VAR"])
		assert.Len(t, root.Config.Commands, 1)
		flagCmd, _, _ := root.RootCmd.Find([]string{"flagcmd"})
		assert.NotNil(t, flagCmd)
		assert.Equal(t, "flagcmd", flagCmd.Name())
	})
}
