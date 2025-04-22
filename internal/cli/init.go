package cli

import (
	"fmt"
	"os"

	"github.com/floppa/yxa-cli/internal/config"
	"github.com/floppa/yxa-cli/internal/executor"
)

// InitializeApp sets up the basic root command structure and loads configuration.
var InitializeApp = func() (*RootCommand, error) {
	// Create a default executor
	exec := executor.NewDefaultExecutor()

	// Create the root command with nil config initially
	root := NewRootCommand(nil, exec)

	// Load configuration and register commands
	localPath := "./yxa.yml"
	if _, statErr := os.Stat(localPath); statErr == nil {
		// Local config file exists, load it
		cfg, err := config.LoadConfigFrom(localPath)
		if err != nil {
			return root, fmt.Errorf("failed to load local configuration: %w", err)
		}
		
		// Store the loaded config
		root.Config = cfg
		
		// Validate command dependencies
		if err = validateCommandDependencies(cfg); err != nil {
			// Return the root command even when there's a validation error
			return root, fmt.Errorf("invalid command dependencies: %w", err)
		}
		
		// Initialize the handler with the config
		root.Handler = NewCommandHandler(cfg, exec)
		
		// Register commands now to ensure they're available
		root.registerCommands()
	}

	return root, nil
}
