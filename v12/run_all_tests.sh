#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

TYPECHECK_FIXTURES_MODE="strict"
FIXTURE_ONLY=false
EXPORT_FIXTURES=false

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
    --help|-h)
      cat <<'EOF'
Usage: run_all_tests.sh [options]

Options:
  --fixture                 Run only Go fixture tests.
  --export-fixtures         Run fixture export step (Go-based exporter).
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
  tmp_gocache="$(mktemp -d)"
  trap 'rm -rf "$tmp_gocache"' EXIT
  if [[ "$FIXTURE_ONLY" == true ]]; then
    ABLE_TYPECHECK_FIXTURES="$TYPECHECK_FIXTURES_MODE" GOCACHE="$tmp_gocache" go test ./pkg/interpreter -run 'Fixture' -count=1
  else
    ABLE_TYPECHECK_FIXTURES="$TYPECHECK_FIXTURES_MODE" GOCACHE="$tmp_gocache" go test ./...
  fi
)

echo "All tests completed successfully."
