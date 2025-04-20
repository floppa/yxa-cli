

<img alt="yxa logo" src="docs/static/images/yxa.svg" height="150" style="float:left;margin-right:20px;" />

<h1 style="border:0;">yxa</h2> 

Yxa is a simple CLI tool that loads a config file (yxa.yml) in the current directory and registers commands defined in it.

Yxa is the word for Axe in Swedish. Let's chop some trees!

[![GitHub release](https://img.shields.io/github/v/release/floppa/yxa-cli?include_prereleases)](https://github.com/floppa/yxa-cli/releases)
[![CI](https://github.com/floppa/yxa-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/floppa/yxa-cli/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/floppa/yxa-cli)](https://goreportcard.com/report/github.com/floppa/yxa-cli)
[![Test Coverage](https://img.shields.io/badge/coverage-86%25-brightgreen.svg)]()
[![gosec](https://img.shields.io/badge/gosec-security-brightgreen)](https://github.com/securego/gosec)
[![Go Version](https://img.shields.io/badge/go-1.24-blue.svg)](https://golang.org/dl/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://github.com/floppa/yxa-cli/pulls)
<!-- [![codecov](https://codecov.io/gh/floppa/yxa-cli/branch/main/graph/badge.svg)](https://codecov.io/gh/floppa/yxa-cli) -->

## Docs

See the docs: https://floppa.github.io/yxa-cli

## Installation

### Using Go Install

```bash
go install github.com/floppa/yxa-cli@latest
```

For other alternatives, see docs.


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

