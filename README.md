

<img alt="golangci-lint logo" src="assets/yxa.svg" height="150" style="float:left;margin-right:20px;" />

<h1 style="border:0;">yxa</h2> 

Yxa is a simple CLI tool that loads a config file (yxa.yml) in the current directory and registers commands defined in it.

Yxa is the word for Axe in Swedish. Let's chop some trees!

[![GitHub release](https://img.shields.io/github/v/release/floppa/yxa-cli?include_prereleases)](https://github.com/floppa/yxa-cli/releases)
[![CI](https://github.com/floppa/yxa-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/floppa/yxa-cli/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/floppa/yxa-cli)](https://goreportcard.com/report/github.com/floppa/yxa-cli)
[![Test Coverage](https://img.shields.io/badge/coverage-86%25-brightgreen.svg)]()
[![gosec](https://img.shields.io/badge/gosec-security-brightgreen)](https://github.com/securego/gosec)
[![Go Version](https://img.shields.io/badge/go-1.24-blue.svg)](https://golang.org/dl/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://github.com/floppa/yxa-cli/pulls)
<!-- [![codecov](https://codecov.io/gh/floppa/yxa-cli/branch/main/graph/badge.svg)](https://codecov.io/gh/floppa/yxa-cli) -->

## Installation

### From GitHub Releases (Recommended)

Download the latest binary for your platform from the [GitHub Releases page](https://github.com/floppa/yxa-cli/releases).

```bash
# Linux/macOS (replace X.Y.Z with the version number and PLATFORM with your platform)
curl -L https://github.com/floppa/yxa-cli/releases/download/vX.Y.Z/yxa-PLATFORM -o yxa
chmod +x yxa
sudo mv yxa /usr/local/bin/

# Windows
# Download the .exe file and add it to your PATH
```

### Using Go Install

```bash
go install github.com/floppa/yxa-cli@latest
```

### Build from Source

```bash
git clone https://github.com/floppa/yxa-cli.git
cd yxa-cli
go build -o yxa
# Optional: move to a directory in your PATH
sudo mv yxa /usr/local/bin/
```

## Usage

### Basic Usage

1. Create a `yxa.yml` file in your project directory with the following structure:

```yaml
name: yxa-cli

variables:
  PROJECT_DIR: .
  BUILD_DIR: ./build

commands:
  build:
    desciption: "Build yxa-cli"
    run: go build -o $BUILD_DIR/app $PROJECT_DIR/...
  test:
    description: "Run tests"
    run: go test -v ./...
  api:
    description: "Call some random api"
    run: curl -H "Authorization: Bearer $API_KEY" $API_URL
  # Add more commands as needed
```

2. Run the CLI tool:

```bash
# List all available commands
yxa

# Run a specific command
yxa build
yxa test
```

## Development

yxa-cli uses itself for its tasks (go figure). The recommended path is to download and install yxa first.

### Development Requirements

- Go 1.18 or higher
- Git

### Test Coverage

The project maintains a high test coverage (currently 85.6%) with comprehensive tests for all components:

```bash
# Run tests with coverage report
yxa check-coverage

# Run tests with race detection
go test -race ./...
```

### Releasing New Versions

The project includes a built-in release command that helps with creating new releases:

```bash
# Run the release command
yxa release
```

This command will:
1. Show the current version tag
2. Suggest the next minor version (e.g., if current is v1.2.0, it will suggest v1.3.0)
3. Create a git tag with the specified version
4. Push the tag to GitHub

Once the tag is pushed, GitHub Actions will automatically:
- Run tests
- Build binaries for multiple platforms (Linux, macOS, Windows)
- Create a GitHub release with the binaries attached
- Generate a changelog based on commits since the last release

### Development Tasks

This project uses itself for development tasks! Here are the available commands:

```bash
# Build the CLI
yxa build

# Run tests
yxa test

# Clean build artifacts
yxa clean

# Install locally
yxa install

# Build for all platforms (outputs to dist/)
yxa dist

# Create a new release (will prompt for version)
yxa release

# Show version information
yxa version
```

This is a perfect example of how you can use `yxa-cli` to replace Makefiles and other build tools with a simple, consistent interface.

### Command Chaining

One of the powerful features of `yxa-cli` is command chaining, which allows you to define dependencies between commands. When you run a command, all its dependencies will be executed first, in the correct order.

For example, in our project:

```yaml
# Install locally
install:
  run: sudo mv ${BINARY_NAME} /usr/local/bin/${BINARY_NAME}
  depends: [build]

# Build for all platforms
dist:
  run: |
    mkdir -p ${DIST_DIR}
    GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o ${DIST_DIR}/${BINARY_NAME}-linux-amd64 .
    # ... more build commands ...
  depends: [clean]
```

When you run `yxa install`, it will first run the `build` command, and then execute the `install` command. Similarly, when you run `yxa dist`, it will first run the `clean` command, and then execute the `dist` command.

This allows you to create complex command chains and dependencies, similar to how Makefiles work, but with a cleaner, more modern syntax.

### Conditional Command Execution

Commands can be configured to run only when certain conditions are met. This is useful for platform-specific commands or commands that should only run in certain environments.

```yaml
commands:
  # Only runs on macOS
  macos-build:
    run: go build -o bin/app-darwin
    condition: "$GOOS == darwin"
    
  # Only runs if the .env file exists
  load-env:
    run: source .env
    condition: "exists .env"
    
  # Only runs if PATH contains /usr/local/bin
  check-path:
    run: echo "Path contains /usr/local/bin"
    condition: "$PATH contains /usr/local/bin"
```

Supported condition operators:
- Equality: `==` (e.g., `$GOOS == darwin`)
- Inequality: `!=` (e.g., `$GOOS != windows`)
- Contains: `contains` (e.g., `$PATH contains /usr/local`)
- Exists: `exists` (e.g., `exists /path/to/file`)

### Command Hooks

You can define pre and post hooks for commands. These are shell commands that run before and after the main command.

```yaml
commands:
  build:
    pre: echo "Starting build..."
    run: go build -o bin/app
    post: echo "Build complete"
    
  deploy:
    pre: go test ./...
    run: scp bin/app user@server:/path
    post: echo "Deployed successfully"
```

Hooks are useful for:
- Setup and cleanup operations
- Validation before running a command
- Notifications after command completion
- Ensuring certain actions always happen around a command

### Command Timeouts

You can specify timeouts for commands to prevent them from running indefinitely. If a command exceeds its timeout, it will be terminated safely with proper cleanup.

```yaml
commands:
  long-process:
    run: sleep 1000
    timeout: 10s
    
  network-operation:
    run: curl -s https://api.example.com
    timeout: 30s
```

Timeout values use Go's duration format:
- `s` for seconds (e.g., `30s`)
- `m` for minutes (e.g., `5m`)
- `h` for hours (e.g., `1h`)

The timeout implementation uses Go's context package for reliable cancellation and resource cleanup, ensuring that timed-out processes don't become orphaned.

### Parallel Command Execution

You can run multiple commands in parallel to speed up execution. This is useful for independent tasks that don't rely on each other's output. The implementation ensures thread-safe execution and proper output synchronization.

```yaml
commands:
  build-all:
    description: Build all components in parallel
    commands:
      frontend: cd frontend && npm run build
      backend: cd backend && go build
      docs: cd docs && mkdocs build
    parallel: true
    
  test-all:
    description: Run tests for all components
    commands:
      unit: go test ./...
      integration: go test -tags=integration ./...
      e2e: cypress run
    parallel: true
    timeout: 5m
```

Features of parallel execution:
- **Thread-safe output**: Each command's output is properly synchronized and prefixed with its name
- **Concurrent execution**: All commands start simultaneously with proper resource management
- **Synchronized completion**: The parent command completes when all parallel commands finish
- **Fail-fast behavior**: If any parallel command fails, the parent command fails
- **Global timeout**: You can combine with timeouts to limit the total execution time
- **Resource cleanup**: All resources are properly cleaned up, even in error cases

## Configuration

The `yxa.yml` file should be placed in the root directory of your project. It has the following structure:

- `name`: The name of your project
- `variables` (optional): A map of variable definitions that can be used in commands
- `commands`: A map of command definitions
  - Each command has a name (key) and the following properties:
    - `run`: The shell command to execute
    - `description` (optional): A short description of what the command does
    - `depends` (optional): A list of command names that should be executed before this command
    - `params` (optional): A list of parameters (flags and positional) for the command

### Command Chaining

One of the powerful features of `yxa-cli` is command chaining, which allows you to define dependencies between commands. When you run a command, all its dependencies will be executed first, in the correct order.

For example:

```yaml
commands:
  build:
    run: go build -o myapp .
    description: Build the application binary
  
  test:
    run: go test ./...
    description: Run all tests
  
  lint:
    run: golangci-lint run
    description: Run linting checks
  
  release:
    run: ./scripts/release.sh
    description: Create a new release
    depends: [build, test, lint]  # These commands will run before release
```

When you run `yxa release`, it will automatically execute the `build`, `test`, and `lint` commands first, and then run the `release` command.

### Command Parameters

Yxa-cli supports defining both flag parameters (options) and positional parameters for commands. This makes commands more flexible and reusable.

```yaml
commands:
  build:
    description: "Build the application"
    run: "go build -o $p_output $p_target"
    params:
      # Flag parameters (options)
      - name: "output|o"
        type: "string"
        default: "app"
        description: "Output filename"
        flag: true
      - name: "verbose|v"
        type: "bool"
        default: "false"
        description: "Enable verbose output"
        flag: true
      
      # Positional parameters
      - name: "target"
        type: "string"
        default: "."
        description: "Build target"
        position: 0
        required: false
```

#### Parameter Properties

- `name`: The parameter name (for flags, you can specify a shorthand with `name|shorthand`)
- `type`: The parameter type (`string`, `bool`, or `int`)
- `default`: The default value if not specified
- `description`: A description of the parameter
- `flag`: Set to `true` for flag parameters (specified with `--name` or `-shorthand`)
- `position`: For positional parameters, the position index (starting at 0)
- `required`: Whether the parameter is required

#### Parameter Variables

All parameters are accessible as variables with the `p_` prefix:

- `$p_output` - Value of the --output flag
- `$p_verbose` - Value of the --verbose flag
- `$p_target` - Value of the first positional parameter

#### Running Commands with Parameters

```bash
# Using flag parameters
yxa build --output=myapp --verbose

# Using shorthand flags
yxa build -o myapp -v

# Using positional parameters
yxa build ./cmd

# Combining flags and positional parameters
yxa build -o myapp ./cmd
```

### Task Aggregator Commands

You can create commands that serve purely as task aggregators by defining dependencies without specifying a `run` or `commands` property. These commands will simply execute all their dependencies in the correct order.

```yaml
commands:
  lint:
    run: golangci-lint run
    description: Run linting checks
    
  test:
    run: go test ./...
    description: Run all tests
    
  check-coverage:
    run: go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out
    description: Check test coverage
    
  # This command only runs its dependencies
  verify:
    description: Run all verification steps
    depends: [lint, test, check-coverage]
```

In this example, running `yxa verify` will execute the `lint`, `test`, and `check-coverage` commands in sequence, without running any additional command itself.

### Variables

The CLI supports four types of variables:

1. **Parameter Variables**: Defined by command parameters (flags and positional arguments)
2. **YAML Variables**: Defined in the `variables` section of the `yxa.yml` file
3. **Environment Variables from .env file**: Defined in a `.env` file in the project root
4. **System Environment Variables**: Available in your shell environment

Variable resolution priority (highest to lowest):
1. Parameter variables
2. YAML variables
3. .env file variables
4. System environment variables

#### Example with Variables

```yaml
name: my-project

variables:
  PROJECT_DIR: .
  BUILD_DIR: ./build
  TEST_FLAGS: -v -race

commands:
  build:
    run: go build -o $BUILD_DIR/app $PROJECT_DIR/...
    description: Build the application
  test:
    run: go test ${TEST_FLAGS} ./...
    description: Run tests with race detection
  env:
    run: echo "GOPATH=$GOPATH"
    description: Show GOPATH environment variable
```

### .env File Support

You can create a `.env` file in the project root to define environment variables that will be available to your commands. This is useful for storing sensitive information or environment-specific configuration.

#### Example .env file

```
# Build settings
GO_LDFLAGS=-ldflags "-s -w"
GOOS=darwin
GOARCH=amd64

# API settings
API_URL=https://api.example.com
API_KEY=your-secret-key-here
```

These variables can be used in your commands just like YAML variables:

## Security Considerations

Yxa CLI is designed to execute shell commands defined in a configuration file. By its nature, this involves certain security implications that users should be aware of:

### Command Execution

The CLI executes shell commands using `sh -c`. This is a core feature of the tool but comes with security implications:

- **Only use yxa.yml files from trusted sources**: Since the CLI will execute any command in the configuration file, you should only use configuration files from trusted sources.
- **Be careful with variable substitution**: Variables can come from the environment or `.env` files, so ensure sensitive data is properly protected.

### File Access

The CLI reads configuration files from the current directory:

- **Config file location**: The tool expects to find `yxa.yml` in the current directory. This is by design but requires users to be in the correct directory when running commands.
- **Environment variables**: Sensitive information can be stored in `.env` files rather than directly in the `yxa.yml` file.

### Best Practices

1. **Review commands before execution**: Always review the commands in `yxa.yml` before running them.
2. **Use `.env` for secrets**: Store sensitive information like API keys in `.env` files which are not committed to version control.
3. **Limit command scope**: Design commands to have the minimum necessary permissions and scope.
4. **Set appropriate timeouts**: Always set reasonable timeouts for commands that might hang or take too long.
5. **Use parallel execution wisely**: While parallel execution can speed up workflows, ensure that concurrent commands don't interfere with each other.
6. **Implement proper error handling**: Design your command chains to handle errors gracefully.

## Project Architecture

The project follows standard Go project layout conventions with a clean separation of concerns:

### Package Structure

- `cmd`: Command-line interface entry points
- `internal`: Private application code not meant to be imported by other projects
  - `cli`: Command handling and execution logic
  - `config`: Configuration loading and processing
  - `errors`: Custom error types
  - `executor`: Command execution implementation
  - `variables`: Variable resolution and substitution

This structure follows Go best practices by:
- Keeping implementation details in the `internal` package
- Separating concerns into focused packages
- Using dependency injection for better testability
- Maintaining a clean API for external consumers


Yxa-cli is designed with a focus on reliability, extensibility, and thread safety. The project follows idiomatic Go practices and is structured to be maintainable and well-tested.

### Core Components

- **Command Execution Engine**: Thread-safe implementation for running shell commands with support for timeouts, output capturing, and error handling.
- **Configuration Parser**: Robust YAML configuration parser with support for variables, command dependencies, and conditional execution.
- **Command Registry**: Dynamic command registration system that creates Cobra commands from YAML configuration.

### Thread Safety

Yxa-cli is designed to be thread-safe, allowing for reliable concurrent command execution:

- **Synchronized Output**: Thread-safe writers ensure that command output is properly synchronized and not interleaved.
- **Mutex Protection**: Critical sections are protected by mutexes to prevent race conditions.
- **Context-Based Timeouts**: Commands use Go's context package for reliable timeout handling.
- **Race-Free Design**: The codebase is regularly tested with Go's race detector to ensure thread safety.

### Testing Approach

The project maintains a high test coverage (currently 85.6%) with a comprehensive testing strategy:

- **Unit Tests**: Each component is thoroughly tested in isolation.
- **Integration Tests**: Command execution and configuration parsing are tested together.
- **Race Detection**: Tests are run with Go's race detector to identify and fix concurrency issues.
- **Mock Objects**: The executor interface is mocked for predictable testing of command execution.

## License

MIT
