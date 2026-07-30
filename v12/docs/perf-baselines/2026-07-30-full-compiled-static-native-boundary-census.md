# Full compiled static native-boundary census closure

Date: 2026-07-30

## Decision

Retain the report-only generated-Go census tooling and no compiler, generated
runtime, runtime, interpreter, bytecode VM, canonical-stdlib, language,
dependency, benchmark, or WASM performance change.

All 66 selected compiled applications generated with `--no-fallbacks` under a
60-second per-module bound. The pass classified `runtime.Value` types,
bridge calls, erased/dynamic calls, interface adapters, union conversions,
struct conversions, callable conversions, control/error transitions, and
heap nominal literals. It found no exact safely removable boundary that is
both material and present in three unlike applications.

No prototype or A/B/Go timing cohort was run because no candidate crossed the
admission gate. The current five-process scorecard and retained normal-binary
profiles remain the timing authorities.

The machine-readable companion is
`2026-07-30-full-compiled-static-native-boundary-census.json`.

## Complete bounded generation

The report-only runner generated and analyzed one module at a time per worker,
then deleted that module before retaining the next compact row result.

| Measure | Result |
| --- | ---: |
| Selected strict modules | 66 |
| Successful / failed | 66 / 0 |
| Generation time | 0.353-23.459 seconds; 8.024-second mean |
| Generated module size | 1,123,264-9,041,898 bytes; 4,373,104-byte mean |
| Generated Go lines | 30,143-213,916; 108,259-line mean |
| Compiled frontier | 66 rows, 60 misses, 8.473895 seconds target excess |

This is static reach evidence, not wall-time evidence. The generation timings
only prove that every row completed within the project guardrail.

## Two complementary scopes

The analyzer records two views:

- `compiled_body` is conservative: it counts every emitted Able-compiled
  function body, including canonical-stdlib bodies that a particular
  application may never call.
- `main_direct_reachable` is a lower bound: it begins at
  `__able_compiled_fn_main` and follows direct calls only to other generated
  compiled bodies. It does not guess targets for Go-interface dispatch.

This distinction prevents support templates from being treated as executed
application boundaries while preserving conservative evidence for unresolved
host-native polymorphism.

The direct-main view contains:

| Semantic category | Applications | Static sites | Disposition |
| --- | ---: | ---: | --- |
| erased/dynamic call | 62 | 182 | 111 are `__able_call_named` host calls; residual method calls are already closed |
| bridge encode | 62 | 307 | primarily host output/argument and runtime-service ABI |
| control/error conversion | 50 | 661 | already closed zero-flat/nil-check parent |
| struct runtime conversion | 38 | 193 | different semantic payloads and lifetimes |
| native union wrap/projection | 41 | 1,810 | successful native Go lowering, not boxing |
| native interface adapter | 22 | 59 | successful host-native polymorphism |
| interface runtime conversion | 6 | 16 | iterator or scheduler-service recovery |
| callable runtime conversion | 5 | 5 | no exact shape reaches three unlike applications |
| union runtime conversion | 1 | 1 | one Option/Result application |
| runtime-value constructor | 0 | 0 | absent from the direct lower-bound paths |

## Exact candidate review

### Explicit host ABI

`__able_call_named` reaches 62 applications at 111 static sites.
`bridge.ToString` reaches 55 at 80 sites, and `bridge.ToFloat64` reaches five
at nine sites. These are one or a few argument/output conversions per
process. They are permitted host boundaries, not a repeated main-phase CPU
owner.

### Closed control and call routes

`__able_control_from_error` reaches 50 applications at 339 direct sites;
the node-attributed form reaches 30 at 148 sites, and
`__able_control_to_error` reaches 26 at 174 sites. Existing telemetry-free
profiles place these at nil checks or zero-flat parents above unlike
descendants.

`__able_method_call_node` reaches 17 applications at 71 sites, concentrated
in concurrency programs. The current call/member/index profile chain already
closes this route with no shared exact child. Static recurrence does not
invalidate those normal-binary profiles.

### Runtime services

`bridge.ToDynamicI64` reaches 30 applications at 173 sites and `bridge.AsInt`
reaches 31 at 98 sites. The retained typed-shape census separates channel
handles, capacities, mutex handles, and host conversions. Current
normal-binary profiles put the material channel-handle helper below one
percent.

`Future<i64>` and `Awaitable<i64>` recovery each reach three applications,
but only related scheduler/mutex rows. They are required service-ABI
conversions, not unlike-program breadth.

### Callable and nominal paths

No exact callable conversion reaches three unlike applications. The only
repeated shape, `void -> i64` to runtime, reaches two related mutex rows.

The exact nominal encoders that reach three or more rows remain confined to
one family:

- `MathDomainError` in N-Body, Distance Field, and RMS Norm;
- `OverflowError` in Fixed Width, Rational Series, and Wide Integer Records;
- `AutomataError` in six regex-family applications.

They encode semantic error payloads on cold/error paths. A nominal-name rule
would also violate the compiler guardrails.

The direct lower-bound graph contains 3,417 heap nominal literal sites across
55 applications. The concrete values have different mutable identity,
escape, return, storage, interface, and scheduler lifetimes. Replacing them
with copied values is not semantics preserving, and grouping them only as
“allocation” does not define a removable boundary.

## Why no production candidate advanced

- Broad `runtime.Value` syntax is dominated by explicit host output/argument
  and runtime-service boundaries.
- Native union projections and Go-interface adapters demonstrate successful
  native lowering and must not be removed as boxing.
- The remaining exact runtime conversion shapes either lack three-unlike-row
  breadth, are already closed by normal profiles, or are required semantic
  payloads.
- Current pure-native misses therefore cannot be explained by one remaining
  compiler/interpreter crossing tax.
- Reopening named containers, non-primitive nominals, broad execution-context
  ABIs, call/member helpers, control/error parents, or nominal identity would
  repeat closed or prohibited work.

## Retained report-only tooling

- `v12/bench_compiled_static_boundary_census` performs the bounded full
  generation, joins current frontier timing, and deletes generated modules
  after each row.
- `cmd/able-generated-boundary-census` parses generated Go and separates
  compiled bodies, entry wrappers, runtime wrappers, support code, and the
  direct-main lower bound.
- A focused synthetic regression test enforces the scope and category
  separation.

The tooling changes no ordinary compiler output or execution path.

## Verification and cleanup

- The four-test census manifest/report contract passes.
- `go test ./cmd/able-generated-boundary-census ./cmd/ablec -count=1
  -timeout=60s` passes.
- The runner passes its shell-syntax and help-contract checks.
- Every retained task source file remains below 1,000 lines.
- The exact 337,268 KiB disk-backed census workspace and its pointer were
  removed after preserving this compact evidence.

## Next

Run one current correctness and release-readiness tranche, then keep
production performance mutation paused until a concrete evidence identity
invalidates a closed owner.

Why: the complete 66-row census finds no new admissible lowering boundary,
and semantic-work, checked arithmetic, Array proofs, call/member,
control/error, execution-context, allocation/GC, and nominal-identity routes
are already closed. Another local performance prototype would be
benchmark-specific or would weaken required semantics.

What it entails: run the ordinary v12 correctness, canonical-stdlib, fixture,
strict dependency, performance-evidence, and release checks; repair only a
real shared failure; verify that the new report-only tooling is reproducible;
and reopen production performance work only if a new application,
correctness fix, stdlib/runtime/compiler change, or observer result produces
one exact owner material in three unlike applications.

Why it matters: this protects the native primitive/Array carriers and
interpreter-free compiled architecture while keeping the 95%-of-Go goal
evidence-driven. It also prevents repeated work on static sites that current
profiles already prove are cold, required, or family-specific. Continue to
defer WASM.
