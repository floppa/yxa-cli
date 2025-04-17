package config

import (
	"os"
	"testing"
)

// TestCommandDependencies tests that command dependencies are correctly loaded from the config file
func TestCommandDependencies(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "yxa-deps-test")
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
			t.Fatalf("Failed to change back to original directory: %v", err)
		}
	}()

	// Change to the temporary directory
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}

	// Create a test config file with command dependencies
	configContent := `name: test-project
commands:
  build:
    run: go build ./...
  test:
    run: go test ./...
  clean:
    run: rm -rf ./build
  install:
    run: cp ./app /usr/local/bin/
    depends: [build]
  dist:
    run: mkdir -p ./dist && go build -o ./dist/app
    depends: [clean]
  release:
    run: git tag -a v1.0.0 -m "Release v1.0.0"
    depends: [build, test]
  complex:
    run: echo "Complex command"
    depends: [clean, build, test]
  circular:
    run: echo "Circular dependency"
    depends: [circular]
  nested:
    run: echo "Nested dependencies"
    depends: [install, dist]
`
	if err := os.WriteFile("yxa.yml", []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Load the config
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test cases for command dependencies
	testCases := []struct {
		name     string
		command  string
		expected []string
	}{
		{
			name:     "No Dependencies",
			command:  "build",
			expected: nil,
		},
		{
			name:     "Single Dependency",
			command:  "install",
			expected: []string{"build"},
		},
		{
			name:     "Multiple Dependencies",
			command:  "release",
			expected: []string{"build", "test"},
		},
		{
			name:     "Complex Dependencies",
			command:  "complex",
			expected: []string{"clean", "build", "test"},
		},
		{
			name:     "Circular Dependency",
			command:  "circular",
			expected: []string{"circular"},
		},
		{
			name:     "Nested Dependencies",
			command:  "nested",
			expected: []string{"install", "dist"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd, ok := cfg.Commands[tc.command]
			if !ok {
				t.Fatalf("Command '%s' not found in config", tc.command)
			}

			// Check if the dependencies match
			if len(cmd.Depends) != len(tc.expected) {
				t.Errorf("Expected command '%s' to have %d dependencies, got %d", tc.command, len(tc.expected), len(cmd.Depends))
			}

			// Check if all expected dependencies are present
			for _, dep := range tc.expected {
				found := false
				for _, actualDep := range cmd.Depends {
					if actualDep == dep {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected dependency '%s' not found for command '%s'", dep, tc.command)
				}
			}
		})
	}
}
