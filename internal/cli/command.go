package cli

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/floppa/yxa-cli/internal/config"
	"github.com/floppa/yxa-cli/internal/executor"
)

// CommandHandler manages command execution with dependencies and variables
type CommandHandler struct {
	Config       *config.ProjectConfig
	Executor     executor.CommandExecutor
	executedCmds map[string]bool
	DryRun       bool
}

// SetDryRun sets the dry-run mode for the handler
func (h *CommandHandler) SetDryRun(dryRun bool) {
	h.DryRun = dryRun
}

// NewCommandHandler creates a new command handler
func NewCommandHandler(cfg *config.ProjectConfig, exec executor.CommandExecutor) *CommandHandler {
	return &CommandHandler{
		Config:       cfg,
		Executor:     exec,
		executedCmds: make(map[string]bool),
	}
}

// ExecuteCommand runs a command with its dependencies using the provided variables
func (h *CommandHandler) ExecuteCommand(cmdName string, cmdVars map[string]string) error {
	// Check if command has already been executed
	if h.executedCmds[cmdName] {
		return nil
	}

	// Mark the command as executed
	h.executedCmds[cmdName] = true

	// Check if this is a subcommand reference (format: parent:subcommandname)
	parts := strings.Split(cmdName, ":")
	if len(parts) > 1 {
		parentName := parts[0]
		subCmdName := parts[1]

		// Get the parent command
		parentCmd, ok := h.Config.Commands[parentName]
		if !ok {
			return fmt.Errorf("parent command '%s' not found", parentName)
		}

		// Get the subcommand by name
		subCmd, ok := parentCmd.Commands[subCmdName]
		if !ok {
			return fmt.Errorf("subcommand '%s' not found in command '%s'", subCmdName, parentName)
		}

		// Execute the subcommand
		return h.executeCommandWithDependencies(cmdName, subCmd, cmdVars)
	}

	// Get the command from the config
	cmd, ok := h.Config.Commands[cmdName]
	if !ok {
		return fmt.Errorf("command '%s' not found", cmdName)
	}

	// Execute the command with proper error handling
	if err := h.executeCommandWithDependencies(cmdName, cmd, cmdVars); err != nil {
		return err
	}

	return nil
}

// executeCommandWithDependencies handles command execution with dependencies
func (h *CommandHandler) executeCommandWithDependencies(cmdName string, cmd config.Command, cmdVars map[string]string) error {
	// Check if the command has a condition
	if err := h.checkCommandCondition(cmdName, cmd, cmdVars); err != nil {
		return err
	}

	// Execute dependencies first
	if err := h.executeDependencies(cmdName, cmd.Depends, cmdVars); err != nil {
		return err
	}

	// If the command has subcommands, it's a command group - just list them
	if len(cmd.Commands) > 0 {
		return h.listSubcommands(cmdName, cmd)
	}

	// Validate the command and determine if it's executable
	if err := h.validateCommandExecutability(cmdName, cmd); err != nil {
		return err
	}

	// Execute the command body (pre-hook, main command, post-hook)
	return h.executeCommandBody(cmdName, cmd, cmdVars)
}

// validateCommandExecutability checks if a command is executable
// A command is executable if it has a run command, tasks, or is a dependency aggregator
func (h *CommandHandler) validateCommandExecutability(cmdName string, cmd config.Command) error {
	// If the command has no run or tasks defined, but has dependencies,
	// it's just a task aggregator, which is fine
	if cmd.Run == "" && len(cmd.Tasks) == 0 {
		if len(cmd.Depends) > 0 {
			// This is a valid dependency aggregator
			return nil
		}
		// Command has no functionality defined
		return fmt.Errorf("command '%s' has no 'run', 'tasks', or 'commands' defined", cmdName)
	}
	
	// Command has either run or tasks defined, so it's executable
	return nil
}

// checkCommandCondition evaluates a command's condition if present
func (h *CommandHandler) checkCommandCondition(cmdName string, cmd config.Command, cmdVars map[string]string) error {
	if cmd.Condition == "" {
		return nil
	}

	// Evaluate the condition with parameter variables
	if !h.Config.EvaluateConditionWithParams(cmd.Condition, cmdVars) {
		fmt.Printf("Skipping command '%s' (condition not met: %s)\n", cmdName, cmd.Condition)
		return nil
	}

	return nil
}

// executeDependencies executes all dependencies for a command
// executeDependencies executes all dependencies for a command
func (h *CommandHandler) executeDependencies(cmdName string, dependencies []string, cmdVars map[string]string) error {
	// If there are no dependencies, return immediately
	if len(dependencies) == 0 {
		return nil
	}

	// Special handling for check-all command
	if cmdName == "check-all" {
		return h.executeCheckAllDependencies(dependencies, cmdVars)
	}

	// Standard behavior for other commands
	return h.executeStandardDependencies(cmdName, dependencies, cmdVars)
}

// executeCheckAllDependencies executes dependencies for the check-all command,
// continuing execution even if some dependencies fail
func (h *CommandHandler) executeCheckAllDependencies(dependencies []string, cmdVars map[string]string) error {
	// Execute all dependencies and collect errors
	errors := h.executeAllDependenciesWithErrorCollection(dependencies, cmdVars)

	// Return a combined error if any dependencies failed
	return h.formatDependencyErrors(errors)
}

// executeAllDependenciesWithErrorCollection executes all dependencies and collects errors
// without stopping on the first error
func (h *CommandHandler) executeAllDependenciesWithErrorCollection(dependencies []string, cmdVars map[string]string) []string {
	var errors []string

	for _, dep := range dependencies {
		// Don't print the execution message here, it will be printed in runMainCommand
		if err := h.ExecuteCommand(dep, cmdVars); err != nil {
			// Log the error but continue with other dependencies
			fmt.Printf("Error executing command '%s': %v\n", dep, err)
			errors = append(errors, fmt.Sprintf("'%s': %v", dep, err))
		}
	}

	return errors
}

// formatDependencyErrors formats a list of dependency errors into a single error
// or returns nil if there are no errors
func (h *CommandHandler) formatDependencyErrors(errors []string) error {
	if len(errors) > 0 {
		return fmt.Errorf("one or more dependencies failed: %s", strings.Join(errors, "; "))
	}

	return nil
}

// executeStandardDependencies executes dependencies for standard commands,
// stopping at the first failure
func (h *CommandHandler) executeStandardDependencies(cmdName string, dependencies []string, cmdVars map[string]string) error {
	// Execute each dependency, stopping at the first error
	return h.executeSequentialDependencies(cmdName, dependencies, cmdVars)
}

// executeSequentialDependencies executes dependencies in sequence and stops at the first error
func (h *CommandHandler) executeSequentialDependencies(cmdName string, dependencies []string, cmdVars map[string]string) error {
	for _, dep := range dependencies {
		if err := h.ExecuteCommand(dep, cmdVars); err != nil {
			return fmt.Errorf("failed to execute dependency '%s' for command '%s': %w", dep, cmdName, err)
		}
	}

	return nil
}

// executeCommandBody executes the main command body including pre/post hooks
func (h *CommandHandler) executeCommandBody(cmdName string, cmd config.Command, cmdVars map[string]string) error {
	if err := h.runPreHook(cmdName, cmd, cmdVars); err != nil {
		return err
	}

	timeout, err := h.parseTimeout(cmdName, cmd.Timeout)
	if err != nil {
		return err
	}

	if err := h.runMainCommand(cmdName, cmd, cmdVars, timeout); err != nil {
		return err
	}

	if err := h.runPostHook(cmdName, cmd, cmdVars); err != nil {
		return err
	}

	return nil
}

// runPreHook executes the pre-hook if defined
func (h *CommandHandler) runPreHook(cmdName string, cmd config.Command, cmdVars map[string]string) error {
	return h.executeHook(cmdName, "pre", cmd.Pre, cmdVars)
}

// runMainCommand handles the main command execution logic
func (h *CommandHandler) runMainCommand(cmdName string, cmd config.Command, cmdVars map[string]string, timeout time.Duration) error {
	fmt.Printf("Executing command '%s'...\n", cmdName)

	// Check for subcommands first
	if len(cmd.Commands) > 0 {
		return h.listSubcommands(cmdName, cmd)
	}

	if cmd.Run != "" {
		return h.runSingleCommand(cmdName, cmd, cmdVars, timeout)
	} else if len(cmd.Tasks) > 0 {
		if cmd.Parallel {
			return h.runParallelCommands(cmdName, cmd, cmdVars, timeout)
		}
		return h.runSequentialCommands(cmdName, cmd, cmdVars, timeout)
	}
	return nil
}

// runSingleCommand executes a single command (Run)
func (h *CommandHandler) runSingleCommand(cmdName string, cmd config.Command, cmdVars map[string]string, timeout time.Duration) error {
	cmdStr := h.replaceVariablesInString(cmd.Run, cmdVars)
	if h.DryRun {
		fmt.Printf("[dry-run] Would execute: %s\n", cmdStr)
		return nil
	}
	if err := h.Executor.Execute(cmdStr, timeout); err != nil {
		return fmt.Errorf("failed to execute command '%s': %w", cmdName, err)
	}
	return nil
}

// runParallelCommands executes tasks in parallel
func (h *CommandHandler) runParallelCommands(cmdName string, cmd config.Command, cmdVars map[string]string, timeout time.Duration) error {
	if h.DryRun {
		for _, subCmd := range cmd.Tasks {
			cmdStr := h.replaceVariablesInString(subCmd, cmdVars)
			fmt.Printf("[dry-run] Would execute (parallel): %s\n", cmdStr)
		}
		return nil
	}
	if err := h.executeParallelCommands(cmdName, cmd, timeout); err != nil {
		return fmt.Errorf("failed to execute parallel commands for '%s': %w", cmdName, err)
	}
	return nil
}

// runSequentialCommands executes tasks sequentially
func (h *CommandHandler) runSequentialCommands(cmdName string, cmd config.Command, cmdVars map[string]string, timeout time.Duration) error {
	if h.DryRun {
		for _, subCmd := range cmd.Tasks {
			cmdStr := h.replaceVariablesInString(subCmd, cmdVars)
			fmt.Printf("[dry-run] Would execute (sequential): %s\n", cmdStr)
		}
		return nil
	}
	if err := h.executeSequentialCommands(cmdName, cmd, timeout); err != nil {
		return fmt.Errorf("failed to execute sequential commands for '%s': %w", cmdName, err)
	}
	return nil
}

// runPostHook executes the post-hook if defined
func (h *CommandHandler) runPostHook(cmdName string, cmd config.Command, cmdVars map[string]string) error {
	return h.executeHook(cmdName, "post", cmd.Post, cmdVars)
}

// executeHook executes a pre or post hook for a command
func (h *CommandHandler) executeHook(cmdName, hookType, hookCmd string, cmdVars map[string]string) error {
	if hookCmd == "" {
		return nil
	}

	fmt.Printf("Executing %s-hook for '%s'...\n", hookType, cmdName)
	hookCmdStr := h.replaceVariablesInString(hookCmd, cmdVars)
	if h.DryRun {
		fmt.Printf("[dry-run] Would execute (%s-hook): %s\n", hookType, hookCmdStr)
		return nil
	}
	if err := h.Executor.Execute(hookCmdStr, 0); err != nil {
		return fmt.Errorf("failed to execute %s-hook for command '%s': %w", hookType, cmdName, err)
	}

	return nil
}

// parseTimeout parses the timeout string into a time.Duration
func (h *CommandHandler) parseTimeout(cmdName, timeoutStr string) (time.Duration, error) {
	if timeoutStr == "" {
		return 0, nil
	}

	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		return 0, fmt.Errorf("invalid timeout '%s' for command '%s': %w", timeoutStr, cmdName, err)
	}

	fmt.Printf("Command '%s' will timeout after %s\n", cmdName, timeout)
	return timeout, nil
}

// replaceVariablesInString replaces variables in a string with their values from the provided map
func (h *CommandHandler) replaceVariablesInString(input string, vars map[string]string) string {
	return h.Config.ReplaceVariablesWithParams(input, vars)
}

// listSubcommands lists all subcommands of a command
func (h *CommandHandler) listSubcommands(cmdName string, cmd config.Command) error {
	stdout := h.Executor.GetStdout()

	// Print header and description
	if err := h.printSubcommandHeader(stdout, cmdName, cmd.Description); err != nil {
		return err
	}

	// Get the maximum length of subcommand names for alignment
	maxLen := h.getMaxSubcommandNameLength(cmd.Commands)

	// List all subcommands with their descriptions
	return h.printSubcommandList(stdout, cmd.Commands, maxLen)
}

// printSubcommandHeader prints the header and description for the subcommand list
func (h *CommandHandler) printSubcommandHeader(stdout io.Writer, cmdName, description string) error {
	// Print command name header
	_, err := fmt.Fprintf(stdout, "Available subcommands for '%s':\n", cmdName)
	if err != nil {
		return fmt.Errorf("failed to write to stdout: %w", err)
	}

	// Print description if available
	if description != "" {
		_, err := fmt.Fprintf(stdout, "Description: %s\n", description)
		if err != nil {
			return fmt.Errorf("failed to write to stdout: %w", err)
		}
	}

	// Add a blank line for readability
	_, err = fmt.Fprintln(stdout)
	if err != nil {
		return fmt.Errorf("failed to write to stdout: %w", err)
	}

	return nil
}

// getMaxSubcommandNameLength calculates the maximum length of subcommand names for alignment
func (h *CommandHandler) getMaxSubcommandNameLength(commands map[string]config.Command) int {
	maxLen := 0
	for subCmdName := range commands {
		if len(subCmdName) > maxLen {
			maxLen = len(subCmdName)
		}
	}
	return maxLen
}

// printSubcommandList prints the list of subcommands with their descriptions
func (h *CommandHandler) printSubcommandList(stdout io.Writer, commands map[string]config.Command, maxLen int) error {
	for subCmdName, subCmd := range commands {
		// Get the description or fallback to run command or default text
		description := h.getSubcommandDescription(subCmd)

		// Print the formatted subcommand line
		_, err := fmt.Fprintf(stdout, "  %-*s  %s\n", maxLen, subCmdName, description)
		if err != nil {
			return fmt.Errorf("failed to write to stdout: %w", err)
		}
	}

	return nil
}

// getSubcommandDescription returns the description for a subcommand, with fallbacks
func (h *CommandHandler) getSubcommandDescription(cmd config.Command) string {
	if cmd.Description != "" {
		return cmd.Description
	}
	
	if cmd.Run != "" {
		return cmd.Run
	}
	
	return "(No description)"
}

// executeSequentialCommands executes multiple tasks sequentially
func (h *CommandHandler) executeSequentialCommands(cmdName string, cmd config.Command, timeout time.Duration) error {
	for i, cmdStr := range cmd.Tasks {
		cmdStr = h.replaceVariablesInString(cmdStr, nil)
		fmt.Printf("Executing sequential sub-command #%d for '%s'...\n", i+1, cmdName)

		err := h.Executor.Execute(cmdStr, timeout)
		if flusher, ok := h.Executor.GetStdout().(interface{ Flush() error }); ok {
			_ = flusher.Flush()
		}
		if err != nil {
			return fmt.Errorf("sub-command #%d for '%s' failed: %w", i+1, cmdName, err)
		}
	}
	return nil
}

// Note: executeParallelCommands is implemented in parallel.go
