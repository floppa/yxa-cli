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
  
  # Create a test config file with various parameter types
  cat > yxa.yml << EOF
name: param-test-project
commands:
  string-param:
    run: echo "String parameter value \$PARAM_VALUE"
    description: "Command with string parameter"
    params:
      - name: PARAM_VALUE
        type: string
        description: "A string parameter"
        default: "default-string"
        flag: true
  
  int-param:
    run: echo "Int parameter value \$PARAM_VALUE"
    description: "Command with integer parameter"
    params:
      - name: PARAM_VALUE
        type: int
        description: "An integer parameter"
        default: "42"
        flag: true
  
  bool-param:
    run: echo "Bool parameter value \$PARAM_VALUE"
    description: "Command with boolean parameter"
    params:
      - name: PARAM_VALUE
        type: bool
        description: "A boolean parameter"
        default: "false"
        flag: true
  
  float-param:
    run: echo "Float parameter value \$PARAM_VALUE"
    description: "Command with float parameter"
    params:
      - name: PARAM_VALUE
        type: float
        description: "A float parameter"
        default: "3.14"
        flag: true
  
  required-param:
    run: echo "Required parameter value \$PARAM_VALUE"
    description: "Command with required parameter"
    params:
      - name: PARAM_VALUE
        type: string
        description: "A required parameter"
        required: true
        flag: true
  
  positional-param:
    run: echo "Positional parameter value \$PARAM_VALUE"
    description: "Command with positional parameter"
    params:
      - name: PARAM_VALUE
        type: string
        description: "A positional parameter"
        position: 0
        flag: false
  
  multiple-params:
    run: echo "First \$FIRST Second \$SECOND"
    description: "Command with multiple parameters"
    params:
      - name: FIRST
        type: string
        description: "First parameter"
        default: "first-default"
        flag: true
      - name: SECOND
        type: string
        description: "Second parameter"
        default: "second-default"
        flag: true
EOF
  
}

teardown() {
  rm -rf "$TEST_DIR"
}

@test "String parameter with default value" {
  run "$YXA_BIN" string-param
  
  # Assert that the command executed successfully
  [ "$status" -eq 0 ]
  
  # Assert that the output contains the default value
  [[ "$output" == *"String parameter value default-string"* ]]
}

@test "String parameter with custom value" {
  run "$YXA_BIN" string-param --PARAM_VALUE "custom-string"
  
  # Assert that the command executed successfully
  [ "$status" -eq 0 ]
  
  # Assert that the output contains the custom value
  [[ "$output" == *"String parameter value custom-string"* ]]
}

@test "Integer parameter with default value" {
  run "$YXA_BIN" int-param
  
  # Assert that the command executed successfully
  [ "$status" -eq 0 ]
  
  # Assert that the output contains the default value
  [[ "$output" == *"Int parameter value 42"* ]]
}

@test "Integer parameter with custom value" {
  run "$YXA_BIN" int-param --PARAM_VALUE 123
  
  # Assert that the command executed successfully
  [ "$status" -eq 0 ]
  
  # Assert that the output contains the custom value
  [[ "$output" == *"Int parameter value 123"* ]]
}

@test "Boolean parameter with default value" {
  run "$YXA_BIN" bool-param
  
  # Assert that the command executed successfully
  [ "$status" -eq 0 ]
  
  # Assert that the output contains the default value
  [[ "$output" == *"Bool parameter value false"* ]]
}

@test "Boolean parameter with custom value" {
  run "$YXA_BIN" bool-param --PARAM_VALUE true
  
  # Assert that the command executed successfully
  [ "$status" -eq 0 ]
  
  # Assert that the output contains the custom value
  [[ "$output" == *"Bool parameter value true"* ]]
}

@test "Float parameter with default value" {
  run "$YXA_BIN" float-param
  
  # Assert that the command executed successfully
  [ "$status" -eq 0 ]
  
  # Assert that the output contains the default value
  [[ "$output" == *"Float parameter value 3.14"* ]]
}

@test "Float parameter with custom value" {
  run "$YXA_BIN" float-param --PARAM_VALUE 2.718
  
  # Assert that the command executed successfully
  [ "$status" -eq 0 ]
  
  # Assert that the output contains the custom value
  [[ "$output" == *"Float parameter value 2.718"* ]]
}

@test "Required parameter without value should fail" {
  run "$YXA_BIN" required-param
  
  # Assert that the command failed
  [ "$status" -ne 0 ]
  
  # Assert that the error message mentions the required flag
  [[ "$output" == *"required"* ]]
}

@test "Required parameter with value should succeed" {
  run "$YXA_BIN" required-param --PARAM_VALUE "provided-value"
  
  # Assert that the command executed successfully
  [ "$status" -eq 0 ]
  
  # Assert that the output contains the provided value
  [[ "$output" == *"Required parameter value provided-value"* ]]
}

@test "Positional parameter with value" {
  run "$YXA_BIN" positional-param "positional-value"
  
  # Assert that the command executed successfully
  [ "$status" -eq 0 ]
  
  # Assert that the output contains the positional value
  [[ "$output" == *"Positional parameter value positional-value"* ]]
}

@test "Multiple parameters with default values" {
  run "$YXA_BIN" multiple-params
  
  # Assert that the command executed successfully
  [ "$status" -eq 0 ]
  
  # Assert that the output contains both default values
  [[ "$output" == *"First first-default"* ]]
  [[ "$output" == *"Second second-default"* ]]
}

@test "Multiple parameters with custom values" {
  run "$YXA_BIN" multiple-params --FIRST "custom-first" --SECOND "custom-second"
  
  # Assert that the command executed successfully
  [ "$status" -eq 0 ]
  
  # Assert that the output contains both custom values
  [[ "$output" == *"First custom-first"* ]]
  [[ "$output" == *"Second custom-second"* ]]
}

@test "Multiple parameters with partial custom values" {
  run "$YXA_BIN" multiple-params --SECOND "custom-second"
  
  # Assert that the command executed successfully
  [ "$status" -eq 0 ]
  
  # Assert that the output contains the default value for first and custom value for second
  [[ "$output" == *"First first-default"* ]]
  [[ "$output" == *"Second custom-second"* ]]
}

@test "Help output should show parameter descriptions" {
  run "$YXA_BIN" string-param --help
  
  # Assert that the command executed successfully
  [ "$status" -eq 0 ]
  
  # Assert that the help output contains the parameter description
  [[ "$output" == *"A string parameter"* ]]
}
