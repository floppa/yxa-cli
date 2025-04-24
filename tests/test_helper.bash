#!/usr/bin/env bash

# Common test helper functions for BATS tests

# Get the directory of this script
TESTS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$( cd "$TESTS_DIR/.." && pwd )"

# Add any common setup or helper functions here

# Helper to check if we're running on macOS
is_macos() {
  [[ "$(uname -s)" == "Darwin" ]]
}

# Helper to check if we're running on Linux
is_linux() {
  [[ "$(uname -s)" == "Linux" ]]
}

# Helper to check if we're running on Windows (in Git Bash or similar)
is_windows() {
  [[ "$(uname -s)" == MINGW* ]] || [[ "$(uname -s)" == CYGWIN* ]]
}

# Helper to skip tests on specific platforms
skip_if_not_macos() {
  if ! is_macos; then
    skip "This test only runs on macOS"
  fi
}

skip_if_not_linux() {
  if ! is_linux; then
    skip "This test only runs on Linux"
  fi
}

skip_if_not_windows() {
  if ! is_windows; then
    skip "This test only runs on Windows"
  fi
}

# Helper to check if a command exists
command_exists() {
  command -v "$1" >/dev/null 2>&1
}

# Helper to create a temporary YXA config file
create_temp_config() {
  local config_content="$1"
  local temp_dir="$(mktemp -d)"
  local config_file="${temp_dir}/yxa.yml"
  
  echo "$config_content" > "$config_file"
  echo "$config_file"
}

# Helper to clean up temporary files
cleanup_temp() {
  local temp_path="$1"
  if [[ -d "$temp_path" ]]; then
    rm -rf "$temp_path"
  elif [[ -f "$temp_path" ]]; then
    rm -f "$temp_path"
  fi
}
