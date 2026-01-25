#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MAX_LINES="${MAX_LINES:-1000}"

paths=(
  "$ROOT/interpreters/go/pkg"
  "$ROOT/parser"
  "$ROOT/stdlib"
  "$ROOT/examples"
)

issue=0
for path in "${paths[@]}"; do
  if [[ ! -d "$path" ]]; then
    continue
  fi
  while IFS= read -r -d '' file; do
    lines=$(wc -l <"$file")
    if (( lines > MAX_LINES )); then
      printf '%s %s\n' "$lines" "$file"
      issue=1
    fi
  done < <(find "$path" -type f \( -name '*.go' -o -name '*.ts' -o -name '*.tsx' -o -name '*.js' \) \
    -not -path '*/node_modules/*' \
    -not -path '*/.gomodcache/*' \
    -not -path '*/tmp/*' \
    -not -path '*/tree-sitter-able/grammar.js' \
    -print0)
done

if (( issue )); then
  echo "One or more files exceed ${MAX_LINES} lines."
  exit 1
fi
