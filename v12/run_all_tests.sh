#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

TYPECHECK_FIXTURES_MODE="strict"
EXPORT_FIXTURES=false
RUN_TREEWALKER=false
RUN_BYTECODE=false
RUN_COMPILER=false
RUN_ALL=true
FILTER=""
GO_TEST_TIMEOUT="${GO_TEST_TIMEOUT:-30m}"

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
    --export-fixtures)
      EXPORT_FIXTURES=true
      shift
      ;;
    --treewalker)
      RUN_TREEWALKER=true
      RUN_ALL=false
      shift
      ;;
    --bytecode)
      RUN_BYTECODE=true
      RUN_ALL=false
      shift
      ;;
    --compiler)
      RUN_COMPILER=true
      RUN_ALL=false
      shift
      ;;
    --filter)
      FILTER="$2"
      shift 2
      ;;
    --filter=*)
      FILTER="${1#*=}"
      shift
      ;;
    --help|-h)
      cat <<'EOF'
Usage: run_all_tests.sh [options]

Options:
  --treewalker              Run only treewalker interpreter tests.
  --bytecode                Run only bytecode interpreter tests.
  --compiler                Run only compiler tests (full matrix).
  --filter PATTERN          Pass -run PATTERN to narrow tests within selected subsets.
  --export-fixtures         Run fixture export step (Go-based exporter).
  --typecheck-fixtures[=MODE]  Set fixture typechecking (MODE: off|warn|strict, default strict).
  --typecheck-fixtures-warn    Shorthand for --typecheck-fixtures=warn.
  --typecheck-fixtures-strict  Shorthand for --typecheck-fixtures=strict.
  -h, --help                   Show this help text.

When no subset flags (--treewalker, --bytecode, --compiler) are given, all tests
run: full package suite plus a bytecode fixture pass.
Subset flags are combinable (e.g. --treewalker --compiler).
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

RUN_FLAG=()
if [[ -n "$FILTER" ]]; then
  RUN_FLAG=(-run "$FILTER")
fi

COMPILER_HEAVY_RELEASE_TESTS_EGREP='^(TestCompilerExecFixtures|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerInterfaceLookupBypassForStaticFixtures(Batch[1-4])?|TestCompilerBoundaryFallbackMarkerForStaticFixtures(Batch[0-9]+)?)$'
COMPILER_CORE_OUTLIER_TESTS_EGREP='^(TestCompiler.*ParityFixtures|TestCompilerExecFixtureFallbacks)$'

echo ">>> Running Go tests"
(
  cd "$ROOT_DIR/interpreters/go"
  mapfile -t all_pkgs < <(go list ./... | grep -Ev '^able/interpreter-go/tmp(/|$)')
  if [[ ${#all_pkgs[@]} -eq 0 ]]; then
    echo "No Go packages found to test." >&2
    exit 1
  fi
  gocache="$ROOT_DIR/interpreters/go/.gocache"
  if [[ "${ABLE_GOCACHE:-}" == "tmp" ]]; then
    gocache="$(mktemp -d)"
    trap 'rm -rf "$gocache"' EXIT
  elif [[ -n "${GOCACHE:-}" ]]; then
    gocache="$GOCACHE"
  fi

  if [[ "$RUN_ALL" == true ]]; then
    echo ">>> Running all packages"
    ABLE_TYPECHECK_FIXTURES="$TYPECHECK_FIXTURES_MODE" \
      GOCACHE="$gocache" \
      ABLE_COMPILER_EXEC_GOCACHE="$gocache" \
      go test -timeout "$GO_TEST_TIMEOUT" "${RUN_FLAG[@]}" "${all_pkgs[@]}"

    echo ">>> Running bytecode fixture pass"
    ABLE_TYPECHECK_FIXTURES="$TYPECHECK_FIXTURES_MODE" \
      GOCACHE="$gocache" \
      ABLE_COMPILER_EXEC_GOCACHE="$gocache" \
      go test -timeout "$GO_TEST_TIMEOUT" "${RUN_FLAG[@]}" ./pkg/interpreter -count=1 -exec-mode=bytecode
  else
    run_go_test_base() {
      ABLE_TYPECHECK_FIXTURES="$TYPECHECK_FIXTURES_MODE" \
        GOCACHE="$gocache" \
        ABLE_COMPILER_EXEC_GOCACHE="$gocache" \
        "$@"
    }

    run_compiler_batched_release_test() {
      local label="$1"
      local fixture_env="$2"
      local batch_index_env="$3"
      local batch_count_env="$4"
      local batch_count="$5"
      local pattern="$6"
      local i
      for ((i=0; i<batch_count; i++)); do
        echo ">>> Running ${label} batch $((i + 1))/${batch_count}"
        env \
          ABLE_TYPECHECK_FIXTURES="$TYPECHECK_FIXTURES_MODE" \
          GOCACHE="$gocache" \
          ABLE_COMPILER_EXEC_GOCACHE="$gocache" \
          "${fixture_env}=all" \
          "${batch_index_env}=${i}" \
          "${batch_count_env}=${batch_count}" \
          go test -timeout "$GO_TEST_TIMEOUT" ./pkg/compiler -run "$pattern" -count=1
      done
    }

    run_compiler_core_batches() {
      local batch_size="$1"
      local -a compiler_tests=()
      local -a batch_tests=()
      local total=0
      local batch_count=0
      local batch_index=0
      local start=0
      local regex=""
      local name=""

      mapfile -t compiler_tests < <(
        env \
          ABLE_TYPECHECK_FIXTURES="$TYPECHECK_FIXTURES_MODE" \
          GOCACHE="$gocache" \
          ABLE_COMPILER_EXEC_GOCACHE="$gocache" \
          go test ./pkg/compiler -list '^Test' |
          grep '^Test' |
          grep -Ev "$COMPILER_HEAVY_RELEASE_TESTS_EGREP" |
          grep -Ev "$COMPILER_CORE_OUTLIER_TESTS_EGREP"
      )
      total=${#compiler_tests[@]}
      if [[ "$total" -eq 0 ]]; then
        echo "No compiler core tests found." >&2
        exit 1
      fi
      batch_count=$(((total + batch_size - 1) / batch_size))

      for ((start=0; start<total; start+=batch_size)); do
        batch_tests=("${compiler_tests[@]:start:batch_size}")
        regex='^('
        for name in "${batch_tests[@]}"; do
          if [[ "$regex" != '^(' ]]; then
            regex+='|'
          fi
          regex+="${name}"
        done
        regex+=')$'
        batch_index=$((start / batch_size))
        echo ">>> Running compiler core batch $((batch_index + 1))/${batch_count}"
        run_go_test_base go test -timeout "$GO_TEST_TIMEOUT" ./pkg/compiler -run "$regex" -count=1
      done
    }

    run_compiler_outlier_tests() {
      local -a outlier_tests=()
      local name=""

      mapfile -t outlier_tests < <(
        env \
          ABLE_TYPECHECK_FIXTURES="$TYPECHECK_FIXTURES_MODE" \
          GOCACHE="$gocache" \
          ABLE_COMPILER_EXEC_GOCACHE="$gocache" \
          go test ./pkg/compiler -list '^Test' |
          grep '^Test' |
          grep -E "$COMPILER_CORE_OUTLIER_TESTS_EGREP"
      )
      for name in "${outlier_tests[@]}"; do
        echo ">>> Running compiler outlier test ${name}"
        run_go_test_base go test -timeout "$GO_TEST_TIMEOUT" ./pkg/compiler -run "^(${name})$" -count=1
      done
    }

    if [[ "$RUN_TREEWALKER" == true ]]; then
      echo ">>> Running treewalker interpreter tests"
      run_go_test_base go test -timeout "$GO_TEST_TIMEOUT" "${RUN_FLAG[@]}" ./pkg/interpreter -count=1
    fi
    if [[ "$RUN_BYTECODE" == true ]]; then
      echo ">>> Running bytecode interpreter tests"
      run_go_test_base go test -timeout "$GO_TEST_TIMEOUT" "${RUN_FLAG[@]}" ./pkg/interpreter -count=1 -exec-mode=bytecode
    fi
    if [[ "$RUN_COMPILER" == true ]]; then
      if [[ -n "$FILTER" ]]; then
        echo ">>> Running compiler tests (filtered)"
        run_go_test_base go test -timeout "$GO_TEST_TIMEOUT" "${RUN_FLAG[@]}" ./pkg/compiler ./pkg/compiler/bridge -count=1
      else
        echo ">>> Running compiler bridge tests"
        run_go_test_base go test -timeout "$GO_TEST_TIMEOUT" ./pkg/compiler/bridge -count=1

        echo ">>> Running compiler core test batches"
        run_compiler_core_batches "25"

        echo ">>> Running compiler core outlier tests"
        run_compiler_outlier_tests

        echo ">>> Running compiler full compiled fixture matrix"
        run_compiler_batched_release_test \
          "compiler exec fixtures" \
          "ABLE_COMPILER_EXEC_FIXTURES" \
          "ABLE_COMPILER_EXEC_FIXTURE_BATCH_INDEX" \
          "ABLE_COMPILER_EXEC_FIXTURE_BATCH_COUNT" \
          "24" \
          '^TestCompilerExecFixtures$'

        echo ">>> Running compiler strict-dispatch audit"
        run_compiler_batched_release_test \
          "compiler strict-dispatch audit" \
          "ABLE_COMPILER_STRICT_DISPATCH_FIXTURES" \
          "ABLE_COMPILER_STRICT_DISPATCH_BATCH_INDEX" \
          "ABLE_COMPILER_STRICT_DISPATCH_BATCH_COUNT" \
          "24" \
          '^TestCompilerStrictDispatchForStdlibHeavyFixtures$'

        echo ">>> Running compiler interface-lookup audit"
        run_compiler_batched_release_test \
          "compiler interface-lookup audit" \
          "ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES" \
          "ABLE_COMPILER_INTERFACE_LOOKUP_BATCH_INDEX" \
          "ABLE_COMPILER_INTERFACE_LOOKUP_BATCH_COUNT" \
          "24" \
          '^TestCompilerInterfaceLookupBypassForStaticFixtures$'

        echo ">>> Running compiler boundary audit"
        run_compiler_batched_release_test \
          "compiler boundary audit" \
          "ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES" \
          "ABLE_COMPILER_BOUNDARY_AUDIT_BATCH_INDEX" \
          "ABLE_COMPILER_BOUNDARY_AUDIT_BATCH_COUNT" \
          "24" \
          '^TestCompilerBoundaryFallbackMarkerForStaticFixtures$'
      fi
    fi
  fi
)

echo "All tests completed successfully."
