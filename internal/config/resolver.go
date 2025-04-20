package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// ResolveConfigPath determines which config file to use based on precedence.
func ResolveConfigPath(flagPath string) (string, error) {
	if flagPath != "" {
		return flagPath, nil
	}
	if envPath := os.Getenv("YXA_CONFIG"); envPath != "" {
		return envPath, nil
	}
	local := filepath.Join(".", "yxa.yml")
	if _, err := os.Stat(local); err == nil {
		return local, nil
	}
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		xdgPath := filepath.Join(xdg, "yxa", "config.yml")
		if _, err := os.Stat(xdgPath); err == nil {
			return xdgPath, nil
		}
	}
	home, err := os.UserHomeDir()
	if err == nil {
		homePath := filepath.Join(home, ".yxa.yml")
		if _, err := os.Stat(homePath); err == nil {
			return homePath, nil
		}
	}
	return "", fmt.Errorf("no yxa config file found")
}
