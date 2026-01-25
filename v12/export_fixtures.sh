#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo ">>> Exporting fixtures (Go-driven)"
(
  cd "$ROOT_DIR/interpreters/go"
  go run ./cmd/fixture-exporter --root "$ROOT_DIR/fixtures/ast"
)
