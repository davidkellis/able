# Concurrency profile and fixed-context ABI gate — 2026-07-13

## Scope

This gate refreshed the verifier-backed `concurrency` suite after adding
Future Pipeline, then profiled Channel Rollup and Future Pipeline separately.
It deliberately distinguishes two questions:

1. Is there a shared bytecode-VM leaf that warrants a VM change?
2. Is there a shared compiled-runtime leaf with a general, already implemented
   candidate that survives non-concurrent controls?

All scored processes used CPU 15, `GOMEMLIMIT=1GiB`, `GOGC=50`, and a
45-second per-run limit. The initial cross-language screen additionally used
`GOMAXPROCS=1`. This makes it a reproducible comparison screen, not a model of
multi-core throughput. The profiles did **not** set `GOMAXPROCS`: forcing one
P changes the goroutine/channel scheduling behavior being examined.

No source change is made in this gate. The candidate tested below is the
existing opt-in `ablec -experimental-execution-context` ABI; the default ABI
remains unchanged.

## Pinned cross-language screen

Every completed row was accepted by its canonical Ruby verifier.

| Workload | Go 1.26.4 | Python 3.14 | Ruby 4.0 | Able compiled | Able bytecode | Interpretation |
| --- | ---: | ---: | ---: | ---: | ---: | --- |
| BinaryTrees | 33.0394s | timeout | timeout | 30.8200s | timeout | compiled is 93.3% of Go; interpreter rows are capped |
| Channel Rollup | 0.0059s | 0.0762s | 0.1051s | 1.3600s | 0.4800s | material miss in both Able modes |
| Future Pipeline | 0.0050s | 0.0900s | 0.1526s | 0.7400s | 0.4700s | material miss in both Able modes |

The two channel/Future applications are intentionally different: Channel
Rollup consumes text through channel workers, while Future Pipeline has a
numeric producer, four workers, completion values, and a cooperative
cancellation/yield handshake. Therefore a descendant that repeats in them is
evidence about the language runtime rather than either program's data shape.

## Profile method and findings

Bytecode was profiled as one normal bytecode process, not with the repeated
`bytecode-runtime` benchmark helper. Reusing that helper retains executor state
and changes observable channel behavior. The retained artifacts are:

- `v12/interpreters/go/.profiles/20260713_channel_rollup_bytecode_process.cpu.pprof`
- `v12/interpreters/go/.profiles/20260713_future_pipeline_bytecode_process.cpu.pprof`
- `v12/interpreters/go/.profiles/20260713_channel_rollup_compiled_phases/main.cpu.pprof`
- `v12/interpreters/go/.profiles/20260713_future_pipeline_compiled_phases/main.cpu.pprof`

The normal bytecode process profiles include loader/bootstrap work, which is
reported rather than hidden. Channel Rollup's 400 ms capture has 160 ms in
loader parsing and 180 ms in `bytecodeVM.runResumable`; its async task parent
is 110 ms. Future Pipeline's 360 ms capture has 300 ms in
`bytecodeVM.runResumable` and 270 ms in `GoroutineExecutor.runTask`, but its
material child is numeric `execBinary` (100 ms cumulative). Conversely,
Channel Rollup has text/call/member work; its channel receive is not a shared
large leaf. The two applications share dispatcher and task **parents**, not a
concrete removable VM descendant. No bytecode interpreter change is justified.

Compiled binaries were profiled with phase-local main CPU profiles, excluding
compiler/bootstrap work. Here the result is materially different:

| Generated main profile | `bridge.currentGID` cumulative | `runtime.Stack` cumulative | Generic task parent |
| --- | ---: | ---: | ---: |
| Channel Rollup (1.20 s samples) | 1.10 s (91.7%) | 1.09 s (90.8%) | `RunFuture` 1.03 s (85.8%) |
| Future Pipeline (650 ms samples) | 610 ms (93.8%) | 600 ms (92.3%) | `RunFuture` 420 ms (64.6%) |

`bridge.currentGID` obtains a goroutine identity through `runtime.Stack`.
This repeated, generic runtime bridge cost is the same in both applications;
it is not tied to `Channel`, `Future`, a task count, a source pattern, or a
named nominal container.

## Existing fixed-context candidate

The opt-in fixed-pointer execution-context ABI propagates an already-known
context through direct generated calls and spawned child contexts, allowing
the concurrency kernel to avoid repeated identity discovery. Dynamic/runtime
boundaries retain their compatibility wrappers. This is a general generated
call ABI; it does not inspect an application, benchmark, task count, or
container type.

Fresh matched default/candidate runs produced identical verifier hashes (or an
identical deterministic stdout hash for Array Fold):

| Workload | Default | Fixed context | Change | Guard |
| --- | ---: | ---: | ---: | --- |
| Channel Rollup, 3 runs | 1.3867s | 1.2267s | 11.5% lower | verifier, goroutine executor |
| Future Pipeline, 3 runs | 0.7367s | 0.6467s | 12.2% lower | verifier, goroutine executor |
| Array Fold i32, 9 runs | 0.1356s | 0.0900s | 33.6% lower | numeric serial output hash |
| Lexical Rollup, 3 runs | 0.1400s | 0.1433s | within this short-run timer resolution | verifier, serial `GOMAXPROCS=1` |

Channel/Future candidate runs did not force `GOMAXPROCS=1`; Array Fold and
Lexical Rollup did, as serial controls. GC averages were neutral or lower:
Channel Rollup 5.00 in both modes, Future Pipeline 4.00 to 3.67, Array Fold
3.44 to 3.00, and Lexical Rollup 3.33 to 3.67. The short Lexical Rollup row is
not evidence of a regression: its 10 ms timing granularity is 7% of the
workload. It is a guard only, and the required broad scorecard remains open.

Focused ABI checks passed:

```sh
go test ./pkg/compiler \
  -run '^(TestCompilerExperimentalExecutionContextUsesFixedPointerForStaticCalls|TestCompilerExperimentalExecutionContextThreadsStaticSpawnKernelCalls|TestCompilerExperimentalExecutionContextBoundMethodValueUsesCompatibilityEntry)$' \
  -count=1 -timeout 60s

go test ./pkg/compiler \
  -run '^TestCompilerExperimentalExecutionContextNestedSpawnExecutes$' \
  -count=1 -timeout 60s
```

The latter builds and executes the nested-spawn channel program with Go's race
detector. The prior complete opt-in fixture gate remains the 114 execution
fixtures recorded in
`2026-07-11-compiled-fixed-context-full-package-gate.md`.

## Decision

- Keep **no** bytecode VM change: the only bytecode overlap is a parent frame,
  so changing it would be speculative and could optimize one workload's text
  or arithmetic subpath at the other's expense.
- Retain the fixed execution-context ABI as an opt-in compiler candidate. Its
  repeated bridge wall and wins across independent concurrent and numeric
  applications make it a valid general candidate.
- Do **not** make it the default yet. The candidate needs a fresh, wider
  default-versus-opt-in generated-binary scorecard across the external
  generality suite plus feature-rich serial controls. That gate must reject a
  material regression, not average it away with concurrency wins.
- No canonical `able-stdlib` change is required; the selected candidate lives
  solely in the generic compiler/bridge ABI.

## Next gate

Run a CPU-pinned default-versus-fixed-context compiled scorecard over the
external generality suite, plus representative numeric, collection, text,
algorithm, and concurrency controls. Preserve canonical verifiers, report
timeouts as caps, and use phase-local profiles only for any material
regression. Enable the ABI by default only if that broad gate is neutral or
better without hiding a meaningful serial loss.
