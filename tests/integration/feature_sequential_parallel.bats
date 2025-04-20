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
    commands:
      subseq1: echo first
      subseq2: echo second
    description: Parent sequential
    parallel: false
EOF

  run "$YXA_BIN" sequential-parent
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
    run: ""
    commands:
      parallel1: echo \"parallel1\"
      parallel2: echo \"parallel2\"
    description: Parent parallel
    parallel: true
EOF

  run "$YXA_BIN" parallel-parent
  [ "$status" -eq 0 ]
  [[ "$output" == *"parallel1"* ]]
  [[ "$output" == *"parallel2"* ]]
}
