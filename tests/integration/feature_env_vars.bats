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

@test "Environment variables from .env are loaded" {
  echo "MY_ENV_VAR=hello" > .env
  cat > yxa.yml <<EOF
commands:
  print-env:
    run: echo "Env \$MY_ENV_VAR"
    description: Prints env var
EOF

  echo ".env contents:" >&2
  cat .env >&2
  echo "yxa.yml contents:" >&2
  cat yxa.yml >&2

  run "$YXA_BIN" print-env
  echo "status=$status" >&2
  echo "output=$output" >&2
  if [ "$status" -ne 0 ]; then
    echo "FAIL: CLI returned nonzero status $status" >&2
    exit 1
  fi
  [[ "$output" == *"Env hello"* ]]
}
