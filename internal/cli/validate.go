package cli

import (
	"fmt"

	"github.com/floppa/yxa-cli/internal/config"
	"github.com/floppa/yxa-cli/internal/errors"
)

// validateCommandDependencies validates that there are no circular dependencies
// in the command configuration
func validateCommandDependencies(cfg *config.ProjectConfig) error {
	// Check if config is nil
	if cfg == nil {
		return nil // No config, no dependencies to validate
	}

	// Create a map to track visited commands during traversal
	visited := make(map[string]bool)

	// Create a map to track commands in the current path
	inPath := make(map[string]bool)

	// Check each command for circular dependencies
	for cmdName := range cfg.Commands {
		// Skip commands that have already been validated
		if visited[cmdName] {
			continue
		}

		// Check this command's dependency tree
		path := []string{}
		if err := validateDependencyTree(cfg, cmdName, visited, inPath, path); err != nil {
			return err
		}
	}

	return nil
}

// validateDependencyTree performs a depth-first traversal of the dependency tree
// to detect circular dependencies
func validateDependencyTree(
	cfg *config.ProjectConfig,
	cmdName string,
	visited map[string]bool,
	inPath map[string]bool,
	path []string,
) error {
	// Check if this command is already in the current path (circular dependency)
	if inPath[cmdName] {
		return errors.NewCircularDependencyConfigError(path, cmdName)
	}

	// Check if this command has already been validated
	if visited[cmdName] {
		return nil
	}

	// Get the command configuration
	cmd, ok := cfg.Commands[cmdName]
	if !ok {
		return fmt.Errorf("command '%s' not found", cmdName)
	}

	// Mark this command as in the current path
	inPath[cmdName] = true
	path = append(path, cmdName)

	// Recursively validate dependencies
	for _, depName := range cmd.Depends {
		if err := validateDependencyTree(cfg, depName, visited, inPath, path); err != nil {
			return err
		}
	}

	// Mark this command as validated
	visited[cmdName] = true

	// Remove this command from the current path
	inPath[cmdName] = false

	return nil
}
