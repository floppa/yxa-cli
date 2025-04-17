package main

import (
	"bytes"
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


