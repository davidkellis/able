# Compiled Goroutine Context Prototype Feasibility

## Decision

The first vertical prototype landed on 2026-07-10 behind the opt-in compiler
flag `ExperimentalExecutionContext` / `ablec -experimental-execution-context`.
It threads a generated execution context through compiled bodies, direct
function and method calls, imported entries, spawned children, and compiled Go
extern entries. The context-aware concurrency kernel consumes the child
payload directly. Default output remains unchanged. The remaining generic
runtime-helper and dynamic boundaries are not complete: the all-helper
variadic experiment regressed Lexical-Rollup and was reverted. The required
fixed-pointer migration audit is recorded in
`design/compiled-fixed-context-abi-audit.md`.

The prototype validates the design boundary: a spawn-only context parameter
would add state without reaching direct generated calls. The retained version
is therefore a complete static vertical slice, not a Channel source or
benchmark exception. The current Flush notifier remains unchanged.

This is progress rather than a blocker: it defines the smallest meaningful
prototype as a general generated-call ABI transform, so later work will not
ship inert plumbing or encode a benchmark-shaped Channel fast path.

## Trace result

The current task boundary is straightforward:

```
compileSpawnExpression / IRSpawn
  -> func(*runtime.Environment) (runtime.Value, error)
  -> __able_spawn / __able_spawn_future
  -> executor.RunFuture
  -> __able_run_compiled_task(payload, env, task)
```

Adding a second task-context argument here would allow the closure to receive
the child environment and async payload. It does not affect the expensive
direct path inside that closure:

```
spawn body
  -> __able_compiled_entry_method_* (no context parameter)
  -> SwapEnvIfNeeded(__able_runtime, packageEnv)
  -> direct runtime helper (no NativeCallContext)
  -> __able_current_payload()
  -> __able_runtime.Env() -> currentGID -> runtime.Stack
```

`runtime.NativeCallContext` already carries `Env` and `State`, but it is only
constructed for named/dynamic native invocation. Direct statically generated
entry methods, wrappers, and concurrency-kernel calls do not receive it.
Therefore neither storing a context on the async payload nor extending the
task closure makes it observable at the material call sites.

## Required meaningful prototype

The prototype must add an internal `*bridge.ExecutionContext` argument to the
context-aware variants of every generated direct-call boundary:

1. Generate `_ctx` variants for compiled functions, methods, package entries,
   thunks, callable dispatch, and direct runtime-kernel helpers. Preserve the
   existing no-context public/generated boundary as a root-context wrapper so
   callers outside compiled code retain their ABI.
2. Extend compile/codegen context with the active execution-context expression.
   Every direct compiled call uses a `_ctx` callee; spawn derives a child
   context from its child environment and async payload.
3. Construct `NativeCallContext{Env, State}` from the execution context for
   runtime/native calls. This lets future/cancellation/await helpers consume
   task state without querying `Runtime.Env`.
4. Move static package environment selection and diagnostic call frames into
   the execution context. Dynamic/interpreter fallback receives explicit
   `...In` environment APIs and may temporarily retain legacy bridge identity
   lookup only outside the normal static path.

This is a generic compiler ABI migration. The implemented slice applies to
compiled direct primitive and nominal calls uniformly; the context-aware
concurrency helpers are a language kernel boundary, not a rule for a nominal
`Channel` type. It does not identify BinaryTrees, a fixture, task counts, or a
named standard-library type.

## Required first guards

Before source migration, add focused compiler tests that lock down:

- root-context wrappers preserve the current generated signatures at external
  boundaries;
- a nested spawn receives a distinct child context/payload while its parent
  context remains intact;
- cross-package entry calls restore the parent context after a child call;
- concurrent diagnostic frames remain isolated; and
- dynamic fallback receives the explicit environment.

Then run existing concurrency parity, blocked-flush, await/re-entry,
cancellation, compiler race, and cross-package fixture coverage. Only a fully
threaded static path is eligible for the Channel-Rollup/roundtrip/BinaryTrees/
Lexical performance comparison.

## Rejected prototype shapes

- **Spawn-only hidden argument:** semantically incomplete and does not reach
  direct static entry methods or runtime helpers.
- **NativeCallContext-only change:** misses direct static calls, which are the
  material profile path.
- **Channel helper parameter:** would be a named-kernel special case and leave
  the same issue in every other concurrent native call.
- **Global active context or a GID cache:** loses parallel/nested isolation or
  still needs `runtime.Stack` to identify the caller.
