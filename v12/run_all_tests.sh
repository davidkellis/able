#!/usr/bin/env bash

set -euo pipefail
export PYTHONDONTWRITEBYTECODE=1

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

echo ">>> Checking fixture target exclusions"
node "$ROOT_DIR/scripts/check-fixture-target-exclusions.mjs"
node --test "$ROOT_DIR/scripts/check-fixture-target-exclusions.test.mjs"

echo ">>> Checking external application scoreboard"
"$ROOT_DIR/bench_external_scoreboard" --check
python3 "$ROOT_DIR/bench_scorecard_evidence_check.py" \
  --scorecard "$ROOT_DIR/docs/perf-baselines/external-scoreboard-current.json" \
  --selection-manifest "$ROOT_DIR/bench-selection-manifest.json" \
  --require-runs 5
python3 "$ROOT_DIR/bench_scorecard_evidence_check_test.py"

echo ">>> Checking performance frontier and architecture evidence"
python3 "$ROOT_DIR/bench_performance_frontier_test.py"
python3 "$ROOT_DIR/bench_performance_frontier.py" --check
python3 "$ROOT_DIR/bench_performance_evidence_ledger_test.py"
"$ROOT_DIR/bench_performance_evidence_ledger" --check
python3 "$ROOT_DIR/bench_compiled_static_boundary_census_test.py"
python3 "$ROOT_DIR/bench_residual_cost_model_test.py"
python3 "$ROOT_DIR/bench_compiled_architecture_budget_test.py"
python3 "$ROOT_DIR/bench_bytecode_architecture_budget_test.py"
python3 "$ROOT_DIR/bench_bytecode_semantic_region_feasibility_test.py"
python3 "$ROOT_DIR/bench_bytecode_native_hot_tier_budget_test.py"
python3 "$ROOT_DIR/bench_cross_engine_architecture_budget_test.py"
python3 "$ROOT_DIR/bench_cross_engine_structural_strategy_test.py"
python3 "$ROOT_DIR/bench_portable_vm_backend_adr_test.py"
python3 "$ROOT_DIR/bench_shared_runtime_semantic_abi_test.py"
python3 "$ROOT_DIR/bench_shared_runtime_closed_region_cutover_test.py"
python3 "$ROOT_DIR/bench_architecture_evidence_chain_test.py"
"$ROOT_DIR/bench_bytecode_semantic_region_feasibility" --check
"$ROOT_DIR/bench_cross_engine_architecture_budget" --check
"$ROOT_DIR/bench_architecture_evidence_chain" --check

echo ">>> Checking feature-to-application coverage"
"$ROOT_DIR/bench_feature_coverage_check"
python3 "$ROOT_DIR/bench_feature_coverage_check_test.py"
python3 "$ROOT_DIR/bench_sustained_workload_depth_test.py"

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
COMPILER_CORE_OUTLIER_TESTS_EGREP='^TestCompiler.*ParityFixtures(Batch[0-9]+)?$'
COMPILER_SHORT_OUTLIER_TESTS_EGREP='^TestCompilerCanonicalStdlibExpectationResultArgumentStaysConcrete$'

echo ">>> Running Go tests"
(
  gocache="$ROOT_DIR/interpreters/go/.gocache"
  if [[ "${ABLE_GOCACHE:-}" == "tmp" ]]; then
    gocache="$(mktemp -d)"
    trap 'rm -rf "$gocache"' EXIT
  elif [[ -n "${GOCACHE:-}" ]]; then
    gocache="$GOCACHE"
  fi

  echo ">>> Running standalone parser Go binding test"
  (
    cd "$ROOT_DIR/parser/tree-sitter-able"
    GOCACHE="$gocache" go test -timeout 1m -count=1 ./bindings/go
  )

  cd "$ROOT_DIR/interpreters/go"
  mapfile -t all_pkgs < <(go list ./... | grep -Ev '^able/interpreter-go/tmp(/|$)')
  if [[ ${#all_pkgs[@]} -eq 0 ]]; then
    echo "No Go packages found to test." >&2
    exit 1
  fi

  run_compiler_short_batches() {
    local batch_size="$1"
    local -a listed_tests=()
    local -a compiler_tests=()
    local -a outlier_tests=()
    local -a batch_tests=()
    local total=0
    local batch_count=0
    local batch_index=0
    local start=0
    local regex=""
    local name=""
    local list_pattern="${FILTER:-^Test}"

    mapfile -t listed_tests < <(
      env \
        ABLE_TYPECHECK_FIXTURES="$TYPECHECK_FIXTURES_MODE" \
        GOCACHE="$gocache" \
        ABLE_COMPILER_EXEC_GOCACHE="$gocache" \
        go test -short ./pkg/compiler -list "$list_pattern" |
        grep '^Test'
    )
    for name in "${listed_tests[@]}"; do
      if [[ "$name" =~ $COMPILER_SHORT_OUTLIER_TESTS_EGREP ]]; then
        outlier_tests+=("$name")
      else
        compiler_tests+=("$name")
      fi
    done
    total=${#compiler_tests[@]}
    if [[ "$total" -eq 0 && ${#outlier_tests[@]} -eq 0 ]]; then
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

    for name in "${outlier_tests[@]}"; do
      echo ">>> Running compiler short-mode outlier ${name}"
      env \
        ABLE_TYPECHECK_FIXTURES="$TYPECHECK_FIXTURES_MODE" \
        GOCACHE="$gocache" \
        ABLE_COMPILER_EXEC_GOCACHE="$gocache" \
        go test -short -timeout "$GO_TEST_TIMEOUT" ./pkg/compiler \
          -run "^${name}$" -count=1
    done
  }

  run_exec_fixture_batches() {
    local label="$1"
    local test_name="$2"
    local exec_mode="$3"
    local batch_count="$4"
    local i=0

    for ((i=0; i<batch_count; i++)); do
      echo ">>> Running ${label} batch $((i + 1))/${batch_count}"
      env \
        ABLE_TYPECHECK_FIXTURES="$TYPECHECK_FIXTURES_MODE" \
        GOCACHE="$gocache" \
        ABLE_COMPILER_EXEC_GOCACHE="$gocache" \
        ABLE_EXEC_FIXTURE_BATCH_INDEX="$i" \
        ABLE_EXEC_FIXTURE_BATCH_COUNT="$batch_count" \
        go test -timeout "$GO_TEST_TIMEOUT" ./pkg/interpreter \
          -run "^${test_name}$" -count=1 -exec-mode="$exec_mode"
    done
  }

  if [[ "$RUN_ALL" == true ]]; then
    mapfile -t fast_pkgs < <(
      printf '%s\n' "${all_pkgs[@]}" |
        grep -Ev '^able/interpreter-go/pkg/(compiler|interpreter|parser)$'
    )

    echo ">>> Running parser package with fixture corpus (full mode)"
    ABLE_TYPECHECK_FIXTURES="$TYPECHECK_FIXTURES_MODE" \
      GOCACHE="$gocache" \
      ABLE_COMPILER_EXEC_GOCACHE="$gocache" \
      go test -timeout 1m ./pkg/parser -count=1

    echo ">>> Running remaining non-compiler packages (short mode)"
    ABLE_TYPECHECK_FIXTURES="$TYPECHECK_FIXTURES_MODE" \
      GOCACHE="$gocache" \
      ABLE_COMPILER_EXEC_GOCACHE="$gocache" \
      go test -short -timeout "$GO_TEST_TIMEOUT" "${RUN_FLAG[@]}" "${fast_pkgs[@]}"

    if [[ -z "$FILTER" ]]; then
      echo ">>> Running interpreter package without aggregate fixture tables"
      ABLE_TYPECHECK_FIXTURES="$TYPECHECK_FIXTURES_MODE" \
        GOCACHE="$gocache" \
        ABLE_COMPILER_EXEC_GOCACHE="$gocache" \
        go test -short -timeout "$GO_TEST_TIMEOUT" ./pkg/interpreter \
          -skip '^(TestExecFixtures|TestExecFixtureParity)$' -count=1

      run_exec_fixture_batches \
        "treewalker exec fixtures" \
        "TestExecFixtures" \
        "treewalker" \
        "8"
      run_exec_fixture_batches \
        "treewalker/bytecode exec parity" \
        "TestExecFixtureParity" \
        "treewalker" \
        "8"
    else
      echo ">>> Running filtered interpreter package"
      ABLE_TYPECHECK_FIXTURES="$TYPECHECK_FIXTURES_MODE" \
        GOCACHE="$gocache" \
        ABLE_COMPILER_EXEC_GOCACHE="$gocache" \
        go test -short -timeout "$GO_TEST_TIMEOUT" "${RUN_FLAG[@]}" \
          ./pkg/interpreter -count=1
    fi

    echo ">>> Running compiler package in bounded short-mode batches"
    run_compiler_short_batches "10"

    echo ">>> Running bytecode fixture pass"
    if [[ -z "$FILTER" ]]; then
      run_exec_fixture_batches \
        "bytecode exec fixtures" \
        "TestExecFixtures" \
        "bytecode" \
        "8"
    else
      ABLE_TYPECHECK_FIXTURES="$TYPECHECK_FIXTURES_MODE" \
        GOCACHE="$gocache" \
        ABLE_COMPILER_EXEC_GOCACHE="$gocache" \
        go test -timeout "$GO_TEST_TIMEOUT" "${RUN_FLAG[@]}" \
          ./pkg/interpreter -count=1 -exec-mode=bytecode
    fi
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
          "128" \
          '^TestCompilerExecFixtureFallbacks$'

        echo ">>> Running compiler full compiled fixture matrix"
        run_compiler_batched_release_test \
          "compiler exec fixtures" \
          "ABLE_COMPILER_EXEC_FIXTURES" \
          "ABLE_COMPILER_EXEC_FIXTURE_BATCH_INDEX" \
          "ABLE_COMPILER_EXEC_FIXTURE_BATCH_COUNT" \
          "128" \
          '^TestCompilerExecFixtures$'

        echo ">>> Running compiler strict-dispatch audit"
        run_compiler_batched_release_test \
          "compiler strict-dispatch audit" \
          "ABLE_COMPILER_STRICT_DISPATCH_FIXTURES" \
          "ABLE_COMPILER_STRICT_DISPATCH_BATCH_INDEX" \
          "ABLE_COMPILER_STRICT_DISPATCH_BATCH_COUNT" \
          "128" \
          '^TestCompilerStrictDispatchForStdlibHeavyFixtures$'

        echo ">>> Running compiler interface-lookup audit"
        run_compiler_batched_release_test \
          "compiler interface-lookup audit" \
          "ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES" \
          "ABLE_COMPILER_INTERFACE_LOOKUP_BATCH_INDEX" \
          "ABLE_COMPILER_INTERFACE_LOOKUP_BATCH_COUNT" \
          "128" \
          '^TestCompilerInterfaceLookupBypassForStaticFixtures$'

        echo ">>> Running compiler boundary audit"
        run_compiler_batched_release_test \
          "compiler boundary audit" \
          "ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES" \
          "ABLE_COMPILER_BOUNDARY_AUDIT_BATCH_INDEX" \
          "ABLE_COMPILER_BOUNDARY_AUDIT_BATCH_COUNT" \
          "128" \
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
