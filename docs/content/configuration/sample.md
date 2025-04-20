---
title: "Sample config"
weight: 4
---

```yaml
# Variables for testing
variables:
  PROJECT_NAME: yxa-test-project
  OUTPUT_DIR: ./output
  GREETING: Hello from Yxa

# Commands for testing all features
commands:
  # Basic command with variable substitution
  hello:
    run: echo "$GREETING, $PROJECT_NAME!"
    description: A simple greeting command

  # Command that creates a directory
  prepare:
    run: mkdir -p $OUTPUT_DIR
    description: Creates the output directory

  # Command with dependencies
  write-file:
    run: echo "Content from write-file" > $OUTPUT_DIR/output.txt
    depends: [prepare]
    description: Writes to a file in the output directory

  # Command with true condition
  conditional:
    run: echo "Condition was met"
    condition: $PROJECT_NAME == yxa-test-project
    description: Only runs if the condition is met

  # Command with false condition
  conditional-false:
    run: echo "This should not run"
    condition: $PROJECT_NAME == wrong-name
    description: Should be skipped due to condition

  # Command with timeout
  timeout:
    run: sleep 5 && echo "This should timeout"
    timeout: 2s
    description: Should timeout after 2 seconds

  # Command with parameters
  with-params:
    run: echo "$PARAM1"
    params:
      - name: PARAM1
        type: string
        description: A test parameter
        default: default-value
        flag: true
    description: Command with parameters

  # Command with parallel execution
  parallel-parent:
    run: echo "Starting parallel execution"
    commands:
      parallel1: sleep 1 && echo "Parallel command 1" > $OUTPUT_DIR/parallel1.txt
      parallel2: sleep 1 && echo "Parallel command 2" > $OUTPUT_DIR/parallel2.txt
    parallel: true
    description: Executes commands in parallel

  # Command with sequential execution
  sequential-parent:
    run: echo "Starting sequential execution"
    commands:
      seq1: sleep 1 && echo "Sequential command 1" > $OUTPUT_DIR/seq1.txt
      seq2: sleep 1 && echo "Sequential command 2" > $OUTPUT_DIR/seq2.txt
    parallel: false
    description: Executes commands sequentially

  # Command with pre and post hooks
  with-hooks:
    run: echo "Main command execution" > $OUTPUT_DIR/main.txt
    pre: echo "Pre-hook execution" > $OUTPUT_DIR/pre.txt
    post: echo "Post-hook execution" > $OUTPUT_DIR/post.txt
    description: Command with pre and post hooks

  # Command that reads environment variables
  env-vars:
    run: echo "ENV_VAR1=$ENV_VAR1, ENV_VAR2=$ENV_VAR2" > $OUTPUT_DIR/env.txt
    description: Reads environment variables from .env file

  # Command that should fail
  failing:
    run: echo "fail"; exit 1
    description: A command that should fail
```