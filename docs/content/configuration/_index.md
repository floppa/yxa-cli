---
title: "Configuration Reference"
weight: 3
---

The minimum viable configuration is the below, it basically adds a command that runs another command.

```yaml
commands:
  <command>:
    run: echo "Hello"
```

A command can be configured with more than this.

```yaml
commands:
  <command>:
    run: echo "Hello"
    timeout: <Timeout>
    depends: [ <command> ]
    parallel: true/false
```

A command could also use predefined variables, environment variables and .env files to add more dynamics to the commands.

## Working Directory (`workingdir`)

You can specify a `workingdir` at either the file (config) level or per command. **The `workingdir` field only affects commands defined in the same config file where it is set. There is no inheritance or merging of `workingdir` between global and project configs.**

- If you set `workingdir` in your project config (`yxa.yml`), it only applies to commands defined in that file.
- If you set `workingdir` in your global config (e.g. `~/.yxa.yml`), it only applies to commands defined in the global config.
- If both a file-level and a command-level `workingdir` are set, the command-level takes precedence for that command.

### Example: Project Config with workingdir

```yaml
workingdir: ./myproject
commands:
  build:
    run: make build
  test:
    run: make test
    workingdir: ./myproject/tests  # Overrides file-level workingdir for this command
```

### Example: Global Config with workingdir

```yaml
workingdir: ~/global-default
commands:
  globalcmd:
    run: echo "This runs in ~/global-default"
  shared:
    run: echo "Shared global command"
```

### Important
- Project commands are **never** affected by the global config's `workingdir`.
- Global commands are **never** affected by the project config's `workingdir`.
- This makes command execution predictable and config files isolated.
