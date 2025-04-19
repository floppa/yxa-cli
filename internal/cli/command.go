package cli

import (
	"fmt"
	"time"

	"github.com/floppa/yxa-cli/internal/config"
	"github.com/floppa/yxa-cli/internal/executor"
)

// CommandHandler manages command execution with dependencies and variables
type CommandHandler struct {
	Config        *config.ProjectConfig
	Executor      executor.CommandExecutor
	executedCmds  map[string]bool
	DryRun        bool
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

	// If the command has no run or commands defined, but has dependencies,
	// it's just a task aggregator, which is fine
	if cmd.Run == "" && len(cmd.Commands) == 0 {
		if len(cmd.Depends) > 0 {
			return nil
		}
		return fmt.Errorf("command '%s' has no 'run' or 'commands' defined", cmdName)
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
	for _, dep := range dependencies {
		if err := h.ExecuteCommand(dep, cmdVars); err != nil {
			return fmt.Errorf("failed to execute dependency '%s' for command '%s': %w", dep, cmdName, err)
		}
	}
	return nil
}

// executeCommandBody executes the main command body including pre/post hooks
func (h *CommandHandler) executeCommandBody(cmdName string, cmd config.Command, cmdVars map[string]string) error {
	// Execute pre-hook if defined
	if err := h.executeHook(cmdName, "pre", cmd.Pre, cmdVars); err != nil {
		return err
	}

	// Parse timeout if specified
	timeout, err := h.parseTimeout(cmdName, cmd.Timeout)
	if err != nil {
		return err
	}

	// Execute the command
	fmt.Printf("Executing command '%s'...\n", cmdName)

	if cmd.Run != "" {
		// Replace variables in the command
		cmdStr := h.replaceVariablesInString(cmd.Run, cmdVars)
		if h.DryRun {
			fmt.Printf("[dry-run] Would execute: %s\n", cmdStr)
			return nil
		}
		if err := h.Executor.Execute(cmdStr, timeout); err != nil {
			return fmt.Errorf("failed to execute command '%s': %w", cmdName, err)
		}
	} else if len(cmd.Commands) > 0 {
		if cmd.Parallel {
			if h.DryRun {
				for _, subCmd := range cmd.Commands {
					cmdStr := h.replaceVariablesInString(subCmd, cmdVars)
					fmt.Printf("[dry-run] Would execute (parallel): %s\n", cmdStr)
				}
				return nil
			}
			if err := h.executeParallelCommands(cmdName, cmd, timeout); err != nil {
				return fmt.Errorf("failed to execute parallel commands for '%s': %w", cmdName, err)
			}
		} else {
			if h.DryRun {
				for _, subCmd := range cmd.Commands {
					cmdStr := h.replaceVariablesInString(subCmd, cmdVars)
					fmt.Printf("[dry-run] Would execute (sequential): %s\n", cmdStr)
				}
				return nil
			}
			if err := h.executeSequentialCommands(cmdName, cmd, timeout); err != nil {
				return fmt.Errorf("failed to execute sequential commands for '%s': %w", cmdName, err)
			}
		}
	}

	// Execute post-hook if defined
	if err := h.executeHook(cmdName, "post", cmd.Post, cmdVars); err != nil {
		return err
	}

	return nil
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
	// Use the ProjectConfig's ReplaceVariablesWithParams for all variable substitution
	return h.Config.ReplaceVariablesWithParams(input, vars)
}

// executeSequentialCommands executes multiple commands sequentially
func (h *CommandHandler) executeSequentialCommands(cmdName string, cmd config.Command, timeout time.Duration) error {
	for name, cmdStr := range cmd.Commands {
		// Replace variables in the command
		cmdStr = h.replaceVariablesInString(cmdStr, nil)
		fmt.Printf("Executing sequential sub-command '%s' for '%s'...\n", name, cmdName)

		// Execute the command with timeout
		err := h.Executor.Execute(cmdStr, timeout)
		// Flush output after each subcommand (no-op for most writers, but ensures buffer is up to date)
		if flusher, ok := h.Executor.GetStdout().(interface{ Flush() error }); ok {
			_ = flusher.Flush()
		}
		if err != nil {
			// Do NOT return early: allow output to be present in the buffer before returning error
			return fmt.Errorf("sub-command '%s' for '%s' failed: %w", name, cmdName, err)
		}
	}

	return nil
}

// Note: executeParallelCommands is implemented in parallel.go
