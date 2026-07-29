# Compiler Callable-Context Default Activation

## Status

Retained on 2026-07-29. The existing generated execution-context ABI now
activates by default for compiled program graphs that require scheduler-aware
callable propagation. The admission result is
`../docs/perf-baselines/2026-07-29-compiled-callable-context-default-activation-retained.md`.

It does not authorize a second callable ABI, a global option flip, a
benchmark-specific path, or a named non-primitive nominal rule. Production
changes beyond this activation still require verifier-backed A/B evidence.

## Problem

Strict compiled applications are interpreter-free, and statically
representable values already use native Go carriers. Several unlike
await-bearing applications nevertheless spend most of their CPU recovering a
goroutine identity through `bridge.currentGID` and `runtime.Stack` when a
runtime callable needs the active environment.

The compiler already has an opt-in solution:

- generated static functions and native interfaces accept
  `*__able_execution_context`;
- native callable types and captured lambdas carry that same pointer;
- runtime-callable adapters translate once through
  `runtime.NativeCallContext`;
- spawned tasks own a child context;
- reverse native-to-generated calls reuse the task context when payload and
  environment identities match; and
- dynamic, host-created, nil-context, and incompatible-environment calls keep
  compatibility entries.

The unresolved question is activation. Enabling the older broad context ABI
for every program improved many applications but materially regressed
unrelated serial and concurrent guards. The later await-gated callable ABI
improved all four reached catalog applications by 55.00%-95.09% and had zero
callable-context reach in the other 59. It remains opt-in.

## Decision

Do not design or implement another execution-context ABI.

The next candidate must reuse the existing ABI and change only its activation
policy:

1. detect a scheduler-context requirement independently of the experimental
   option;
2. enable the existing context ABI by default only when that requirement is
   present;
3. retain the experimental option as a force-on diagnostic for await-free
   programs during the admission period; and
4. keep the compatibility ABI at dynamic and host entry points.

A simple global default for `ExperimentalExecutionContext` is forbidden. The
July 24 full-scorecard regressions already reject that route.

## Current context ownership

The existing generated carrier separates three identities:

| Field or view | Owner | Meaning |
| --- | --- | --- |
| `env` | current task or compatibility caller | Environment visible to runtime services and explicit dynamic work |
| `packageEnv` | current compiled package entry | Package whose proven body may run without another compatibility swap |
| `payload` | spawned asynchronous task | Future handle, await state, cancellation, and scheduler state |
| embedded `NativeCallContext` | execution context | Allocation-free runtime-callable view of `env` and `payload` |

These identities must not be collapsed. A spawned task has a child `env` but
inherits its caller's `packageEnv`. A cross-package dependent body localizes
the context to the callee package while retaining the task payload. A
generated Go closure captures lexical native values normally; it receives the
caller's task context at invocation. An interpreter `FunctionValue` continues
to own its separate lexical `Closure` environment.

## Required invariants

### Native static execution

- Primitive, String, Array, nominal, union, interface, result, and callable
  values remain on their existing native Go carriers.
- Adding the context pointer must not box arguments, results, captures, or
  receivers.
- A context-aware static or native-callable call performs no heap allocation
  solely to transport context.
- A context-aware internal call must not consult goroutine identity.

### Package and lexical environments

- Environment-independent bodies may use their raw generated entry.
- A dependent body must establish its own package environment before reading
  package state.
- Cross-package invocation must never infer the callee package from the
  caller's task environment.
- Generated lambdas retain ordinary Go lexical capture. Dynamic
  `FunctionValue` callables retain their runtime `Closure`.
- Context localization preserves the task payload and constructs a matching
  embedded native view.

### Scheduler state

- Every spawned task owns one primary context before user task code runs.
- Nested spawn creates a child environment and a new task payload; it does not
  reuse the parent's payload.
- Native-to-generated reverse calls may reuse a context only when both payload
  and environment identity match.
- Await, cancellation, blocking/unblocking, stale-waker, and executor
  semantics remain unchanged.

### Compatibility boundaries

- `runtime.NativeCallContext` remains the host/runtime compatibility surface.
- Nil contexts, host-created contexts, dynamic values, interpreter callables,
  and mismatched captured/package environments retain the established
  reconstruction or bridge path.
- Compatibility entries remain generated for old/dynamic callers even when
  all known static callers use the context entry.
- Explicit dynamic boundaries may use the bridge; their surrounding static
  code returns immediately to native carriers.

## Activation rule

The activation fact must be computed before code generation and must not
depend on `executionContextsEnabled()`. The current
`programNeedsCallableExecutionContext` check is option-dependent and filters
to the entry package, so it is not sufficient as a production selection
rule.

The first candidate should conservatively activate when any statically loaded
module in the compiled program graph contains an `await` expression. This
includes imported-package await bodies and nested/captured lambdas. Scanning
the loaded graph may activate unreachable await code, but it cannot miss a
reachable static await. A later reachability refinement is admissible only if
generated-source evidence proves the conservative rule causes a material
regression.

Conceptually:

```text
schedulerContextRequired = loadedProgramContainsAwait(modules)
executionContextsEnabled =
    forceExperimentalExecutionContext || schedulerContextRequired
callableExecutionContextsEnabled =
    executionContextsEnabled && schedulerContextRequired
```

The force-on option must not cause the scheduler-only callable surface to
appear in await-free programs unless a separate diagnostic explicitly
requests that surface. This preserves the current zero-reach guard and avoids
reviving the broad all-callable experiment accidentally.

## Implementation seams

The candidate should remain small and general:

1. Make scheduler-context detection independent of compiler options and cover
   all loaded modules.
2. Store the resulting fact on the generator before any pass consults context
   activation.
3. Separate "context machinery enabled" from "scheduler callable context
   required" so option forcing and production selection cannot become
   circular.
4. Update CLI/help wording only after the candidate clears admission. During
   A/B work, the existing option remains an experimental force-on control.
5. Add default-selection tests without weakening existing compatibility,
   dynamic-boundary, or await-free source-shape guards.

No changes are required to `runtime.NativeCallContext`, interpreter callable
types, Able syntax, the canonical stdlib, or public language semantics.

## Failure modes to guard

- **Global activation:** serial programs gain context siblings and repeat the
  July 24 allocation/GC regressions.
- **Entry-package-only detection:** a reachable imported await body uses the
  compatibility ABI and recovers goroutine identity.
- **Environment conflation:** task-local state or a foreign package binding is
  read through the wrong environment.
- **Payload loss:** localization or a runtime adapter drops cancellation or
  await state.
- **Unsafe reverse reuse:** a host or cross-package call receives a task
  context belonging to another environment.
- **Hidden allocation:** a localized, native, callable, or bound-method entry
  allocates a context or argument prefix per call.
- **Compatibility removal:** dynamic or host callbacks lose a valid entry.
- **Static bridge fallback:** a known generated call reaches
  `currentGID`/`runtime.Stack`.

## Focused proof

Before application timing, bounded tests must prove:

- default await-bearing code emits context-aware static, interface, callable,
  captured-lambda, bound-method, and runtime-adapter entries;
- default await-free code remains byte-identical at every generated Go source
  file;
- an imported-package await activates the context path;
- package-dependent and package-independent cross-package callbacks preserve
  their distinct entry behavior;
- nested spawn, cancellation, Future, Channel, Mutex, stale waker, error
  control, and both executors retain parity;
- nil/host-created native contexts and explicit dynamic named/value calls use
  compatibility paths correctly;
- strict generated graphs omit `pkg/interpreter`; and
- escape analysis and allocation counters show no per-call context allocation.

Each test command must remain below one minute; broader matrices must be
partitioned.

## Application admission protocol

Build frozen baseline and candidate compilers. Use disk-backed tranche
workspaces, five or more order-balanced processes per side, catalog CPU and
executor settings, public verifiers, and matching Go 1.26.5 references.

The governing cohort is all 63 compiled applications:

- all outputs and verifiers must match;
- all strict graphs must remain interpreter-free;
- the 59 currently await-unreached programs must retain identical generated
  Go source;
- Future Await Race, Await Channel Mux, Mutex Await Journal, and Mutex Work
  Queue must execute the context path;
- unrelated serial controls, especially N-body, numeric records, Unicode, and
  text/map applications, must not execute or emit it; and
- no established 95%-of-Go guard may be lost.

For the four reached applications, also collect repeated CPU, allocation, and
exact boundary counts. The candidate advances only if:

- every reached application is non-regressing under paired measurement;
- the reached geometric mean improves materially;
- `currentGID`/`runtime.Stack` falls by the expected exact call class;
- context bytes/objects do not merely move to another owner; and
- the full 63-application geometric mean and summed target excess do not
  regress.

If any condition fails, retain no production activation change. The existing
opt-in ABI and this design record remain useful evidence.

## Why this matters

This activation change is the shortest general route from existing generated
Go carriers to native Go call behavior in await-bearing applications. It can
remove tens of thousands of `runtime.Stack` calls without changing Able data
representation or crossing into the interpreter. The zero-reach requirement
is equally important: native performance is a product-wide goal, so the
concurrency win cannot be purchased with new overhead in ordinary compiled
programs.
