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

func TestProcessFlagParameter_Success(t *testing.T) {
	cmd := &cobra.Command{
		Use: "test",
	}

	// Register all types of flags
	cmd.Flags().String("string-param", "string-value", "string param")
	cmd.Flags().Int("int-param", 42, "int param")
	cmd.Flags().Float64("float-param", 3.14, "float param")
	cmd.Flags().Bool("bool-param", true, "bool param")
	cmd.Flags().String("default-param", "default-value", "default param with unknown type")

	// Test string parameter
	param := config.Param{
		Name:        "string-param",
		Type:        "string",
		Description: "A string param",
		Flag:        true,
	}
	val, err := processFlagParameter(cmd, param)
	assert.NoError(t, err)
	assert.Equal(t, "string-value", val)

	// Test int parameter
	param = config.Param{
		Name:        "int-param",
		Type:        "int",
		Description: "An int param",
		Flag:        true,
	}
	val, err = processFlagParameter(cmd, param)
	assert.NoError(t, err)
	assert.Equal(t, "42", val)

	// Test float parameter
	param = config.Param{
		Name:        "float-param",
		Type:        "float",
		Description: "A float param",
		Flag:        true,
	}
	val, err = processFlagParameter(cmd, param)
	assert.NoError(t, err)
	assert.Equal(t, "3.14", val)

	// Test bool parameter
	param = config.Param{
		Name:        "bool-param",
		Type:        "bool",
		Description: "A bool param",
		Flag:        true,
	}
	val, err = processFlagParameter(cmd, param)
	assert.NoError(t, err)
	assert.Equal(t, "true", val)

	// Test default type (unknown type defaults to string)
	param = config.Param{
		Name:        "default-param",
		Type:        "unknown",
		Description: "A param with unknown type",
		Flag:        true,
	}
	val, err = processFlagParameter(cmd, param)
	assert.NoError(t, err)
	assert.Equal(t, "default-value", val)

	// Test parameter with shorthand notation
	cmd.Flags().String("shorthand", "shorthand-value", "param with shorthand")
	param = config.Param{
		Name:        "shorthand|s",
		Type:        "string",
		Description: "A param with shorthand",
		Flag:        true,
	}
	val, err = processFlagParameter(cmd, param)
	assert.NoError(t, err)
	assert.Equal(t, "shorthand-value", val)
}

func TestProcessFlagParameter_Errors(t *testing.T) {
	cmd := &cobra.Command{
		Use: "test",
	}
	param := config.Param{
		Name:        "missing-int",
		Type:        "int",
		Description: "An int param that is not set",
		Flag:        true,
	}
	// Do not register the flag
	_, err := processFlagParameter(cmd, param)
	if err == nil {
		t.Errorf("Should error if flag is missing")
	}

	// Register as string, but expect int
	cmd.Flags().String("wrong-type", "", "wrong type")
	param.Name = "wrong-type"
	param.Type = "int"
	_, err = processFlagParameter(cmd, param)
	if err == nil {
		t.Errorf("Should error if flag type is wrong")
	}

	// Unknown type should default to string, but if not present should error
	param.Name = "not-present"
	param.Type = "unknown"
	_, err = processFlagParameter(cmd, param)
	if err == nil {
		t.Errorf("Should error for unknown type if flag is missing")
	}
}

func TestValidateRequiredPositionalParameters_Errors(t *testing.T) {
	tests := []struct {
		name      string
		posParams map[int]config.Param
		args      []string
		hasError  bool
	}{
		{
			name: "all required present",
			posParams: map[int]config.Param{
				0: {Name: "first", Required: true},
				1: {Name: "second", Required: true},
			},
			args:     []string{"v1", "v2"},
			hasError: false,
		},
		{
			name: "one required missing",
			posParams: map[int]config.Param{
				0: {Name: "first", Required: true},
				1: {Name: "second", Required: true},
			},
			args:     []string{"v1"},
			hasError: true,
		},
		{
			name: "multiple required missing",
			posParams: map[int]config.Param{
				0: {Name: "first", Required: true},
				1: {Name: "second", Required: true},
				2: {Name: "third", Required: true},
			},
			args:     []string{},
			hasError: true,
		},
		{
			name: "no required",
			posParams: map[int]config.Param{
				0: {Name: "first", Required: false},
				1: {Name: "second", Required: false},
			},
			args:     []string{},
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRequiredPositionalParameters(tt.posParams, tt.args)
			if tt.hasError && err == nil {
				t.Errorf("Expected error but got nil")
			}
			if !tt.hasError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
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
