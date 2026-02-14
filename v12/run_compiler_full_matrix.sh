#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

TYPECHECK_FIXTURES_MODE="strict"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --typecheck-fixtures)
      if [[ $# -lt 2 ]]; then
        echo "Missing value for --typecheck-fixtures" >&2
        exit 1
      fi
      TYPECHECK_FIXTURES_MODE="$2"
      shift 2
      ;;
    --typecheck-fixtures=*)
      TYPECHECK_FIXTURES_MODE="${1#*=}"
      shift
      ;;
    --help|-h)
      cat <<'EOF'
Usage: run_compiler_full_matrix.sh [options]

Runs full compiler fixture matrix sweeps intended for nightly/manual validation.

Options:
  --typecheck-fixtures[=MODE]  Set fixture typechecking (MODE: off|warn|strict, default strict).
  -h, --help                   Show this help text.

Environment overrides (defaults shown):
  ABLE_COMPILER_EXEC_FIXTURES=all
  ABLE_COMPILER_STRICT_DISPATCH_FIXTURES=all
  ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES=all
EOF
      exit 0
      ;;
    *)
      echo "Unknown option: $1" >&2
      exit 1
      ;;
  esac
done

exec_fixtures="${ABLE_COMPILER_EXEC_FIXTURES:-all}"
strict_fixtures="${ABLE_COMPILER_STRICT_DISPATCH_FIXTURES:-all}"
boundary_fixtures="${ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES:-all}"

echo ">>> Compiler full matrix (typecheck=$TYPECHECK_FIXTURES_MODE)"
echo ">>> ABLE_COMPILER_EXEC_FIXTURES=$exec_fixtures"
echo ">>> ABLE_COMPILER_STRICT_DISPATCH_FIXTURES=$strict_fixtures"
echo ">>> ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES=$boundary_fixtures"

(
  cd "$ROOT_DIR/interpreters/go"
  gocache="$ROOT_DIR/interpreters/go/.gocache"
  if [[ "${ABLE_GOCACHE:-}" == "tmp" ]]; then
    gocache="$(mktemp -d)"
    trap 'rm -rf "$gocache"' EXIT
  elif [[ -n "${GOCACHE:-}" ]]; then
    gocache="$GOCACHE"
  fi

  ABLE_TYPECHECK_FIXTURES="$TYPECHECK_FIXTURES_MODE" \
    ABLE_COMPILER_EXEC_FIXTURES="$exec_fixtures" \
    GOCACHE="$gocache" \
    ABLE_COMPILER_EXEC_GOCACHE="$gocache" \
    go test ./pkg/compiler -run TestCompilerExecFixtures -count=1

  ABLE_TYPECHECK_FIXTURES="$TYPECHECK_FIXTURES_MODE" \
    ABLE_COMPILER_STRICT_DISPATCH_FIXTURES="$strict_fixtures" \
    GOCACHE="$gocache" \
    ABLE_COMPILER_EXEC_GOCACHE="$gocache" \
    go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1

  ABLE_TYPECHECK_FIXTURES="$TYPECHECK_FIXTURES_MODE" \
    ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES="$boundary_fixtures" \
    GOCACHE="$gocache" \
    ABLE_COMPILER_EXEC_GOCACHE="$gocache" \
    go test ./pkg/compiler -run TestCompilerBoundaryFallbackMarkerForStaticFixtures -count=1
)

echo "Compiler full-matrix sweep completed successfully."

