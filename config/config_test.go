package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReplaceVariables(t *testing.T) {
	// Create a ProjectConfig with variables
	config := &ProjectConfig{
		Name: "test-project",
		Variables: map[string]string{
			"PROJECT_NAME": "test-project",
			"BUILD_DIR":   "./build",
		},
		envVars: map[string]string{
			"ENV_VAR":  "env-value",
			"API_KEY":  "secret-key",
			"API_HOST": "api.example.com",
		},
	}

	// Test cases for variable replacement
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple variable",
			input:    "echo $PROJECT_NAME",
			expected: "echo test-project",
		},
		{
			name:     "Curly braces variable",
			input:    "mkdir -p ${BUILD_DIR}",
			expected: "mkdir -p ./build",
		},
		{
			name:     "Environment variable from .env",
			input:    "curl -H \"Authorization: Bearer $API_KEY\" $API_HOST",
			expected: "curl -H \"Authorization: Bearer secret-key\" api.example.com",
		},
		{
			name:     "Multiple variables",
			input:    "echo $PROJECT_NAME is using $ENV_VAR with ${API_KEY}",
			expected: "echo test-project is using env-value with secret-key",
		},
		{
			name:     "Undefined variable",
			input:    "echo $UNDEFINED_VAR",
			expected: "echo $UNDEFINED_VAR", // Should remain unchanged
		},
		{
			name:     "Mixed defined and undefined variables",
			input:    "echo $PROJECT_NAME and $UNDEFINED_VAR",
			expected: "echo test-project and $UNDEFINED_VAR",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := config.ReplaceVariables(tc.input)
			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "yxa-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Save current directory to return to it later
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(currentDir)

	// Change to the temporary directory
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}

	// Test case: Missing config file
	t.Run("MissingConfigFile", func(t *testing.T) {
		_, err := LoadConfig()
		if err == nil {
			t.Error("Expected error for missing config file, got nil")
		}
	})

	// Create a valid config file
	validConfig := `name: test-project
variables:
  PROJECT_NAME: test-project
  BUILD_DIR: ./build
commands:
  echo:
    run: echo "Hello, $PROJECT_NAME!"
  list:
    run: ls -la ${BUILD_DIR}
  env:
    run: echo "Using $ENV_VAR"
`
	if err := os.WriteFile(filepath.Join(tempDir, "yxa.yml"), []byte(validConfig), 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}
	
	// Create a .env file
	envContent := `ENV_VAR=env-value
API_KEY=secret-key
API_HOST=api.example.com
`
	if err := os.WriteFile(filepath.Join(tempDir, ".env"), []byte(envContent), 0644); err != nil {
		t.Fatalf("Failed to write test .env file: %v", err)
	}

	// Test case: Valid config file with variables and .env
	t.Run("ValidConfigFileWithVariables", func(t *testing.T) {
		config, err := LoadConfig()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if config.Name != "test-project" {
			t.Errorf("Expected name 'test-project', got '%s'", config.Name)
		}

		// Check variables
		if len(config.Variables) != 2 {
			t.Errorf("Expected 2 variables, got %d", len(config.Variables))
		}

		if config.Variables["PROJECT_NAME"] != "test-project" {
			t.Errorf("Expected PROJECT_NAME to be 'test-project', got '%s'", config.Variables["PROJECT_NAME"])
		}

		if config.Variables["BUILD_DIR"] != "./build" {
			t.Errorf("Expected BUILD_DIR to be './build', got '%s'", config.Variables["BUILD_DIR"])
		}

		// Check environment variables from .env
		if len(config.envVars) != 3 {
			t.Errorf("Expected 3 environment variables, got %d", len(config.envVars))
		}

		if config.envVars["ENV_VAR"] != "env-value" {
			t.Errorf("Expected ENV_VAR to be 'env-value', got '%s'", config.envVars["ENV_VAR"])
		}

		// Check commands with variable substitution
		if len(config.Commands) != 3 {
			t.Errorf("Expected 3 commands, got %d", len(config.Commands))
		}

		echoCmd, ok := config.Commands["echo"]
		if !ok {
			t.Error("Expected 'echo' command, not found")
		} else if echoCmd.Run != `echo "Hello, test-project!"` {
			t.Errorf("Expected echo command with substituted variable, got '%s'", echoCmd.Run)
		}

		listCmd, ok := config.Commands["list"]
		if !ok {
			t.Error("Expected 'list' command, not found")
		} else if listCmd.Run != "ls -la ./build" {
			t.Errorf("Expected list command with substituted variable, got '%s'", listCmd.Run)
		}

		envCmd, ok := config.Commands["env"]
		if !ok {
			t.Error("Expected 'env' command, not found")
		} else if envCmd.Run != "echo \"Using env-value\"" {
			t.Errorf("Expected env command with substituted .env variable, got '%s'", envCmd.Run)
		}
	})

	// Create an invalid config file
	invalidConfig := `name: test-project
commands:
  - echo
  - list
`
	if err := os.WriteFile(filepath.Join(tempDir, "yxa.yml"), []byte(invalidConfig), 0644); err != nil {
		t.Fatalf("Failed to write invalid test config file: %v", err)
	}

	// Test case: Invalid config file
	t.Run("InvalidConfigFile", func(t *testing.T) {
		_, err := LoadConfig()
		if err == nil {
			t.Error("Expected error for invalid config file, got nil")
		}
	})
}
