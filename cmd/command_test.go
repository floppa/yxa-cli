package cmd

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/magnuseriksson/yxa-cli/config"
)

// TestCommandChaining tests the command chaining functionality
func TestCommandChaining(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "yxa-cmd-chain-test")
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

	// Create a test log file to track command execution
	logFile := "execution.log"
	if err := os.WriteFile(logFile, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create log file: %v", err)
	}

	// Setup test commands with dependencies
	cfg = &config.ProjectConfig{
		Name: "test-project",
		Commands: map[string]config.Command{
			"cmd1": {
				Run: "echo 'cmd1' >> " + logFile,
			},
			"cmd2": {
				Run:     "echo 'cmd2' >> " + logFile,
				Depends: []string{"cmd1"},
			},
			"cmd3": {
				Run:     "echo 'cmd3' >> " + logFile,
				Depends: []string{"cmd2"},
			},
			"parallel1": {
				Run:     "echo 'parallel1' >> " + logFile,
				Depends: []string{"cmd1"},
			},
			"parallel2": {
				Run:     "echo 'parallel2' >> " + logFile,
				Depends: []string{"cmd1"},
			},
			"combined": {
				Run:     "echo 'combined' >> " + logFile,
				Depends: []string{"parallel1", "parallel2"},
			},
			"circular": {
				Run:     "echo 'circular' >> " + logFile,
				Depends: []string{"circular"}, // Circular dependency
			},
		},
	}

	// Reset executed commands
	executedCommands = make(map[string]bool)

	// Test cases for command chaining
	testCases := []struct {
		name           string
		command        string
		expectedOutput []string
	}{
		{
			name:           "Simple Chain",
			command:        "cmd3",
			expectedOutput: []string{"cmd1", "cmd2", "cmd3"},
		},
		{
			name:           "Multiple Dependencies",
			command:        "combined",
			expectedOutput: []string{"cmd1", "parallel1", "parallel2", "combined"},
		},
		{
			name:           "Circular Dependency",
			command:        "circular",
			expectedOutput: []string{"circular"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset log file
			if err := os.WriteFile(logFile, []byte(""), 0644); err != nil {
				t.Fatalf("Failed to reset log file: %v", err)
			}

			// Reset executed commands
			executedCommands = make(map[string]bool)

			// Capture stdout to prevent output during tests
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Execute the command
			err := executeCommand(tc.command)
			if err != nil {
				t.Fatalf("Failed to execute command: %v", err)
			}

			// Restore stdout
			if err := w.Close(); err != nil {
				t.Fatalf("Failed to close pipe writer: %v", err)
			}
			os.Stdout = oldStdout
			_, err = io.Copy(io.Discard, r)
			if err != nil {
				t.Fatalf("Failed to discard output: %v", err)
			}

			// Read the log file to check execution order
			logContent, err := os.ReadFile(logFile)
			if err != nil {
				t.Fatalf("Failed to read log file: %v", err)
			}

			// Split the log content into lines
			lines := bytes.Split(bytes.TrimSpace(logContent), []byte("\n"))
			if len(lines) != len(tc.expectedOutput) {
				t.Errorf("Expected %d commands to be executed, got %d", len(tc.expectedOutput), len(lines))
			}

			// Check the execution order
			for i, expectedCmd := range tc.expectedOutput {
				if i >= len(lines) {
					t.Errorf("Missing expected command: %s", expectedCmd)
					continue
				}
				if string(bytes.TrimSpace(lines[i])) != expectedCmd {
					t.Errorf("Expected command %d to be '%s', got '%s'", i+1, expectedCmd, string(bytes.TrimSpace(lines[i])))
				}
			}
		})
	}
}
