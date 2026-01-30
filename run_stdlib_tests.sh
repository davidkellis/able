#!/usr/bin/env bash
set -u
set -o pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEST_DIR="$ROOT/v12/stdlib/tests"
GO_DIR="$ROOT/v12/interpreters/go"
GO_CACHE="$GO_DIR/.gocache"
BIN_DIR="$(mktemp -d)"
ABLE_BIN="$BIN_DIR/able"

cleanup() {
  rm -rf "$BIN_DIR"
}
trap cleanup EXIT

run_cmd() {
  local label="$1"
  local cmd="$2"
  local start=0
  local end=0
  local elapsed=0
  echo ">>> $label"
  echo "\$ $cmd"
  start=$(date +%s)
  eval "$cmd"
  local rc=$?
  end=$(date +%s)
  elapsed=$((end - start))
  echo ">>> $label completed in ${elapsed}s"
  return $rc
}

status=0

echo ">>> Building Able v12 CLI"
echo "\$ go build -o \"$ABLE_BIN\" ./cmd/able"
mkdir -p "$GO_CACHE"
if ! env GOCACHE="$GO_CACHE" go build -C "$GO_DIR" -o "$ABLE_BIN" ./cmd/able; then
  echo "Able v12 CLI build failed."
  exit 1
fi

if run_cmd "Able v12 Go treewalker stdlib tests" "\"$ABLE_BIN\" --exec-mode=treewalker test \"$TEST_DIR\""; then
  echo "Treewalker stdlib tests passed."
else
  echo "Treewalker stdlib tests failed."
  status=1
fi

if run_cmd "Able v12 Go bytecode stdlib tests" "\"$ABLE_BIN\" --exec-mode=bytecode test \"$TEST_DIR\""; then
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
