package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMainPackage tests the main package functionality through integration tests
func TestMainPackage(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "yxa-main-test")
	require.NoError(t, err, "Failed to create temp dir")
	defer func() {
		err := os.RemoveAll(tempDir)
		if err != nil {
			t.Logf("Warning: Failed to remove temp dir %s: %v", tempDir, err)
		}
	}()

	// Save current directory to return to it later
	currentDir, err := os.Getwd()
	require.NoError(t, err, "Failed to get current directory")
	defer func() {
		err := os.Chdir(currentDir)
		if err != nil {
			t.Logf("Warning: Failed to change back to original directory: %v", err)
		}
	}()

	// Change to the temporary directory
	require.NoError(t, os.Chdir(tempDir), "Failed to change to temp dir")

	// Create a simple yxa.yml file
	yxaConfig := `name: yxa-test-project

variables:
  PROJECT_NAME: yxa-test-project
  OUTPUT_DIR: ./output

commands:
  hello:
    run: echo "Hello, $PROJECT_NAME!"
    description: A simple greeting command

  prepare:
    run: mkdir -p $OUTPUT_DIR
    description: Creates the output directory

  write-file:
    run: echo "Content from write-file" > $OUTPUT_DIR/output.txt
    depends: [prepare]
    description: Writes to a file in the output directory
`
	require.NoError(t, os.WriteFile("yxa.yml", []byte(yxaConfig), 0644), "Failed to write yxa.yml file")

	// Build the yxa CLI binary path
	// This assumes the binary is in the project root or a known location
	yxaBinary := filepath.Join(currentDir, "..", "yxa")
	if _, err := os.Stat(yxaBinary); os.IsNotExist(err) {
		// Try to find the binary in the current directory
		yxaBinary = filepath.Join(currentDir, "yxa")
		if _, err := os.Stat(yxaBinary); os.IsNotExist(err) {
			// Fall back to assuming it's in PATH
			yxaBinary = "yxa"
		}
	}

	// Test the help command to list available commands
	t.Run("Help Command", func(t *testing.T) {
		cmd := exec.Command(yxaBinary, "--help")
		output, err := cmd.CombinedOutput()
		assert.NoError(t, err, "Help command failed")
		
		outputStr := string(output)
		assert.Contains(t, outputStr, "Available Commands", "Help should list available commands")
		assert.Contains(t, outputStr, "hello", "Help should include hello command")
		assert.Contains(t, outputStr, "prepare", "Help should include prepare command")
		assert.Contains(t, outputStr, "write-file", "Help should include write-file command")
	})

	// Test executing a simple command
	t.Run("Execute Simple Command", func(t *testing.T) {
		cmd := exec.Command(yxaBinary, "hello")
		output, err := cmd.CombinedOutput()
		assert.NoError(t, err, "Hello command failed")
		assert.Contains(t, string(output), "Hello, yxa-test-project!", "Command output should contain greeting")
	})

	// Test command with dependencies
	t.Run("Command With Dependencies", func(t *testing.T) {
		cmd := exec.Command(yxaBinary, "write-file")
		output, err := cmd.CombinedOutput()
		assert.NoError(t, err, "write-file command failed: %s", string(output))
		
		// Verify the output directory was created
		outputDir := filepath.Join(tempDir, "output")
		_, err = os.Stat(outputDir)
		assert.NoError(t, err, "Output directory should be created")
		
		// Verify the output file was created
		outputFile := filepath.Join(outputDir, "output.txt")
		content, err := os.ReadFile(outputFile)
		assert.NoError(t, err, "Should be able to read output file")
		assert.Contains(t, string(content), "Content from write-file", "File should contain expected content")
	})

	// Test completion command
	t.Run("Completion Command", func(t *testing.T) {
		cmd := exec.Command(yxaBinary, "completion", "--help")
		output, err := cmd.CombinedOutput()
		assert.NoError(t, err, "Completion command failed")
		
		outputStr := string(output)
		assert.Contains(t, outputStr, "completion", "Completion help should include information about the command")
	})

	// Test non-existent command
	t.Run("Non-existent Command", func(t *testing.T) {
		cmd := exec.Command(yxaBinary, "non-existent-command")
		err := cmd.Run()
		assert.Error(t, err, "Non-existent command should fail")
	})

	// Test command help
	t.Run("Command Help", func(t *testing.T) {
		cmd := exec.Command(yxaBinary, "help", "hello")
		output, err := cmd.CombinedOutput()
		assert.NoError(t, err, "Command help failed")
		
		outputStr := string(output)
		assert.Contains(t, outputStr, "hello", "Command help should include command name")
		assert.Contains(t, outputStr, "A simple greeting command", "Command help should include description")
	})
}
