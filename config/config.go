package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// ProjectConfig represents the structure of the yxa.yml file
type ProjectConfig struct {
	Name      string             `yaml:"name"`
	Variables map[string]string  `yaml:"variables,omitempty"`
	Commands  map[string]Command `yaml:"commands"`
	// Internal field to store environment variables (not from YAML)
	envVars   map[string]string
}

// Command represents a command defined in the project.yml file
type Command struct {
	Run         string            `yaml:"run"`                    // Main command to execute
	Commands    map[string]string `yaml:"commands,omitempty"`     // Multiple commands for parallel execution
	Depends     []string          `yaml:"depends,omitempty"`      // Dependencies to execute first
	Description string            `yaml:"description,omitempty"`  // Command description
	Condition   string            `yaml:"condition,omitempty"`   // Condition to evaluate before running
	Pre         string            `yaml:"pre,omitempty"`         // Command to run before the main command
	Post        string            `yaml:"post,omitempty"`        // Command to run after the main command
	Timeout     string            `yaml:"timeout,omitempty"`     // Timeout for command execution (e.g. "30s", "5m")
	Parallel    bool              `yaml:"parallel,omitempty"`    // Whether to run commands in parallel
}

// LoadConfig loads the project configuration from the yxa.yml file
func LoadConfig() (*ProjectConfig, error) {
	// Find the yxa.yml file in the current directory
	configPath := filepath.Join(".", "yxa.yml")
	
	// Check if the file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("yxa.yml not found in the current directory")
	}

	// Read the file
	// #nosec G304 -- This is intentional as reading the config file is the core functionality
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read yxa.yml: %w", err)
	}

	// Parse the YAML data
	var config ProjectConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse yxa.yml: %w", err)
	}

	// Initialize the environment variables map
	config.envVars = make(map[string]string)

	// Load environment variables from .env file if it exists
	envPath := filepath.Join(".", ".env")
	if _, err := os.Stat(envPath); err == nil {
		envVars, err := godotenv.Read(envPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read .env file: %w", err)
		}
		
		// Store the environment variables
		for key, value := range envVars {
			config.envVars[key] = value
		}
	}

	// Process the commands to replace variables
	for name, cmd := range config.Commands {
		cmd.Run = config.ReplaceVariables(cmd.Run)
		config.Commands[name] = cmd
	}

	return &config, nil
}

// ReplaceVariables replaces variables in the given string with their values
func (c *ProjectConfig) ReplaceVariables(input string) string {
	// Define regex pattern for variables: $VAR or ${VAR}
	pattern := regexp.MustCompile(`\$(\w+|\{\w+\})`)

	// Replace all occurrences
	result := pattern.ReplaceAllStringFunc(input, func(match string) string {
		// Extract variable name (remove $ and {} if present)
		varName := match[1:] // Remove $
		if strings.HasPrefix(varName, "{") && strings.HasSuffix(varName, "}") {
			varName = varName[1 : len(varName)-1] // Remove { and }
		}

		// Try to get value from different sources in order of priority
		// 1. YAML variables
		if value, ok := c.Variables[varName]; ok {
			return value
		}

		// 2. Environment variables from .env file
		if value, ok := c.envVars[varName]; ok {
			return value
		}

		// 3. System environment variables
		if value, ok := os.LookupEnv(varName); ok {
			return value
		}

		// If variable not found, return the original match
		return match
	})

	return result
}

// EvaluateCondition evaluates a condition string and returns whether it's true
func (c *ProjectConfig) EvaluateCondition(condition string) bool {
	if condition == "" {
		// Empty condition is always true
		return true
	}

	// Replace variables in the condition
	condition = c.ReplaceVariables(condition)

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
