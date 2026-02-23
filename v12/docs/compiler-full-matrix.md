# Compiler Full-Matrix Sweeps (v12)

This workflow runs the full compiler fixture matrix intentionally separated from the default fast gates.
It also runs the compiler fallback audit by default (`TestCompilerExecFixtureFallbacks`) to verify no silent fallback paths.

## When To Run
- Nightly CI monitoring.
- Manual pre-release confidence sweeps.
- Investigating regressions that only appear outside reduced default fixture sets.

## Local Commands
- Direct compiler sweep:
  - `./v12/run_compiler_full_matrix.sh`
- Via the standard v12 test runner:
  - `./run_all_tests.sh --version=v12 --compiler-full-matrix`

If you need a faster local iteration and want to skip fallback auditing:
- `./v12/run_compiler_full_matrix.sh --skip-fallback-audit`

## Environment Overrides
These default to `all` when unset:
- `ABLE_COMPILER_EXEC_FIXTURES`
- `ABLE_COMPILER_STRICT_DISPATCH_FIXTURES`
- `ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES`
- `ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES`

This defaults to `1` when unset:
- `ABLE_COMPILER_GLOBAL_LOOKUP_STRICT_TOTAL`

Timeout controls (defaults shown):
- `ABLE_COMPILER_SUITE_TIMEOUT=25m` (passed to each `go test -timeout`)
- `ABLE_COMPILER_SUITE_WALL_TIMEOUT=30m` (hard wall clock per suite via `timeout(1)` when available)

Example narrowed local run:
```bash
ABLE_COMPILER_EXEC_FIXTURES=06_12_26_stdlib_test_harness_reporters \
ABLE_COMPILER_STRICT_DISPATCH_FIXTURES=06_12_26_stdlib_test_harness_reporters \
ABLE_COMPILER_INTERFACE_LOOKUP_FIXTURES=10_17_interface_overload_dispatch,14_01_language_interfaces_index_apply_iterable \
ABLE_COMPILER_BOUNDARY_AUDIT_FIXTURES=06_12_26_stdlib_test_harness_reporters \
./v12/run_compiler_full_matrix.sh --typecheck-fixtures=strict
```

## CI Workflow
- File: `.github/workflows/compiler-full-matrix-nightly.yml`
- Triggers:
  - Daily schedule.
  - `workflow_dispatch` (manual run with inputs).
- Manual dispatch inputs:
  - `typecheck_mode` (`off|warn|strict`)
  - `exec_fixtures`
  - `strict_dispatch_fixtures`
  - `interface_lookup_fixtures`
  - `global_lookup_strict_total`
  - `boundary_audit_fixtures`
  - `suite_timeout`
  - `suite_wall_timeout`

## Runtime Profile (Current Baseline)
Approximate durations on current fixture corpus:
- `TestCompilerExecFixtures` with `...=all`: ~506s
- `TestCompilerStrictDispatchForStdlibHeavyFixtures` with `...=all`: ~533s
- `TestCompilerInterfaceLookupBypassForStaticFixtures` with `...=all` and global strict-total: ~539s
- `TestCompilerBoundaryFallbackMarkerForStaticFixtures` with `...=all`: ~463s

Total sweep runtime is expected to be materially longer than default PR-speed gates.
