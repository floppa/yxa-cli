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
	Run         string   `yaml:"run"`
	Depends     []string `yaml:"depends,omitempty"`
	Description string   `yaml:"description,omitempty"`
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
	// Regular expression to match ${VARIABLE} or $VARIABLE patterns
	varPattern := regexp.MustCompile(`\$\{([^}]+)\}|\$([A-Za-z0-9_]+)`)
	
	// Replace all occurrences of variables
	result := varPattern.ReplaceAllStringFunc(input, func(match string) string {
		// Extract variable name
		var varName string
		if strings.HasPrefix(match, "${") && strings.HasSuffix(match, "}") {
			varName = match[2 : len(match)-1]
		} else {
			varName = match[1:]
		}
		
		// Check if the variable is defined in the YAML variables
		if value, ok := c.Variables[varName]; ok {
			return value
		}
		
		// Check if the variable is defined in the .env file
		if value, ok := c.envVars[varName]; ok {
			return value
		}
		
		// Check if the variable is defined in the environment
		if value, ok := os.LookupEnv(varName); ok {
			return value
		}
		
		// If the variable is not found, return the original match
		return match
	})
	
	return result
}
