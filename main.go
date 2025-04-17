package main

import (
	"fmt"
	"os"

	"github.com/magnuseriksson/yxa-cli/cmd"
)

// Version information - these will be set during build by the Makefile
var (
	version   = "dev"
	buildTime = "unknown"
)

func main() {
	// Check if version flag is provided
	if len(os.Args) > 1 && (os.Args[1] == "-v" || os.Args[1] == "--version") {
		fmt.Printf("yxa version %s (built at %s)\n", version, buildTime)
		os.Exit(0)
	}

	// Execute the CLI
	cmd.Execute()
}
