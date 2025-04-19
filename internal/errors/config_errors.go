package errors

import (
	"fmt"
	"strings"
)

// ConfigError represents an error that occurred during configuration processing
type ConfigError struct {
	Section string // The section of the config where the error occurred (e.g., "command", "parameter")
	Name    string // The name of the item with the error (e.g., command name, parameter name)
	Message string // A descriptive error message
	Err     error  // The underlying error, if any
}

// Error implements the error interface
func (e *ConfigError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("config error in %s '%s': %s: %v", e.Section, e.Name, e.Message, e.Err)
	}
	return fmt.Sprintf("config error in %s '%s': %s", e.Section, e.Name, e.Message)
}

// Unwrap returns the underlying error
func (e *ConfigError) Unwrap() error {
	return e.Err
}

// NewConfigError creates a new ConfigError
func NewConfigError(section, name, message string, err error) *ConfigError {
	return &ConfigError{
		Section: section,
		Name:    name,
		Message: message,
		Err:     err,
	}
}

// NewCommandConfigError creates a new error for command configuration issues
func NewCommandConfigError(cmdName, message string, err error) *ConfigError {
	return &ConfigError{
		Section: "command",
		Name:    cmdName,
		Message: message,
		Err:     err,
	}
}

// NewParameterConfigError creates a new error for parameter configuration issues
func NewParameterConfigError(cmdName, paramName, message string, err error) *ConfigError {
	return &ConfigError{
		Section: fmt.Sprintf("command '%s' parameter", cmdName),
		Name:    paramName,
		Message: message,
		Err:     err,
	}
}

// NewDependencyConfigError creates a new error for dependency configuration issues
func NewDependencyConfigError(cmdName, depName, message string, err error) *ConfigError {
	return &ConfigError{
		Section: fmt.Sprintf("command '%s' dependency", cmdName),
		Name:    depName,
		Message: message,
		Err:     err,
	}
}

// NewCircularDependencyConfigError creates a new error for circular dependencies
func NewCircularDependencyConfigError(path []string, cmdName string) *ConfigError {
	// Create a readable path string showing the circular dependency
	pathStr := strings.Join(path, " -> ") + " -> " + cmdName

	return &ConfigError{
		Section: "command dependency",
		Name:    cmdName,
		Message: fmt.Sprintf("circular dependency detected: %s", pathStr),
	}
}

// NewTimeoutConfigError creates a new error for timeout configuration issues
func NewTimeoutConfigError(cmdName, timeoutStr string, err error) *ConfigError {
	return &ConfigError{
		Section: "command timeout",
		Name:    cmdName,
		Message: fmt.Sprintf("invalid timeout '%s'", timeoutStr),
		Err:     err,
	}
}

// NewParallelConfigError creates a new error for parallel command configuration issues
func NewParallelConfigError(cmdName, message string) *ConfigError {
	return &ConfigError{
		Section: "command parallel",
		Name:    cmdName,
		Message: message,
	}
}

// NewDuplicateParameterError creates a new error for duplicate parameter names
func NewDuplicateParameterError(cmdName, paramName string) *ConfigError {
	return &ConfigError{
		Section: fmt.Sprintf("command '%s' parameter", cmdName),
		Name:    paramName,
		Message: "duplicate parameter name",
	}
}

// NewInvalidParameterTypeError creates a new error for invalid parameter types
func NewInvalidParameterTypeError(cmdName, paramName, paramType string) *ConfigError {
	return &ConfigError{
		Section: fmt.Sprintf("command '%s' parameter", cmdName),
		Name:    paramName,
		Message: fmt.Sprintf("unsupported parameter type '%s'", paramType),
	}
}

// NewParameterPositionError creates a new error for parameter position issues
func NewParameterPositionError(cmdName, paramName1, paramName2 string, position int) *ConfigError {
	return &ConfigError{
		Section: fmt.Sprintf("command '%s' parameter position", cmdName),
		Name:    fmt.Sprintf("%s, %s", paramName1, paramName2),
		Message: fmt.Sprintf("conflicting position %d", position),
	}
}

// NewParameterGapError creates a new error for gaps in positional parameters
func NewParameterGapError(cmdName string, position int) *ConfigError {
	return &ConfigError{
		Section: fmt.Sprintf("command '%s' parameter position", cmdName),
		Name:    fmt.Sprintf("position %d", position),
		Message: "gap in positional parameters",
	}
}
