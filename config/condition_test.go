package config

import (
	"os"
	"testing"
)

func TestEvaluateCondition(t *testing.T) {
	// Create a test config
	cfg := &ProjectConfig{
		Variables: map[string]string{
			"TEST_VAR":  "test_value",
			"GOOS":      "darwin",
			"GOARCH":    "amd64",
			"TEST_PATH": "/usr/local/bin:/usr/bin",
		},
		envVars: map[string]string{
			"ENV_VAR": "env_value",
		},
	}

	// Set up environment variable for testing
	if err := os.Setenv("SYS_VAR", "sys_value"); err != nil {
		t.Fatalf("Failed to set environment variable: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("SYS_VAR"); err != nil {
			t.Logf("Failed to unset environment variable: %v", err)
		}
	}()

	// Create a temporary file for the exists test
	tmpFile, err := os.CreateTemp("", "condition-test")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			t.Logf("Failed to remove temp file: %v", err)
		}
	}()

	// Test cases
	tests := []struct {
		name      string
		condition string
		expected  bool
	}{
		{
			name:      "Empty condition",
			condition: "",
			expected:  true,
		},
		{
			name:      "Equal condition with YAML variable",
			condition: "$TEST_VAR == test_value",
			expected:  true,
		},
		{
			name:      "Equal condition with environment variable",
			condition: "$ENV_VAR == env_value",
			expected:  true,
		},
		{
			name:      "Equal condition with system variable",
			condition: "$SYS_VAR == sys_value",
			expected:  true,
		},
		{
			name:      "Equal condition that is false",
			condition: "$GOOS == windows",
			expected:  false,
		},
		{
			name:      "Not equal condition that is true",
			condition: "$GOOS != windows",
			expected:  true,
		},
		{
			name:      "Not equal condition that is false",
			condition: "$GOOS != darwin",
			expected:  false,
		},
		{
			name:      "Contains condition that is true",
			condition: "$TEST_PATH contains /usr/local",
			expected:  true,
		},
		{
			name:      "Contains condition that is false",
			condition: "$TEST_PATH contains /opt",
			expected:  false,
		},
		{
			name:      "Exists condition that is true",
			condition: "exists " + tmpFile.Name(),
			expected:  true,
		},
		{
			name:      "Exists condition that is false",
			condition: "exists /path/that/does/not/exist",
			expected:  false,
		},
		{
			name:      "Unknown condition format",
			condition: "invalid condition",
			expected:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := cfg.EvaluateCondition(tc.condition)
			if result != tc.expected {
				t.Errorf("Expected %v for condition '%s', got %v", tc.expected, tc.condition, result)
			}
		})
	}
}
