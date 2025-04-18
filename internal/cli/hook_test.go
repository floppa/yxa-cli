package cli

import (
	"fmt"
	"strings"
	"testing"

	"github.com/floppa/yxa-cli/internal/config"
	"github.com/floppa/yxa-cli/internal/executor"
)

func TestCommandHandler_ExecuteHook(t *testing.T) {
	// Create a mock executor
	mockExec := executor.NewMockExecutor()
	
	// Create a test config
	cfg := &config.ProjectConfig{
		Name: "test-project",
		Variables: map[string]string{
			"PROJECT_NAME": "test-project",
		},
		Commands: map[string]config.Command{
			"with-hooks": {
				Run:         "echo 'main command'",
				Description: "Command with hooks",
				Pre:         "echo 'pre-hook'",
				Post:        "echo 'post-hook'",
			},
		},
	}

	// Add command results
	mockExec.AddCommandResult("echo 'pre-hook'", "pre-hook", nil)
	mockExec.AddCommandResult("echo 'post-hook'", "post-hook", nil)
	mockExec.AddCommandResult("echo 'main command'", "main command", nil)
	mockExec.AddCommandResult("echo 'failing-hook'", "", fmt.Errorf("hook failed"))

	// Create a command handler
	handler := NewCommandHandler(cfg, mockExec)

	// Test executing a pre-hook
	err := handler.executeHook("with-hooks", "pre", "echo 'pre-hook'", nil)
	if err != nil {
		t.Errorf("executeHook() pre-hook error = %v", err)
	}

	// Verify pre-hook was executed
	output := mockExec.GetOutput()
	if !strings.Contains(output, "pre-hook") {
		t.Errorf("Expected output to contain 'pre-hook', got '%s'", output)
	}

	// Clear output for next test
	mockExec.ClearOutput()

	// Test executing a post-hook
	err = handler.executeHook("with-hooks", "post", "echo 'post-hook'", nil)
	if err != nil {
		t.Errorf("executeHook() post-hook error = %v", err)
	}

	// Verify post-hook was executed
	output = mockExec.GetOutput()
	if !strings.Contains(output, "post-hook") {
		t.Errorf("Expected output to contain 'post-hook', got '%s'", output)
	}

	// Clear output for next test
	mockExec.ClearOutput()

	// Test executing a hook with variables
	vars := map[string]string{"PARAM": "param-value"}
	err = handler.executeHook("with-hooks", "pre", "echo 'pre-hook'", vars)
	if err != nil {
		t.Errorf("executeHook() with vars error = %v", err)
	}

	// Clear output for next test
	mockExec.ClearOutput()

	// Test executing a failing hook
	err = handler.executeHook("with-hooks", "pre", "echo 'failing-hook'", nil)
	if err == nil {
		t.Errorf("Expected error for failing hook, got nil")
	}

	// Verify error message contains hook type
	if err != nil && !strings.Contains(err.Error(), "pre-hook") {
		t.Errorf("Expected error to contain 'pre-hook', got '%v'", err)
	}
}
