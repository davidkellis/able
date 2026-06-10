#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

TYPECHECK_FIXTURES_MODE="strict"
EXPORT_FIXTURES=false
RUN_TREEWALKER=false
RUN_BYTECODE=false
RUN_COMPILER=false
RUN_COMPILED_CLI=false
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
    --compiled-cli)
      RUN_COMPILED_CLI=true
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
  --compiled-cli            Run generated-Go CLI integration tests (release lane).
  --filter PATTERN          Pass -run PATTERN to narrow tests within selected subsets.
  --export-fixtures         Run fixture export step (Go-based exporter).
  --typecheck-fixtures[=MODE]  Set fixture typechecking (MODE: off|warn|strict, default strict).
  --typecheck-fixtures-warn    Shorthand for --typecheck-fixtures=warn.
  --typecheck-fixtures-strict  Shorthand for --typecheck-fixtures=strict.
  -h, --help                   Show this help text.

When no subset flags (--treewalker, --bytecode, --compiler, --compiled-cli) are
given, the fast suite runs: all packages in Go short mode plus a complete
bytecode fixture pass. Use `--compiler` for the separately batched release
compiler audits.
`--compiled-cli` is an explicit release lane for CLI tests that compile and run
generated Go. Subset flags are combinable (e.g. --treewalker --compiler).
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

echo ">>> Checking external application scoreboard"
"$ROOT_DIR/bench_external_scoreboard" --check
python3 "$ROOT_DIR/bench_scorecard_evidence_check.py" \
  --scorecard "$ROOT_DIR/docs/perf-baselines/external-scoreboard-current.json" \
  --selection-manifest "$ROOT_DIR/bench-selection-manifest.json" \
  --require-runs 5
python3 "$ROOT_DIR/bench_scorecard_evidence_check_test.py"

echo ">>> Checking feature-to-application coverage"
"$ROOT_DIR/bench_feature_coverage_check"
python3 "$ROOT_DIR/bench_feature_coverage_check_test.py"

echo ">>> Checking external scorecard selection contract"
"$ROOT_DIR/bench_selection_manifest_check"
python3 "$ROOT_DIR/bench_scorecard_selection_test.py"
python3 "$ROOT_DIR/bench_execution_contract_test.py"
python3 "$ROOT_DIR/bench_refresh_external_scorecard_test.py"

echo ">>> Checking external threshold controls"
"$ROOT_DIR/bench_external_threshold_controls" --check

echo ">>> Checking generated-artifact cleanup policy"
bash "$ROOT_DIR/../scripts/cleanup_test.sh"

echo ">>> Checking canonical and embedded kernel synchronization"
bash "$ROOT_DIR/../scripts/check-embedded-kernel.sh"

RUN_FLAG=()
if [[ -n "$FILTER" ]]; then
  RUN_FLAG=(-run "$FILTER")
fi

COMPILER_HEAVY_RELEASE_TESTS_EGREP='^(TestCompilerExecFixtures|TestCompilerExecFixtureFallbacks|TestCompilerStrictDispatchForStdlibHeavyFixtures|TestCompilerInterfaceLookupBypassForStaticFixtures(Batch[1-4])?|TestCompilerBoundaryFallbackMarkerForStaticFixtures(Batch[0-9]+)?)$'
COMPILER_CORE_OUTLIER_TESTS_EGREP='^TestCompiler.*ParityFixtures$'

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

  run_compiler_short_batches() {
    local batch_size="$1"
    local -a compiler_tests=()
    local -a batch_tests=()
    local total=0
    local batch_count=0
    local batch_index=0
    local start=0
    local regex=""
    local name=""
    local list_pattern="${FILTER:-^Test}"

    mapfile -t compiler_tests < <(
      env \
        ABLE_TYPECHECK_FIXTURES="$TYPECHECK_FIXTURES_MODE" \
        GOCACHE="$gocache" \
        ABLE_COMPILER_EXEC_GOCACHE="$gocache" \
        go test -short ./pkg/compiler -list "$list_pattern" |
        grep '^Test'
    )
    total=${#compiler_tests[@]}
    if [[ "$total" -eq 0 ]]; then
      echo "No compiler tests found." >&2
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
      echo ">>> Running compiler short-mode batch $((batch_index + 1))/${batch_count}"
      env \
        ABLE_TYPECHECK_FIXTURES="$TYPECHECK_FIXTURES_MODE" \
        GOCACHE="$gocache" \
        ABLE_COMPILER_EXEC_GOCACHE="$gocache" \
        go test -short -timeout "$GO_TEST_TIMEOUT" ./pkg/compiler -run "$regex" -count=1
    done
  }

  if [[ "$RUN_ALL" == true ]]; then
    mapfile -t fast_pkgs < <(printf '%s\n' "${all_pkgs[@]}" | grep -v '^able/interpreter-go/pkg/compiler$')
    echo ">>> Running non-compiler packages (short mode)"
    ABLE_TYPECHECK_FIXTURES="$TYPECHECK_FIXTURES_MODE" \
      GOCACHE="$gocache" \
      ABLE_COMPILER_EXEC_GOCACHE="$gocache" \
      go test -short -timeout "$GO_TEST_TIMEOUT" "${RUN_FLAG[@]}" "${fast_pkgs[@]}"

    echo ">>> Running compiler package in bounded short-mode batches"
    run_compiler_short_batches "25"

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

        echo ">>> Running compiler fallback audit"
        run_compiler_batched_release_test \
          "compiler fallback audit" \
          "ABLE_COMPILER_EXEC_FIXTURES" \
          "ABLE_COMPILER_EXEC_FIXTURE_BATCH_INDEX" \
          "ABLE_COMPILER_EXEC_FIXTURE_BATCH_COUNT" \
          "24" \
          '^TestCompilerExecFixtureFallbacks$'

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
    if [[ "$RUN_COMPILED_CLI" == true ]]; then
      echo ">>> Running generated-Go CLI integration tests"
      env \
        ABLE_TYPECHECK_FIXTURES="$TYPECHECK_FIXTURES_MODE" \
        GOCACHE="$gocache" \
        ABLE_COMPILER_EXEC_GOCACHE="$gocache" \
        ABLE_RUN_COMPILED_CLI_INTEGRATION=1 \
        go test -timeout "$GO_TEST_TIMEOUT" "${RUN_FLAG[@]}" ./cmd/able -count=1
    fi
  fi
)

echo "All tests completed successfully."
