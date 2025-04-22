package cli

import (
	"fmt"
	"strconv"
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

	// Check if this is a subcommand reference (format: parent:index)
	parts := strings.Split(cmdName, ":")
	if len(parts) > 1 {
		parentName := parts[0]
		subCmdIndex, err := strconv.Atoi(parts[1])
		if err != nil {
			return fmt.Errorf("invalid subcommand index in '%s': %w", cmdName, err)
		}

		// Get the parent command
		parentCmd, ok := h.Config.Commands[parentName]
		if !ok {
			return fmt.Errorf("parent command '%s' not found", parentName)
		}

		// Check if the subcommand index is valid
		if subCmdIndex < 1 || subCmdIndex > len(parentCmd.Commands) {
			return fmt.Errorf("subcommand index %d out of range for command '%s' (valid range: 1-%d)", 
				subCmdIndex, parentName, len(parentCmd.Commands))
		}

		// Get the subcommand (adjusting for 1-based indexing in the UI)
		subCmd := parentCmd.Commands[subCmdIndex-1]

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

	// If the command has no run or tasks defined, but has dependencies,
	// it's just a task aggregator, which is fine
	if cmd.Run == "" && len(cmd.Tasks) == 0 {
		if len(cmd.Depends) > 0 {
			return nil
		}
		return fmt.Errorf("command '%s' has no 'run', 'tasks', or 'commands' defined", cmdName)
	}

	// Execute the command body (pre-hook, main command, post-hook)
	return h.executeCommandBody(cmdName, cmd, cmdVars)
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
func (h *CommandHandler) executeDependencies(cmdName string, dependencies []string, cmdVars map[string]string) error {
	// Special handling for check-all command to continue execution even if dependencies fail
	if cmdName == "check-all" {
		var errors []string
		for _, dep := range dependencies {
			// Don't print the execution message here, it will be printed in runMainCommand
			if err := h.ExecuteCommand(dep, cmdVars); err != nil {
				// Log the error but continue with other dependencies
				fmt.Printf("Error executing command '%s': %v\n", dep, err)
				errors = append(errors, fmt.Sprintf("'%s': %v", dep, err))
			}
		}
		
		// Return a combined error if any dependencies failed
		if len(errors) > 0 {
			return fmt.Errorf("one or more dependencies failed: %s", strings.Join(errors, "; "))
		}
		return nil
	}
	
	// Standard behavior for other commands
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
	_, err := fmt.Fprintf(stdout, "Available subcommands for '%s':\n", cmdName)
	if err != nil {
		return fmt.Errorf("failed to write to stdout: %w", err)
	}
	
	if cmd.Description != "" {
		_, err := fmt.Fprintf(stdout, "Description: %s\n", cmd.Description)
		if err != nil {
			return fmt.Errorf("failed to write to stdout: %w", err)
		}
	}
	
	_, err = fmt.Fprintln(stdout)
	if err != nil {
		return fmt.Errorf("failed to write to stdout: %w", err)
	}

	// Get the maximum length of subcommand names for alignment
	maxLen := 0
	for i := range cmd.Commands {
		subCmdName := fmt.Sprintf("%s:%d", cmdName, i+1)
		if len(subCmdName) > maxLen {
			maxLen = len(subCmdName)
		}
	}

	// List all subcommands with their descriptions
	for i, subCmd := range cmd.Commands {
		subCmdName := fmt.Sprintf("%s:%d", cmdName, i+1)
		description := subCmd.Description
		if description == "" {
			if subCmd.Run != "" {
				description = subCmd.Run
			} else {
				description = "(No description)"
			}
		}
		
		_, err := fmt.Fprintf(stdout, "  %-*s  %s\n", maxLen, subCmdName, description)
		if err != nil {
			return fmt.Errorf("failed to write to stdout: %w", err)
		}
	}

	return nil
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
