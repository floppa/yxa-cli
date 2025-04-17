package main

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

// TestRunWithVersionFlag tests the run function with the version flag
func TestRunWithVersionFlag(t *testing.T) {
	// Save original variables
	origVersion := version
	origBuildTime := buildTime
	
	// Set test values
	version = "test-version"
	buildTime = "test-time"
	
	// Restore original values after test
	defer func() {
		version = origVersion
		buildTime = origBuildTime
	}()
	
	// Create a buffer to capture output
	buf := new(bytes.Buffer)
	
	// Test with -v flag
	args := []string{"yxa", "-v"}
	exitCode := run(args, buf)
	
	// Check exit code
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
	
	// Check output
	output := buf.String()
	expected := "yxa version test-version (built at test-time)\n"
	if output != expected {
		t.Errorf("Expected output %q, got %q", expected, output)
	}
	
	// Test with error writing to output
	buf.Reset()
	errWriter := &errorWriter{}
	args = []string{"yxa", "-v"}
	exitCode = run(args, errWriter)
	
	// Check exit code for error case
	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for write error, got %d", exitCode)
	}
	
	// Test with --version flag
	buf.Reset()
	args = []string{"yxa", "--version"}
	exitCode = run(args, buf)
	
	// Check exit code
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
	
	// Check output
	output = buf.String()
	if output != expected {
		t.Errorf("Expected output %q, got %q", expected, output)
	}
}

// TestRunWithoutFlags tests the run function without any flags
func TestRunWithoutFlags(t *testing.T) {
	// Create a buffer to capture output
	buf := new(bytes.Buffer)
	
	// Save original cmdExecute function
	origExecute := cmdExecute
	
	// Mock cmdExecute to avoid actual execution
	cmdExecute = func() {}
	
	// Restore original function after test
	defer func() { 
		cmdExecute = origExecute 
	}()
	
	// Test with no flags
	args := []string{"yxa"}
	exitCode := run(args, buf)
	
	// Check exit code
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
	
	// Test with empty args
	args = []string{}
	exitCode = run(args, buf)
	
	// Check exit code
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

// TestMain tests the main function
func TestMain(t *testing.T) {
	// Save original functions
	origOsExit := osExit
	origOsArgs := osArgs
	origStdout := stdout
	origExecute := cmdExecute
	
	// Create a buffer to capture output
	buf := new(bytes.Buffer)
	
	// Mock functions
	exitCode := 0
	osExit = func(code int) { exitCode = code }
	osArgs = []string{"yxa", "--version"}
	
	// We can't directly assign buf to stdout (type mismatch)
	// Instead, we'll modify run to use our buffer
	origRun := run
	run = func(args []string, out io.Writer) int {
		return origRun(args, buf)
	}
	
	cmdExecute = func() {}
	
	// Restore original functions after test
	defer func() {
		osExit = origOsExit
		osArgs = origOsArgs
		stdout = origStdout
		cmdExecute = origExecute
		run = origRun
	}()
	
	// Run main
	main()
	
	// Check exit code
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

// errorWriter is a mock writer that always returns an error
type errorWriter struct{}

func (w *errorWriter) Write(p []byte) (n int, err error) {
	return 0, errors.New("mock write error")
}


