# Able v12

go_dir := "v12/interpreters/go"
cmd_able_dir := go_dir / "cmd/able"

# Copy kernel source into the embedded/ directory for go:embed
embed:
    mkdir -p {{cmd_able_dir}}/embedded/kernel/src
    cp v12/kernel/package.yml {{cmd_able_dir}}/embedded/kernel/
    cp v12/kernel/src/kernel.able {{cmd_able_dir}}/embedded/kernel/src/

# Build the able binary
build: embed
    cd {{go_dir}} && go build -o able ./cmd/able
    cp {{go_dir}}/able ./able

# Build with optimizations stripped for smaller binary
build-small: embed
    cd {{go_dir}} && go build -ldflags="-s -w" -o able ./cmd/able
    cp {{go_dir}}/able ./able

# Run all tests
test: embed
    cd {{go_dir}} && go test ./pkg/runtime/... ./pkg/interpreter/... ./pkg/compiler/...

# Run CLI tests
test-cli: embed
    cd {{go_dir}} && go test ./cmd/able/...

# Clean build artifacts
clean:
    rm -f able {{go_dir}}/able v12/wasm/ablewasm.wasm
    rm -rf {{cmd_able_dir}}/embedded/

# Build the pre-parsed-AST JS/WASM prototype. The output is ignored and is
# removable with just clean or just cleanup-apply.
wasm-build:
    cd {{go_dir}} && GOOS=js GOARCH=wasm go build -o ../../wasm/ablewasm.wasm ./cmd/ablewasm

# Link and execute the prototype through Node without requiring tree-sitter or
# npm dependencies; the checked-in sample is already fixture-style AST JSON.
wasm-smoke: wasm-build
    cd v12/wasm && node module_loader.test.mjs
    cd v12/wasm && node source_request.test.mjs
    cd v12/wasm && node run_prototype.mjs --module-json ./samples/addition.ast.json --wasm ./ablewasm.wasm --exec-mode treewalker
    cd v12/wasm && node run_prototype.mjs --module-json ./samples/addition.ast.json --wasm ./ablewasm.wasm --exec-mode bytecode
    cd v12/wasm && node run_prototype.mjs --module-json ./samples/host-output.ast.json --wasm ./ablewasm.wasm --exec-mode treewalker --expect-host-stdout 'wasm host output\n' --expect-host-stderr ''
    cd v12/wasm && node run_prototype.mjs --module-json ./samples/host-output.ast.json --wasm ./ablewasm.wasm --exec-mode bytecode --expect-host-stdout 'wasm host output\n' --expect-host-stderr ''
    cd v12/wasm && node run_prototype.mjs --module-json ./samples/host-error.ast.json --wasm ./ablewasm.wasm --exec-mode treewalker --expect-host-stdout '' --expect-host-stderr 'decode module json: decode Module: invalid argument\n' --expect-response-ok false
    cd v12/wasm && node run_prototype.mjs --module-json ./samples/extern-go-unavailable.ast.json --wasm ./ablewasm.wasm --exec-mode treewalker --expect-host-stdout '' --expect-host-stderr 'evaluate module: extern function host_value is unavailable on js/wasm; browser host callbacks are not implemented\n' --expect-response-ok false
    cd v12/wasm && node run_prototype.mjs --module-json ./samples/extern-go-unavailable.ast.json --wasm ./ablewasm.wasm --exec-mode bytecode --expect-host-stdout '' --expect-host-stderr 'evaluate module: extern function host_value is unavailable on js/wasm; browser host callbacks are not implemented\n' --expect-response-ok false

# Exercise browser source parsing plus a dependency-first static import closure.
# Requires `cd v12/wasm && npm ci` because the parser dependency is optional.
wasm-source-module-smoke: wasm-build
    cd v12/wasm && node run_prototype.mjs --source ./samples/modules/main.able --module-root ./samples/modules --wasm ./ablewasm.wasm --exec-mode treewalker --expect-host-stdout '42\n' --expect-host-stderr ''
    cd v12/wasm && node run_prototype.mjs --source ./samples/modules/main.able --module-root ./samples/modules --wasm ./ablewasm.wasm --exec-mode bytecode --expect-host-stdout '42\n' --expect-host-stderr ''

# Runs actual Able source through a real headless Firefox + Go/WASM runtime.
# Requires Firefox, geckodriver, and `cd v12/wasm && npm ci` for tree-sitter.
wasm-browser-smoke: wasm-build
    cd v12/wasm && node browser_smoke.test.mjs

# Preview removable v12-generated artifacts without changing the workspace.
cleanup:
    ./scripts/cleanup.sh

# Remove generated v12 caches, benchmark workspaces, fixture targets, and binaries.
cleanup-apply:
    ./scripts/cleanup.sh --apply

# Preview removal of generated artifacts and archived local profiling evidence.
cleanup-archives-preview:
    ./scripts/cleanup.sh --include-profiles

# Also remove archived local profiling artifacts; they are not needed to build or test.
cleanup-archives:
    ./scripts/cleanup.sh --apply --include-profiles

# Refresh pinned, verifier-backed sibling Go reference measurements.
bench-go-reference *args:
    ./v12/bench_refresh_go_refs {{args}}

# Refresh pinned, verifier-backed sibling Python/Ruby reference measurements.
bench-interpreter-reference *args:
    ./v12/bench_refresh_interpreter_refs {{args}}

# Verify that one CPU is quiet enough before a pinned performance measurement.
bench-host-check *args:
    ./v12/bench_host_cpu_check {{args}}

# Statically validate every active portable benchmark/reference lane and local fixture.
bench-catalog-check *args:
    ./v12/bench_catalog_check {{args}}
    python3 ./v12/bench_feature_coverage_check_test.py
    python3 ./v12/bench_feature_interaction_matrix_test.py
    python3 ./v12/bench_feature_interaction_triples_test.py
    python3 ./v12/bench_operation_depth_check_test.py

# Validate the reviewed mode-aware scorecard selection and its fast protocol tests.
bench-selection-check:
    ./v12/bench_selection_manifest_check
    python3 ./v12/bench_execution_contract_test.py
    python3 ./v12/bench_preserved_compiled_report_test.py
    python3 ./v12/bench_scorecard_selection_test.py
    python3 ./v12/bench_refresh_external_scorecard_test.py

# Refresh fresh references and all compiled/bytecode rows in bounded groups,
# defaulting to five independent verifier-backed samples per row before writing
# one dated aggregate scorecard; pass --no-promote to retain the current report.
bench-scorecard-refresh *args:
    ./v12/bench_refresh_external_scorecard {{args}}

# Emit debug-only compiled boundary counters across verifier-backed applications.
bench-compiled-boundary-audit *args:
    ./v12/bench_compiled_boundary_audit {{args}}

# Emit debug-only residual generated call-path counters across verifier-backed applications.
bench-compiled-call-path-audit *args:
    ./v12/bench_compiled_boundary_audit --telemetry call-path {{args}}

# Build all selected applications first, then time the unchanged binaries in
# forward/reverse verifier-backed cohorts with an explicit variance rejection.
bench-preserved-compiled *args:
    ./v12/bench_compare_preserved_compiled {{args}}

# Rebuild the checked-in report from existing verifier-backed application runs.
bench-scoreboard *args:
    ./v12/bench_external_scoreboard --write-current {{args}}

# Fast CI-safe check: verifies the report is synchronized; never executes benchmarks.
bench-scoreboard-check:
    ./v12/bench_external_scoreboard --check
    python3 ./v12/bench_scorecard_evidence_check.py \
        --scorecard ./v12/docs/perf-baselines/external-scoreboard-current.json \
        --selection-manifest ./v12/bench-selection-manifest.json \
        --require-runs 5
    python3 ./v12/bench_scorecard_evidence_check_test.py
    python3 ./v12/bench_composite_interface_contract_reconciliation_test.py
    ./v12/bench_composite_interface_contract_reconciliation --check
    ./v12/bench_performance_evidence_ledger --check

# Require complete repeated Able/reference evidence for every reviewed row.
bench-scorecard-evidence-check *args:
    python3 ./v12/bench_scorecard_evidence_check.py \
        --scorecard ./v12/docs/perf-baselines/external-scoreboard-current.json \
        --selection-manifest ./v12/bench-selection-manifest.json \
        --require-runs 5 {{args}}

# Rebuild the complete selected-row ownership/disposition ledger without timing.
bench-frontier *args:
    python3 ./v12/bench_performance_frontier.py {{args}}

# Fast CI-safe check that the frontier still matches its pinned inputs.
bench-frontier-check:
    python3 ./v12/bench_performance_frontier_test.py
    python3 ./v12/bench_performance_frontier.py --check

# Reconcile the current frontier with optimistic per-engine architecture bounds.
# This is deterministic and never executes benchmark workloads.
bench-architecture-budget-check:
    cd v12/interpreters/go && go test ./internal/semanticabi/... -count=1 -timeout 60s
    cd v12/interpreters/go && go run ./internal/semanticabi/cmd/manifestgen -runtime pkg/runtime/values.go -out internal/semanticabi/manifest_generated.go -check
    cd v12/interpreters/go && go run ./internal/semanticabi/cmd/flowreport -project-root ../../.. -out ../../docs/perf-baselines/2026-07-22-semantic-abi-shadow-image-lowering.json -check
    cd v12/interpreters/go && go run ./internal/semanticabi/cmd/heapreport -out ../../docs/perf-baselines/2026-07-22-shared-value-heap-conformance-contract.json -check
    cd v12/interpreters/go && go run ./internal/semanticabi/cmd/bindingreport -out ../../docs/perf-baselines/2026-07-22-shared-value-heap-contract-reconciliation.json -check
    python3 ./v12/bench_runtime_contract_reconciliation_test.py
    ./v12/bench_runtime_contract_reconciliation --check
    python3 ./v12/bench_shared_value_heap_production_pilot_test.py
    ./v12/bench_shared_value_heap_production_pilot --check
    python3 ./v12/bench_shared_runtime_closed_region_cutover_test.py
    ./v12/bench_shared_runtime_closed_region_cutover --check
    python3 ./v12/bench_performance_evidence_ledger_test.py
    ./v12/bench_performance_evidence_ledger --check
    python3 ./v12/bench_bytecode_architecture_budget_test.py
    python3 ./v12/bench_compiled_architecture_budget_test.py
    python3 ./v12/bench_residual_cost_model_test.py
    python3 ./v12/bench_semantic_work_amplification_test.py
    python3 ./v12/bench_cross_engine_architecture_budget_test.py
    python3 ./v12/bench_cross_engine_structural_strategy_test.py
    python3 ./v12/bench_portable_vm_backend_adr_test.py
    python3 ./v12/bench_shared_runtime_semantic_abi_test.py
    python3 ./v12/bench_bytecode_semantic_region_feasibility_test.py
    python3 ./v12/bench_bytecode_native_hot_tier_budget_test.py
    python3 ./v12/bench_architecture_evidence_chain_test.py
    ./v12/bench_cross_engine_architecture_budget --check
    ./v12/bench_bytecode_semantic_region_feasibility --check
    ./v12/bench_architecture_evidence_chain --check

# Check or safely refresh the topologically ordered architecture evidence.
bench-architecture-evidence-chain-check:
    python3 ./v12/bench_architecture_evidence_chain_test.py
    ./v12/bench_architecture_evidence_chain --check

bench-architecture-evidence-chain-refresh:
    ./v12/bench_architecture_evidence_chain --refresh

# Compare the remaining cross-engine structural routes without running benchmarks.
bench-structural-strategy-check:
    python3 ./v12/bench_cross_engine_structural_strategy_test.py
    ./v12/bench_cross_engine_structural_strategy --check

# Check the foreign backend ownership/ABI decision without compiling native code.
bench-portable-vm-backend-adr-check:
    python3 ./v12/bench_portable_vm_backend_adr_test.py
    ./v12/bench_portable_vm_backend_adr --check

# Check the backend-neutral program/value/effect ABI feasibility decision.
bench-semantic-abi-feasibility-check:
    python3 ./v12/bench_shared_runtime_semantic_abi_test.py
    ./v12/bench_shared_runtime_semantic_abi --check

# Decide whether a shared semantic runtime has a bounded, adapter-free
# production ownership cut across three unlike hot-function families.
bench-shared-runtime-cutover-check:
    python3 ./v12/bench_shared_runtime_closed_region_cutover_test.py
    ./v12/bench_shared_runtime_closed_region_cutover --check

# Validate the non-executing shared semantic ABI cell, codec, manifests, and
# representative whole-function shadow images.
bench-semantic-abi-codec-check:
    cd v12/interpreters/go && go test ./internal/semanticabi/... -count=1 -timeout 60s
    cd v12/interpreters/go && go run ./internal/semanticabi/cmd/manifestgen -runtime pkg/runtime/values.go -out internal/semanticabi/manifest_generated.go -check

# Check execution-complete but non-executing register/CFG shadow images for
# three unlike whole functions.
bench-semantic-abi-flow-check:
    cd v12/interpreters/go && go test ./internal/semanticabi/... -count=1 -timeout 60s
    cd v12/interpreters/go && go run ./internal/semanticabi/cmd/manifestgen -runtime pkg/runtime/values.go -out internal/semanticabi/manifest_generated.go -check
    cd v12/interpreters/go && go run ./internal/semanticabi/cmd/flowreport -project-root ../../.. -out ../../docs/perf-baselines/2026-07-22-semantic-abi-shadow-image-lowering.json -check

# Check the generated shared-value layouts and deterministic identity/lifetime
# model without changing or executing the production runtime.
bench-semantic-abi-heap-contract-check:
    cd v12/interpreters/go && go test ./internal/semanticabi/... -count=1 -timeout 60s
    cd v12/interpreters/go && go run ./internal/semanticabi/cmd/manifestgen -runtime pkg/runtime/values.go -out internal/semanticabi/manifest_generated.go -check
    cd v12/interpreters/go && go run ./internal/semanticabi/cmd/heapreport -out ../../docs/perf-baselines/2026-07-22-shared-value-heap-conformance-contract.json -check

# Check the reconciled test-only binding between current Go runtime graphs and
# the shared heap contract. This does not import it into production execution.
bench-semantic-abi-go-binding-check:
    cd v12/interpreters/go && go test ./internal/semanticabi/... -count=1 -timeout 60s
    cd v12/interpreters/go && go run ./internal/semanticabi/cmd/manifestgen -runtime pkg/runtime/values.go -out internal/semanticabi/manifest_generated.go -check
    cd v12/interpreters/go && go run ./internal/semanticabi/cmd/bindingreport -out ../../docs/perf-baselines/2026-07-22-shared-value-heap-contract-reconciliation.json -check

# Report closures whose evidence was invalidated; this never runs benchmarks.
bench-evidence-ledger *args:
    ./v12/bench_performance_evidence_ledger {{args}}

# Fast selector contract and checked-artifact integrity gate.
bench-evidence-ledger-check:
    python3 ./v12/bench_performance_evidence_ledger_test.py
    ./v12/bench_performance_evidence_ledger --check

# Check the repeated-process review that permits the current shared-runtime
# scope snapshot to replace the pre-contract-reconciliation closure baseline.
bench-runtime-contract-reconciliation-check:
    python3 ./v12/bench_runtime_contract_reconciliation_test.py
    ./v12/bench_runtime_contract_reconciliation --check

# Check the reviewed v12-spec scope rebase after the composite-interface
# Self-pattern contract became canonical.
bench-composite-interface-contract-reconciliation-check:
    python3 ./v12/bench_composite_interface_contract_reconciliation_test.py
    ./v12/bench_composite_interface_contract_reconciliation --check

# Check the repeated call/return cell pilot and prove its rejected live path
# remains fully reverted.
bench-shared-value-production-pilot-check:
    python3 ./v12/bench_shared_value_heap_production_pilot_test.py
    ./v12/bench_shared_value_heap_production_pilot --check

# Aggregate independent verifier-backed comparison runs; this reports spread only.
bench-variance-report *args:
    ./v12/bench_variance_report {{args}}

# Refresh the report-only guard-band classification from retained five-pair controls.
bench-threshold-controls *args:
    ./v12/bench_external_threshold_controls --write-current {{args}}

# Fast CI-safe integrity check; validates evidence classification but never runs timings.
bench-threshold-controls-check:
    ./v12/bench_external_threshold_controls --check
