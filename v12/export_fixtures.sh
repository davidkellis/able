#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo ">>> Exporting fixtures (TS-driven)"
(
  cd "$ROOT_DIR/interpreters/ts"
  bun run scripts/export-fixtures.ts
)
