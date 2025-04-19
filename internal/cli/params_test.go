package cli

import (
	"testing"

	"github.com/floppa/yxa-cli/internal/config"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestProcessParamName(t *testing.T) {
	tests := []struct {
		name          string
		paramName     string
		expectedName  string
		expectedShort string
	}{
		{
			name:          "simple name",
			paramName:     "test",
			expectedName:  "test",
			expectedShort: "",
		},
		{
			name:          "name with shorthand",
			paramName:     "test|t",
			expectedName:  "test",
			expectedShort: "t",
		},
		{
			name:          "name with multiple pipes",
			paramName:     "test|t|extra",
			expectedName:  "test",
			expectedShort: "t|extra",
		},
		{
			name:          "empty name",
			paramName:     "",
			expectedName:  "",
			expectedShort: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, short := processParamName(tt.paramName)
			assert.Equal(t, tt.expectedName, name)
			assert.Equal(t, tt.expectedShort, short)
		})
	}
}

func TestAddParametersToCommand(t *testing.T) {
	// Create a test command
	cmd := &cobra.Command{
		Use:   "test",
		Short: "Test command",
		Run:   func(cmd *cobra.Command, args []string) {},
	}

	// Define test parameters
	params := []config.Param{
		{
			Name:        "string-param",
			Type:        "string",
			Description: "A string parameter",
			Default:     "default-value",
			Required:    false,
			Flag:        true,
		},
		{
			Name:        "int-param",
			Type:        "int",
			Description: "An integer parameter",
			Default:     "42",
			Required:    false,
			Flag:        true,
		},
		{
			Name:        "float-param",
			Type:        "float",
			Description: "A float parameter",
			Default:     "3.14",
			Required:    false,
			Flag:        true,
		},
		{
			Name:        "bool-param",
			Type:        "bool",
			Description: "A boolean parameter",
			Default:     "true",
			Required:    false,
			Flag:        true,
		},
		{
			Name:        "required-param",
			Type:        "string",
			Description: "A required parameter",
			Default:     "",
			Required:    true,
			Flag:        true,
		},
		{
			Name:        "positional-param",
			Type:        "string",
			Description: "A positional parameter",
			Default:     "pos-default",
			Required:    false,
			Flag:        false,
			Position:    0,
		},
		{
			Name:        "invalid-int",
			Type:        "int",
			Description: "Parameter with invalid int default",
			Default:     "not-an-int",
			Required:    false,
			Flag:        true,
		},
		{
			Name:        "invalid-float",
			Type:        "float",
			Description: "Parameter with invalid float default",
			Default:     "not-a-float",
			Required:    false,
			Flag:        true,
		},
		{
			Name:        "invalid-bool",
			Type:        "bool",
			Description: "Parameter with invalid bool default",
			Default:     "not-a-bool",
			Required:    false,
			Flag:        true,
		},
		{
			Name:        "unknown-type",
			Type:        "unknown",
			Description: "Parameter with unknown type",
			Default:     "default",
			Required:    false,
			Flag:        true,
		},
	}

	// Add parameters to the command
	addParametersToCommand(cmd, params)

	// Verify flags were added correctly
	stringFlag, err := cmd.Flags().GetString("string-param")
	assert.NoError(t, err)
	assert.Equal(t, "default-value", stringFlag)

	intFlag, err := cmd.Flags().GetInt("int-param")
	assert.NoError(t, err)
	assert.Equal(t, 42, intFlag)

	floatFlag, err := cmd.Flags().GetFloat64("float-param")
	assert.NoError(t, err)
	assert.Equal(t, 3.14, floatFlag)

	boolFlag, err := cmd.Flags().GetBool("bool-param")
	assert.NoError(t, err)
	assert.Equal(t, true, boolFlag)

	// Check that required flag is marked as required
	// The pflag.Flag type doesn't expose Required field directly, so we'll verify
	// that the flag exists and is registered correctly
	requiredFlag := cmd.Flags().Lookup("required-param")
	assert.NotNil(t, requiredFlag)

	// We can't directly check if it's required, but we can verify it was registered
	assert.Equal(t, "", requiredFlag.DefValue)

	// Verify invalid defaults were handled gracefully
	invalidIntFlag, err := cmd.Flags().GetInt("invalid-int")
	assert.NoError(t, err)
	assert.Equal(t, 0, invalidIntFlag)

	invalidFloatFlag, err := cmd.Flags().GetFloat64("invalid-float")
	assert.NoError(t, err)
	assert.Equal(t, 0.0, invalidFloatFlag)

	invalidBoolFlag, err := cmd.Flags().GetBool("invalid-bool")
	assert.NoError(t, err)
	assert.Equal(t, false, invalidBoolFlag)

	// Verify unknown type defaults to string
	unknownTypeFlag, err := cmd.Flags().GetString("unknown-type")
	assert.NoError(t, err)
	assert.Equal(t, "default", unknownTypeFlag)
}

func TestProcessParameters(t *testing.T) {
	// Create a test command
	cmd := &cobra.Command{
		Use:   "test",
		Short: "Test command",
		Run:   func(cmd *cobra.Command, args []string) {},
	}

	// Define test parameters
	params := []config.Param{
		{
			Name:        "string-param",
			Type:        "string",
			Description: "A string parameter",
			Default:     "default-value",
			Required:    false,
			Flag:        true,
		},
		{
			Name:        "int-param",
			Type:        "int",
			Description: "An integer parameter",
			Default:     "42",
			Required:    false,
			Flag:        true,
		},
		{
			Name:        "positional-param",
			Type:        "string",
			Description: "A positional parameter",
			Default:     "pos-default",
			Required:    false,
			Flag:        false,
			Position:    0,
		},
		{
			Name:        "second-pos",
			Type:        "int",
			Description: "Second positional parameter",
			Default:     "0",
			Required:    false,
			Flag:        false,
			Position:    1,
		},
	}

	// Add parameters to the command
	addParametersToCommand(cmd, params)

	// Set flag values
	if err := cmd.Flags().Set("string-param", "flag-value"); err != nil {
		t.Fatalf("Failed to set string-param flag: %v", err)
	}
	if err := cmd.Flags().Set("int-param", "100"); err != nil {
		t.Fatalf("Failed to set int-param flag: %v", err)
	}

	// Test with positional args
	args := []string{"pos-value", "200"}
	paramVars, err := processParameters(cmd, args, params)
	assert.NoError(t, err)

	// Verify parameter values
	assert.Equal(t, "flag-value", paramVars["string-param"])
	assert.Equal(t, "100", paramVars["int-param"])
	assert.Equal(t, "pos-value", paramVars["positional-param"])
	assert.Equal(t, "200", paramVars["second-pos"])

	// Test with missing positional args (the implementation doesn't use defaults for positional params)
	args = []string{}
	paramVars, err = processParameters(cmd, args, params)
	assert.NoError(t, err)

	// Verify flag parameter values are present
	assert.Equal(t, "flag-value", paramVars["string-param"])
	assert.Equal(t, "100", paramVars["int-param"])

	// Positional parameters won't be set when args are missing
	_, posParamPresent := paramVars["positional-param"]
	assert.False(t, posParamPresent, "Positional parameter should not be set when no args provided")
	_, secondPosPresent := paramVars["second-pos"]
	assert.False(t, secondPosPresent, "Second positional parameter should not be set when no args provided")

	// Test with invalid positional parameter type
	params[3].Type = "invalid"
	args = []string{"pos-value", "not-an-int"}
	paramVars, err = processParameters(cmd, args, params)
	assert.NoError(t, err) // Should not error, just use the value as-is
	assert.Equal(t, "not-an-int", paramVars["second-pos"])
}
