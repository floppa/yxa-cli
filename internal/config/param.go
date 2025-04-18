package config

// Param represents a command parameter, which can be either a flag or a positional parameter
type Param struct {
	Name        string `yaml:"name"`
	Type        string `yaml:"type"`
	Default     string `yaml:"default,omitempty"`
	Description string `yaml:"description"`
	Required    bool   `yaml:"required,omitempty"`
	Flag        bool   `yaml:"flag,omitempty"`     // Is this a flag parameter?
	Position    int    `yaml:"position,omitempty"` // Position for positional params (-1 means not positional)
}

// ProcessParamDefinition extracts name and shorthand from the parameter definition
func ProcessParamDefinition(paramDef string) (name, shorthand string) {
	parts := []string{paramDef}
	if len(paramDef) > 0 {
		if idx := indexOf(paramDef, "|"); idx >= 0 {
			parts = []string{paramDef[:idx], paramDef[idx+1:]}
		}
	}
	
	name = parts[0]
	
	if len(parts) > 1 {
		shorthand = parts[1]
	}
	
	return name, shorthand
}

// indexOf returns the index of the first instance of sep in s, or -1 if sep is not present in s.
func indexOf(s, sep string) int {
	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			return i
		}
	}
	return -1
}
