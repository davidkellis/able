# Post-nullable compiled concurrency reconciliation

## Decision

Reconcile `compiled-concurrency` as causally current and retain no production
change.

All 23 current strict applications passed their public verifiers, contain 96
packages, and omit `pkg/interpreter`. The primitive nullable carrier is
materially reached in four applications. Three already consume it as a native
Go `value + valid` result. The only remaining nullable-to-runtime cast shape is
Await Channel Mux's explicit runtime-callable callback ABI, so its exact
breadth is one application rather than the required three.

The broad channel- and mutex-handle boundaries are unchanged from the
pre-nullable census. They remain explicit scheduler-service ABIs, not
accidental compiler/interpreter crossings, and the retained normal-binary
profiles keep channel-handle recovery below one percent. No changed shared
owner qualifies for fresh profiling or an A/B implementation.

## Strict 23-application gate

The diagnostic compiler was built from the retained production compiler with
typed-boundary telemetry enabled. Every application was generated with
`--no-fallbacks` into a disk-backed `/var/tmp` workspace and run once through
its public verifier.

| Gate | Result |
| --- | ---: |
| Strict builds | 23 / 23 |
| Public verifier smokes | 23 / 23 |
| Dependency packages | 96 in every graph |
| Interpreter dependencies | 0 |
| Build or smoke failures | 0 |
| Timing samples admitted | 0 |

The smokes are reachability diagnostics, not performance evidence. The
authoritative frontier retains five successful Able and Go processes for each
row. Its 23 concurrency rows span 5.333333x to 19.649123x Go, have an
8.310084x geometric ratio, and contribute 0.882632 seconds of positive target
excess.

The rows are Await Channel Mux, Channel Rollup, Concurrent Audio Voices,
Concurrent Document Pipeline, Concurrent Event Routing, Concurrent Graph
Visitors, Concurrent Packet Codecs, Concurrent Policy Callbacks, Concurrent
Scene Tiles, Concurrent Signal Dispatch, Concurrent State Machines,
Concurrent Stateful Pipeline, Concurrent Stencil Reduction, Concurrent Text
Index, Concurrent Transform Chain, Concurrent Tree Folds, Dependency Wave
Validation, Future Await Race, Future Pipeline, Mutex Await Journal, Mutex
Ledger, Mutex Work Queue, and Validated Job Pipeline.

## Primitive nullable reach

Whole-module support definitions are not treated as application reach. The
fresh generated material paths classify as follows:

| Application | Material carrier path | Remaining changed residual |
| --- | --- | --- |
| Await Channel Mux | `fn(i64?) -> i64` callback parameter | Four generated callback sites cross the explicit runtime-callable ABI; 2,048 transitions |
| Channel Rollup | `Channel<i64>.receive()` result | None; match reads `.valid` directly |
| Future Await Race | ready-channel `receive()` result | None; native result is intentionally ignored |
| Future Pipeline | job, result, and cancellation `receive()` results | None; matches read `.valid` directly |
| Other 19 rows | No material primitive-nullable path | None |

Await's callback body converts its native nullable argument only because the
callback itself is exposed through the runtime callable protocol. That exact
box/cast pattern occurs in no other current concurrency row. In particular,
the superficially similar casts in Mutex Await Journal and Mutex Work Queue
start from dynamic `Future.value()` results held as runtime values; they are
not native nullable values and cannot establish breadth for this candidate.

The retained carrier already removed 2,048 Await control-to-error
materializations relative to the prior census while preserving its 2,048
callable transitions. This confirms causal reach and improvement without
turning the remaining explicit ABI into a three-application rule.

## Stable service boundaries

Typed-boundary event counts are reachability evidence, not CPU or allocation
attribution:

| Boundary | Applications | Current events | Pre-nullable events | Change |
| --- | ---: | ---: | ---: | ---: |
| Channel handle recovery | 14 | 238,838 | 238,838 | 0 |
| Mutex handle recovery | 3 | 43,016 | 43,016 | 0 |
| Native-nullable callback callable | 1 | 2,048 | not a separate pre-carrier shape | new representation, same explicit callable ABI |

The channel and mutex handles are opaque identifiers owned by the scheduler
service. Recovering them at that ABI is a real static/runtime boundary, but it
does not link or invoke the interpreter. Retained normal-binary profiles put
channel-handle recovery at only 0.11% flat CPU in Channel Rollup and 0.22% in
Future Pipeline. The exact event counts are identical before and after the
nullable change, so the carrier did not invalidate that ownership evidence.

The full fresh diagnostic totals include 281,884 runtime-to-integer, 108,951
runtime-to-struct, 100,291 struct-to-runtime, 18,948 runtime-to-union, 12,296
union-to-runtime, 10,248 runtime-to-interface, 8,704
interface-to-runtime, and 8,704 callable-to-runtime events. These aggregate
different semantic shapes. None is evidence for flattening Future, Channel,
Mutex, interface, union, callable, or nominal semantics wholesale.

## Retained profile and admission gate

The retained post-spawn profile intersection remains Await Channel Mux,
Validated Job Pipeline, and Concurrent Stateful Pipeline:

| Owner | Await | Validated | Stateful | Disposition |
| --- | ---: | ---: | ---: | --- |
| `bridge.ToInt`, objects/run | 4,098 | 4,107 | 65,554.3 | Three-row leaf, but global cache slowed TapeLang 4.17% |
| `bridge.currentGID` | material | material | not material | Breadth two; broad context ABI closed |
| Nominal struct construction | absent | material | material | Breadth two; nominal special cases forbidden |

The nullable carrier changes Await's primitive callback representation but
does not change the `ToInt` leaf in the other two applications, make
`currentGID` material in Stateful, or add a third nominal owner. Fresh
profiles cannot increase the exact one-application breadth of Await's
nullable callback cast, and repeating the stable service-handle profiles
would not test a changed mechanism.

Therefore no CPU profile, allocation profile, timing cohort, or A/B
implementation was admitted. No compiler, generated runtime, runtime package,
interpreter, bytecode VM, canonical stdlib, benchmark, language, dependency,
nominal special case, or WASM source changed.

## Evidence, verification, and cleanup

The machine-readable companion is
`2026-07-30-post-nullable-compiled-concurrency-reconciliation.json`. It records
the complete row set, aggregate telemetry, retained profile intersection,
artifact identities for the four carrier-reaching applications, and the
closed alternatives.

Six focused native-nullable, Awaitable, Future, and mutex guards passed in
15.787 seconds. `go test ./cmd/ablec` passed in 7.209 seconds. All five
frontier tests and seven evidence-ledger tests passed; the regenerated
130-row frontier has zero actionable groups, and its direct check passes. The
ledger check passes with all 23 closures accounted for and the expected 12
compiler-scope invalidations retained.

The exact 2,563 MiB disk-backed compiler, generated-source, binary,
smoke-output, telemetry, and Go-cache workspace is then removed; no matching
artifact is retained under `/var/tmp` or `/tmp`.

## Next

Reconcile `cross-family-architecture-ownership`, then advance the compiler-
scope closures together if that final causal review passes.

Why: concurrency was the last unreviewed compiled family after the primitive
nullable compiler change. The ledger intentionally keeps all 12 closures
sharing that compiler-production drift invalidated until their family and
cross-family ownership claims are jointly current.

What it entails: merge the post-nullable causal records across all compiled
families, audit whether any exact changed owner spans three unlike families,
refresh the cross-family architecture record, and atomically rebase the 12
shared-scope closures without changing their dispositions unless the evidence
requires it.

Why it matters: this prevents stale compiler identities from masquerading as
open optimization work, verifies that no shared boxing or
compiler/interpreter boundary was hidden by family-local analysis, and makes
the evidence selector trustworthy for choosing the next real performance
target.
