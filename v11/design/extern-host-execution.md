# Extern Host Execution (v11)

Status: Draft

## Goal
Make extern host function bodies (§16.1.2) execute as host code, per spec, with
no per-function interpreter wiring. Preserve the existing kernel bridges for
arrays/proc/channel/string/hasher/ratio, and keep the kernel surface unchanged.

## Problem
The current interpreters treat extern bodies as declarations and map a few names
(`now_nanos`, `read_text`) to native handlers. This violates the v11 spec, which
requires executing the host code in extern bodies (and in preludes) directly.

## Spec Requirements (Summary)
- Extern bodies must execute as host code for the target language.
- Preludes run before extern bodies for the same target.
- `host_error(message: String)` is available in the host context and produces an
  Able `Error` value.
- Type mapping must follow §16.2/§16.3 (copy-in/copy-out for arrays, `!T` error
  mapping, `?T` as null/nil, no pointers).
- Extern bodies are not permitted inside dynamic packages.
- If a pure Able body exists and no extern matches the target, use the Able body.

## Design Overview
Each package that contains extern bodies for a target is compiled into a single
host module. The module includes the target prelude(s), defines all extern
functions for that package, and exposes callable host function objects. The
interpreter then registers a generic Able stub for each extern that marshals
arguments, invokes the host function, and marshals the result back.

### Per-Package Host Modules
- Build a host module per Able package and target.
- The module concatenates all matching `prelude <target>` blocks in order.
- Each extern body becomes a host function with the same name.
- All functions are exported so the interpreter can bind them.
- A generated host prologue injects `host_error(message: String)` plus type
  aliases for `IoHandle`/`ProcHandle` so extern bodies compile consistently.

### Host Execution Engines
- Go target: generate a Go module and compile it with
  `go build -buildmode=plugin`, then load it via `plugin.Open`.
- TypeScript target: generate a `.ts` module and `import()` it in Bun (which
  supports TS natively).

### Generated Host Helpers (No Registry)
The interpreter does not maintain a handle registry. `IoHandle`/`ProcHandle`
values are host objects passed through extern boundaries as opaque values.
Handles from one package can flow into another because the interpreter stores
the host object directly on the Able value and marshals it back unchanged.

Generated helpers per host module:
- `host_error(message: string)` helper (Go: returns `error`, TS: throws).
- `type IoHandle = any` and `type ProcHandle = any` (or equivalent) aliases so
  extern signatures compile without extra imports.

### Native Kernel Bridges (Unchanged)
The following native bridges remain as-is and are not reimplemented as extern
bodies:
- Array buffer hooks: `__able_array_*`
- Hash map hooks: `__able_hash_map_*`
- Concurrency: `__able_channel_*`, `__able_mutex_*`, `__able_await_*`
- String/char bridges: `__able_String_*`, `__able_char_*`
- Hasher: `__able_hasher_*`
- Ratio: `__able_ratio_from_float`
- Scheduler globals: `print`, `proc_yield`, `proc_cancelled`, `proc_flush`,
  `proc_pending_tasks`

### Extern Resolution Rules
- If an extern body is non-empty, compile and bind via the host engine.
- Empty extern bodies are reserved for kernel bridges only. If a non-kernel
  extern is empty, raise a runtime error and require a real host body.

## Value Marshaling
- Primitive types follow §16.2.
- `Array T` is copied in/out to avoid sharing pointers.
- `?T` maps to nil/None/null.
- `!T` maps to `(T, error)` in Go and exceptions in TS; exceptions map to Able
  `Error` via `host_error` or the default error conversion rules.
- `IoHandle`/`ProcHandle` are passed as opaque host objects with identity
  semantics; no registries or integer IDs.

## Prelude and Host Helpers
- Preludes are concatenated and evaluated once per package + target.
- `host_error(message: String)` is injected in every host module.
- Host modules are self-contained; no interpreter state is injected by default.

## Cooperative Suspension (TypeScript)
The TS interpreter remains synchronous at the Able surface, but extern bodies
may complete asynchronously. The host execution engine treats extern results as
either:
- Immediate values (current behavior), or
- A suspended operation that resumes the current task when the host callback or
  Promise resolves.

When a host extern suspends, the interpreter marks the current task as blocked,
returns control to the scheduler, and resumes the task with the result/error
once available. This preserves the synchronous Able API while ensuring other
procs keep running, matching the spec requirement that blocking operations only
block the calling task.

To ensure entrypoint code can suspend without stalling other procs, the
interpreter runs top-level evaluation (including `main`) as an implicit task in
the cooperative scheduler.

## Caching
- Host modules are cached by a hash of:
  - target + prelude text + extern bodies + version
- Go plugins are stored in a temp cache directory to avoid recompilation.
- TS modules are written to a cache dir and imported by absolute path.

## Security / Trust
Extern bodies run with host privileges. This mirrors the spec and existing
prelude behavior. Tooling may add sandboxing flags later, but not in this scope.

## Compatibility Notes
- Externs remain disallowed in dynamic packages (spec).
- If host execution is unavailable (missing Go toolchain or Bun), the runtime
  should report a clear error indicating extern support is not available.
