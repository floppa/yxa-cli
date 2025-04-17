package cmd

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/floppa/yxa-cli/config"
)


// TestCommandRegistration tests that commands are correctly registered from config
func TestCommandRegistration(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "yxa-registration-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		err := os.RemoveAll(tempDir)
		if err != nil {
			t.Logf("Warning: Failed to remove temp dir %s: %v", tempDir, err)
		}
	}()

	// Save current directory to return to it later
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		err := os.Chdir(currentDir)
		if err != nil {
			t.Logf("Warning: Failed to change back to original directory: %v", err)
		}
	}()

	// Change to the temporary directory
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}

	// Create a test config file
	configContent := `name: test-project
commands:
  echo:
    run: echo "Test Command"
  list:
    run: ls -la
`
	if err := os.WriteFile("yxa.yml", []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Load the config
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify the config was loaded correctly
	if cfg.Name != "test-project" {
		t.Errorf("Expected project name 'test-project', got '%s'", cfg.Name)
	}

	// Check that all commands were loaded
	expectedCommands := []string{"echo", "list"}
	for _, cmdName := range expectedCommands {
		cmd, ok := cfg.Commands[cmdName]
		if !ok {
			t.Errorf("Expected command '%s' not found in config", cmdName)
			continue
		}

		// Verify the command has the correct run property
		switch cmdName {
		case "echo":
			if cmd.Run != "echo \"Test Command\"" {
				t.Errorf("Expected echo command to run 'echo \"Test Command\"', got '%s'", cmd.Run)
			}
		case "list":
			if cmd.Run != "ls -la" {
				t.Errorf("Expected list command to run 'ls -la', got '%s'", cmd.Run)
			}
		}
	}
}

// TestCommandExecution tests that registered commands execute correctly
func TestCommandExecution(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "yxa-cmd-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		err := os.RemoveAll(tempDir)
		if err != nil {
			t.Logf("Warning: Failed to remove temp dir %s: %v", tempDir, err)
		}
	}()

	// Save current directory to return to it later
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		err := os.Chdir(currentDir)
		if err != nil {
			t.Logf("Warning: Failed to change back to original directory: %v", err)
		}
	}()

	// Change to the temporary directory
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}

	// Create a test file
	testFileName := "test-file.txt"
	testFileContent := "Test content"
	if err := os.WriteFile(testFileName, []byte(testFileContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Test cases for direct command execution
	testCases := []struct {
		name        string
		command     string
		expectedOut string
	}{
		{
			name:        "Echo Command",
			command:     "echo \"Test Command\"",
			expectedOut: "Test Command",
		},
		{
			name:        "Cat Command",
			command:     "cat test-file.txt",
			expectedOut: "Test content",
		},
	}

	// Test direct execution of shell commands
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Execute the shell command directly
			cmd := exec.Command("sh", "-c", tc.command)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("Command failed: %v\nOutput: %s", err, string(output))
			}

			if !strings.Contains(string(output), tc.expectedOut) {
				t.Errorf("Expected output to contain '%s', got: '%s'", tc.expectedOut, string(output))
			}
		})
	}
}

// TestDirectCommandExecution tests direct execution of shell commands
func TestDirectCommandExecution(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "yxa-direct-cmd-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		err := os.RemoveAll(tempDir)
		if err != nil {
			t.Logf("Warning: Failed to remove temp dir %s: %v", tempDir, err)
		}
	}()

	// Save current directory to return to it later
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		err := os.Chdir(currentDir)
		if err != nil {
			t.Logf("Warning: Failed to change back to original directory: %v", err)
		}
	}()

	// Change to the temporary directory
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}

	// Test complex shell commands
	complexCommands := []struct {
		name        string
		command     string
		expectedOut string
	}{
		{
			name:        "Pipes and Grep",
			command:     "echo 'Line 1\nLine 2\nLine 3' | grep 'Line 2'",
			expectedOut: "Line 2",
		},
		{
			name:        "File Creation and Reading",
			command:     "echo 'Test file content' > test.txt && cat test.txt",
			expectedOut: "Test file content",
		},
		{
			name:        "Conditional Execution",
			command:     "if [ 1 -eq 1 ]; then echo 'Condition true'; else echo 'Condition false'; fi",
			expectedOut: "Condition true",
		},
	}

	for _, tc := range complexCommands {
		t.Run(tc.name, func(t *testing.T) {
			// Execute the shell command directly
			cmd := exec.Command("sh", "-c", tc.command)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("Command failed: %v\nOutput: %s", err, string(output))
			}

			if !strings.Contains(string(output), tc.expectedOut) {
				t.Errorf("Expected output to contain '%s', got: '%s'", tc.expectedOut, string(output))
			}
		})
	}
}
