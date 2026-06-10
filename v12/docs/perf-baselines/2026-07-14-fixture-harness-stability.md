# Fixture harness stability — 2026-07-14

## Decision

Keep no compiler, bytecode VM, runtime, canonical-stdlib, or benchmark
performance change. This tranche repairs test-harness lifecycle and source-root
selection so semantic fixtures remain trustworthy evidence for later,
cross-application performance work.

## Changes

- The recursive quicksort fixture is now a 200-element semantic-parity case.
  Its former 2,000-element workload moved to interpreter testdata and remains
  the input of `BenchmarkBytecodeQuicksortHotloopRuntime`. This prevents a
  profiling workload from making a correctness subtest exceed its time budget.
- Serial executors created by execution-fixture, parity-fixture, and stdlib
  source-test helpers are closed at the end of each test. This prevents an
  accumulated set of idle executor goroutines from affecting later tests.
- Fixture source discovery selects the first valid stdlib root found from the
  entry context. Working-directory and executable discovery are fallbacks only.
  In this workspace that selects the sibling canonical `able-stdlib` checkout
  for fixtures, while temporary source tests retain their installed-cache root.
  The driver's intentional collision error for genuinely simultaneous stdlib
  roots remains unchanged.
- The blocking-I/O fixture resets its host pipe for each execution, so its
  tree-walker and bytecode parity runs do not reuse a writer closed by the
  preceding execution.

## Verification

All commands used the project-local Go cache:

```text
go test ./pkg/interpreter -run '^(TestFindStdlibRoots.*|TestFixtureStdlibRootSource.*|TestBuildExecSearchPathsKeepsEntryStdlibRoot)$' -count=1 -timeout 60s
go test ./pkg/interpreter -run '^TestExecFixtures$' -count=1 -timeout 60s
go test ./pkg/interpreter -run '^TestExecFixtureParity$' -count=1 -timeout 60s
go test ./pkg/interpreter -run '^TestExecFixtureParity/07_10_bytecode_quicksort_hotloop$' -count=1 -timeout 60s
go test ./pkg/interpreter -run '^$' -bench '^BenchmarkBytecodeQuicksortHotloopRuntime$' -benchtime=1x -count=1 -timeout 90s
```

The source-root tests passed in 0.033s; execution fixtures in 27.804s; parity
fixtures in 51.831s; quicksort parity in 3.151s; and the retained full
benchmark in 7,298,917 ns/op for one iteration. The canonical string source
fast-path regression test also passes after the root-selection change.

No file in the external `able-stdlib` checkout was edited.

## Next recommendation

Return to performance selection only when the verifier-backed application suite
reveals a concrete leaf shared by unlike real programs. The harness changes
make that evidence cleaner; another isolated call-frame or quicksort-specific
micro-optimization would not move the project toward general application
performance.
