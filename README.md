# yxa

Yxa is a simple CLI tool that loads a config file (yxa.yml) in the current directory and registers commands defined in it.

Yxa is the word for Axe is swedish. Lets chop some trees.

## Overview

`yxa` is a command-line tool that allows you to define project-specific commands in a YAML configuration file (yxa.yml). It dynamically registers these commands and executes them using the shell.

## Installation

```bash
go install github.com/magnuseriksson/yxa-cli@latest
```

Or build from source:

```bash
git clone https://github.com/magnuseriksson/yxa-cli.git
cd yxa-cli
go build -o yxa
```

## Usage

### Basic Usage

1. Create a `yxa.yml` file in your project directory with the following structure:

```yaml
name: my-project

variables:
  PROJECT_DIR: .
  BUILD_DIR: ./build

commands:
  build:
    run: go build -o $BUILD_DIR/app $PROJECT_DIR/...
  test:
    run: go test -v ./...
  api:
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

## Configuration

The `yxa.yml` file should be placed in the root directory of your project. It has the following structure:

- `name`: The name of your project
- `variables` (optional): A map of variable definitions that can be used in commands
- `commands`: A map of command definitions
  - Each command has a name (key) and the following properties:
    - `run`: The shell command to execute
    - `depends` (optional): A list of command names that should be executed before this command

### Variables

You can define variables in the `yxa.yml` file and use them in your commands. Variables can be referenced using `$VAR_NAME` or `${VAR_NAME}` syntax.

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
  test:
    run: go test ${TEST_FLAGS} ./...
  env:
    run: echo "GOPATH=$GOPATH"
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

## License

MIT
