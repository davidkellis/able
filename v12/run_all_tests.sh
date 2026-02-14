#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

TYPECHECK_FIXTURES_MODE="strict"
FIXTURE_ONLY=false
EXPORT_FIXTURES=false
COMPILER_FULL_MATRIX=false

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
    --fixture)
      FIXTURE_ONLY=true
      shift
      ;;
    --export-fixtures)
      EXPORT_FIXTURES=true
      shift
      ;;
    --compiler-full-matrix)
      COMPILER_FULL_MATRIX=true
      shift
      ;;
    --help|-h)
      cat <<'EOF'
Usage: run_all_tests.sh [options]

Options:
  --fixture                 Run only Go fixture tests.
  --export-fixtures         Run fixture export step (Go-based exporter).
  --compiler-full-matrix    Run full compiler fixture matrix sweep (`...=all`).
  --typecheck-fixtures[=MODE]  Set fixture typechecking (MODE: off|warn|strict, default strict).
  --typecheck-fixtures-warn    Shorthand for --typecheck-fixtures=warn.
  --typecheck-fixtures-strict  Shorthand for --typecheck-fixtures=strict.
  -h, --help                   Show this help text.
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

echo ">>> Running Go tests"
(
  cd "$ROOT_DIR/interpreters/go"
  gocache="$ROOT_DIR/interpreters/go/.gocache"
  if [[ "${ABLE_GOCACHE:-}" == "tmp" ]]; then
    gocache="$(mktemp -d)"
    trap 'rm -rf "$gocache"' EXIT
  elif [[ -n "${GOCACHE:-}" ]]; then
    gocache="$GOCACHE"
  fi
  if [[ "$FIXTURE_ONLY" == true ]]; then
    ABLE_TYPECHECK_FIXTURES="$TYPECHECK_FIXTURES_MODE" GOCACHE="$gocache" ABLE_COMPILER_EXEC_GOCACHE="$gocache" go test ./pkg/interpreter -run 'Fixture' -count=1 -exec-mode=treewalker
    ABLE_TYPECHECK_FIXTURES="$TYPECHECK_FIXTURES_MODE" GOCACHE="$gocache" ABLE_COMPILER_EXEC_GOCACHE="$gocache" go test ./pkg/interpreter -run 'Fixture' -count=1 -exec-mode=bytecode
  else
    ABLE_TYPECHECK_FIXTURES="$TYPECHECK_FIXTURES_MODE" GOCACHE="$gocache" ABLE_COMPILER_EXEC_GOCACHE="$gocache" go test ./...
    ABLE_TYPECHECK_FIXTURES="$TYPECHECK_FIXTURES_MODE" GOCACHE="$gocache" ABLE_COMPILER_EXEC_GOCACHE="$gocache" go test ./pkg/interpreter -run 'Fixture' -count=1 -exec-mode=bytecode
  fi
)

if [[ "$COMPILER_FULL_MATRIX" == true ]]; then
  echo ">>> Running compiler full matrix sweep"
  ABLE_COMPILER_EXEC_FIXTURES="${ABLE_COMPILER_EXEC_FIXTURES:-all}" \
    ABLE_COMPILER_STRICT_DISPATCH_FIXTURES="${ABLE_COMPILER_STRICT_DISPATCH_FIXTURES:-all}" \
    ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES="${ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES:-all}" \
    "$ROOT_DIR/run_compiler_full_matrix.sh" --typecheck-fixtures="$TYPECHECK_FIXTURES_MODE"
fi

echo "All tests completed successfully."
