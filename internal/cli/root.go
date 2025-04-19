package cli

import (
	"fmt"
	"os"
	"sync"

	"github.com/floppa/yxa-cli/internal/config"
	"github.com/floppa/yxa-cli/internal/executor"
	"github.com/spf13/cobra"
)

// RootCommand manages the root command and its subcommands
type RootCommand struct {
	Config   *config.ProjectConfig
	Executor executor.CommandExecutor
	Handler  *CommandHandler
	RootCmd  *cobra.Command
	DryRun   bool // global dry-run flag
}

// NewRootCommand creates a new root command
func NewRootCommand(cfg *config.ProjectConfig, exec executor.CommandExecutor) *RootCommand {
	root := &RootCommand{
		Config:   cfg,
		Executor: exec,
		Handler:  NewCommandHandler(cfg, exec),
	}

	// Create the root command
	root.RootCmd = &cobra.Command{
		Use:   "yxa",
		Short: "the morakniv of cli´s",
		Long:  `yxa is a CLI tool that is defined by a config file - yxa.yml`,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd:     false,
			DisableNoDescFlag:     false,
			DisableDescriptions:   false,
		},
	}

	// Add persistent dry-run flag
	root.RootCmd.PersistentFlags().BoolVarP(&root.DryRun, "dry-run", "d", false, "Show commands to be executed without running them")

	// Register commands from configuration
	root.registerCommands()

	// Setup command completion
	root.setupCompletion()

	return root
}

// Execute executes the root command
func (r *RootCommand) Execute() error {
	return r.RootCmd.Execute()
}

// registerCommands registers all commands from the configuration
func (r *RootCommand) registerCommands() {
	// Skip if no config is available
	if r.Config == nil {
		fmt.Println("No configuration loaded, no commands will be registered")
		return
	}

	// Register each command from the configuration
	for name, cmd := range r.Config.Commands {
		// Create a local copy of the variables for the closure
		cmdName := name
		cmdConfig := cmd

		// Create a new cobra command
		cobraCmd := &cobra.Command{
			Use:   cmdName,
			Short: cmdConfig.Description,
			Long:  cmdConfig.Description,
			Run: func(cmd *cobra.Command, args []string) {
				// Create a variables map with global variables
				cmdVars := make(map[string]string)
				if r.Config != nil {
					for k, v := range r.Config.Variables {
						cmdVars[k] = v
					}
				}

				// Process parameters if defined
				if len(cmdConfig.Params) > 0 {
					// Process parameters and update cmdVars
					paramVars, err := processParameters(cmd, args, cmdConfig.Params)
					if err != nil {
						fmt.Printf("Error processing parameters: %v\n", err)
						os.Exit(1)
					}

					// Add parameter variables to the command variables
					for k, v := range paramVars {
						cmdVars[k] = v
					}
				}

				// Set dry-run flag on the handler
				r.Handler.SetDryRun(r.DryRun)

				// Execute the command with variables
				if err := r.Handler.ExecuteCommand(cmdName, cmdVars); err != nil {
					fmt.Printf("Error executing command '%s': %v\n", cmdName, err)
					os.Exit(1)
				}
			},
		}

		// Add parameters as flags if defined
		if len(cmdConfig.Params) > 0 {
			addParametersToCommand(cobraCmd, cmdConfig.Params)
		}

		// Add the command to the root command
		r.RootCmd.AddCommand(cobraCmd)
	}
}

// setupCompletion sets up command completion
func (r *RootCommand) setupCompletion() {
	// Add completion command
	completionCmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate completion script",
		Long: `To load completions:

Bash:
  $ source <(yxa completion bash)

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ yxa completion zsh > "${fpath[1]}/_yxa"

  # You will need to start a new shell for this setup to take effect.

Fish:
  $ yxa completion fish | source

  # To load completions for each session, execute once:
  $ yxa completion fish > ~/.config/fish/completions/yxa.fish

PowerShell:
  PS> yxa completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> yxa completion powershell > yxa.ps1
  # and source this file from your PowerShell profile.
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			switch args[0] {
			case "bash":
				err = r.RootCmd.GenBashCompletion(os.Stdout)
			case "zsh":
				err = r.RootCmd.GenZshCompletion(os.Stdout)
			case "fish":
				err = r.RootCmd.GenFishCompletion(os.Stdout, true)
			case "powershell":
				err = r.RootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error generating completion: %v\n", err)
				os.Exit(1)
			}
		},
	}

	r.RootCmd.AddCommand(completionCmd)
}

// Global mutexes for synchronizing output to writers
var (
	writerMutexes = make(map[interface{}]*sync.Mutex)
	writerMutexLock sync.Mutex
)

// getWriterMutex returns a mutex for the given writer
func getWriterMutex(writer interface{}) *sync.Mutex {
	writerMutexLock.Lock()
	defer writerMutexLock.Unlock()
	
	if mutex, ok := writerMutexes[writer]; ok {
		return mutex
	}
	
	mutex := &sync.Mutex{}
	writerMutexes[writer] = mutex
	return mutex
}
