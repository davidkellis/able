#!/usr/bin/env bash
set -euo pipefail

# Reclaim only generated, ignored workspace state. This is intentionally more
# conservative than `git clean`: it refuses any path that contains tracked
# files and never touches source fixtures, the canonical stdlib, or v10/v11.

usage() {
  cat <<'EOF'
Usage: scripts/cleanup.sh [--apply] [--include-profiles]

Without --apply, print the project-local generated artifacts that would be
removed. --apply performs the deletion after checking that no selected path
contains Git-tracked files.

--include-profiles additionally removes .profiles/, which is archived
diagnostic evidence rather than a build dependency. It is excluded by default.

The default cleanup covers v12 build caches, temporary benchmark workspaces,
generated fixture target trees, generated binaries, local Go test binaries,
and Python bytecode/test caches. It deliberately leaves source fixtures,
able-stdlib, and v10/v11 untouched.
EOF
}

apply=false
include_profiles=false
while (($# > 0)); do
  case "$1" in
    --apply)
      apply=true
      ;;
    --include-profiles)
      include_profiles=true
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      printf 'cleanup: unknown option: %s\n' "$1" >&2
      usage >&2
      exit 2
      ;;
  esac
  shift
done

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "$script_dir/.." && pwd)"
if ! git -C "$repo_root" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  printf 'cleanup: %s is not a Git worktree\n' "$repo_root" >&2
  exit 2
fi

declare -a candidates=()
declare -A seen=()

add_candidate() {
  local rel="$1"
  local abs="$repo_root/$rel"
  [[ -e "$abs" || -L "$abs" ]] || return 0
  [[ "$abs" == "$repo_root/"* ]] || {
    printf 'cleanup: refusing path outside repository: %s\n' "$abs" >&2
    exit 2
  }
  [[ -z "${seen[$rel]:-}" ]] || return 0
  seen[$rel]=1
  candidates+=("$rel")
}

# Project-local compiler/build caches and temporary harness workspaces.
for rel in \
  .gocache .gotmp .tmp tmp target \
  v12/.gocache v12/.gotmp v12/.tmp v12/tmp v12/target \
  v12/examples/benchmarks/target \
  v12/interpreters/go/.gocache v12/interpreters/go/.gotmp \
  v12/interpreters/go/target v12/interpreters/go/tmp \
  v12/interpreters/go/v12/tmp \
  v12/wasm/ablewasm.wasm \
  able v12/interpreters/go/able v12/interpreters/go/ablec \
  v12/interpreters/go/able.test v12/interpreters/go/interpreter.test \
  v12/interpreters/go/.stats_persistent_sorted_set_enabled.json; do
  add_candidate "$rel"
done

# Compiler fixture builds are intentionally kept next to their source fixture.
# Select only directories named target; no fixture source directory is removed.
while IFS= read -r -d '' target_dir; do
  add_candidate "${target_dir#"$repo_root/"}"
done < <(find "$repo_root/v12/fixtures" -type d -name target -prune -print0)

# Python tooling and unit tests leave safe, reproducible caches beside the
# benchmark scripts. Restrict recursive discovery to active v12 so cleanup
# never mutates the frozen v10/v11 workspaces.
for rel in .pytest_cache .mypy_cache .ruff_cache; do
  add_candidate "$rel"
done
while IFS= read -r -d '' python_cache; do
  add_candidate "${python_cache#"$repo_root/"}"
done < <(find "$repo_root/v12" -type d \
  \( -name __pycache__ -o -name .pytest_cache -o -name .mypy_cache -o -name .ruff_cache \) \
  -prune -print0)

# A few older profiling commands wrote their test-only artifacts directly into
# the Go workspace rather than v12/tmp or .profiles.
while IFS= read -r -d '' artifact; do
  add_candidate "${artifact#"$repo_root/"}"
done < <(find "$repo_root/v12/interpreters/go" -maxdepth 1 -type f \
  \( -name '*.test' -o -name '.profile_*' -o -name '.tmp_*.pprof' \) -print0)

if "$include_profiles"; then
  # CPU/heap captures live beside the v12 Go interpreter. Keep them by
  # default because the decision records reference them, but let the archive
  # cleanup remove every project-local profile directory when requested.
  for rel in .profiles v12/.profiles v12/interpreters/go/.profiles; do
    add_candidate "$rel"
  done
fi

if ((${#candidates[@]} == 0)); then
  printf 'cleanup: no generated project artifacts found\n'
  exit 0
fi

total_kib=0
for rel in "${candidates[@]}"; do
  # A directory pathspec includes all descendants. Do not silently delete a
  # future tracked artifact just because the current generated directories are
  # ignored.
  if git -C "$repo_root" ls-files -- "$rel" "$rel/**" | grep -q .; then
    printf 'cleanup: refusing to remove tracked path: %s\n' "$rel" >&2
    exit 2
  fi
  size_kib="$(du -sk -- "$repo_root/$rel" | awk '{print $1}')"
  total_kib=$((total_kib + size_kib))
  if "$apply"; then
    printf 'remove %s (%s)\n' "$rel" "$(du -sh -- "$repo_root/$rel" | awk '{print $1}')"
    rm -rf -- "$repo_root/$rel"
  else
    printf 'would remove %s (%s)\n' "$rel" "$(du -sh -- "$repo_root/$rel" | awk '{print $1}')"
  fi
done

printf 'cleanup: %s reclaimable across %d generated paths\n' \
  "$(awk -v kib="$total_kib" 'BEGIN { printf "%.2f GiB", kib / 1024 / 1024 }')" \
  "${#candidates[@]}"
if ! "$apply"; then
  printf 'cleanup: dry run only; use just cleanup-apply to remove these paths\n'
fi
