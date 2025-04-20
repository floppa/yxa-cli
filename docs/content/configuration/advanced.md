---
title: "Advanced Configuration"
weight: 3
---

## Advanced configuration


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

```bash
# Build settings
GO_LDFLAGS=-ldflags "-s -w"
GOOS=darwin
GOARCH=amd64

# API settings
API_URL=https://api.example.com
API_KEY=your-secret-key-here
```

These variables can be used in your commands just like YAML variables:


### Parameters

Commands can accept parameters using `${PARAM}` syntax:

```yaml
commands:
  echo:
    run: echo "${MESSAGE}"
```

Run with:
```bash
yxa echo --MESSAGE="Hello from param!"
```

### Command chaining

One of the powerful features of `yxa-cli` is command chaining, which allows you to define dependencies between commands. When you run a command, all its dependencies will be executed first, in the correct order.

```yaml
commands:
  build:
    run: go build ./...
  test:
    run: go test ./...
  all:
    depends:
      - build
      - test
```

Running `yxa all` will execute `build` and then `test` in order.

### Sequential subcommands

You can define subcommands that run in sequence:

```yaml
commands:
  setup:
    depends:
      - clean
      - build
      - migrate
  clean:
    run: rm -rf ./build
  build:
    run: go build ./...
  migrate:
    run: ./migrate.sh
```

### Parallel subcommands

To run subcommands in parallel, use the `parallel` flag:

```yaml
commands:
  test-all:
    parallel: true
    depends:
      - test-unit
      - test-integration
  test-unit:
    run: go test ./unit/...
  test-integration:
    run: go test ./integration/...
```

This will run `test-unit` and `test-integration` at the same time. Parallel execution is thread-safe and handles timeouts gracefully.

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


