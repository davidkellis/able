# Compiled Goroutine Bridge Third Profile

## Decision

Keep no implementation change in this profiling tranche. The third workload
does confirm `bridge.currentGID` as the next eligible *generic design target*:
it recurs materially in two independent compiled goroutine programs after the
shared `Flush` polling wall was removed. It is not yet safe to replace—the
bridge uses that identity to preserve per-goroutine runtime environments and
call-frame state—so the next tranche must establish a semantics-preserving
environment-propagation design before changing it.

## Method

- Built the current compiled `channel_roundtrip_i32_small` fixture and ran it
  with `ABLE_EXECUTOR=goroutine`, `GOMEMLIMIT=1GiB`, and `GOGC=50`.
- The fixture performs eight independent buffered-channel rounds. Each round
  spawns a producer, flushes it, and then receives 10,000 values; it therefore
  exercises real concurrent blocking without reusing Channel-Rollup's text,
  filesystem, reduction, or worker-pipeline source shape.
- Five normal collector-free CPU-profiled processes all printed the expected
  `1859646040`; their profile is retained as
  `.profiles/20260710_channel_roundtrip_i32_small_compiled_flush_notifier.cpu.pprof`.
- Compared this current candidate state with the retained post-notifier
  Channel-Rollup and BinaryTrees profiles under the same goroutine selection.

## Result

| Workload | `currentGID` cumulative samples | Material bridge path |
| --- | ---: | --- |
| Channel-Rollup | 95.1% of 11.03 s | `Runtime.Env`, environment swap/restore, spawned task calls |
| Channel roundtrip | 93.9% of 16.49 s | `Runtime.Env` 36.2%, swap/restore 30.8%, Channel send/receive calls |
| BinaryTrees | 0.7% of 2.81 s | Not material; tree construction/allocation dominates |

In both material workloads, `currentGID` spends nearly all of its time in
`runtime.Stack` to derive a goroutine identifier. The exact bridge helper is
shared; it is reached from environment lookup, swapping, and per-goroutine
call-frame work, not from a special compiler rule for `Channel`. BinaryTrees
shows that merely selecting the goroutine executor is insufficient—the cost is
material when generated concurrent tasks make native bridge calls.

The retained Flush notifier remains absent as a material CPU leaf, so this is
new residual work rather than a stale pre-change profile. No stdlib source or
benchmark program was changed.

## Next recommendation

Perform a design and call-boundary audit for generic goroutine-local runtime
environment propagation. Why: two independent concurrent-blocking programs
now establish `currentGID`/`runtime.Stack` as a shared cost, but its existing
identity map protects re-entrant environment and call-frame semantics. The
work entails tracing `Runtime.Env`, `SetEnv`, `SwapEnv`,
`SwapEnvIfNeeded`, and generated native-call/task boundaries; identifying a
context or explicit-environment propagation mechanism that does not derive
goroutine IDs; and writing semantic tests for nested tasks, environment
restore, and concurrent call frames before any candidate benchmark. Do not
special-case Channel operations, this fixture, task counts, or a nominal type.
