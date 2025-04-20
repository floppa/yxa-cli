---
title: "Usage"
weight: 2
---

## Usage

`yxa [command]`

### Basic Usage

1. Create a `yxa.yml` file in your project directory with the following structure:

```yaml
commands:
  hello:
    description: "Says hello"
    run: echo "hello"
  # Add more commands as needed
```

2. Run the CLI tool:

```bash
# List all available commands
yxa

# Run a specific command
yxa hello
```

Yxa also has some default parameters

#### --help / -h

Just outputs same as if only call `yxa`

#### --dry-run / d

Dry run just outputs what will be called.


