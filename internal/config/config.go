package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/floppa/yxa-cli/internal/variables"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// ProjectConfig represents the structure of the yxa.yml file
type ProjectConfig struct {
	Name       string             `yaml:"name"`
	Variables  map[string]string  `yaml:"variables,omitempty"`
	Commands   map[string]Command `yaml:"commands"`
	WorkingDir string             `yaml:"workingdir,omitempty"` // Directory-level workingdir
	// Internal field to store environment variables (not from YAML)
	envVars map[string]string
}

// Command represents a command defined in the project.yml file
type Command struct {
	Run         string            `yaml:"run"`                   // Main command to execute
	Tasks       []string          `yaml:"tasks,omitempty"`       // Multiple tasks for parallel or sequential execution
	Commands    []Command         `yaml:"commands,omitempty"`    // Subcommands for hierarchical command structures
	Depends     []string          `yaml:"depends,omitempty"`     // Dependencies to execute first
	Description string            `yaml:"description,omitempty"` // Command description
	Condition   string            `yaml:"condition,omitempty"`   // Condition to evaluate before running
	Pre         string            `yaml:"pre,omitempty"`         // Command to run before the main command
	Post        string            `yaml:"post,omitempty"`        // Command to run after the main command
	Timeout     string            `yaml:"timeout,omitempty"`     // Timeout for command execution (e.g. "30s", "5m")
	Parallel    bool              `yaml:"parallel,omitempty"`    // Whether to run tasks in parallel
	Params      []Param           `yaml:"params,omitempty"`      // Command parameters (flags and positional)
	WorkingDir  string            `yaml:"workingdir,omitempty"`  // Command-level workingdir
}

// LoadConfig loads the project configuration from the yxa.yml file (legacy, cwd)
func LoadConfig() (*ProjectConfig, error) {
	return LoadConfigFrom(filepath.Join(".", "yxa.yml"))
}

// MergeConfigs merges global and project configs. Project config values take precedence.
func MergeConfigs(global, project *ProjectConfig) *ProjectConfig {
	if global == nil && project == nil {
		return &ProjectConfig{}
	}
	if global == nil {
		return project
	}
	if project == nil {
		return global
	}
	merged := *global // shallow copy
	if project.Name != "" {
		merged.Name = project.Name
	}

	// Merge variables
	merged.Variables = map[string]string{}
	for k, v := range global.Variables {
		merged.Variables[k] = v
	}
	for k, v := range project.Variables {
		merged.Variables[k] = v
	}
	// Merge commands
	merged.Commands = map[string]Command{}
	for k, v := range global.Commands {
		merged.Commands[k] = v
	}
	for k, v := range project.Commands {
		merged.Commands[k] = v
	}
	return &merged
}

// LoadConfigFrom loads the project configuration from the specified file path, merging with global config if present.
func LoadConfigFrom(configPath string) (*ProjectConfig, error) {
	// Check if the file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", configPath)
	}

	// Read the file
	// #nosec G304 -- This is intentional as reading the config file is the core functionality
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse the YAML data
	var config ProjectConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Initialize the environment variables map
	config.envVars = make(map[string]string)

	// Load environment variables from .env file if it exists (always relative to cwd)
	envPath := filepath.Join(".", ".env")
	if _, err := os.Stat(envPath); err == nil {
		envVars, err := godotenv.Read(envPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read .env file: %w", err)
		}
		for key, value := range envVars {
			config.envVars[key] = value
		}
	}

	// Process the commands to replace variables
	for name, cmd := range config.Commands {
		cmd.Run = config.ReplaceVariables(cmd.Run)
		config.Commands[name] = cmd
	}

	// Try to load and merge global config if present
	globalConfigPath, err := getGlobalConfigPath(configPath)
	if err == nil {
		globalConfig, err := LoadConfigFrom(globalConfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load global config: %w", err)
		}
		config = *MergeConfigs(globalConfig, &config)
	}

	return &config, nil
}

// getGlobalConfigPath returns the path to the global config, or error if not found or not applicable.
func getGlobalConfigPath(currentPath string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	globalCandidates := []string{}
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		globalCandidates = append(globalCandidates, filepath.Join(xdg, "yxa", "config.yml"))
	}
	globalCandidates = append(globalCandidates, filepath.Join(home, ".yxa.yml"))
	for _, p := range globalCandidates {
		if p == currentPath {
			continue
		}
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("no global config found")
}

// ReplaceVariables replaces variables in the given string with their values
func (c *ProjectConfig) ReplaceVariables(input string) string {
	// Create a variable resolver with the project's variables
	resolver := variables.NewResolver().
		WithConfigVars(c.Variables).
		WithEnvFileVars(c.envVars)

	// Resolve variables in the input string
	return resolver.Resolve(input)
}

// ReplaceVariablesWithParams replaces variables in the given string with their values,
// including parameter variables
func (c *ProjectConfig) ReplaceVariablesWithParams(input string, paramVars map[string]string) string {
	// Create a variable resolver with all variable sources
	resolver := variables.NewResolver().
		WithParamVars(paramVars).
		WithConfigVars(c.Variables).
		WithEnvFileVars(c.envVars)

	// Resolve variables in the input string
	return resolver.Resolve(input)
}

// EvaluateCondition evaluates a condition string and returns whether it's true
func (c *ProjectConfig) EvaluateCondition(condition string) bool {
	return c.EvaluateConditionWithParams(condition, nil)
}

// EvaluateConditionWithParams evaluates a condition string with parameter variables
func (c *ProjectConfig) EvaluateConditionWithParams(condition string, paramVars map[string]string) bool {
	if condition == "" {
		// Empty condition is always true
		return true
	}

	// Replace variables in the condition using all variable sources
	resolver := variables.NewResolver().
		WithParamVars(paramVars).
		WithConfigVars(c.Variables).
		WithEnvFileVars(c.envVars)
	condition = resolver.Resolve(condition)

	// Evaluate the resolved condition
	return evaluateConditionString(condition)
}

func evaluateConditionString(condition string) bool {
	// Simple equality check (e.g., "$GOOS == darwin")
	equalityPattern := regexp.MustCompile(`^\s*(.+?)\s*==\s*(.+?)\s*$`)
	if matches := equalityPattern.FindStringSubmatch(condition); len(matches) == 3 {
		return strings.TrimSpace(matches[1]) == strings.TrimSpace(matches[2])
	}

	// Simple inequality check (e.g., "$GOOS != darwin")
	inequalityPattern := regexp.MustCompile(`^\s*(.+?)\s*!=\s*(.+?)\s*$`)
	if matches := inequalityPattern.FindStringSubmatch(condition); len(matches) == 3 {
		return strings.TrimSpace(matches[1]) != strings.TrimSpace(matches[2])
	}

	// Contains check (e.g., "$PATH contains /usr/local")
	containsPattern := regexp.MustCompile(`^\s*(.+?)\s+contains\s+(.+?)\s*$`)
	if matches := containsPattern.FindStringSubmatch(condition); len(matches) == 3 {
		return strings.Contains(strings.TrimSpace(matches[1]), strings.TrimSpace(matches[2]))
	}

	// Exists check (e.g., "exists /path/to/file")
	existsPattern := regexp.MustCompile(`^\s*exists\s+(.+?)\s*$`)
	if matches := existsPattern.FindStringSubmatch(condition); len(matches) == 2 {
		_, err := os.Stat(strings.TrimSpace(matches[1]))
		return err == nil
	}

	// If we can't parse the condition, default to false
	return false
}
