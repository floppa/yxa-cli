package cli

import (
	"fmt"
	"os"

	"github.com/floppa/yxa-cli/internal/config"
	"github.com/floppa/yxa-cli/internal/executor"
)

// InitializeConfig loads the project configuration and validates it
func InitializeConfig() (*config.ProjectConfig, error) {
	// Load the configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Validate command dependencies
	if err := validateCommandDependencies(cfg); err != nil {
		return nil, fmt.Errorf("invalid command dependencies: %w", err)
	}

	return cfg, nil
}

// InitializeApp is a variable that holds the function to initialize the application
// This allows it to be mocked for testing
var InitializeApp = func() (*RootCommand, error) {
	// Load and validate configuration
	cfg, err := InitializeConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing application: %v\n", err)
		return nil, err
	}

	// Create a default executor
	exec := executor.NewDefaultExecutor()

	// Create the root command
	root := NewRootCommand(cfg, exec)

	return root, nil
}
