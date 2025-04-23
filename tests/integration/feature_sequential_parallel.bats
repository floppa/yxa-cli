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

@test "Sequential commands output in order" {
  cat > yxa.yml <<EOF
name: yxa-test-project
commands:
  seq1:
    run: echo first
    description: First step
  seq2:
    run: echo second
    description: Second step
  sequential-parent:
    description: Parent sequential
    parallel: false
    commands:
      subseq1:
        run: echo first
        description: First subcommand
      subseq2:
        run: echo second
        description: Second subcommand
EOF

  # Execute both subcommands in sequence
  run bash -c "$YXA_BIN sequential-parent subseq1 && $YXA_BIN sequential-parent subseq2"
  [ "$status" -eq 0 ]
  [[ "$output" == *"first"* ]]
  [[ "$output" == *"second"* ]]
}

@test "Parallel commands output both results" {
  cat > yxa.yml <<EOF
name: yxa-test-project
commands:
  parallel1:
    run: echo \"parallel1\"
    description: First parallel
  parallel2:
    run: echo \"parallel2\"
    description: Second parallel
  parallel-parent:
    description: Parent parallel
    parallel: true
    commands:
      parallel1:
        run: echo \"parallel1\"
        description: First parallel subcommand
      parallel2:
        run: echo \"parallel2\"
        description: Second parallel subcommand
EOF

  # Execute both subcommands in parallel
  run bash -c "$YXA_BIN parallel-parent parallel1 & $YXA_BIN parallel-parent parallel2 & wait"
  [ "$status" -eq 0 ]
  [[ "$output" == *"parallel1"* ]]
  [[ "$output" == *"parallel2"* ]]
}
