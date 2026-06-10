# Compiled discarded pattern-result liveness retained

Date: 2026-07-25

## Decision

Retain the general discarded pattern-assignment result lowering.

When a pattern assignment is compiled in a context that discards its result,
the compiler now evaluates the right-hand side once, performs the complete
pattern test, commits bindings only on success, and transfers the same mismatch
error on failure without first converting the successful right-hand value to
`runtime.Value`.

Pattern assignments used as expressions or conditions retain their original
success value and error-valued mismatch semantics. The rule is based only on
expression-result liveness. It does not name Array, String, a nominal type, a
stdlib container, or an application.

Twenty order-balanced baseline/candidate/Go cohorts improved all three unlike
applications: Concurrent Event Routing by 16.51%, Sensor Calibration by
27.11%, and Manifest Normalization by 7.57%, for a 17.45% geometric-mean
improvement. Exact main-phase allocation bytes fell 14.42%-36.83%, while
allocation counts fell 22.32%-52.07%.

No canonical stdlib, generated runtime, interpreter, tree-walker, bytecode VM,
language, dependency, or WASM change was needed.

Machine-readable evidence is in
`2026-07-25-compiled-discarded-pattern-result-liveness-retained.json`.

## Three-application admission

The previous exact caller analysis found that Event Routing and Sensor
Calibration converted successful native `Array String` pattern subjects to
runtime Arrays only so a surrounding statement could probe the result for a
mismatch error and then discard it.

A current benchmark-source census found two other applications with the same
five-field pattern:

- Concurrent Policy Callbacks executes it only 32 times during input loading;
  this is not material enough to govern an optimization.
- Manifest Normalization executes it 4,096 times in a serial
  Option/Result/captured-callback workload.

A fresh strict Manifest binary verified and omitted `pkg/interpreter`. Its
exact main allocation profile attributed 19,072 `bridge.ToString` objects to
`__able_array_String_to`; the complete Array conversion owned 26,752
allocations. Generated `parse_manifest` then probed the runtime result for an
error and discarded it. This is the same semantic owner as Event and Sensor,
not merely the same leaf name.

The admitted applications are unlike:

- Event Routing is a four-worker channel/Result/interface application.
- Sensor Calibration is a serial numeric validation and error-union
  application.
- Manifest Normalization is a serial Option/Result/captured-callback
  application.

## General lowering

`compileAssignmentMode` already knows whether a surrounding statement needs an
assignment result. Non-simple patterns now pass that fact into the shared
pattern-assignment lowering.

For a discarded result:

1. The RHS is evaluated exactly once into its existing native or runtime
   carrier.
2. All declarations are staged before the match, as before.
3. Pattern conditions run before binding assignments, preserving atomic
   mismatch behavior.
4. A mismatch transfers
   `runtime.ErrorValue{Message: "pattern assignment mismatch"}` through the
   existing control envelope.
5. A success commits the bindings but does not synthesize a result temporary,
   lower the subject to `runtime.Value`, or probe that value for an error.

Observed pattern assignments continue through the old result-producing path.
The pattern implementation was moved to
`generator_pattern_assignments.go` so both touched generator files remain
below 1,000 lines.

## Repeated wall time

After two verified warmups per lane, twenty order-balanced
baseline/candidate/Go process cohorts ran on CPU 9 with `GOMAXPROCS=1`,
`GOMEMLIMIT=1GiB`, `GOGC=50`, and the application-appropriate executor. Every
timed and warmup process passed its public verifier.

| Application | Baseline mean | Candidate mean | Change | Go mean | Candidate / Go | Go performance |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Concurrent Event Routing | 32.681 ms | 27.286 ms | -16.51% | 3.333 ms | 8.188x | 12.21% |
| Sensor Calibration | 21.068 ms | 15.356 ms | -27.11% | 4.458 ms | 3.445x | 29.03% |
| Manifest Normalization | 15.484 ms | 14.312 ms | -7.57% | 2.931 ms | 4.882x | 20.48% |

The geometric-mean candidate/Go ratio is 5.164x, or 19.37% of Go
performance. This tranche removes one broad compiled/runtime boundary, but all
three applications remain outside the 1.052632x compiled target.

## Exact main-phase allocation

Three exact baseline and candidate samples were captured per application.

| Application | Bytes baseline -> candidate | Change | Allocations baseline -> candidate | Change | GC baseline -> candidate |
| --- | ---: | ---: | ---: | ---: | ---: |
| Concurrent Event Routing | 11,607,965 -> 8,960,771 | -22.81% | 155,660 -> 104,600 | -32.80% | 6.67 -> 6.33 |
| Sensor Calibration | 6,394,939 -> 4,039,739 | -36.83% | 125,377 -> 60,097 | -52.07% | 6.00 -> 3.33 |
| Manifest Normalization | 6,805,400 -> 5,824,403 | -14.42% | 119,878 -> 93,126 | -22.32% | 6.33 -> 5.00 |

The geometric-mean reductions are 25.27% for allocated bytes and 36.99% for
allocation count. The selected `__able_array_String_to` pattern-result owner
disappears from all three candidate parsers. Event retains a separate
`bridge.ToString` channel-payload boundary; Manifest retains nominal Result
callback and nullable String conversions; Sensor retains neither String
conversion.

## Candidate CPU profiles and next-owner selection

Ten verified candidate main-phase CPU profiles were merged per application.
The short processes make individual 10 ms samples coarse, but the selected
pattern-result conversion is absent from generated source and exact allocation
profiles as well as CPU stacks.

There is no new legal semantic owner shared by these same three applications.
The exact profiles do expose a narrower two-application route:

- Event Routing converts a native `EventRecord` while a nested `Result.map`
  callback invokes the captured `scorer`.
- Manifest Normalization converts a native `ManifestRecord` while a nested
  `Result.map` callback invokes the captured `normalizer`.

The generated `Result.map` specializations themselves already use native union
and callable carriers. The conversion occurs because the outer captured
callable remains erased inside the nested lambda. Binary Event Log provides an
independent source-level third candidate with the same captured-scorer shape,
but it must be freshly profiled before any implementation is admissible.

## Verification

- Generated-source guards require discarded Array and struct patterns to omit
  their runtime result conversions while requiring an observed conditional
  pattern to retain its result conversion and truthiness check.
- Executable guards cover a computed RHS evaluated once, nested struct and
  Array patterns, successful binding, observed conditional success, standalone
  mismatch propagation, and rescue handling.
- All 180 timed and 18 warmup processes passed public verification.
- All 30 candidate CPU-profile and 18 baseline/candidate exact-allocation
  profile processes passed public verification.
- All three fresh `--no-fallbacks` dependency graphs omit `pkg/interpreter`.
- All pattern-focused compiler tests pass in 6.145 seconds.
- Six no-bootstrap assignment-pattern fixtures pass in 7.782 seconds.
- `go test ./cmd/ablec -count=1 -timeout 60s` passes in 5.645 seconds.
- `generator_assignments.go` is 852 lines and
  `generator_pattern_assignments.go` is 193 lines.

## Next

Refresh exact profiles across Concurrent Event Routing, Manifest
Normalization, and Binary Event Log to qualify the captured-callable carrier
round trip inside nested generic callbacks.

This is next because Event and Manifest now show the same remaining semantic
sequence: an already-native nominal record enters an erased captured callable
from inside a fully native generic method, forcing a runtime conversion and
immediate recovery. Binary Event Log is an unlike binary-parsing application
with the same source shape and can supply or reject the third evidence point.

The work entails tracing the captured callable's inferred signature and escape
uses, proving whether one concrete carrier is valid for every capture,
preserving erased carriers for dynamic/conflicting uses, and admitting at most
one general captured-callable specialization rule. Any candidate must pass
callable escape/dynamic guards, strict dependency checks, exact profiles, and
repeated three-application A/B/Go measurement.

This is important because closures should capture a concrete generated Go
function when all of their statically known uses agree, rather than boxing
native arguments solely because the callable crosses a nested-lambda boundary.
The rule must be generic across callable and nominal types and must not name
`Result`, Event records, manifests, or any container. If Binary Event Log does
not reproduce the exact owner, retain no code and document the closure.

Do not begin WASM work.
