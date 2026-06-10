# Compiled Execution-Context Prototype

## Scope

This is an opt-in compiler prototype (`ablec -experimental-execution-context`),
not the default generated ABI. It carries an internal execution context through
generated compiled bodies, direct same-package and imported calls, spawned
closures, generated package entries, and compiled Go extern entries. Spawn
creates a child environment so async payload data is not shared between tasks.

The context-aware concurrency kernel receives the child payload directly for
channel and mutex operations. This is a language concurrency-kernel boundary,
not a `Channel` nominal-type rule: user source still uses ordinary generated
calls, and no fixture, task count, container, or benchmark name is inspected.
The dynamic bridge remains on its legacy compatibility path, so the option is
intentionally not enabled by default.

## Semantic verification

- `TestCompilerExperimentalExecutionContextThreadsStaticSpawnKernelCalls`
  asserts the generated root/body/spawn/kernel chain structurally.
- `TestCompilerExperimentalExecutionContextNestedSpawnExecutes` builds and
  runs nested spawned channel work under Go's race detector and prints `42`.
- The full compiler concurrency parity set, including cancellation, awaits,
  blocked flush, I/O, and the nested-spawn fixture, passes on the default ABI.
- Experimental output checks passed for `channel_roundtrip_i32_small`
  (`1859646040`), Channel-Rollup (`16384:4828:502100`), BinaryTrees small,
  and Lexical-Rollup (`16384:4828:502100`).

## Bounded performance evidence

All runs used `ABLE_EXECUTOR=goroutine`, `GOMEMLIMIT=1GiB`, and `GOGC=50`.
The compared binaries were freshly generated from the same current source;
they differ only by the experimental compiler option.

| Workload | Default | Execution context | Result |
| --- | ---: | ---: | --- |
| Channel roundtrip CPU-profile duration | 3.23 s | 1.03 s | 68.1% lower |
| Channel-Rollup CPU-profile duration | 3.42 s | 2.25 s | 34.2% lower |
| Channel-Rollup ordinary wall (3 runs) | 5.39, 4.58, 5.38 s | 2.26, 2.29, 2.27 s | mean 5.12 → 2.27 s, 55.6% lower |

The default roundtrip profile sampled 3.32 CPU seconds dominated by Go stack
traceback work; the experimental run sampled 1.04 CPU seconds and no longer
showed that bridge identity wall materially. The Channel-Rollup profiles move
to allocation/GC work rather than bridge goroutine identity. These are bounded
profiles, not a final scorecard or a default-rollout authorization.

## Rejected uniform-helper ABI experiment

The next generic candidate made every generated runtime helper accept and
receive the execution context, and added explicit-environment bridge fallbacks
for dynamic named/value calls. It did not inspect a benchmark, task count, or
nominal container. It was nevertheless rejected and fully reverted: passing a
variadic context through every helper made the serial control materially slower.

The compared binaries were freshly generated from the same candidate source;
the only difference was the experimental context option. All runs used
`ABLE_EXECUTOR=goroutine`, `GOMEMLIMIT=1GiB`, and `GOGC=50`.

| Workload | Default (3 runs) | Uniform-helper candidate (3 runs) | Decision |
| --- | ---: | ---: | --- |
| Channel-Rollup | 3.45, 3.56, 3.54 s (mean 3.52) | 2.26, 2.29, 2.24 s (mean 2.26) | 35.7% lower, retained prototype benefit confirmed |
| Lexical-Rollup serial control | 2.03, 2.01, 2.08 s (mean 2.04) | 2.23, 2.16, 2.30 s (mean 2.23) | 9.4% regression; reject |
| BinaryTrees small control | 0.04, 0.04, 0.04 s | 0.04, 0.04, 0.04 s | timer-resolution neutral |

All candidate launches produced their expected output. The concurrency win is
not sufficient to accept a compiler-wide ABI that makes an independent serial
application slower. The retained option therefore still limits explicit
propagation to direct generated calls and the concurrency kernel; the dynamic
bridge and non-concurrency helpers remain on their compatibility path.

## Retained fixed-pointer ABI follow-up

The audited follow-up is retained behind the same opt-in flag. Experimental
compiled function, method, package-entry, native Array, Go extern, and
monomorphized `Iterator.collect` bodies now have a fixed-pointer `_ctx` entry.
Their existing callable names remain no-context wrappers for dynamic/runtime
boundaries. Direct source-level static lowering selects `_ctx` only when it has
a lexical execution context; native callable wrappers derive one from their
`NativeCallContext`.

The 40 runtime helpers reachable from source-level lowering now expose a
fixed-pointer `_ctx` surface as well. The current migration adapters forward to
the existing `_impl` bodies, passing the pointer through for the concurrency
kernel and preserving the old behavior elsewhere. The two await helpers and
dynamic bridge remain on compatibility paths. This deliberately avoids the
rejected uniform variadic ABI and does not add a benchmark, task-count, or
named-container rule.

Generated-source guards cover fixed core/wrapper pairs, named static calls,
all 40 helper entry points, map-literal helper calls, nested spawn, and native
context construction. A Lexical-Rollup build exposed the generic
monomorphized `Iterator.collect` renderer as a missed paired entry; it now
emits the same `_ctx` core plus legacy wrappers. Focused compiler/bridge,
native-interface/container, race-build nested-spawn, and concurrency parity
checks pass.

Fresh matched binaries were built from this source. Each run used
`ABLE_EXECUTOR=goroutine`, `GOMEMLIMIT=1GiB`, and `GOGC=50`; Channel-Rollup
and Lexical-Rollup received the canonical 1,743,363-byte ENABLE word-list path.
All output controls matched (`16384:4828:502100` for both Rollups; matching
BinaryTrees output).

| Workload | Default (3 runs) | Fixed-pointer context (3 runs) | Decision |
| --- | ---: | ---: | --- |
| Channel-Rollup | 3.65, 3.70, 3.63 s (mean 3.66) | 2.03, 2.04, 2.06 s (mean 2.04) | 44.2% lower; retain |
| Lexical-Rollup serial control | 2.08, 2.01, 2.07 s (mean 2.05) | 2.09, 2.02, 2.03 s (mean 2.05) | 0.3% lower; no serial regression |
| BinaryTrees small | 0.04, 0.04, 0.05 s | 0.04, 0.04, 0.04 s | timer-resolution neutral |

This retains the candidate only as opt-in evidence. It does not yet authorize
making the execution-context ABI the default: the remaining dynamic bridge and
compatibility adapters must be profiled across additional application-shaped
workloads before that decision.

## Next boundary

Refresh bounded profiles for the retained fixed-pointer candidate on
Channel-Rollup, Lexical-Rollup, Word-Frequency, and Document-Audit, separating
generated main time from bootstrap. Select a further implementation only if a
single generic residual repeats across the serial and concurrent applications.
Do not replace compatibility adapters speculatively, or add a container-,
task-, or benchmark-specific branch.
