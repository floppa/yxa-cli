package cli

import (
	"github.com/floppa/yxa-cli/internal/executor"
)

// InitializeApp sets up the basic root command structure without loading configuration.
// Configuration loading and command registration will happen in PersistentPreRunE.
var InitializeApp = func() (*RootCommand, error) {
	// Create a default executor
	exec := executor.NewDefaultExecutor()

	// Create the root command with nil config initially
	// Config will be loaded and commands registered in PersistentPreRunE
	root := NewRootCommand(nil, exec)

	return root, nil
}
