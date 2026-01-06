#!/usr/bin/env bash
set -u
set -o pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEST_DIR="$ROOT/v11/stdlib/tests"

run_cmd() {
  local label="$1"
  local cmd="$2"
  echo ">>> $label"
  echo "\$ $cmd"
  eval "$cmd"
  return $?
}

status=0

if run_cmd "Able v11 TypeScript stdlib tests" "\"$ROOT/v11/ablets\" test \"$TEST_DIR\""; then
  echo "TypeScript stdlib tests passed."
else
  echo "TypeScript stdlib tests failed."
  status=1
fi

if run_cmd "Able v11 Go stdlib tests" "\"$ROOT/v11/ablego\" test \"$TEST_DIR\""; then
  echo "Go stdlib tests passed."
else
  echo "Go stdlib tests failed."
  status=1
fi

if [ "$status" -ne 0 ]; then
  echo "Stdlib test run failed."
else
  echo "Stdlib test run succeeded."
fi

exit "$status"
