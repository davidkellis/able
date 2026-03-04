#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

TYPECHECK_FIXTURES_MODE="strict"
EXPORT_FIXTURES=false
RUN_TREEWALKER=false
RUN_BYTECODE=false
RUN_COMPILER=false
RUN_ALL=true
FILTER=""
GO_TEST_TIMEOUT="${GO_TEST_TIMEOUT:-30m}"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --typecheck-fixtures)
      TYPECHECK_FIXTURES_MODE="warn"
      shift
      ;;
    --typecheck-fixtures=*)
      TYPECHECK_FIXTURES_MODE="${1#*=}"
      shift
      ;;
    --typecheck-fixtures-warn)
      TYPECHECK_FIXTURES_MODE="warn"
      shift
      ;;
    --typecheck-fixtures-strict)
      TYPECHECK_FIXTURES_MODE="strict"
      shift
      ;;
    --export-fixtures)
      EXPORT_FIXTURES=true
      shift
      ;;
    --treewalker)
      RUN_TREEWALKER=true
      RUN_ALL=false
      shift
      ;;
    --bytecode)
      RUN_BYTECODE=true
      RUN_ALL=false
      shift
      ;;
    --compiler)
      RUN_COMPILER=true
      RUN_ALL=false
      shift
      ;;
    --filter)
      FILTER="$2"
      shift 2
      ;;
    --filter=*)
      FILTER="${1#*=}"
      shift
      ;;
    --help|-h)
      cat <<'EOF'
Usage: run_all_tests.sh [options]

Options:
  --treewalker              Run only treewalker interpreter tests.
  --bytecode                Run only bytecode interpreter tests.
  --compiler                Run only compiler tests (full matrix).
  --filter PATTERN          Pass -run PATTERN to narrow tests within selected subsets.
  --export-fixtures         Run fixture export step (Go-based exporter).
  --typecheck-fixtures[=MODE]  Set fixture typechecking (MODE: off|warn|strict, default strict).
  --typecheck-fixtures-warn    Shorthand for --typecheck-fixtures=warn.
  --typecheck-fixtures-strict  Shorthand for --typecheck-fixtures=strict.
  -h, --help                   Show this help text.

When no subset flags (--treewalker, --bytecode, --compiler) are given, all tests
run: full package suite plus a bytecode fixture pass.
Subset flags are combinable (e.g. --treewalker --compiler).
EOF
      exit 0
      ;;
    *)
      echo "Unknown option: $1" >&2
      exit 1
      ;;
  esac
done

echo ">>> Fixture typechecking mode: $TYPECHECK_FIXTURES_MODE"

if [[ "$EXPORT_FIXTURES" == true ]]; then
  echo ">>> Exporting fixtures"
  "$ROOT_DIR/export_fixtures.sh"
fi

echo ">>> Checking exec coverage index"
node "$ROOT_DIR/scripts/check-exec-coverage.mjs"

RUN_FLAG=()
if [[ -n "$FILTER" ]]; then
  RUN_FLAG=(-run "$FILTER")
fi

echo ">>> Running Go tests"
(
  cd "$ROOT_DIR/interpreters/go"
  mapfile -t all_pkgs < <(go list ./... | grep -Ev '^able/interpreter-go/tmp(/|$)')
  if [[ ${#all_pkgs[@]} -eq 0 ]]; then
    echo "No Go packages found to test." >&2
    exit 1
  fi
  gocache="$ROOT_DIR/interpreters/go/.gocache"
  if [[ "${ABLE_GOCACHE:-}" == "tmp" ]]; then
    gocache="$(mktemp -d)"
    trap 'rm -rf "$gocache"' EXIT
  elif [[ -n "${GOCACHE:-}" ]]; then
    gocache="$GOCACHE"
  fi

  if [[ "$RUN_ALL" == true ]]; then
    echo ">>> Running all packages"
    ABLE_TYPECHECK_FIXTURES="$TYPECHECK_FIXTURES_MODE" \
      GOCACHE="$gocache" \
      ABLE_COMPILER_EXEC_GOCACHE="$gocache" \
      go test -timeout "$GO_TEST_TIMEOUT" "${RUN_FLAG[@]}" "${all_pkgs[@]}"

    echo ">>> Running bytecode fixture pass"
    ABLE_TYPECHECK_FIXTURES="$TYPECHECK_FIXTURES_MODE" \
      GOCACHE="$gocache" \
      ABLE_COMPILER_EXEC_GOCACHE="$gocache" \
      go test -timeout "$GO_TEST_TIMEOUT" "${RUN_FLAG[@]}" ./pkg/interpreter -count=1 -exec-mode=bytecode
  else
    if [[ "$RUN_TREEWALKER" == true ]]; then
      echo ">>> Running treewalker interpreter tests"
      ABLE_TYPECHECK_FIXTURES="$TYPECHECK_FIXTURES_MODE" \
        GOCACHE="$gocache" \
        ABLE_COMPILER_EXEC_GOCACHE="$gocache" \
        go test -timeout "$GO_TEST_TIMEOUT" "${RUN_FLAG[@]}" ./pkg/interpreter -count=1
    fi
    if [[ "$RUN_BYTECODE" == true ]]; then
      echo ">>> Running bytecode interpreter tests"
      ABLE_TYPECHECK_FIXTURES="$TYPECHECK_FIXTURES_MODE" \
        GOCACHE="$gocache" \
        ABLE_COMPILER_EXEC_GOCACHE="$gocache" \
        go test -timeout "$GO_TEST_TIMEOUT" "${RUN_FLAG[@]}" ./pkg/interpreter -count=1 -exec-mode=bytecode
    fi
    if [[ "$RUN_COMPILER" == true ]]; then
      echo ">>> Running compiler tests (full matrix)"
      ABLE_TYPECHECK_FIXTURES="$TYPECHECK_FIXTURES_MODE" \
        GOCACHE="$gocache" \
        ABLE_COMPILER_EXEC_GOCACHE="$gocache" \
        ABLE_COMPILER_EXEC_FIXTURES=all \
        ABLE_COMPILER_STRICT_DISPATCH_FIXTURES=all \
        ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all \
        ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES=all \
        go test -timeout "$GO_TEST_TIMEOUT" "${RUN_FLAG[@]}" ./pkg/compiler ./pkg/compiler/bridge -count=1
    fi
  fi
)

echo "All tests completed successfully."
