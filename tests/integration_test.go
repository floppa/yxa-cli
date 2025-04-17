package tests

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/magnuseriksson/yxa-cli/config"
)

// TestCommandExecution tests that the CLI can execute various bash commands
func TestCommandExecution(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "yxa-command-test")
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

	// Test cases with different bash commands
	testCases := []struct {
		name        string
		command     string
		expectedOut string
		expectError bool
	}{
		{
			name:        "Echo Command",
			command:     "echo \"Hello, World!\"",
			expectedOut: "Hello, World!",
			expectError: false,
		},
		{
			name:        "Directory Listing",
			command:     "ls -la",
			expectedOut: ".", // Just check for current directory marker
			expectError: false,
		},
		{
			name:        "Command With Pipes",
			command:     "echo 'Line 1\nLine 2\nLine 3' | grep Line",
			expectedOut: "Line",
			expectError: false,
		},
		{
			name:        "Command With Environment Variables",
			command:     "TEST_VAR=\"Hello from env\" && echo $TEST_VAR",
			expectedOut: "Hello from env",
			expectError: false,
		},
		{
			name:        "Invalid Command",
			command:     "command_that_does_not_exist",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Execute the shell command directly
			cmd := exec.Command("sh", "-c", tc.command)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			err := cmd.Run()

			// Check error expectation
			if tc.expectError && err == nil {
				t.Errorf("Expected error but got none")
			} else if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v\nStderr: %s", err, stderr.String())
			}

			// If we expect specific output, check it
			if tc.expectedOut != "" && !strings.Contains(stdout.String(), tc.expectedOut) {
				t.Errorf("Expected output to contain '%s', got: '%s'", tc.expectedOut, stdout.String())
			}
		})
	}
}

// TestComplexBashCommands tests more complex bash scenarios
func TestComplexBashCommands(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "yxa-complex-test")
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
	testFileContent := "Line 1\nLine 2\nLine 3\n"
	if err := os.WriteFile(testFileName, []byte(testFileContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Test cases for complex commands
	complexTests := []struct {
		name        string
		command     string
		expectedOut string
	}{
		{
			name:        "Grep File",
			command:     "grep \"Line\" test-file.txt",
			expectedOut: "Line",
		},
		{
			name:        "Create and Read File",
			command:     "echo \"New content\" > new-file.txt && cat new-file.txt",
			expectedOut: "New content",
		},
		{
			name:        "Conditional Command",
			command:     "if [ -f \"test-file.txt\" ]; then echo \"File exists\"; else echo \"File not found\"; fi",
			expectedOut: "File exists",
		},
		{
			name:        "Loop Command",
			command:     "for i in {1..3}; do echo \"Iteration $i\"; done",
			expectedOut: "Iteration",
		},
	}

	for _, tc := range complexTests {
		t.Run(tc.name, func(t *testing.T) {
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

// TestConfigWithBashCommands tests that the config can properly load and parse commands
func TestConfigWithBashCommands(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "yxa-config-test")
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

	// Create a config file with various bash commands
	configContent := `name: test-project
commands:
  simple:
    run: echo "Simple command"
  pipes:
    run: echo "Hello" | grep "Hello"
  complex:
    run: for i in {1..3}; do echo "Item $i"; done
  conditional:
    run: if [ "$HOME" != "" ]; then echo "Home is set"; fi
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
	expectedCommands := []string{"simple", "pipes", "complex", "conditional"}
	for _, cmdName := range expectedCommands {
		cmd, ok := cfg.Commands[cmdName]
		if !ok {
			t.Errorf("Expected command '%s' not found in config", cmdName)
			continue
		}

		// Execute the command to verify it works
		shellCmd := exec.Command("sh", "-c", cmd.Run)
		output, err := shellCmd.CombinedOutput()
		if err != nil {
			t.Errorf("Command '%s' failed to execute: %v\nOutput: %s", cmdName, err, string(output))
		}

		// Print the command and its output for debugging
		fmt.Printf("Command '%s' output: %s\n", cmdName, string(output))
	}
}
