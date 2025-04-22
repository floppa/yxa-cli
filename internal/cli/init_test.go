package cli

import (
	"os"
	"path/filepath"
	"testing"

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

		// Initialize app (config is nil initially)
		root, err := InitializeApp()
		assert.NoError(t, err)
		assert.NotNil(t, root)
		assert.Nil(t, root.Config) // Config not loaded yet
		assert.NotNil(t, root.Executor)
		assert.NotNil(t, root.Handler)
		assert.NotNil(t, root.RootCmd)

		// Manually call the loading logic (simulates PersistentPreRunE)
		// Pass empty string for config flag value to test default loading
		err = root.loadConfigAndRegisterCommands("") 

		// Check state AFTER loading
		assert.NoError(t, err) // Loading should succeed
		assert.NotNil(t, root.Config) // Config should be loaded now
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

		// Initialize app
		root, err := InitializeApp()
		assert.NoError(t, err)
		assert.NotNil(t, root)
		assert.Nil(t, root.Config)

		// Manually call the loading logic with empty flag value
		err = root.loadConfigAndRegisterCommands("")

		// Check state AFTER loading attempt
		assert.Error(t, err) // Loading should fail
		assert.Nil(t, root.Config) // Config should still be nil
		assert.Contains(t, err.Error(), "failed to resolve config path") // Check for expected error
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

		// Initialize app
		root, err := InitializeApp()
		assert.NoError(t, err)
		assert.NotNil(t, root)
		assert.Nil(t, root.Config)

		// Manually call the loading logic
		err = root.loadConfigAndRegisterCommands("") 

		// Check state AFTER loading attempt
		assert.Error(t, err)
		// Config might be partially loaded before validation fails
		// assert.Nil(t, root.Config) 
		assert.Contains(t, err.Error(), "circular dependency")
	})

	// Test using --config flag during load
	t.Run("config flag during load", func(t *testing.T) {
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

		// Initialize app
		root, err := InitializeApp()
		assert.NoError(t, err)
		assert.NotNil(t, root)
		assert.Nil(t, root.Config)

		// Manually call the loading logic with the specific config path
		err = root.loadConfigAndRegisterCommands(configPath)

		// Check state AFTER loading
		assert.NoError(t, err)
		assert.NotNil(t, root.Config)
		assert.Equal(t, "flag-value", root.Config.Variables["FLAG_VAR"])
		assert.Len(t, root.Config.Commands, 1)
		flagCmd, _, _ := root.RootCmd.Find([]string{"flagcmd"})
		assert.NotNil(t, flagCmd)
		assert.Equal(t, "flagcmd", flagCmd.Name())
	})
}
