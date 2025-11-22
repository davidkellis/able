#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PARITY_REPORT_DIR="$ROOT_DIR/tmp"
PARITY_REPORT_PATH="$PARITY_REPORT_DIR/parity-report.json"
declare -a PARITY_REPORT_COPIES=()

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

mkdir -p "$PARITY_REPORT_DIR"

copy_parity_report() {
  local target="$1"
  if [[ -z "$target" ]]; then
    return
  fi
  for recorded in "${PARITY_REPORT_COPIES[@]}"; do
    if [[ "$recorded" == "$target" ]]; then
      return
    fi
  done
  PARITY_REPORT_COPIES+=("$target")
  local dest_dir
  dest_dir="$(dirname "$target")"
  mkdir -p "$dest_dir"
  cp "$PARITY_REPORT_PATH" "$target"
  echo ">>> Parity JSON report copied to $target"
}

echo ">>> Exporting fixtures"
"$ROOT_DIR/export_fixtures.sh"

echo ">>> Running TypeScript unit tests"
(
  cd "$ROOT_DIR/interpreters/ts"
  bun test
)

echo ">>> Running Able CLI tests"
(
  cd "$ROOT_DIR/interpreters/ts"
  bun test test/cli
)

echo ">>> Running TypeScript fixture suite"
(
  cd "$ROOT_DIR/interpreters/ts"
  if [[ -n "$TYPECHECK_FIXTURES_MODE" ]]; then
    ABLE_TYPECHECK_FIXTURES="$TYPECHECK_FIXTURES_MODE" bun run scripts/run-fixtures.ts
  else
    bun run scripts/run-fixtures.ts
  fi
)

echo ">>> Running cross-interpreter parity harness"
(
  cd "$ROOT_DIR/interpreters/ts"
  if [[ -n "$TYPECHECK_FIXTURES_MODE" ]]; then
    echo ">>> Parity diagnostics disabled (ABLE_TYPECHECK_FIXTURES=off) to focus on runtime parity"
  fi
  ABLE_TYPECHECK_FIXTURES="off" bun run scripts/run-parity.ts --suite fixtures --suite examples --report "$PARITY_REPORT_PATH"
)
echo ">>> Parity JSON report written to $PARITY_REPORT_PATH"
if [[ -n "${ABLE_PARITY_REPORT_DEST:-}" ]]; then
  copy_parity_report "$ABLE_PARITY_REPORT_DEST"
fi
if [[ -n "${CI_ARTIFACTS_DIR:-}" ]]; then
  copy_parity_report "$CI_ARTIFACTS_DIR/parity-report.json"
fi

echo ">>> Running Go unit tests"
(
  cd "$ROOT_DIR/interpreters/go"
  tmp_gocache="$(mktemp -d)"
  trap 'rm -rf "$tmp_gocache"' EXIT
  if [[ -n "$TYPECHECK_FIXTURES_MODE" ]]; then
    ABLE_TYPECHECK_FIXTURES="$TYPECHECK_FIXTURES_MODE" GOCACHE="$tmp_gocache" go test ./...
  else
    GOCACHE="$tmp_gocache" go test ./...
  fi
)

echo "All tests completed successfully."
