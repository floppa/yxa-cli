#!/usr/bin/env bats

setup() {
  TEST_DIR="$(mktemp -d)"
  cd "$TEST_DIR"
  export YXA_BIN="${BATS_TEST_DIRNAME}/../../yxa"
  if [ ! -x "$YXA_BIN" ]; then
    echo "WARN: yxa binary not found at $YXA_BIN, falling back to 'yxa' in PATH" >&2
    YXA_BIN="yxa"
  fi
  echo "Using YXA_BIN: $YXA_BIN" >&2
  which "$YXA_BIN" || echo "WARN: yxa binary not found in PATH" >&2
}

teardown() {
  rm -rf "$TEST_DIR"
}

@test "Variable substitution in command output" {
  cat > yxa.yml <<EOF
variables:
  PROJECT_NAME: yxa-test-project
commands:
  print-var:
    run: echo "Project \$PROJECT_NAME"
    description: Prints the project name
EOF

  run "$YXA_BIN" print-var
  echo "status=$status" >&2
  echo "output=$output" >&2
  if [ "$status" -ne 0 ]; then
    echo "FAIL: CLI returned nonzero status $status" >&2
    exit 1
  fi
  [[ "$output" == *"Project yxa-test-project"* ]]
}
