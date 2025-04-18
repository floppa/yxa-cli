package errors

import (
	"fmt"
)

// VariableError represents an error that occurred during variable resolution
type VariableError struct {
	VariableName string // The name of the variable
	Context      string // The context where the error occurred (e.g., "command", "condition")
	Message      string // A descriptive error message
	Err          error  // The underlying error, if any
}

// Error implements the error interface
func (e *VariableError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("variable error for '%s' in %s: %s: %v", e.VariableName, e.Context, e.Message, e.Err)
	}
	return fmt.Sprintf("variable error for '%s' in %s: %s", e.VariableName, e.Context, e.Message)
}

// Unwrap returns the underlying error
func (e *VariableError) Unwrap() error {
	return e.Err
}

// NewVariableError creates a new VariableError
func NewVariableError(varName, context, message string, err error) *VariableError {
	return &VariableError{
		VariableName: varName,
		Context:      context,
		Message:      message,
		Err:          err,
	}
}

// NewVariableNotFoundError creates a new error for when a variable is not found
func NewVariableNotFoundError(varName, context string) *VariableError {
	return &VariableError{
		VariableName: varName,
		Context:      context,
		Message:      "variable not found",
	}
}

// NewVariableSubstitutionError creates a new error for when variable substitution fails
func NewVariableSubstitutionError(varName, context string, err error) *VariableError {
	return &VariableError{
		VariableName: varName,
		Context:      context,
		Message:      "substitution failed",
		Err:          err,
	}
}

// NewVariableTypeError creates a new error for when a variable has the wrong type
func NewVariableTypeError(varName, context, expectedType string) *VariableError {
	return &VariableError{
		VariableName: varName,
		Context:      context,
		Message:      fmt.Sprintf("expected type %s", expectedType),
	}
}

// NewVariableCircularReferenceError creates a new error for circular variable references
func NewVariableCircularReferenceError(varName, context string, path []string) *VariableError {
	pathStr := fmt.Sprintf("%s", path)
	return &VariableError{
		VariableName: varName,
		Context:      context,
		Message:      fmt.Sprintf("circular reference detected: %s", pathStr),
	}
}

// NewVariableResolutionLimitError creates a new error for when variable resolution exceeds the limit
func NewVariableResolutionLimitError(varName, context string, limit int) *VariableError {
	return &VariableError{
		VariableName: varName,
		Context:      context,
		Message:      fmt.Sprintf("resolution exceeded limit of %d iterations", limit),
	}
}

// NewConditionEvaluationError creates a new error for when condition evaluation fails
func NewConditionEvaluationError(condition, context string, err error) *VariableError {
	return &VariableError{
		VariableName: condition,
		Context:      context,
		Message:      "condition evaluation failed",
		Err:          err,
	}
}

// NewInvalidConditionError creates a new error for when a condition is invalid
func NewInvalidConditionError(condition, context, reason string) *VariableError {
	return &VariableError{
		VariableName: condition,
		Context:      context,
		Message:      fmt.Sprintf("invalid condition: %s", reason),
	}
}
