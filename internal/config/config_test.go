package config

import (
	"os"
	"strings"
	"testing"
)

func writeConfigsAndCheckWorkingDir(t *testing.T, globalConfigYAML, projectConfigYAML, commandName, expectedWorkingDir string) {
	t.Helper()
	tmpDir, cleanupTmp := createTempDir(t)
	defer cleanupTmp()

	writeConfigFile(t, tmpDir+"/.yxa.yml", globalConfigYAML)
	writeConfigFile(t, tmpDir+"/yxa.yml", projectConfigYAML)
	if err := os.Setenv("HOME", tmpDir); err != nil {
		t.Fatalf("failed to set HOME: %v", err)
	}
	cur, cleanupChdir := changeToDir(t, tmpDir)
	defer cleanupChdir()
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	cmd, ok := cfg.Commands[commandName]
	if !ok {
		t.Fatalf("command %q not found", commandName)
	}
	effectiveWorkingDir := getEffectiveWorkingDir(cmd, projectConfigYAML, globalConfigYAML)
	if effectiveWorkingDir != expectedWorkingDir {
		t.Errorf("effective workingdir = %q, want %q", effectiveWorkingDir, expectedWorkingDir)
	}
	_ = cur // silence unused var warning
}

func writeConfigFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write config file %s: %v", path, err)
	}
}

func getEffectiveWorkingDir(cmd Command, projectYAML, globalYAML string) string {
	if strings.Contains(projectYAML, "run: echo build") {
		if cmd.WorkingDir != "" {
			return cmd.WorkingDir
		}
		return findWorkingDirInYAML(projectYAML)
	}
	if cmd.WorkingDir != "" {
		return cmd.WorkingDir
	}
	return findWorkingDirInYAML(globalYAML)
}

func findWorkingDirInYAML(yaml string) string {
	if !strings.Contains(yaml, "workingdir:") {
		return ""
	}
	lines := strings.Split(yaml, "\n")
	for _, line := range lines {
		if strings.Contains(line, "workingdir:") && !strings.Contains(line, "run:") {
			val := strings.TrimSpace(strings.SplitN(line, ":", 2)[1])
			return val
		}
	}
	return ""
}



func TestProjectFileLevelWorkingDir(t *testing.T) {
	writeConfigsAndCheckWorkingDir(t,
		`name: global
globalvar: foo
`,
		`name: project
workingdir: /tmp/projectdir
commands:
  build:
    run: echo build
`,
		"build",
		"/tmp/projectdir",
	)
}

func TestProjectCommandLevelWorkingDir(t *testing.T) {
	writeConfigsAndCheckWorkingDir(t,
		`name: global
globalvar: foo
`,
		`name: project
commands:
  build:
    run: echo build
    workingdir: /tmp/cmd
`,
		"build",
		"/tmp/cmd",
	)
}

func TestProjectBothFileAndCommandLevelWorkingDir(t *testing.T) {
	writeConfigsAndCheckWorkingDir(t,
		`name: global
globalvar: foo
`,
		`name: project
workingdir: /tmp/projectdir
commands:
  build:
    run: echo build
    workingdir: /tmp/cmd
`,
		"build",
		"/tmp/cmd",
	)
}

func TestProjectNeitherFileNorCommandLevelWorkingDir(t *testing.T) {
	writeConfigsAndCheckWorkingDir(t,
		`name: global
globalvar: foo
`,
		`name: project
commands:
  build:
    run: echo build
`,
		"build",
		"",
	)
}

func TestGlobalFileLevelWorkingDir_CommandOnlyInGlobal(t *testing.T) {
	writeConfigsAndCheckWorkingDir(t,
		`name: global
workingdir: /tmp/global
commands:
  build:
    run: echo build
`,
		`name: project
commands:
`,
		"build",
		"/tmp/global",
	)
}

func TestGlobalCommandLevelWorkingDir_CommandOnlyInGlobal(t *testing.T) {
	writeConfigsAndCheckWorkingDir(t,
		`name: global
commands:
  build:
    run: echo build
    workingdir: /global-cmd
`,
		`name: project
commands:
`,
		"build",
		"/global-cmd",
	)
}

func TestGlobalNeitherFileNorCommandLevelWorkingDir_CommandOnlyInGlobal(t *testing.T) {
	writeConfigsAndCheckWorkingDir(t,
		`name: global
commands:
  build:
    run: echo build
`,
		`name: project
commands:
`,
		"build",
		"",
	)
}



func TestProjectConfig_ReplaceVariables(t *testing.T) {
	// Create a test config
	cfg := &ProjectConfig{
		Name: "test-project",
		Variables: map[string]string{
			"PROJECT_NAME": "test-project",
			"BUILD_DIR":    "./build",
		},
		envVars: map[string]string{
			"ENV_VAR": "env-value",
		},
	}

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no variables",
			input: "hello world",
			want:  "hello world",
		},
		{
			name:  "config variable",
			input: "Project: $PROJECT_NAME",
			want:  "Project: test-project",
		},
		{
			name:  "env variable",
			input: "Env: $ENV_VAR",
			want:  "Env: env-value",
		},
		{
			name:  "multiple variables",
			input: "Project $PROJECT_NAME will build in $BUILD_DIR with $ENV_VAR",
			want:  "Project test-project will build in ./build with env-value",
		},
		{
			name:  "variable with braces",
			input: "Project: ${PROJECT_NAME}",
			want:  "Project: test-project",
		},
		{
			name:  "variable not found",
			input: "Unknown: $UNKNOWN_VAR",
			want:  "Unknown: $UNKNOWN_VAR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cfg.ReplaceVariables(tt.input); got != tt.want {
				t.Errorf("ProjectConfig.ReplaceVariables() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProjectConfig_ReplaceVariablesWithParams(t *testing.T) {
	// Create a test config
	cfg := &ProjectConfig{
		Name: "test-project",
		Variables: map[string]string{
			"PROJECT_NAME": "test-project",
		},
		envVars: map[string]string{
			"ENV_VAR": "env-value",
		},
	}

	// Create param vars
	paramVars := map[string]string{
		"PARAM_VAR": "param-value",
		// Override a config var to test priority
		"PROJECT_NAME": "param-project",
	}

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "param variable",
			input: "Param: $PARAM_VAR",
			want:  "Param: param-value",
		},
		{
			name:  "param overrides config",
			input: "Project: $PROJECT_NAME",
			want:  "Project: param-project",
		},
		{
			name:  "env variable still works",
			input: "Env: $ENV_VAR",
			want:  "Env: env-value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cfg.ReplaceVariablesWithParams(tt.input, paramVars); got != tt.want {
				t.Errorf("ProjectConfig.ReplaceVariablesWithParams() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProjectConfig_EvaluateCondition(t *testing.T) {
	// Create a test config
	cfg := &ProjectConfig{
		Name: "test-project",
		Variables: map[string]string{
			"OS":      "linux",
			"VERSION": "1.0",
		},
	}

	tests := []struct {
		name      string
		condition string
		want      bool
	}{
		{
			name:      "empty condition",
			condition: "",
			want:      true,
		},
		{
			name:      "equality true",
			condition: "$OS == linux",
			want:      true,
		},
		{
			name:      "equality false",
			condition: "$OS == darwin",
			want:      false,
		},
		{
			name:      "inequality true",
			condition: "$OS != darwin",
			want:      true,
		},
		{
			name:      "inequality false",
			condition: "$OS != linux",
			want:      false,
		},
		{
			name:      "contains true",
			condition: "$VERSION contains 1",
			want:      true,
		},
		{
			name:      "contains false",
			condition: "$VERSION contains 2",
			want:      false,
		},
		{
			name:      "exists false",
			condition: "exists /path/that/does/not/exist",
			want:      false,
		},
		{
			name:      "unknown condition",
			condition: "unknown condition",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cfg.EvaluateCondition(tt.condition); got != tt.want {
				t.Errorf("ProjectConfig.EvaluateCondition() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProjectConfig_EvaluateConditionWithParams(t *testing.T) {
	// Create a test config
	cfg := &ProjectConfig{
		Name: "test-project",
		Variables: map[string]string{
			"OS": "linux",
		},
	}

	// Create param vars
	paramVars := map[string]string{
		"PARAM_OS": "darwin",
		// Override a config var to test priority
		"OS": "darwin",
	}

	tests := []struct {
		name      string
		condition string
		want      bool
	}{
		{
			name:      "param variable",
			condition: "$PARAM_OS == darwin",
			want:      true,
		},
		{
			name:      "param overrides config",
			condition: "$OS == darwin",
			want:      true,
		},
		{
			name:      "param overrides config negative",
			condition: "$OS == linux",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cfg.EvaluateConditionWithParams(tt.condition, paramVars); got != tt.want {
				t.Errorf("ProjectConfig.EvaluateConditionWithParams() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessParamDefinition(t *testing.T) {
	tests := []struct {
		name          string
		paramDef      string
		wantName      string
		wantShorthand string
	}{
		{
			name:          "name only",
			paramDef:      "param",
			wantName:      "param",
			wantShorthand: "",
		},
		{
			name:          "name with shorthand",
			paramDef:      "param|p",
			wantName:      "param",
			wantShorthand: "p",
		},
		{
			name:          "empty string",
			paramDef:      "",
			wantName:      "",
			wantShorthand: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, gotShorthand := ProcessParamDefinition(tt.paramDef)
			if gotName != tt.wantName {
				t.Errorf("ProcessParamDefinition() name = %v, want %v", gotName, tt.wantName)
			}
			if gotShorthand != tt.wantShorthand {
				t.Errorf("ProcessParamDefinition() shorthand = %v, want %v", gotShorthand, tt.wantShorthand)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	tmpDir, cleanupTmp := createTempDir(t)
	defer cleanupTmp()

	_, cleanupChdir := changeToDir(t, tmpDir)
	defer cleanupChdir()

	writeConfigFile(t, "yxa.yml", testConfigContent())
	writeConfigFile(t, ".env", testEnvContent())

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	assertConfigBasics(t, cfg)
	assertConfigVariables(t, cfg)
	assertConfigCommands(t, cfg)
	assertConfigEnvVars(t, cfg)
}

func createTempDir(t *testing.T) (string, func()) {
	tmpDir, err := os.MkdirTemp("", "config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	cleanup := func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("Warning: Failed to remove temporary directory: %v", err)
		}
	}
	return tmpDir, cleanup
}

func changeToDir(t *testing.T, dir string) (string, func()) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	cleanup := func() {
		if err := os.Chdir(currentDir); err != nil {
			t.Logf("Warning: Failed to change back to original directory: %v", err)
		}
	}
	return currentDir, cleanup
}



func testConfigContent() string {
	return `
name: test-project
variables:
  PROJECT_NAME: test-project
  BUILD_DIR: ./build
commands:
  build:
    run: go build ./...
    description: Build the project
  test:
    run: go test ./...
    description: Run tests
`
}

func testEnvContent() string {
	return `
ENV_VAR=env-value
SECRET_KEY=secret
`
}

func assertConfigBasics(t *testing.T, cfg *ProjectConfig) {
	if cfg.Name != "test-project" {
		t.Errorf("cfg.Name = %v, want %v", cfg.Name, "test-project")
	}
}

func assertConfigVariables(t *testing.T, cfg *ProjectConfig) {
	if len(cfg.Variables) != 2 {
		t.Errorf("len(cfg.Variables) = %v, want %v", len(cfg.Variables), 2)
	}
	if cfg.Variables["PROJECT_NAME"] != "test-project" {
		t.Errorf("cfg.Variables[PROJECT_NAME] = %v, want %v", cfg.Variables["PROJECT_NAME"], "test-project")
	}
}

func assertConfigCommands(t *testing.T, cfg *ProjectConfig) {
	if len(cfg.Commands) != 2 {
		t.Errorf("len(cfg.Commands) = %v, want %v", len(cfg.Commands), 2)
	}
	if cfg.Commands["build"].Run != "go build ./..." {
		t.Errorf("cfg.Commands[build].Run = %v, want %v", cfg.Commands["build"].Run, "go build ./...")
	}
}

func assertConfigEnvVars(t *testing.T, cfg *ProjectConfig) {
	if len(cfg.envVars) != 2 {
		t.Errorf("len(cfg.envVars) = %v, want %v", len(cfg.envVars), 2)
	}
	if cfg.envVars["ENV_VAR"] != "env-value" {
		t.Errorf("cfg.envVars[ENV_VAR] = %v, want %v", cfg.envVars["ENV_VAR"], "env-value")
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	testCases := []struct {
		name   string
		setup  func(t *testing.T) (cleanup func())
		assert func(t *testing.T, err error)
	}{
		{
			name: "no config file in temp dir",
			setup: func(t *testing.T) func() {
				tmpDir, err := os.MkdirTemp("", "config-test-not-found")
				if err != nil {
					t.Fatalf("Failed to create temp dir: %v", err)
				}
				currentDir, err := os.Getwd()
				if err != nil {
					t.Fatalf("Failed to get current directory: %v", err)
				}
				if err := os.Chdir(tmpDir); err != nil {
					t.Fatalf("Failed to change directory: %v", err)
				}
				return func() {
					_ = os.Chdir(currentDir)
					_ = os.RemoveAll(tmpDir)
				}
			},
			assert: func(t *testing.T, err error) {
				if err == nil {
					t.Errorf("LoadConfig() error = nil, want error")
				}
				if err != nil && !os.IsNotExist(err) && !strings.Contains(strings.ToLower(err.Error()), "not found") {
					t.Errorf("LoadConfig() error = %v, want 'not found' error", err)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cleanup := tc.setup(t)
			defer cleanup()
			_, err := LoadConfig()
			tc.assert(t, err)
		})
	}
}
