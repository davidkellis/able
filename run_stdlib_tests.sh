#!/usr/bin/env bash
set -u
set -o pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GO_DIR="$ROOT/v12/interpreters/go"
GO_CACHE="$GO_DIR/.gocache"
BIN_DIR="$(mktemp -d)"
ABLE_BIN="$BIN_DIR/able"
TEMP_ABLE_HOME=""
SCRIPT_ABLE_HOME="${ABLE_HOME:-}"
TEST_DIR=""

cleanup() {
  rm -rf "$BIN_DIR"
  if [ -n "$TEMP_ABLE_HOME" ] && [ -d "$TEMP_ABLE_HOME" ]; then
    rm -rf "$TEMP_ABLE_HOME"
  fi
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

find_cached_stdlib_tests() {
  local able_home="$1"
  if [ -z "$able_home" ] || [ ! -d "$able_home/pkg/src/able" ]; then
    return 1
  fi
  find "$able_home/pkg/src/able" -maxdepth 2 -name tests -type d 2>/dev/null | head -1
}

stdlib_tests_from_source_path() {
  local source_path="$1"
  if [ -z "$source_path" ]; then
    return 1
  fi
  if [ -d "$source_path/tests" ]; then
    printf '%s\n' "$source_path/tests"
    return 0
  fi
  local source_parent
  source_parent="$(dirname "$source_path")"
  if [ -d "$source_parent/tests" ]; then
    printf '%s\n' "$source_parent/tests"
    return 0
  fi
  return 1
}

find_setup_lock_stdlib_tests() {
  local able_home="$1"
  local lock_path="$able_home/setup.lock"
  if [ -z "$able_home" ] || [ ! -f "$lock_path" ]; then
    return 1
  fi
  local stdlib_source
  stdlib_source="$(
    awk '
      $1 == "-" && $2 == "name:" && $3 == "able" { in_able = 1; next }
      in_able && $1 == "source:" {
        sub(/^[[:space:]]*source:[[:space:]]*/, "", $0)
        print $0
        exit
      }
      in_able && $1 == "-" { exit }
    ' "$lock_path"
  )"
  stdlib_source="${stdlib_source#path:}"
  stdlib_tests_from_source_path "$stdlib_source"
}

resolve_stdlib_tests_dir() {
  if [ -n "${ABLE_STDLIB_ROOT:-}" ] && [ -d "$ABLE_STDLIB_ROOT/tests" ]; then
    TEST_DIR="$ABLE_STDLIB_ROOT/tests"
    return 0
  fi

  if [ "${ABLE_STDLIB_SKIP_SIBLING:-0}" != "1" ]; then
    if [ -d "$ROOT/../able-stdlib/tests" ]; then
      TEST_DIR="$(cd "$ROOT/../able-stdlib/tests" && pwd)"
      return 0
    fi
    if [ -d "$ROOT/able-stdlib/tests" ]; then
      TEST_DIR="$(cd "$ROOT/able-stdlib/tests" && pwd)"
      return 0
    fi
  fi

  local cached_home="${SCRIPT_ABLE_HOME:-${HOME:-}/.able}"
  TEST_DIR="$(find_cached_stdlib_tests "$cached_home")"
  if [ -n "$TEST_DIR" ]; then
    return 0
  fi
  TEST_DIR="$(find_setup_lock_stdlib_tests "$cached_home")"
  [ -n "$TEST_DIR" ]
}

bootstrap_stdlib_tests_dir() {
  if [ -n "$SCRIPT_ABLE_HOME" ]; then
    mkdir -p "$SCRIPT_ABLE_HOME"
  else
    TEMP_ABLE_HOME="$(mktemp -d)"
    SCRIPT_ABLE_HOME="$TEMP_ABLE_HOME"
  fi

  echo ">>> Bootstrapping stdlib cache with 'able setup'"
  echo "\$ ABLE_HOME=\"$SCRIPT_ABLE_HOME\" \"$ABLE_BIN\" setup"
  if ! env ABLE_HOME="$SCRIPT_ABLE_HOME" "$ABLE_BIN" setup; then
    echo "Error: unable to bootstrap stdlib tests directory." >&2
    exit 1
  fi

  TEST_DIR="$(find_setup_lock_stdlib_tests "$SCRIPT_ABLE_HOME")"
  if [ -z "$TEST_DIR" ]; then
    TEST_DIR="$(find_cached_stdlib_tests "$SCRIPT_ABLE_HOME")"
  fi
  if [ -z "$TEST_DIR" ]; then
    echo "Error: unable to locate stdlib tests directory after setup." >&2
    exit 1
  fi
}

status=0

echo ">>> Building Able v12 CLI"
echo "\$ go build -o \"$ABLE_BIN\" ./cmd/able"
mkdir -p "$GO_CACHE"
if ! env GOCACHE="$GO_CACHE" go build -C "$GO_DIR" -o "$ABLE_BIN" ./cmd/able; then
  echo "Able v12 CLI build failed."
  exit 1
fi

if ! resolve_stdlib_tests_dir; then
  bootstrap_stdlib_tests_dir
fi

# Run from the stdlib root so the loader cwd doesn't traverse the entire able repo.
STDLIB_ROOT="$(dirname "$TEST_DIR")"
ABLE_RUN_ENV=""
if [ -n "$SCRIPT_ABLE_HOME" ]; then
  ABLE_RUN_ENV="ABLE_HOME=\"$SCRIPT_ABLE_HOME\" "
fi

if run_cmd "Able v12 Go treewalker stdlib tests" "cd \"$STDLIB_ROOT\" && ${ABLE_RUN_ENV}\"$ABLE_BIN\" --exec-mode=treewalker test \"$TEST_DIR\""; then
  echo "Treewalker stdlib tests passed."
else
  echo "Treewalker stdlib tests failed."
  status=1
fi

if run_cmd "Able v12 Go bytecode stdlib tests" "cd \"$STDLIB_ROOT\" && ${ABLE_RUN_ENV}\"$ABLE_BIN\" --exec-mode=bytecode test \"$TEST_DIR\""; then
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
