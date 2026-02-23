#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

TYPECHECK_FIXTURES_MODE="strict"
RUN_FALLBACK_AUDIT=true

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
    --skip-fallback-audit)
      RUN_FALLBACK_AUDIT=false
      shift
      ;;
    --help|-h)
      cat <<'EOF'
Usage: run_compiler_full_matrix.sh [options]

Runs full compiler fixture matrix sweeps intended for nightly/manual validation.

Options:
  --typecheck-fixtures[=MODE]  Set fixture typechecking (MODE: off|warn|strict, default strict).
  --skip-fallback-audit        Skip TestCompilerExecFixtureFallbacks (enabled by default).
  -h, --help                   Show this help text.

Environment overrides (defaults shown):
  ABLE_COMPILER_EXEC_FIXTURES=all
  ABLE_COMPILER_STRICT_DISPATCH_FIXTURES=all
  ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=all
  ABLE_COMPILER_GLOBAL_LOOKUP_STRICT_TOTAL=1
  ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES=all
  ABLE_COMPILER_SUITE_TIMEOUT=25m
  ABLE_COMPILER_SUITE_WALL_TIMEOUT=30m
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
interface_lookup_fixtures="${ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES:-all}"
global_lookup_strict_total="${ABLE_COMPILER_GLOBAL_LOOKUP_STRICT_TOTAL:-1}"
boundary_fixtures="${ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES:-all}"
suite_timeout="${ABLE_COMPILER_SUITE_TIMEOUT:-25m}"
suite_wall_timeout="${ABLE_COMPILER_SUITE_WALL_TIMEOUT:-30m}"

echo ">>> Compiler full matrix (typecheck=$TYPECHECK_FIXTURES_MODE)"
echo ">>> ABLE_COMPILER_EXEC_FIXTURES=$exec_fixtures"
echo ">>> ABLE_COMPILER_STRICT_DISPATCH_FIXTURES=$strict_fixtures"
echo ">>> ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=$interface_lookup_fixtures"
echo ">>> ABLE_COMPILER_GLOBAL_LOOKUP_STRICT_TOTAL=$global_lookup_strict_total"
echo ">>> ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES=$boundary_fixtures"
echo ">>> ABLE_COMPILER_SUITE_TIMEOUT=$suite_timeout"
echo ">>> ABLE_COMPILER_SUITE_WALL_TIMEOUT=$suite_wall_timeout"
echo ">>> RUN_FALLBACK_AUDIT=$RUN_FALLBACK_AUDIT"

(
  cd "$ROOT_DIR/interpreters/go"
  gocache="$ROOT_DIR/interpreters/go/.gocache"
  if [[ "${ABLE_GOCACHE:-}" == "tmp" ]]; then
    gocache="$(mktemp -d)"
    trap 'rm -rf "$gocache"' EXIT
  elif [[ -n "${GOCACHE:-}" ]]; then
    gocache="$GOCACHE"
  fi

  run_suite() {
    local label="$1"
    shift
    echo ">>> running $label"
    if command -v timeout >/dev/null 2>&1; then
      timeout "$suite_wall_timeout" "$@"
      return
    fi
    "$@"
  }

  run_suite "TestCompilerExecFixtures" \
    env ABLE_TYPECHECK_FIXTURES="$TYPECHECK_FIXTURES_MODE" \
      ABLE_COMPILER_EXEC_FIXTURES="$exec_fixtures" \
      GOCACHE="$gocache" \
      ABLE_COMPILER_EXEC_GOCACHE="$gocache" \
      go test ./pkg/compiler -run TestCompilerExecFixtures -count=1 -timeout="$suite_timeout"

  run_suite "TestCompilerStrictDispatchForStdlibHeavyFixtures" \
    env ABLE_TYPECHECK_FIXTURES="$TYPECHECK_FIXTURES_MODE" \
      ABLE_COMPILER_STRICT_DISPATCH_FIXTURES="$strict_fixtures" \
      GOCACHE="$gocache" \
      ABLE_COMPILER_EXEC_GOCACHE="$gocache" \
      go test ./pkg/compiler -run TestCompilerStrictDispatchForStdlibHeavyFixtures -count=1 -timeout="$suite_timeout"

  run_suite "TestCompilerInterfaceLookupBypassForStaticFixtures" \
    env ABLE_TYPECHECK_FIXTURES="$TYPECHECK_FIXTURES_MODE" \
      ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES="$interface_lookup_fixtures" \
      ABLE_COMPILER_GLOBAL_LOOKUP_STRICT_TOTAL="$global_lookup_strict_total" \
      GOCACHE="$gocache" \
      ABLE_COMPILER_EXEC_GOCACHE="$gocache" \
      go test ./pkg/compiler -run TestCompilerInterfaceLookupBypassForStaticFixtures -count=1 -timeout="$suite_timeout"

  run_suite "TestCompilerBoundaryFallbackMarkerForStaticFixtures" \
    env ABLE_TYPECHECK_FIXTURES="$TYPECHECK_FIXTURES_MODE" \
      ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES="$boundary_fixtures" \
      GOCACHE="$gocache" \
      ABLE_COMPILER_EXEC_GOCACHE="$gocache" \
      go test ./pkg/compiler -run TestCompilerBoundaryFallbackMarkerForStaticFixtures -count=1 -timeout="$suite_timeout"

  if [[ "$RUN_FALLBACK_AUDIT" == true ]]; then
    run_suite "TestCompilerExecFixtureFallbacks" \
      env ABLE_TYPECHECK_FIXTURES="$TYPECHECK_FIXTURES_MODE" \
        ABLE_COMPILER_FALLBACK_AUDIT=1 \
        ABLE_COMPILER_EXEC_FIXTURES="$exec_fixtures" \
        GOCACHE="$gocache" \
        ABLE_COMPILER_EXEC_GOCACHE="$gocache" \
        go test ./pkg/compiler -run TestCompilerExecFixtureFallbacks -count=1 -timeout="$suite_timeout"
  fi
)

echo "Compiler full-matrix sweep completed successfully."
