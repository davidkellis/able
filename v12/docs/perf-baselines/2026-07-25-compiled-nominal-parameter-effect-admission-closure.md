# Compiled nominal-parameter effect admission closure

Date: 2026-07-25

## Decision

Retain no compiler or runtime code for read-only, non-escaping nominal
parameter effects or runtime-origin writeback suppression.

The required three-unlike-application admission gate failed. Fresh strict
generated code contains the exact

```text
runtime.Value -> native nominal -> compiled static call -> *_apply
```

shape in Manifest Normalization and Binary Event Log, but not in either of the
closest current controls, Concurrent Event Routing or Policy Record Dispatch.
Building a transitive mutation/escape effect system for only two applications
would violate the project's broad-evidence gate.

Machine-readable evidence is in
`2026-07-25-compiled-nominal-parameter-effect-admission-closure.json`.

## Current-source survey

The survey covered all 51 external benchmark applications with `run.able`
sources and inspected their 19 generic collection/result callback sites.
Four applications carry a nominal record through the closest matching
`Result.map` callback shape:

| Application | Captured callable carrier | Application-level runtime-origin writeback | Admission |
| --- | --- | ---: | --- |
| Manifest Normalization | erased `runtime.Value, runtime.Value -> NormalizedManifest` | 1 (`ManifestRecord`) | qualifies |
| Binary Event Log | erased `runtime.Value, runtime.Value -> i64` | 1 (`EventRecord`) | qualifies |
| Concurrent Event Routing | native `EventRecord, i64 -> AcceptedRoute` | 0 | disqualified |
| Policy Record Dispatch | native `PolicyRecord, i64 -> AcceptedDecision` | 0 | disqualified |

Only calls using `__able_runtime` from generated application control flow were
counted. Generated runtime ABI wrappers use separate `rt`-based writeback and
are required dynamic/host boundaries, not this candidate.

The two disqualified rows are not historical assumptions. The current
compiler's retained forward callable-binding inference gives both scorers
concrete generated Go signatures. Their nested callbacks therefore invoke
native record and integer carriers directly and never recover or write back a
runtime-origin record.

## Materiality and stopping rule

The immediately preceding exact profiles already establish materiality for
the two qualifying rows:

- Manifest performs 3,072 successful record recoveries and writebacks.
- Binary performs 53,248 successful record recoveries and writebacks.
- caller-owned recovery merely relocated those objects because the mandatory
  `*_apply` call made the recovered record escape.

No third current application has the same generated owner. Per the
authoritative PLAN gate, effect-summary implementation, semantic guard
construction, candidate profiling, and twenty-cohort A/B/Go measurement were
not started. Those later steps are warranted only after three unlike material
applications qualify.

This also closes the historical Policy route. An older generated artifact had
an erased Policy scorer and writeback, but a fresh current build does not.
Historical generated code cannot admit a present compiler optimization.

## Verification and provenance

- Manifest, Binary, Event Routing, and Policy were freshly built with the
  retained compiler under `--no-fallbacks`.
- All four public verifiers pass.
- All four final Go dependency graphs contain 96 packages and omit
  `able/interpreter-go/pkg/interpreter`.
- The fresh generated application sources contain two total candidate
  writebacks: one in Manifest and one in Binary.
- `go test ./cmd/ablec -count=1 -timeout=60s` passes in 5.515 seconds.
- No compiler, runtime, interpreter, VM, stdlib, language, dependency, or WASM
  code was changed.
- Retained compiler SHA-256:
  `8a64cddbb3c20b341ea20205c75257b558ac05cbdfe4369c06157a00381cc30e`.
- Raw artifacts:
  `/tmp/able-nominal-param-effects-20260725.ACWCNo`.

## Next

Refresh the verifier-backed compiled external scorecard from the current
retained compiler, then profile at least three unlike material target misses
and select their largest shared generated-code or generated-runtime owner.

This is next because the checked-in scoreboard stops before the many retained
2026-07-25 lowering improvements, while the current candidate inventory has
now exhausted the only known shared runtime-origin record writeback route.
Continuing from stale ratios or a two-program pattern would misdirect effort.

The work entails rebuilding the broad strict compiled application set,
repeating Able and equivalent Go measurements with averages, ranking current
misses by both ratio and material wall time, and collecting exact allocation
and CPU profiles for at least three unlike applications. A candidate advances
only if the same semantic owner appears in all three and permits a general
primitive/native-carrier, nominal, callable, interface, control-flow, or
runtime-boundary correction.

This is important because a current scoreboard reconnects the immediate work
to the actual goal: compiled Able should remain on native Go carriers and
approach or exceed equivalent Go performance across broad programs. It also
prevents retrying closed benchmark-specific, one-family, named-container,
execution-context ABI, or mandatory dynamic-boundary routes.

Do not begin WASM work.
