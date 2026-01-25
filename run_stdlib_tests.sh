#!/usr/bin/env bash
set -u
set -o pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEST_DIR="$ROOT/v12/stdlib/tests"

run_cmd() {
  local label="$1"
  local cmd="$2"
  echo ">>> $label"
  echo "\$ $cmd"
  eval "$cmd"
  return $?
}

status=0

if run_cmd "Able v12 Go treewalker stdlib tests" "\"$ROOT/v12/abletw\" test \"$TEST_DIR\""; then
  echo "Treewalker stdlib tests passed."
else
  echo "Treewalker stdlib tests failed."
  status=1
fi

if run_cmd "Able v12 Go bytecode stdlib tests" "\"$ROOT/v12/ablebc\" test \"$TEST_DIR\""; then
  echo "Bytecode stdlib tests passed."
else
  echo "Bytecode stdlib tests failed."
  status=1
fi

if [ "$status" -ne 0 ]; then
  echo "Stdlib test run failed."
else
  echo "Stdlib test run succeeded."
fi

exit "$status"
