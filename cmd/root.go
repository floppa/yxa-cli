package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/magnuseriksson/yxa-cli/config"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "yxa",
	Short: "A CLI tool that executes commands defined in yxa.yml",
	Long: `yxa is a CLI tool that loads a yxa.yml file in the current directory
and registers commands defined in it. Each command has a name and a shell command to execute.`,
}

// Execute executes the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// Global variables to track command execution and dependencies
var (
	cfg              *config.ProjectConfig
	executedCommands = make(map[string]bool)
)

// executeCommand runs a command with its dependencies
func executeCommand(cmdName string) error {
	// Check if command has already been executed
	if executedCommands[cmdName] {
		return nil
	}

	// Get the command
	cmd, ok := cfg.Commands[cmdName]
	if !ok {
		return fmt.Errorf("command '%s' not found", cmdName)
	}

	// Execute dependencies first
	for _, depName := range cmd.Depends {
		// Skip circular dependencies
		if depName == cmdName {
			fmt.Printf("Warning: Skipping circular dependency '%s' -> '%s'\n", cmdName, depName)
			continue
		}

		// Execute the dependency
		fmt.Printf("Executing dependency '%s' for command '%s'...\n", depName, cmdName)
		if err := executeCommand(depName); err != nil {
			return fmt.Errorf("failed to execute dependency '%s' for command '%s': %w", depName, cmdName, err)
		}
	}

	// Execute the command
	fmt.Printf("Executing command '%s'...\n", cmdName)
	shellCmd := exec.Command("sh", "-c", cmd.Run)
	shellCmd.Stdout = os.Stdout
	shellCmd.Stderr = os.Stderr
	shellCmd.Stdin = os.Stdin

	if err := shellCmd.Run(); err != nil {
		return fmt.Errorf("error executing command '%s': %w", cmdName, err)
	}

	// Mark command as executed
	executedCommands[cmdName] = true
	return nil
}

// init function is called when the package is initialized
func init() {
	// Load the project configuration
	var err error
	cfg, err = config.LoadConfig()
	if err != nil {
		fmt.Printf("Warning: %v\n", err)
		return
	}

	// Register commands from the configuration
	for name, cmd := range cfg.Commands {
		// Create a copy of the variables for the closure
		cmdName := name
		cmdRun := cmd.Run

		// Format the short description to show dependencies
		shortDesc := fmt.Sprintf("Run '%s'", cmdRun)
		if len(cmd.Depends) > 0 {
			shortDesc += fmt.Sprintf(" (depends on: %s)", strings.Join(cmd.Depends, ", "))
		}

		// Create a new cobra command
		command := &cobra.Command{
			Use:   cmdName,
			Short: shortDesc,
			Run: func(cmd *cobra.Command, args []string) {
				// Reset executed commands for each run
				executedCommands = make(map[string]bool)

				// Execute the command with its dependencies
				if err := executeCommand(cmdName); err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			},
		}

		// Add the command to the root command
		rootCmd.AddCommand(command)
	}
}
