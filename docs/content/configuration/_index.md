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

A command could also use predefined variables, environement variables and .env files to add more dynamics to the commands.
