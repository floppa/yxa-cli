package main

import (
	"fmt"
	"io"
	"os"

	"github.com/floppa/yxa-cli/internal/cli"
)

// osExit is a variable that holds os.Exit function
// This allows it to be mocked for testing
var osExit = os.Exit

// Version information - these will be set during build by the Makefile
var (
	version   = "dev"
	buildTime = "unknown"
)

// main is the entry point for the application
func main() {
	// Run the application and exit with the returned code
	code := run(os.Args, os.Stdout)
	osExit(code)
}

// run executes the application logic and returns an exit code
// This function is separate from main to make it testable
func run(args []string, out io.Writer) int {
	// Check if version flag is provided
	if len(args) > 1 && (args[1] == "-v" || args[1] == "--version") {
		_, err := fmt.Fprintf(out, "yxa version %s (built at %s)\n", version, buildTime)
		if err != nil {
			// If we can't write to output, return a non-zero exit code
			return 1
		}
		return 0
	}

	// Initialize the application
	rootCmd, err := cli.InitializeApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing application: %v\n", err)
		return 1
	}

	// Execute the command
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		return 1
	}
	
	return 0
}
