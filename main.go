package main

import (
	"fmt"
	"io"
	"os"

	"github.com/floppa/yxa-cli/cmd"
)

// Version information - these will be set during build by the Makefile
var (
	version   = "dev"
	buildTime = "unknown"
)

// For testing purposes
var (
	osExit   = os.Exit
	osArgs   = os.Args
	stdout   = os.Stdout
	cmdExecute = cmd.Execute
	run       = runImpl
)

// main is the entry point for the application
func main() {
	// Run the application and exit with the returned code
	code := run(osArgs, stdout)
	osExit(code)
}

// runImpl executes the application logic and returns an exit code
// This function is separate from main to make it testable
func runImpl(args []string, out io.Writer) int {
	// Check if version flag is provided
	if len(args) > 1 && (args[1] == "-v" || args[1] == "--version") {
		_, err := fmt.Fprintf(out, "yxa version %s (built at %s)\n", version, buildTime)
		if err != nil {
			// If we can't write to output, return a non-zero exit code
			return 1
		}
		return 0
	}

	// Execute the CLI
	cmdExecute()
	return 0
}
