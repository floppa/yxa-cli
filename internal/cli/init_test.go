package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitializeConfig(t *testing.T) {
	// Save current working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalWd); err != nil {
			t.Logf("Warning: Failed to restore original working directory: %v", err)
		}
	}() // Restore original working directory

	// Create a temporary directory for test files
	tempDir := t.TempDir()

	// Create a valid yxa.yml file in the temporary directory
	validConfig := `
variables:
  PROJECT_NAME: test-project
commands:
  test:
    run: echo "Test command"
    description: A test command
  with-deps:
    run: echo "Command with dependencies"
    depends: [test]
`
	err = os.WriteFile(filepath.Join(tempDir, "yxa.yml"), []byte(validConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Test with valid configuration
	t.Run("valid config", func(t *testing.T) {
		// Change to the temporary directory
		err = os.Chdir(tempDir)
		if err != nil {
			t.Fatalf("Failed to change to temp directory: %v", err)
		}

		// Initialize config
		cfg, err := InitializeConfig()
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "test-project", cfg.Variables["PROJECT_NAME"])
		assert.Len(t, cfg.Commands, 2) // test and with-deps
	})

	// Test with circular dependencies
	t.Run("circular dependencies", func(t *testing.T) {
		// Change to the temporary directory
		err = os.Chdir(tempDir)
		if err != nil {
			t.Fatalf("Failed to change to temp directory: %v", err)
		}

		// Create a config with circular dependencies
		circularConfig := `
variables:
  PROJECT_NAME: test-project
commands:
  test:
    run: echo "Test command"
    description: A test command
  circular1:
    depends: [circular2]
  circular2:
    depends: [circular1]
`
		err = os.WriteFile(filepath.Join(tempDir, "yxa.yml"), []byte(circularConfig), 0644)
		if err != nil {
			t.Fatalf("Failed to create circular config file: %v", err)
		}

		// Initialize config (should detect circular dependencies)
		cfg, err := InitializeConfig()
		assert.Error(t, err)
		assert.Nil(t, cfg)
		assert.Contains(t, err.Error(), "circular dependency")
	})

	// Test with missing config file
	t.Run("missing config", func(t *testing.T) {
		// Create a new empty directory
		emptyDir := filepath.Join(tempDir, "empty")
		err = os.Mkdir(emptyDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create empty directory: %v", err)
		}

		// Change to the empty directory
		err = os.Chdir(emptyDir)
		if err != nil {
			t.Fatalf("Failed to change to empty directory: %v", err)
		}

		// Initialize config (should fail)
		cfg, err := InitializeConfig()
		assert.Error(t, err)
		assert.Nil(t, cfg)
	})
}

func TestInitializeApp(t *testing.T) {
	// Save current working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalWd); err != nil {
			t.Logf("Warning: Failed to restore original working directory: %v", err)
		}
	}() // Restore original working directory

	// Create a temporary directory for test files
	tempDir := t.TempDir()

	// Create a valid yxa.yml file in the temporary directory
	validConfig := `
variables:
  PROJECT_NAME: test-project
commands:
  test:
    run: echo "Test command"
    description: A test command
`
	err = os.WriteFile(filepath.Join(tempDir, "yxa.yml"), []byte(validConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Test with valid configuration
	t.Run("valid config", func(t *testing.T) {
		// Change to the temporary directory
		err = os.Chdir(tempDir)
		if err != nil {
			t.Fatalf("Failed to change to temp directory: %v", err)
		}

		// Initialize app
		root, err := InitializeApp()
		assert.NoError(t, err)
		assert.NotNil(t, root)
		assert.NotNil(t, root.Config)
		assert.NotNil(t, root.Executor)
		assert.NotNil(t, root.Handler)
		assert.NotNil(t, root.RootCmd)
		assert.Equal(t, "yxa", root.RootCmd.Use)
	})

	// Test with missing config file
	t.Run("missing config", func(t *testing.T) {
		// Create a new empty directory
		emptyDir := filepath.Join(tempDir, "empty")
		err = os.Mkdir(emptyDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create empty directory: %v", err)
		}

		// Change to the empty directory
		err = os.Chdir(emptyDir)
		if err != nil {
			t.Fatalf("Failed to change to empty directory: %v", err)
		}

		// Initialize app (should fail)
		root, err := InitializeApp()
		assert.Error(t, err)
		assert.Nil(t, root)
	})
}
