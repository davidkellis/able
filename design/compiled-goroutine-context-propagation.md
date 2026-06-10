# Compiled Goroutine Context Propagation

## Status

Design audit completed on 2026-07-10. No implementation is selected yet.

## Problem

Compiled goroutine programs store the active runtime environment and
diagnostic call frames in `bridge.Runtime` maps keyed by an identifier parsed
from `runtime.Stack`. After the retained event-driven `Flush` change, this is
the dominant residual cost in two independent concurrent-blocking programs:

| Workload | `bridge.currentGID` cumulative CPU |
| --- | ---: |
| Channel-Rollup | 95.1% |
| Channel roundtrip | 93.9% |
| BinaryTrees goroutine control | 0.7% |

The cost is generic bridge machinery, not a `Channel` lowering rule. It is
material when generated concurrent tasks make native/runtime calls; selecting
the goroutine executor alone does not make it material.

## Current boundary

The bridge has nine `currentGID` call sites. They protect all three pieces of
per-execution state:

- `Runtime.Env`, `SetEnv`, `SwapEnv`, and `currentEnv` select an environment;
- `SwapEnvIfNeeded` / `RestoreEnvIfNeeded` bracket generated package and
  wrapper calls; and
- bridge diagnostic push/pop/snapshot calls keep call frames separate across
  concurrent tasks.

The compiler emits twelve `__able_runtime.Env()` and twenty-two
`SwapEnvIfNeeded` sites across generated runtime helpers, call paths, wrappers,
thunks, and package entry methods. The hot roundtrip attribution is concrete:

```
Channel send/receive
  -> __able_current_payload()                 36.5%
  -> Runtime.Env() -> currentGID/Stack

compiled package entry method
  -> SwapEnvIfNeeded / RestoreEnvIfNeeded      30.8%
  -> currentGID/Stack
```

`runtime.NativeCallContext` already carries `Env` and `State`, and the
compiler already writes the async payload into task-environment runtime data.
That is useful but insufficient by itself: direct statically generated calls
such as compiled Channel methods call runtime helpers without a native context.
They therefore cannot recover task-local payload or environment without the
global bridge lookup.

## Rejected shortcuts

- A last-goroutine-ID cache cannot avoid calculating the ID and is unsafe for
  concurrent callers.
- A global active-environment pointer loses isolation between parallel tasks
  and fails nested swap/restore.
- Mapping from Future to environment is circular: generated code must already
  know which Future/task is active to perform that lookup.
- Replacing the `runtime.Stack` parser with another undocumented goroutine-ID
  mechanism changes no semantic boundary and leaves the unsupported identity
  dependency in the hot path.
- Removing package swaps without carrying the target package environment would
  break cross-package resolution, dynamic fallback, and diagnostics.

## Viable direction

Introduce an internal generated `__able_execution_context` that owns:

- the active `*runtime.Environment`;
- the active `*__able_async_payload`, when executing a spawned task; and
- task-local diagnostic call-frame state.

The compiler must thread this as a hidden parameter through generated entry
methods, wrappers, thunks, callable dispatch, and runtime-kernel calls. Direct
compiled code then obtains package environment and payload from the explicit
context, not `bridge.Runtime.Env()`. Spawn creates a child context carrying the
child environment and payload. Nested calls reuse or derive that context;
serial execution may use a root context.

Dynamic/interpreter fallback must receive the explicit environment through
new bridge `...In` APIs. The old goroutine-ID map may remain only for legacy
bridge callers until all generated paths are migrated; it must not remain in
the normal static concurrent path if the optimization is kept.

This is a general compiler ABI/runtime-context change. It applies to every
generated call and concurrency kernel, rather than recognizing Channels,
fixtures, task counts, or nominal types.

## Required semantic guards

Before benchmarking a prototype, add/extend coverage for:

1. nested spawned tasks that make native calls and retain distinct payloads;
2. cross-package wrapper entry and nested environment swap/restore;
3. concurrent diagnostic call-frame isolation and error attribution;
4. cancellation, blocked-task `future_flush`, await re-entry, and future value
   evaluation; and
5. serial executor behavior and dynamic bridge fallback.

Existing compiler concurrency parity fixtures, blocked-task flush parity, and
generated race coverage remain mandatory. A candidate must also preserve
Channel-Rollup, Channel roundtrip, BinaryTrees, and Lexical-Rollup output
before any performance result is considered.

## Next experiment

The spawn-boundary feasibility trace is in
`design/compiled-goroutine-context-prototype-feasibility.md`. It rules out a
partial task-only or Channel-only prototype: direct static generated calls do
not receive task context. The next implementation must instead use generated
context-aware ABI variants across functions, methods, package entries, thunks,
dispatch, and runtime kernels, while retaining root wrappers and explicit
dynamic fallback. Add the required semantic guards before that migration; only
the fully threaded static path should be profiled across the material workloads
and BinaryTrees/Lexical controls.
