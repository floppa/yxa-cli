package cli

import (
	"fmt"
	"os"
	"sync"

	"github.com/floppa/yxa-cli/internal/config"
	"github.com/floppa/yxa-cli/internal/executor"
	"github.com/spf13/cobra"
)

var exitFunc = os.Exit

// RootCommand manages the root command and its subcommands
type RootCommand struct {
	Config   *config.ProjectConfig
	Executor executor.CommandExecutor
	Handler  *CommandHandler
	RootCmd  *cobra.Command
	DryRun   bool // global dry-run flag
}

// NewRootCommand creates a new root command
var ConfigFlag string

func NewRootCommand(cfg *config.ProjectConfig, exec executor.CommandExecutor) *RootCommand {
	root := &RootCommand{
		Config:   cfg,
		Executor: exec,
		Handler:  NewCommandHandler(cfg, exec),
	}

	// Capture root for use in the closure and method below
	r := root

	// Create the root command
	r.RootCmd = &cobra.Command{
		Use:   "yxa",
		Short: "the morakniv of cli´s",
		Long:  `yxa is a CLI tool that is defined by a config file - yxa.yml`,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd:   false,
			DisableNoDescFlag:   false,
			DisableDescriptions: false,
		},
		// PersistentPreRunE now delegates to the dedicated loading method.
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// ConfigFlag is populated by Cobra before this hook runs.
			return r.loadConfigAndRegisterCommands(ConfigFlag)
		},
		// Add RunE to ensure configuration is loaded even when no command is specified
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no command is specified, just show the help
			return cmd.Help()
		},
	}

	// Add persistent config flag
	r.RootCmd.PersistentFlags().StringVar(&ConfigFlag, "config", "", "config file (default: yxa.yml in current directory, or global config)")
	// Add persistent dry-run flag
	r.RootCmd.PersistentFlags().BoolVarP(&r.DryRun, "dry-run", "d", false, "Show commands to be executed without running them")

	// Setup command completion
	r.setupCompletion()

	return r
}

// loadConfigAndRegisterCommands loads the configuration and registers commands
// loadConfigAndRegisterCommands loads the configuration and registers commands
// This is the main entry point for loading configuration and setting up commands
func (r *RootCommand) loadConfigAndRegisterCommands(configFlagValue string) error {
	// Check if we need to load a new configuration
	if !r.shouldLoadNewConfig(configFlagValue) {
		return nil
	}

	// Load the configuration
	config, err := r.loadConfiguration(configFlagValue)
	if err != nil {
		return err
	}

	// Set up the CLI with the loaded configuration
	return r.setupWithConfig(config)
}

// shouldLoadNewConfig determines if a new configuration should be loaded
func (r *RootCommand) shouldLoadNewConfig(configFlagValue string) bool {
	// If configFlagValue is provided, always try to load it
	if configFlagValue != "" {
		return true
	}

	// If no config is loaded yet, try to load from default locations
	if r.Config == nil {
		return true
	}

	return false
}

// loadConfiguration loads the configuration from the specified path or default locations
func (r *RootCommand) loadConfiguration(configFlagValue string) (*config.ProjectConfig, error) {
	// If configFlagValue is provided, use it directly
	if configFlagValue != "" {
		return r.loadConfigFromPath(configFlagValue)
	}

	// Try to load from the current directory first
	localPath := "./yxa.yml"
	if _, statErr := os.Stat(localPath); statErr == nil {
		// Local config file exists, load it
		return r.loadConfigFromPath(localPath)
	}

	// No local config, try to resolve using standard paths
	path, err := config.ResolveConfigPath(configFlagValue)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve config path: %w", err)
	}

	return r.loadConfigFromPath(path)
}

// loadConfigFromPath loads configuration from a specific path
func (r *RootCommand) loadConfigFromPath(path string) (*config.ProjectConfig, error) {
	loadedConfig, err := config.LoadConfigFrom(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration from '%s': %w", path, err)
	}
	return loadedConfig, nil
}

// setupWithConfig sets up the CLI with the loaded configuration
func (r *RootCommand) setupWithConfig(loadedConfig *config.ProjectConfig) error {
	// Store the loaded config
	r.Config = loadedConfig

	// Validate command dependencies
	if err := validateCommandDependencies(r.Config); err != nil {
		return fmt.Errorf("invalid command dependencies: %w", err)
	}

	// Re-initialize the handler with the loaded config
	r.Handler = NewCommandHandler(r.Config, r.Executor)

	// Clear existing user commands before registering new ones
	r.clearUserCommands()

	// Register commands based on the loaded config
	r.registerCommands()

	return nil
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
						exitFunc(1)
					}

					// Add parameter variables to the command variables
					for k, v := range paramVars {
						cmdVars[k] = v
					}
				}

				// Check if this command has a subcommand specified
				if len(args) > 0 && len(cmdConfig.Commands) > 0 {
					subCmdName := args[0]
					_, ok := cmdConfig.Commands[subCmdName]
					if ok {
						// Execute the subcommand
						fullCmdName := fmt.Sprintf("%s:%s", cmdName, subCmdName)

						// Set dry-run flag on the handler
						r.Handler.SetDryRun(r.DryRun)

						// Use ExecuteCommand which will internally call executeCommandWithDependencies
						if err := r.Handler.ExecuteCommand(fullCmdName, cmdVars); err != nil {
							fmt.Printf("Error executing subcommand '%s': %v\n", fullCmdName, err)
							exitFunc(1)
						}
						return
					}
				}

				// Set dry-run flag on the handler
				r.Handler.SetDryRun(r.DryRun)

				// Execute the command with variables
				if err := r.Handler.ExecuteCommand(cmdName, cmdVars); err != nil {
					fmt.Printf("Error executing command '%s': %v\n", cmdName, err)
					exitFunc(1)
				}
			},
		}

		// Add parameters if defined
		if len(cmdConfig.Params) > 0 {
			addParametersToCommand(cobraCmd, cmdConfig.Params)
		}

		// Add subcommands if defined
		if len(cmdConfig.Commands) > 0 {
			for subName, subCmd := range cmdConfig.Commands {
				// Create a local copy for the closure
				subCmdName := subName
				subCmdConfig := subCmd

				// Create the subcommand
				subCobraCmd := &cobra.Command{
					Use:   subCmdName,
					Short: subCmdConfig.Description,
					Long:  subCmdConfig.Description,
					Run: func(cmd *cobra.Command, args []string) {
						// Create variables map with parent command variables
						cmdVars := make(map[string]string)
						if r.Config != nil {
							for k, v := range r.Config.Variables {
								cmdVars[k] = v
							}
						}

						// Execute the subcommand
						fullCmdName := fmt.Sprintf("%s:%s", cmdName, subCmdName)

						// Set dry-run flag on the handler
						r.Handler.SetDryRun(r.DryRun)

						// Execute the command
						if err := r.Handler.ExecuteCommand(fullCmdName, cmdVars); err != nil {
							fmt.Printf("Error executing subcommand '%s': %v\n", fullCmdName, err)
							exitFunc(1)
						}
					},
				}

				// Add parameters to the subcommand if defined
				if len(subCmdConfig.Params) > 0 {
					addParametersToCommand(subCobraCmd, subCmdConfig.Params)
				}

				// Add the subcommand to the parent command
				cobraCmd.AddCommand(subCobraCmd)
			}
		}

		// Add the command to the root command
		r.RootCmd.AddCommand(cobraCmd)
	}
}

// while preserving built-in commands like help and completion
func (r *RootCommand) clearUserCommands() {
	// Get all commands
	commands := r.RootCmd.Commands()

	// Create a copy of the commands slice to avoid modification during iteration
	commandsCopy := make([]*cobra.Command, len(commands))
	copy(commandsCopy, commands)

	// Remove user-defined commands
	for _, cmd := range commandsCopy {
		// Skip built-in commands
		if cmd.Name() == "help" || cmd.Name() == "completion" {
			continue
		}

		// Remove the command
		r.RootCmd.RemoveCommand(cmd)
	}
}

// setupCompletion sets up command completion
func (r *RootCommand) setupCompletion() {
	// Check if completion command already exists
	for _, cmd := range r.RootCmd.Commands() {
		if cmd.Name() == "completion" {
			// Completion command already exists, don't add it again
			return
		}
	}

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
				exitFunc(1)
			}
		},
	}

	r.RootCmd.AddCommand(completionCmd)
}

// Global mutexes for synchronizing output to writers
var (
	WriterMutexes   = make(map[interface{}]*sync.Mutex)
	WriterMutexLock sync.Mutex
)

// GetWriterMutex returns a mutex for the given writer
func GetWriterMutex(writer interface{}) *sync.Mutex {
	WriterMutexLock.Lock()
	defer WriterMutexLock.Unlock()

	if mutex, ok := WriterMutexes[writer]; ok {
		return mutex
	}

	// Create a new mutex for this writer
	mutex := &sync.Mutex{}
	WriterMutexes[writer] = mutex
	return mutex
}
