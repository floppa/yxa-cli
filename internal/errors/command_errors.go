package errors

import (
	"fmt"
	"strings"
)

// CommandError represents an error that occurred during command execution
type CommandError struct {
	CommandName string
	Message     string
	Err         error
}

// Error implements the error interface
func (e *CommandError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("command '%s': %s: %v", e.CommandName, e.Message, e.Err)
	}
	return fmt.Sprintf("command '%s': %s", e.CommandName, e.Message)
}

// Unwrap returns the underlying error
func (e *CommandError) Unwrap() error {
	return e.Err
}

// NewCommandError creates a new CommandError
func NewCommandError(cmdName, message string, err error) *CommandError {
	return &CommandError{
		CommandName: cmdName,
		Message:     message,
		Err:         err,
	}
}

// NewCommandNotFoundError creates a new error for when a command is not found
func NewCommandNotFoundError(cmdName string) *CommandError {
	return &CommandError{
		CommandName: cmdName,
		Message:     "command not found",
	}
}

// NewDependencyError creates a new error for when a dependency fails
func NewDependencyError(cmdName, depName string, err error) *CommandError {
	return &CommandError{
		CommandName: cmdName,
		Message:     fmt.Sprintf("failed to execute dependency '%s'", depName),
		Err:         err,
	}
}

// NewCircularDependencyError creates a new error for circular dependencies
func NewCircularDependencyError(path []string, cmdName string) *CommandError {
	// Create a readable path string showing the circular dependency
	pathStr := strings.Join(path, " -> ") + " -> " + cmdName
	
	return &CommandError{
		CommandName: cmdName,
		Message:     fmt.Sprintf("circular dependency detected: %s", pathStr),
	}
}

// NewExecutionError creates a new error for when command execution fails
func NewExecutionError(cmdName string, err error) *CommandError {
	return &CommandError{
		CommandName: cmdName,
		Message:     "execution failed",
		Err:         err,
	}
}

// NewHookError creates a new error for when a pre or post hook fails
func NewHookError(cmdName, hookType string, err error) *CommandError {
	return &CommandError{
		CommandName: cmdName,
		Message:     fmt.Sprintf("%s-hook execution failed", hookType),
		Err:         err,
	}
}

// NewTimeoutError creates a new error for when a timeout is invalid
func NewTimeoutError(cmdName, timeoutStr string, err error) *CommandError {
	return &CommandError{
		CommandName: cmdName,
		Message:     fmt.Sprintf("invalid timeout '%s'", timeoutStr),
		Err:         err,
	}
}

// NewParameterError creates a new error for parameter-related issues
func NewParameterError(cmdName, paramName, message string) *CommandError {
	return &CommandError{
		CommandName: cmdName,
		Message:     fmt.Sprintf("parameter '%s': %s", paramName, message),
	}
}
