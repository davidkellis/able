#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo ">>> Running TypeScript unit tests"
(
  cd "$ROOT_DIR/interpreter10"
  bun test
)

echo ">>> Running TypeScript fixture suite"
(
  cd "$ROOT_DIR/interpreter10"
  bun run scripts/run-fixtures.ts
)

echo ">>> Running Go unit tests"
(
  cd "$ROOT_DIR/interpreter10-go"
  tmp_gocache="$(mktemp -d)"
  trap 'rm -rf "$tmp_gocache"' EXIT
  GOCACHE="$tmp_gocache" go test ./...
)

echo "All tests completed successfully."
