package variables

import (
	"os"
	"regexp"
	"strings"
)

// Resolver handles variable resolution from multiple sources
type Resolver struct {
	// Sources of variables in order of priority (highest first)
	ConfigVars   map[string]string // Variables from config file
	EnvFileVars  map[string]string // Variables from .env file
	ParamVars    map[string]string // Variables from command parameters
	SystemEnvVar bool              // Whether to check system environment variables
}

// NewResolver creates a new variable resolver
func NewResolver() *Resolver {
	return &Resolver{
		ConfigVars:   make(map[string]string),
		EnvFileVars:  make(map[string]string),
		ParamVars:    make(map[string]string),
		SystemEnvVar: true,
	}
}

// WithConfigVars adds config variables to the resolver
func (r *Resolver) WithConfigVars(vars map[string]string) *Resolver {
	// Range over map is safe even if map is nil
	for k, v := range vars {
		r.ConfigVars[k] = v
	}
	return r
}

// WithEnvFileVars adds .env file variables to the resolver
func (r *Resolver) WithEnvFileVars(vars map[string]string) *Resolver {
	// Range over map is safe even if map is nil
	for k, v := range vars {
		r.EnvFileVars[k] = v
	}
	return r
}

// WithParamVars adds parameter variables to the resolver
func (r *Resolver) WithParamVars(vars map[string]string) *Resolver {
	// Range over map is safe even if map is nil
	for k, v := range vars {
		r.ParamVars[k] = v
	}
	return r
}

// WithSystemEnvVar sets whether to check system environment variables
func (r *Resolver) WithSystemEnvVar(check bool) *Resolver {
	r.SystemEnvVar = check
	return r
}

// Resolve resolves variables in the given string
func (r *Resolver) Resolve(input string) string {
	if input == "" {
		return input
	}

	// Define regex pattern for variables: $VAR or ${VAR}
	pattern := regexp.MustCompile(`\$(\w+|\{\w+\})`)

	// Replace all occurrences
	result := pattern.ReplaceAllStringFunc(input, func(match string) string {
		// Extract variable name (remove $ and {} if present)
		varName := match[1:] // Remove $
		if strings.HasPrefix(varName, "{") && strings.HasSuffix(varName, "}") {
			varName = varName[1 : len(varName)-1] // Remove { and }
		}

		// Try to get value from different sources in order of priority
		// 1. Parameter variables (highest priority)
		if value, ok := r.ParamVars[varName]; ok {
			return value
		}

		// 2. Config variables
		if value, ok := r.ConfigVars[varName]; ok {
			return value
		}

		// 3. Environment variables from .env file
		if value, ok := r.EnvFileVars[varName]; ok {
			return value
		}

		// 4. System environment variables (if enabled)
		if r.SystemEnvVar {
			if value, ok := os.LookupEnv(varName); ok {
				return value
			}
		}

		// If variable not found, return the original match
		return match
	})

	return result
}

// ResolveAll resolves variables in all the given strings
func (r *Resolver) ResolveAll(inputs ...string) []string {
	results := make([]string, len(inputs))
	for i, input := range inputs {
		results[i] = r.Resolve(input)
	}
	return results
}

// GetVariableValue gets the value of a variable from all sources
func (r *Resolver) GetVariableValue(varName string) (string, bool) {
	// Check sources in order of priority
	// 1. Parameter variables (highest priority)
	if value, ok := r.ParamVars[varName]; ok {
		return value, true
	}

	// 2. Config variables
	if value, ok := r.ConfigVars[varName]; ok {
		return value, true
	}

	// 3. Environment variables from .env file
	if value, ok := r.EnvFileVars[varName]; ok {
		return value, true
	}

	// 4. System environment variables (if enabled)
	if r.SystemEnvVar {
		if value, ok := os.LookupEnv(varName); ok {
			return value, true
		}
	}

	return "", false
}
