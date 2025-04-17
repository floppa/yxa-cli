package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/floppa/yxa-cli/config"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "yxa",
	Short: "the morakniv of cli´s",
	Long: `yxa is a CLI tool that is defined by a config file - yxa.yml`,
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: false,
		DisableNoDescFlag: false,
		DisableDescriptions: false,
	},
}

// Execute executes the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// Global variables to track command execution and dependencies
var (
	cfg              *config.ProjectConfig
	executedCommands = make(map[string]bool)
)

// executeCommand runs a command with its dependencies
func executeCommand(cmdName string) error {
	// Check if command has already been executed
	if executedCommands[cmdName] {
		return nil
	}

	// Get the command from the config
	cmd, ok := cfg.Commands[cmdName]
	if !ok {
		return fmt.Errorf("command '%s' not found", cmdName)
	}

	// Evaluate condition if present
	if cmd.Condition != "" {
		conditionStr := cfg.ReplaceVariables(cmd.Condition)
		if !cfg.EvaluateCondition(conditionStr) {
			fmt.Printf("Skipping command '%s' (condition not met: %s)\n", cmdName, conditionStr)
			// Mark as executed even though we're skipping it
			executedCommands[cmdName] = true
			return nil
		}
	}

	// Execute dependencies first
	for _, dep := range cmd.Depends {
		// Check for circular dependencies
		if dep == cmdName {
			return fmt.Errorf("circular dependency detected for command '%s'", cmdName)
		}

		// Execute the dependency
		if err := executeCommand(dep); err != nil {
			return err
		}
	}

	// Execute pre-hook if present
	if cmd.Pre != "" {
		preCmd := cfg.ReplaceVariables(cmd.Pre)
		fmt.Printf("Executing pre-hook for '%s'...\n", cmdName)
		preCmdExec := exec.Command("sh", "-c", preCmd) // #nosec G204
		preCmdExec.Stdout = os.Stdout
		preCmdExec.Stderr = os.Stderr
		if err := preCmdExec.Run(); err != nil {
			return fmt.Errorf("pre-hook for command '%s' failed: %v", cmdName, err)
		}
	}

	// Parse timeout if specified
	var timeoutDuration time.Duration
	var err error
	if cmd.Timeout != "" {
		timeoutStr := cfg.ReplaceVariables(cmd.Timeout)
		timeoutDuration, err = time.ParseDuration(timeoutStr)
		if err != nil {
			return fmt.Errorf("invalid timeout format for command '%s': %v", cmdName, err)
		}
		fmt.Printf("Command '%s' will timeout after %s\n", cmdName, timeoutDuration)
	}

	// Check if we need to run multiple commands in parallel
	if cmd.Parallel && len(cmd.Commands) > 0 {
		// Run commands in parallel
		fmt.Printf("Executing parallel commands for '%s'...\n", cmdName)
		err = executeParallelCommands(cmdName, cmd, timeoutDuration)
	} else if cmd.Run != "" {
		// Execute single command with timeout if specified
		cmdStr := cfg.ReplaceVariables(cmd.Run)
		fmt.Printf("Executing command '%s'...\n", cmdName)
		err = executeSingleCommand(cmdStr, timeoutDuration)
	} else if len(cmd.Commands) > 0 {
		// Execute multiple commands sequentially
		fmt.Printf("Executing sequential commands for '%s'...\n", cmdName)
		err = executeSequentialCommands(cmdName, cmd, timeoutDuration)
	} else if len(cmd.Depends) == 0 {
		// Only show warning if the command has no dependencies
		return fmt.Errorf("command '%s' has no 'run' or 'commands' defined", cmdName)
	}
	// If we get here, the command has dependencies but no run or commands
	// This is fine - the command just serves as a task aggregator

	// Execute post-hook if present (even if main command failed)
	if cmd.Post != "" {
		postCmd := cfg.ReplaceVariables(cmd.Post)
		fmt.Printf("Executing post-hook for '%s'...\n", cmdName)
		postCmdExec := exec.Command("sh", "-c", postCmd) // #nosec G204
		postCmdExec.Stdout = os.Stdout
		postCmdExec.Stderr = os.Stderr
		if postErr := postCmdExec.Run(); postErr != nil {
			// If main command succeeded but post-hook failed, return post-hook error
			if err == nil {
				err = fmt.Errorf("post-hook for command '%s' failed: %v", cmdName, postErr)
			} else {
				// Both failed, mention both errors
				err = fmt.Errorf("command '%s' failed: %v (and post-hook failed: %v)", cmdName, err, postErr)
			}
		}
	}

	// Check if there was an error in the main command
	if err != nil && cmd.Post == "" {
		return fmt.Errorf("command '%s' failed: %v", cmdName, err)
	} else if err != nil {
		// Error already formatted above if post-hook was present
		return err
	}

	// Mark command as executed
	executedCommands[cmdName] = true
	return nil
}

// executeSingleCommand executes a single command with an optional timeout
func executeSingleCommand(cmdStr string, timeout time.Duration) error {
	// Create a command
	cmdExec := exec.Command("sh", "-c", cmdStr) // #nosec G204
	cmdExec.Stdout = os.Stdout
	cmdExec.Stderr = os.Stderr

	// If no timeout is specified, just run the command
	if timeout == 0 {
		return cmdExec.Run()
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Set the context on the command
	err := cmdExec.Start()
	if err != nil {
		return err
	}

	// Create a channel for the command completion
	done := make(chan error, 1)
	go func() {
		done <- cmdExec.Wait()
	}()

	// Wait for either command completion or timeout
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		// Command timed out, kill it
		if err := cmdExec.Process.Kill(); err != nil {
			return fmt.Errorf("command timed out after %s and failed to kill process: %v", timeout, err)
		}
		return fmt.Errorf("command timed out after %s", timeout)
	}
}

// executeParallelCommands executes multiple commands in parallel
func executeParallelCommands(cmdName string, cmd config.Command, timeout time.Duration) error {
	// Create a wait group to wait for all commands to complete
	var wg sync.WaitGroup
	// Create a channel to collect errors
	errChan := make(chan error, len(cmd.Commands))

	// Create a context with timeout if specified
	var ctx context.Context
	var cancel context.CancelFunc
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
		defer cancel()
	} else {
		ctx, cancel = context.WithCancel(context.Background())
		defer cancel()
	}

	// Start all commands in parallel
	for name, cmdStr := range cmd.Commands {
		wg.Add(1)
		go func(name, cmdStr string) {
			defer wg.Done()

			// Replace variables in the command
			cmdStr = cfg.ReplaceVariables(cmdStr)
			fmt.Printf("Executing parallel sub-command '%s' for '%s'...\n", name, cmdName)

			// Create and start the command
			cmdExec := exec.Command("sh", "-c", cmdStr) // #nosec G204
			
			// Create prefixed writers for stdout and stderr
			prefixedStdout := newPrefixedWriter(os.Stdout, fmt.Sprintf("[%s] ", name))
			prefixedStderr := newPrefixedWriter(os.Stderr, fmt.Sprintf("[%s] ", name))
			cmdExec.Stdout = prefixedStdout
			cmdExec.Stderr = prefixedStderr

			// Start the command
			err := cmdExec.Start()
			if err != nil {
				errChan <- fmt.Errorf("failed to start sub-command '%s' for '%s': %v", name, cmdName, err)
				return
			}

			// Create a channel for command completion
			done := make(chan error, 1)
			go func() {
				done <- cmdExec.Wait()
			}()

			// Wait for command completion or timeout
			select {
			case err := <-done:
				if err != nil {
					errChan <- fmt.Errorf("sub-command '%s' for '%s' failed: %v", name, cmdName, err)
				}
			case <-ctx.Done():
				// Command timed out or context was canceled
				if err := cmdExec.Process.Kill(); err != nil {
					errChan <- fmt.Errorf("sub-command '%s' for '%s' timed out after %s and failed to kill process: %v", name, cmdName, timeout, err)
				} else {
					errChan <- fmt.Errorf("sub-command '%s' for '%s' timed out after %s", name, cmdName, timeout)
				}
			}
		}(name, cmdStr)
	}

	// Wait for all commands to complete
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// Collect errors
	var errors []string
	for err := range errChan {
		errors = append(errors, err.Error())
	}

	// Return combined errors if any
	if len(errors) > 0 {
		return fmt.Errorf("one or more parallel commands failed:\n%s", strings.Join(errors, "\n"))
	}

	return nil
}

// executeSequentialCommands executes multiple commands sequentially
func executeSequentialCommands(cmdName string, cmd config.Command, timeout time.Duration) error {
	for name, cmdStr := range cmd.Commands {
		// Replace variables in the command
		cmdStr = cfg.ReplaceVariables(cmdStr)
		fmt.Printf("Executing sequential sub-command '%s' for '%s'...\n", name, cmdName)

		// Execute the command with timeout
		err := executeSingleCommand(cmdStr, timeout)
		if err != nil {
			return fmt.Errorf("sub-command '%s' for '%s' failed: %v", name, cmdName, err)
		}
	}

	return nil
}

// newPrefixedWriter creates a writer that prefixes each line with the given prefix
type prefixedWriter struct {
	writer io.Writer
	prefix string
	buffer []byte
}

func newPrefixedWriter(writer io.Writer, prefix string) *prefixedWriter {
	return &prefixedWriter{
		writer: writer,
		prefix: prefix,
		buffer: make([]byte, 0, 1024),
	}
}

func (w *prefixedWriter) Write(p []byte) (n int, err error) {
	n = len(p)
	for _, b := range p {
		if b == '\n' {
			// Write the prefix and the buffered content
			_, err = fmt.Fprintf(w.writer, "%s%s\n", w.prefix, w.buffer)
			if err != nil {
				return
			}
			w.buffer = w.buffer[:0]
		} else {
			w.buffer = append(w.buffer, b)
		}
	}
	return
}

// init function is called when the package is initialized
func init() {
	// Load config file
	var err error
	cfg, err = config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
	}

	// Register commands from config
	if cfg != nil {
		registerCommands()
	}

	// Add custom completion function
	rootCmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var completions []string
		if cfg == nil {
			return completions, cobra.ShellCompDirectiveNoFileComp
		}

		// Add all commands from config
		for cmdName := range cfg.Commands {
			if strings.HasPrefix(cmdName, toComplete) {
				completions = append(completions, cmdName)
			}
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}

func registerCommands() {
	// Register commands from the configuration
	for name, cmd := range cfg.Commands {
		// Create a copy of the variables for the closure
		cmdName := name

		// Use the command description if available, otherwise use a generic description
		shortDesc := cmd.Description
		if shortDesc == "" {
			shortDesc = fmt.Sprintf("Execute the '%s' command", cmdName)
		}

		// Add dependency information if present
		if len(cmd.Depends) > 0 {
			shortDesc += fmt.Sprintf(" (depends on: %s)", strings.Join(cmd.Depends, ", "))
		}

		// Create a new cobra command
		command := &cobra.Command{
			Use:   cmdName,
			Short: shortDesc,
			Run: func(cmd *cobra.Command, args []string) {
				// Reset executed commands for each run
				executedCommands = make(map[string]bool)

				// Execute the command with its dependencies
				if err := executeCommand(cmdName); err != nil {
					// Check if this is the "no run or commands" error for a command with dependencies
					command := cfg.Commands[cmdName]
					if len(command.Depends) > 0 && strings.Contains(err.Error(), "has no 'run' or 'commands' defined") {
						// This is a command that only has dependencies - this is fine
						// Don't show the error message
					} else {
						fmt.Println(err)
						os.Exit(1)
					}
				}
			},
		}

		// Add the command to the root command
		rootCmd.AddCommand(command)
	}
}
