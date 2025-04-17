# yxa

Yxa is a simple CLI tool that loads a config file (yxa.yml) in the current directory and registers commands defined in it.

Yxa is the word for Axe is swedish. Lets chop some trees!

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

yxa-cli uses it self for its tasks (go figure). The recommended path is to download and install yxa first.

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

You can specify timeouts for commands to prevent them from running indefinitely. If a command exceeds its timeout, it will be terminated.

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

### Parallel Command Execution

You can run multiple commands in parallel to speed up execution. This is useful for independent tasks that don't rely on each other's output.

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
- Each command's output is prefixed with its name for easy identification
- All commands start simultaneously
- The parent command completes when all parallel commands finish
- If any parallel command fails, the parent command fails
- You can combine with timeouts to limit the total execution time

## Configuration

The `yxa.yml` file should be placed in the root directory of your project. It has the following structure:

- `name`: The name of your project
- `variables` (optional): A map of variable definitions that can be used in commands
- `commands`: A map of command definitions
  - Each command has a name (key) and the following properties:
    - `run`: The shell command to execute
    - `description` (optional): A short description of what the command does
    - `depends` (optional): A list of command names that should be executed before this command

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

The CLI supports three types of variables:

1. **YAML Variables**: Defined in the `variables` section of the `yxa.yml` file
2. **Environment Variables from .env file**: Defined in a `.env` file in the project root
3. **System Environment Variables**: Available in your shell environment

Variable resolution priority (highest to lowest):
1. YAML variables
2. .env file variables
3. System environment variables

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

## License

MIT
