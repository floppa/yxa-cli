package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// ResolveConfigPath determines which config file to use based on precedence.
func ResolveConfigPath(flagPath string) (string, error) {
	// Add debug information about the current working directory
	cwd, _ := os.Getwd()
	fmt.Printf("Resolving config path. Current working directory: %s\n", cwd)

	if flagPath != "" {
		fmt.Printf("Using config path from flag: %s\n", flagPath)
		return flagPath, nil
	}

	if envPath := os.Getenv("YXA_CONFIG"); envPath != "" {
		fmt.Printf("Using config path from YXA_CONFIG environment variable: %s\n", envPath)
		return envPath, nil
	}

	local := filepath.Join(".", "yxa.yml")
	fmt.Printf("Checking for local config at: %s\n", local)
	if _, err := os.Stat(local); err == nil {
		fmt.Printf("Found local config file at: %s\n", local)
		return local, nil
	} else {
		fmt.Printf("Local config file not found: %v\n", err)
	}

	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		xdgPath := filepath.Join(xdg, "yxa", "config.yml")
		fmt.Printf("Checking for XDG config at: %s\n", xdgPath)
		if _, err := os.Stat(xdgPath); err == nil {
			fmt.Printf("Found XDG config file at: %s\n", xdgPath)
			return xdgPath, nil
		} else {
			fmt.Printf("XDG config file not found: %v\n", err)
		}
	}

	home, err := os.UserHomeDir()
	if err == nil {
		homePath := filepath.Join(home, ".yxa.yml")
		fmt.Printf("Checking for home config at: %s\n", homePath)
		if _, err := os.Stat(homePath); err == nil {
			fmt.Printf("Found home config file at: %s\n", homePath)
			return homePath, nil
		} else {
			fmt.Printf("Home config file not found: %v\n", err)
		}
	}

	fmt.Println("No yxa config file found")
	return "", fmt.Errorf("no yxa config file found")
}
