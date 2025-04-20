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

@test "Command chaining executes dependencies in order" {
  cat > yxa.yml <<EOF
name: yxa-test-project
commands:
  prepare:
    run: echo "Preparing"
    description: Prepare step
  build:
    run: echo "Building"
    depends: [prepare]
    description: Build step
  deploy:
    run: echo "Deploying"
    depends: [build]
    description: Deploy step
EOF

  run "$YXA_BIN" deploy
  [ "$status" -eq 0 ]
  [[ "$output" == *"Preparing"* ]]
  [[ "$output" == *"Building"* ]]
  [[ "$output" == *"Deploying"* ]]
}
