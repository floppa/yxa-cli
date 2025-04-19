package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/floppa/yxa-cli/internal/config"
	"github.com/spf13/cobra"
)

// addParametersToCommand adds parameters as flags to a cobra command
func addParametersToCommand(cmd *cobra.Command, params []config.Param) {
	// Helper: add string flag
	addStringFlag := func(cmd *cobra.Command, name, shorthand, def, desc string) {
		cmd.Flags().StringP(name, shorthand, def, desc)
	}
	// Helper: add int flag
	addIntFlag := func(cmd *cobra.Command, name, shorthand, def, desc string, paramName string) {
		defaultVal := 0
		if def != "" {
			var err error
			defaultVal, err = strconv.Atoi(def)
			if err != nil {
				fmt.Printf("Warning: Invalid default value '%s' for int parameter '%s', using 0\n", def, paramName)
			}
		}
		cmd.Flags().IntP(name, shorthand, defaultVal, desc)
	}
	// Helper: add float flag
	addFloatFlag := func(cmd *cobra.Command, name, shorthand, def, desc string, paramName string) {
		defaultVal := 0.0
		if def != "" {
			var err error
			defaultVal, err = strconv.ParseFloat(def, 64)
			if err != nil {
				fmt.Printf("Warning: Invalid default value '%s' for float parameter '%s', using 0.0\n", def, paramName)
			}
		}
		cmd.Flags().Float64P(name, shorthand, defaultVal, desc)
	}
	// Helper: add bool flag
	addBoolFlag := func(cmd *cobra.Command, name, shorthand, def, desc string, paramName string) {
		defaultVal := false
		if def != "" {
			var err error
			defaultVal, err = strconv.ParseBool(def)
			if err != nil {
				fmt.Printf("Warning: Invalid default value '%s' for bool parameter '%s', using false\n", def, paramName)
			}
		}
		cmd.Flags().BoolP(name, shorthand, defaultVal, desc)
	}

	// Create maps to track parameter positions and names
	posParams := make(map[int]config.Param)

	// Process each parameter
	for _, param := range params {
		// Skip positional parameters for flag registration
		if !param.Flag && param.Position >= 0 {
			posParams[param.Position] = param
			continue
		}

		name, shorthand := processParamName(param.Name)
		switch strings.ToLower(param.Type) {
		case "string":
			addStringFlag(cmd, name, shorthand, param.Default, param.Description)
		case "int":
			addIntFlag(cmd, name, shorthand, param.Default, param.Description, param.Name)
		case "float":
			addFloatFlag(cmd, name, shorthand, param.Default, param.Description, param.Name)
		case "bool":
			addBoolFlag(cmd, name, shorthand, param.Default, param.Description, param.Name)
		default:
			addStringFlag(cmd, name, shorthand, param.Default, param.Description)
		}

		// Mark required parameters
		if param.Required {
			if err := cmd.MarkFlagRequired(name); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to mark flag '%s' as required: %v\n", name, err)
			}
		}
	}
}


// processParameters processes command parameters and returns a map of parameter values
func processParameters(cmd *cobra.Command, args []string, params []config.Param) (map[string]string, error) {
	// Helper: process a single flag parameter and return its value as a string
	processFlagParameter := func(cmd *cobra.Command, param config.Param) (string, error) {
		name, _ := processParamName(param.Name)
		switch strings.ToLower(param.Type) {
		case "string":
			if val, err := cmd.Flags().GetString(name); err == nil {
				return val, nil
			} else {
				return "", fmt.Errorf("error getting string parameter '%s': %w", name, err)
			}
		case "int":
			if val, err := cmd.Flags().GetInt(name); err == nil {
				return strconv.Itoa(val), nil
			} else {
				return "", fmt.Errorf("error getting int parameter '%s': %w", name, err)
			}
		case "float":
			if val, err := cmd.Flags().GetFloat64(name); err == nil {
				return strconv.FormatFloat(val, 'f', -1, 64), nil
			} else {
				return "", fmt.Errorf("error getting float parameter '%s': %w", name, err)
			}
		case "bool":
			if val, err := cmd.Flags().GetBool(name); err == nil {
				return strconv.FormatBool(val), nil
			} else {
				return "", fmt.Errorf("error getting bool parameter '%s': %w", name, err)
			}
		default:
			if val, err := cmd.Flags().GetString(name); err == nil {
				return val, nil
			} else {
				return "", fmt.Errorf("error getting parameter '%s': %w", name, err)
			}
		}
	}
	// Create a map to store parameter values
	paramVars := make(map[string]string)

	// Create a map to track positional parameters
	posParams := make(map[int]config.Param)

	// First pass: identify positional parameters
	for _, param := range params {
		if !param.Flag && param.Position >= 0 {
			posParams[param.Position] = param
		}
	}

	// Second pass: process flag parameters
	for _, param := range params {
		// Skip positional parameters for now
		if !param.Flag && param.Position >= 0 {
			continue
		}

		value, err := processFlagParameter(cmd, param)
		if err != nil {
			return nil, err
		}
		paramVars[param.Name] = value
	}

	// Helper: process a single positional parameter
	processPositionalParameter := func(arg string, param config.Param) string {
		// For now, just return the arg as-is (could add type conversion if needed)
		return arg
	}

	// Helper: validate required positional parameters
	validateRequiredPositionalParameters := func(posParams map[int]config.Param, args []string) error {
		for pos, param := range posParams {
			if param.Required && pos >= len(args) {
				return fmt.Errorf("required positional parameter '%s' not provided", param.Name)
			}
		}
		return nil
	}

	// Third pass: process positional parameters
	for i, arg := range args {
		if param, ok := posParams[i]; ok {
			paramVars[param.Name] = processPositionalParameter(arg, param)
		}
	}

	// Validate required positional parameters
	if err := validateRequiredPositionalParameters(posParams, args); err != nil {
		return nil, err
	}

	return paramVars, nil
}

// processParamName extracts name and shorthand from the parameter name
func processParamName(paramName string) (name, shorthand string) {
	parts := []string{paramName}
	if len(paramName) > 0 {
		if idx := strings.Index(paramName, "|"); idx >= 0 {
			parts = []string{paramName[:idx], paramName[idx+1:]}
		}
	}

	name = parts[0]

	if len(parts) > 1 {
		shorthand = parts[1]
	}

	return name, shorthand
}
