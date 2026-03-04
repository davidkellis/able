#!/usr/bin/env bash
set -u
set -o pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Resolve stdlib tests directory: env override → sibling able-stdlib → $ABLE_HOME cache
if [ -n "${ABLE_STDLIB_ROOT:-}" ] && [ -d "$ABLE_STDLIB_ROOT/tests" ]; then
  TEST_DIR="$ABLE_STDLIB_ROOT/tests"
elif [ -d "$ROOT/../able-stdlib/tests" ]; then
  TEST_DIR="$(cd "$ROOT/../able-stdlib/tests" && pwd)"
elif [ -d "$ROOT/able-stdlib/tests" ]; then
  TEST_DIR="$(cd "$ROOT/able-stdlib/tests" && pwd)"
else
  ABLE_HOME="${ABLE_HOME:-$HOME/.able}"
  TEST_DIR=$(find "$ABLE_HOME/pkg/src/able" -maxdepth 2 -name tests -type d 2>/dev/null | head -1)
  if [ -z "$TEST_DIR" ]; then
    echo "Error: unable to locate stdlib tests directory."
    echo "Set ABLE_STDLIB_ROOT or run 'able setup' first."
    exit 1
  fi
fi
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

# Run from the stdlib root so the loader cwd doesn't traverse the entire able repo.
STDLIB_ROOT="$(dirname "$TEST_DIR")"

if run_cmd "Able v12 Go treewalker stdlib tests" "cd \"$STDLIB_ROOT\" && \"$ABLE_BIN\" --exec-mode=treewalker test \"$TEST_DIR\""; then
  echo "Treewalker stdlib tests passed."
else
  echo "Treewalker stdlib tests failed."
  status=1
fi

if run_cmd "Able v12 Go bytecode stdlib tests" "cd \"$STDLIB_ROOT\" && \"$ABLE_BIN\" --exec-mode=bytecode test \"$TEST_DIR\""; then
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
