#!/usr/bin/env bats

setup() {
  TEST_DIR="$(mktemp -d)"
  cd "$TEST_DIR"
  export YXA_BIN="${BATS_TEST_DIRNAME}/../../yxa"
  [ -x "$YXA_BIN" ] || YXA_BIN="yxa"
}

teardown() {
  rm -rf "$TEST_DIR"
}

@test "Failing command returns error and outputs error message" {
  cat > yxa.yml <<EOF
name: yxa-test-project
commands:
  sequential-with-error:
    description: Parent with failing sequential command
    parallel: false
    commands:
      seq1:
        run: "echo 'seq1'"
        description: First sequence command
      seq-fail:
        run: "exit 1"
        description: Failing command
EOF

  # Run the failing subcommand directly
  run "$YXA_BIN" sequential-with-error seq-fail
  echo "status=$status" >&2
  echo "output=$output" >&2
  if [ "$status" -eq 0 ]; then
    echo "FAIL: CLI should have returned nonzero status for error" >&2
    exit 1
  fi
  [[ "$output" == *"Error executing subcommand 'sequential-with-error:seq-fail'"* ]]
}
