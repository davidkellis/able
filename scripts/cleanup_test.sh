#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source_script="$script_dir/cleanup.sh"
test_root="$(mktemp -d "${TMPDIR:-/var/tmp}/able-cleanup-test.XXXXXX")"
trap 'rm -rf -- "$test_root"' EXIT

mkdir -p \
  "$test_root/scripts" \
  "$test_root/v12/__pycache__" \
  "$test_root/v12/.profiles" \
  "$test_root/v12/fixtures/bench/sample/target" \
  "$test_root/v12/interpreters/go"
cp "$source_script" "$test_root/scripts/cleanup.sh"
chmod +x "$test_root/scripts/cleanup.sh"

git -C "$test_root" init -q
printf 'tracked\n' > "$test_root/tracked.txt"
git -C "$test_root" add tracked.txt

printf 'binary\n' > "$test_root/v12/interpreters/go/ablec"
printf '{}\n' > "$test_root/v12/interpreters/go/.stats_persistent_sorted_set_enabled.json"
printf 'bytecode\n' > "$test_root/v12/__pycache__/bench.pyc"
printf 'fixture build\n' > "$test_root/v12/fixtures/bench/sample/target/output"
printf 'profile evidence\n' > "$test_root/v12/.profiles/keep.pprof"

preview="$("$test_root/scripts/cleanup.sh")"
for expected in \
  "v12/interpreters/go/ablec" \
  "v12/interpreters/go/.stats_persistent_sorted_set_enabled.json" \
  "v12/__pycache__" \
  "v12/fixtures/bench/sample/target"; do
  if ! grep -Fq "would remove $expected " <<<"$preview"; then
    printf 'cleanup test: dry run omitted %s\n' "$expected" >&2
    exit 1
  fi
done
if grep -Fq "v12/.profiles" <<<"$preview"; then
  printf 'cleanup test: default dry run selected profile evidence\n' >&2
  exit 1
fi

"$test_root/scripts/cleanup.sh" --apply >/dev/null
for removed in \
  "$test_root/v12/interpreters/go/ablec" \
  "$test_root/v12/interpreters/go/.stats_persistent_sorted_set_enabled.json" \
  "$test_root/v12/__pycache__" \
  "$test_root/v12/fixtures/bench/sample/target"; do
  if [[ -e "$removed" ]]; then
    printf 'cleanup test: apply retained %s\n' "$removed" >&2
    exit 1
  fi
done
for retained in "$test_root/tracked.txt" "$test_root/v12/.profiles/keep.pprof"; do
  if [[ ! -f "$retained" ]]; then
    printf 'cleanup test: apply removed protected path %s\n' "$retained" >&2
    exit 1
  fi
done

printf 'cleanup policy test passed\n'
