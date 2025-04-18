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
	// Create maps to track parameter positions and names
	posParams := make(map[int]config.Param)
	
	// Process each parameter
	for _, param := range params {
		// Skip positional parameters for flag registration
		if !param.Flag && param.Position >= 0 {
			posParams[param.Position] = param
			continue
		}
		
		// Process flag parameters
		name, shorthand := processParamName(param.Name)
		
		// Register the flag based on its type
		switch strings.ToLower(param.Type) {
		case "string":
			cmd.Flags().StringP(name, shorthand, param.Default, param.Description)
		case "int":
			defaultVal := 0
			if param.Default != "" {
				var err error
				defaultVal, err = strconv.Atoi(param.Default)
				if err != nil {
					fmt.Printf("Warning: Invalid default value '%s' for int parameter '%s', using 0\n", 
						param.Default, param.Name)
				}
			}
			cmd.Flags().IntP(name, shorthand, defaultVal, param.Description)
		case "float":
			defaultVal := 0.0
			if param.Default != "" {
				var err error
				defaultVal, err = strconv.ParseFloat(param.Default, 64)
				if err != nil {
					fmt.Printf("Warning: Invalid default value '%s' for float parameter '%s', using 0.0\n", 
						param.Default, param.Name)
				}
			}
			cmd.Flags().Float64P(name, shorthand, defaultVal, param.Description)
		case "bool":
			defaultVal := false
			if param.Default != "" {
				var err error
				defaultVal, err = strconv.ParseBool(param.Default)
				if err != nil {
					fmt.Printf("Warning: Invalid default value '%s' for bool parameter '%s', using false\n", 
						param.Default, param.Name)
				}
			}
			cmd.Flags().BoolP(name, shorthand, defaultVal, param.Description)
		default:
			// Default to string for unknown types
			cmd.Flags().StringP(name, shorthand, param.Default, param.Description)
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
		
		// Process flag parameters
		name, _ := processParamName(param.Name)
		
		// Get the flag value based on its type
		var value string
		switch strings.ToLower(param.Type) {
		case "string":
			if val, err := cmd.Flags().GetString(name); err == nil {
				value = val
			} else {
				return nil, fmt.Errorf("error getting string parameter '%s': %w", name, err)
			}
		case "int":
			if val, err := cmd.Flags().GetInt(name); err == nil {
				value = strconv.Itoa(val)
			} else {
				return nil, fmt.Errorf("error getting int parameter '%s': %w", name, err)
			}
		case "float":
			if val, err := cmd.Flags().GetFloat64(name); err == nil {
				value = strconv.FormatFloat(val, 'f', -1, 64)
			} else {
				return nil, fmt.Errorf("error getting float parameter '%s': %w", name, err)
			}
		case "bool":
			if val, err := cmd.Flags().GetBool(name); err == nil {
				value = strconv.FormatBool(val)
			} else {
				return nil, fmt.Errorf("error getting bool parameter '%s': %w", name, err)
			}
		default:
			// Default to string for unknown types
			if val, err := cmd.Flags().GetString(name); err == nil {
				value = val
			} else {
				return nil, fmt.Errorf("error getting parameter '%s': %w", name, err)
			}
		}
		
		// Store the parameter value
		paramVars[param.Name] = value
	}
	
	// Third pass: process positional parameters
	for i, arg := range args {
		// Find the parameter for this position
		if param, ok := posParams[i]; ok {
			// Store the parameter value
			paramVars[param.Name] = arg
		}
	}
	
	// Check if all required positional parameters are provided
	for pos, param := range posParams {
		if param.Required && pos >= len(args) {
			return nil, fmt.Errorf("required positional parameter '%s' not provided", param.Name)
		}
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
