# Fixed-context full compiler fixture gate — 2026-07-11

## Scope

This tranche completed the remaining semantic gate for the opt-in
`ExperimentalExecutionContext` compiler ABI. Production defaults remain
unchanged. The goal was broad parity, not benchmark-specific acceleration.

The monolithic serial package command reached the `TestCompilerExecFixtures`
parallel-test queue and hit its aggregate 30-minute timeout without reporting
a semantic assertion failure. The fixture harness already supports disjoint
batches, so the complete corpus was instead run in eight serial, bounded
processes:

```sh
for batch in 0 1 2 3 4 5 6 7; do
  ABLE_COMPILER_FIXTURE_EXPERIMENTAL_EXECUTION_CONTEXT=1 \
  ABLE_COMPILER_EXEC_FIXTURE_BATCH_INDEX=$batch \
  ABLE_COMPILER_EXEC_FIXTURE_BATCH_COUNT=8 \
  GOMEMLIMIT=1GiB GOGC=50 \
  go test ./pkg/compiler -run '^TestCompilerExecFixtures$' \
    -count=1 -parallel=1 -timeout 8m
done
```

## Result

All 114 compiler execution fixtures passed on the final source. The eight
batches took 555.875 seconds total, with individual batches between 52.299 and
88.101 seconds. This includes imports, dynamic boundaries, host externs,
stdlib I/O/filesystem/OS behavior, generic and nominal values, concurrency,
package initialization, and language error/control-flow cases.

The generated nested-spawn binary also passed a Go race build:

```sh
GOMEMLIMIT=1GiB GOGC=50 \
go test ./pkg/compiler \
  -run '^TestCompilerExperimentalExecutionContextNestedSpawnExecutes$' \
  -count=1 -parallel=1 -timeout 90s
```

Focused VM raw-scalar, compiler bridge, fixed-context source guard, native
bound-method, dynamic-boundary, and host-extern checks passed as well.

## Shared defects found and repaired

The gate found correctness gaps at general runtime/compiler boundaries. None
is keyed to a benchmark or a named stdlib container.

- Completed VM results and control payloads now materialize VM-only raw scalar
  carriers before leaving a run. The compiler bridge also materializes those
  carriers at its value boundary; VM slots and stacks retain raw execution.
- Generated named-struct conversion now accepts both interpreter storage
  layouts: field maps and positional named slots. This preserves arrays and
  ordinary nominal values created during interpreter bootstrap when compiled
  code later receives them.
- Every generated host-extern return shape now emits the stable compatibility
  wrappers required by the fixed-context ABI. Static extern calls continue to
  use their context-aware package entry rather than falling through a mixed
  dynamic path.
- An implicit final expression takes precedence over the synthetic success
  value for `T | void` return types. This preserves failures returned by a
  forwarded host extern such as `IOError | void`.

The new `TestCompilerGoExternVoidUnionImplicitReturnPreservesFailure` exercises
the final rule in both the default and fixed-context ABI modes.

No `able-stdlib` source changed. The external stdlib was inspected only; the
fixes belong in the generic compiler/VM boundary machinery.

## Next gate

Keep the ABI opt-in and measure a fresh default-versus-candidate compiled
scorecard before considering a default flip. The benchmark harness needs a
small, explicit passthrough for `ablec --experimental-execution-context`, then
the candidate should be compared with the default on the broad generality,
numeric, collection, text, algorithm, and concurrency suites and against the
checked-in Go references. The bytecode side should be measured against the
same generality/external set for Python and Ruby comparison. This verifies that
the semantic compatibility work produces a broad throughput/memory result and
does not simply move cost between workloads.
