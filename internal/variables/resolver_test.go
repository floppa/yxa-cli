package variables

import (
	"os"
	"testing"
)

func TestResolver_Resolve(t *testing.T) {
	// Set up environment variable for testing
	if err := os.Setenv("TEST_ENV_VAR", "env_value"); err != nil {
		t.Fatalf("Failed to set environment variable: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("TEST_ENV_VAR"); err != nil {
			t.Logf("Warning: Failed to unset environment variable: %v", err)
		}
	}()

	tests := []struct {
		name        string
		input       string
		configVars  map[string]string
		envFileVars map[string]string
		paramVars   map[string]string
		systemEnv   bool
		want        string
	}{
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "no variables",
			input: "hello world",
			want:  "hello world",
		},
		{
			name:       "config variable",
			input:      "hello $CONFIG_VAR",
			configVars: map[string]string{"CONFIG_VAR": "config_value"},
			want:       "hello config_value",
		},
		{
			name:        "env file variable",
			input:       "hello $ENV_FILE_VAR",
			envFileVars: map[string]string{"ENV_FILE_VAR": "env_file_value"},
			want:        "hello env_file_value",
		},
		{
			name:      "param variable",
			input:     "hello $PARAM_VAR",
			paramVars: map[string]string{"PARAM_VAR": "param_value"},
			want:      "hello param_value",
		},
		{
			name:      "system env variable",
			input:     "hello $TEST_ENV_VAR",
			systemEnv: true,
			want:      "hello env_value",
		},
		{
			name:      "system env variable disabled",
			input:     "hello $TEST_ENV_VAR",
			systemEnv: false,
			want:      "hello $TEST_ENV_VAR",
		},
		{
			name:       "variable with braces",
			input:      "hello ${CONFIG_VAR}",
			configVars: map[string]string{"CONFIG_VAR": "config_value"},
			want:       "hello config_value",
		},
		{
			name:        "multiple variables",
			input:       "$CONFIG_VAR $ENV_FILE_VAR $PARAM_VAR",
			configVars:  map[string]string{"CONFIG_VAR": "config_value"},
			envFileVars: map[string]string{"ENV_FILE_VAR": "env_file_value"},
			paramVars:   map[string]string{"PARAM_VAR": "param_value"},
			want:        "config_value env_file_value param_value",
		},
		{
			name:  "variable not found",
			input: "hello $NOT_FOUND",
			want:  "hello $NOT_FOUND",
		},
		{
			name:        "priority order",
			input:       "$PRIORITY",
			configVars:  map[string]string{"PRIORITY": "config"},
			envFileVars: map[string]string{"PRIORITY": "env_file"},
			paramVars:   map[string]string{"PRIORITY": "param"},
			systemEnv:   true,
			want:        "param", // Param vars have highest priority
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewResolver()
			if tt.configVars != nil {
				r.WithConfigVars(tt.configVars)
			}
			if tt.envFileVars != nil {
				r.WithEnvFileVars(tt.envFileVars)
			}
			if tt.paramVars != nil {
				r.WithParamVars(tt.paramVars)
			}
			r.WithSystemEnvVar(tt.systemEnv)

			if got := r.Resolve(tt.input); got != tt.want {
				t.Errorf("Resolver.Resolve() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolver_ResolveAll(t *testing.T) {
	r := NewResolver().WithConfigVars(map[string]string{"VAR": "value"})
	inputs := []string{"$VAR", "hello", "$VAR world"}
	want := []string{"value", "hello", "value world"}

	got := r.ResolveAll(inputs...)
	if len(got) != len(want) {
		t.Errorf("Resolver.ResolveAll() returned %d items, want %d", len(got), len(want))
		return
	}

	for i := range got {
		if got[i] != want[i] {
			t.Errorf("Resolver.ResolveAll()[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestResolver_GetVariableValue(t *testing.T) {
	// Set up environment variable for testing
	if err := os.Setenv("TEST_ENV_VAR", "env_value"); err != nil {
		t.Fatalf("Failed to set environment variable: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("TEST_ENV_VAR"); err != nil {
			t.Logf("Warning: Failed to unset environment variable: %v", err)
		}
	}()

	r := NewResolver().
		WithConfigVars(map[string]string{"CONFIG_VAR": "config_value"}).
		WithEnvFileVars(map[string]string{"ENV_FILE_VAR": "env_file_value"}).
		WithParamVars(map[string]string{"PARAM_VAR": "param_value"})

	tests := []struct {
		name    string
		varName string
		want    string
		found   bool
	}{
		{
			name:    "config variable",
			varName: "CONFIG_VAR",
			want:    "config_value",
			found:   true,
		},
		{
			name:    "env file variable",
			varName: "ENV_FILE_VAR",
			want:    "env_file_value",
			found:   true,
		},
		{
			name:    "param variable",
			varName: "PARAM_VAR",
			want:    "param_value",
			found:   true,
		},
		{
			name:    "system env variable",
			varName: "TEST_ENV_VAR",
			want:    "env_value",
			found:   true,
		},
		{
			name:    "variable not found",
			varName: "NOT_FOUND",
			want:    "",
			found:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, found := r.GetVariableValue(tt.varName)
			if got != tt.want {
				t.Errorf("Resolver.GetVariableValue() value = %v, want %v", got, tt.want)
			}
			if found != tt.found {
				t.Errorf("Resolver.GetVariableValue() found = %v, want %v", found, tt.found)
			}
		})
	}
}
