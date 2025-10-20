#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

TYPECHECK_FIXTURES_MODE=""

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
    --help|-h)
      cat <<'EOF'
Usage: run_all_tests.sh [options]

Options:
  --typecheck-fixtures[=MODE]  Enable Go fixture typechecking (MODE: warn|strict, default warn).
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

if [[ -n "$TYPECHECK_FIXTURES_MODE" ]]; then
  echo ">>> Enabling Go fixture typechecking: $TYPECHECK_FIXTURES_MODE"
fi

echo ">>> Running TypeScript unit tests"
(
  cd "$ROOT_DIR/interpreter10"
  bun test
)

echo ">>> Running TypeScript fixture suite"
(
  cd "$ROOT_DIR/interpreter10"
  if [[ -n "$TYPECHECK_FIXTURES_MODE" ]]; then
    ABLE_TYPECHECK_FIXTURES="$TYPECHECK_FIXTURES_MODE" bun run scripts/run-fixtures.ts
  else
    bun run scripts/run-fixtures.ts
  fi
)

echo ">>> Running Go unit tests"
(
  cd "$ROOT_DIR/interpreter10-go"
  tmp_gocache="$(mktemp -d)"
  trap 'rm -rf "$tmp_gocache"' EXIT
  if [[ -n "$TYPECHECK_FIXTURES_MODE" ]]; then
    ABLE_TYPECHECK_FIXTURES="$TYPECHECK_FIXTURES_MODE" GOCACHE="$tmp_gocache" go test ./...
  else
    GOCACHE="$tmp_gocache" go test ./...
  fi
)

echo "All tests completed successfully."
