package config

import (
	"os"
	"testing"
)

// TestVariableSubstitutionInCommands tests that variables are correctly substituted in commands
func TestVariableSubstitutionInCommands(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "yxa-vars-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Save current directory to return to it later
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		err := os.Chdir(currentDir)
		if err != nil {
			t.Fatalf("Failed to change back to original directory: %v", err)
		}
	}()

	// Change to the temporary directory
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}

	// Create a test config file with variables and commands that use them
	configContent := `name: test-project
variables:
  PROJECT_NAME: test-project
  BUILD_DIR: ./build
  TEST_FLAGS: -v -race
commands:
  build:
    run: go build -o $BUILD_DIR/$PROJECT_NAME
  test:
    run: go test ${TEST_FLAGS} ./...
  combined:
    run: echo "Building $PROJECT_NAME with flags ${TEST_FLAGS}"
  env_var:
    run: echo "Using $ENV_VAR"
`
	if err := os.WriteFile("yxa.yml", []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Create a .env file
	envContent := `ENV_VAR=env-value
API_KEY=secret-key
`
	if err := os.WriteFile(".env", []byte(envContent), 0644); err != nil {
		t.Fatalf("Failed to write test .env file: %v", err)
	}

	// Set a system environment variable for testing
	os.Setenv("SYSTEM_ENV_VAR", "system-value")
	defer os.Unsetenv("SYSTEM_ENV_VAR")

	// Load the config
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test cases for variable substitution
	testCases := []struct {
		name     string
		command  string
		expected string
	}{
		{
			name:     "Simple Variable",
			command:  "build",
			expected: "go build -o ./build/test-project",
		},
		{
			name:     "Curly Braces Variable",
			command:  "test",
			expected: "go test -v -race ./...",
		},
		{
			name:     "Multiple Variables",
			command:  "combined",
			expected: "echo \"Building test-project with flags -v -race\"",
		},
		{
			name:     "Environment Variable from .env",
			command:  "env_var",
			expected: "echo \"Using env-value\"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd, ok := cfg.Commands[tc.command]
			if !ok {
				t.Fatalf("Command '%s' not found in config", tc.command)
			}

			if cmd.Run != tc.expected {
				t.Errorf("Expected command '%s' to be '%s', got '%s'", tc.command, tc.expected, cmd.Run)
			}
		})
	}

	// Test direct variable substitution
	t.Run("Direct Variable Substitution", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected string
		}{
			{
				input:    "echo $PROJECT_NAME",
				expected: "echo test-project",
			},
			{
				input:    "mkdir -p ${BUILD_DIR}",
				expected: "mkdir -p ./build",
			},
			{
				input:    "curl -H \"Authorization: Bearer $API_KEY\"",
				expected: "curl -H \"Authorization: Bearer secret-key\"",
			},
			{
				input:    "echo $SYSTEM_ENV_VAR",
				expected: "echo system-value",
			},
			{
				input:    "echo $UNDEFINED_VAR",
				expected: "echo $UNDEFINED_VAR", // Should remain unchanged
			},
		}

		for _, tc := range testCases {
			result := cfg.ReplaceVariables(tc.input)
			if result != tc.expected {
				t.Errorf("Expected '%s' to be replaced with '%s', got '%s'", tc.input, tc.expected, result)
			}
		}
	})

	// Test variable resolution priority
	t.Run("Variable Resolution Priority", func(t *testing.T) {
		// Create a config with overlapping variable names
		priorityConfig := &ProjectConfig{
			Name: "priority-test",
			Variables: map[string]string{
				"PRIORITY_VAR": "from-yaml",
				"COMMON_VAR":   "from-yaml",
			},
			envVars: map[string]string{
				"PRIORITY_VAR": "from-env-file", // Should be overridden by YAML
				"COMMON_VAR":   "from-env-file", // Should be overridden by YAML
				"ENV_ONLY_VAR": "from-env-file",
			},
		}

		// Set system environment variable
		os.Setenv("PRIORITY_VAR", "from-system") // Should be overridden by YAML and .env
		os.Setenv("COMMON_VAR", "from-system")   // Should be overridden by YAML and .env
		os.Setenv("ENV_ONLY_VAR", "from-system") // Should be overridden by .env
		os.Setenv("SYSTEM_ONLY_VAR", "from-system")
		defer func() {
			os.Unsetenv("PRIORITY_VAR")
			os.Unsetenv("COMMON_VAR")
			os.Unsetenv("ENV_ONLY_VAR")
			os.Unsetenv("SYSTEM_ONLY_VAR")
		}()

		testCases := []struct {
			varName  string
			expected string
		}{
			{
				varName:  "PRIORITY_VAR",
				expected: "from-yaml", // YAML has highest priority
			},
			{
				varName:  "COMMON_VAR",
				expected: "from-yaml", // YAML has highest priority
			},
			{
				varName:  "ENV_ONLY_VAR",
				expected: "from-env-file", // .env has priority over system
			},
			{
				varName:  "SYSTEM_ONLY_VAR",
				expected: "from-system", // System is used when not in YAML or .env
			},
		}

		for _, tc := range testCases {
			result := priorityConfig.ReplaceVariables("$" + tc.varName)
			if result != tc.expected {
				t.Errorf("Expected '$%s' to be replaced with '%s', got '%s'", tc.varName, tc.expected, result)
			}
		}
	})
}
