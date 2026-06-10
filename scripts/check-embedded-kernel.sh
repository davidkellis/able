#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "$script_dir/.." && pwd)"
canonical_root="$repo_root/v12/kernel"
embedded_root="$repo_root/v12/interpreters/go/cmd/able/embedded/kernel"

for relative_path in package.yml src/kernel.able; do
  canonical_path="$canonical_root/$relative_path"
  embedded_path="$embedded_root/$relative_path"
  if [[ ! -f "$canonical_path" ]]; then
    printf 'embedded kernel check: missing canonical %s\n' "$relative_path" >&2
    exit 1
  fi
  if [[ ! -f "$embedded_path" ]]; then
    printf 'embedded kernel check: missing embedded %s\n' "$relative_path" >&2
    exit 1
  fi
  if ! cmp -s "$canonical_path" "$embedded_path"; then
    printf 'embedded kernel check: %s differs; run `just embed`\n' "$relative_path" >&2
    exit 1
  fi
done

printf 'embedded kernel check passed\n'
