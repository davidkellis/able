# Compiled native String.split retained

Date: 2026-07-25

## Decision

Retain the guarded canonical primitive `String.split` call-site lowering.

An exact statically resolved
`able.text.string.String.split(String) -> Array String` call now evaluates its
receiver and delimiter once, validates both native Go strings, and returns a
fresh native `Array String` backed by `strings.Split`. Invalid UTF-8 falls
through to the unchanged canonical Able method. Bound-method, erased, dynamic,
user-defined, and otherwise unproven calls do not receive the lowering.

Twenty order-balanced baseline/candidate/Go cohorts improved all three unlike
applications: Concurrent Event Routing by 70.31%, Word Frequency by 14.39%,
and Sensor Calibration by 37.52%, for a 45.85% geometric-mean improvement.
Exact main-phase allocation bytes fell 20.96%-37.64%, while allocation counts
fell 41.22%-61.40%.

No benchmark, application, named container, or non-primitive nominal type is
selected by this rule. No canonical stdlib, generated runtime, interpreter,
tree-walker, bytecode VM, language, dependency, or WASM change was needed.

Machine-readable evidence is in
`2026-07-25-compiled-string-split-native-retained.json`.

## Lowering and semantic contract

The optimization is attached to the resolved canonical call site, not to the
generated method body. Admission requires all of the following:

- package `able.text.string`;
- primitive target `String`;
- method `split`;
- receiver and delimiter represented as Go `string`;
- return represented as `*__able_array_String`; and
- a concrete canonical definition and ordinary self receiver.

The native IIFE receives already-lowered receiver and delimiter expressions as
arguments, preserving one-time left-to-right evaluation. It takes the native
path only when `utf8.ValidString` accepts both values. Otherwise it invokes the
same compiled canonical method entry that the call used before this change.

This preserves:

- the canonical invalid-receiver and invalid-delimiter UTF-8 errors;
- Unicode-codepoint splitting for an empty delimiter;
- exact multibyte delimiter, leading, trailing, and repeated-delimiter
  behavior;
- a fresh mutable result Array for every invocation; and
- the original Able implementation for bound-method/dynamic compatibility.

Go strings are immutable, so result elements may safely share their immutable
backing bytes. `strings.Split` creates a new result slice, and each invocation
wraps it in a new native Able Array carrier.

## Repeated wall time

After two verified warmups per lane, twenty order-balanced
baseline/candidate/Go process cohorts ran on CPU 9 with `GOMAXPROCS=1`,
`GOMEMLIMIT=1GiB`, `GOGC=50`, and the application-appropriate executor. Every
timed process passed its public verifier.

| Application | Baseline mean | Candidate mean | Change | Go mean | Candidate / Go | Go performance |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Concurrent Event Routing | 100.081 ms | 29.714 ms | -70.31% | 3.151 ms | 9.431x | 10.60% |
| Word Frequency | 25.673 ms | 21.979 ms | -14.39% | 5.174 ms | 4.248x | 23.54% |
| Sensor Calibration | 22.538 ms | 14.082 ms | -37.52% | 3.036 ms | 4.638x | 21.56% |

The geometric-mean candidate/Go ratio is 5.707x, or 17.52% of Go performance.
This tranche materially narrows the gap but does not meet the 1.052632x
compiled target.

## Exact main-phase allocation

Three exact candidate samples per application were compared with the
immediately preceding retained three-sample means.

| Application | Bytes baseline -> candidate | Change | Allocations baseline -> candidate | Change | GC baseline -> candidate |
| --- | ---: | ---: | ---: | ---: | ---: |
| Concurrent Event Routing | 14,687,085 -> 11,608,056 | -20.96% | 264,834 -> 155,663 | -41.22% | 9.33 -> 7.33 |
| Word Frequency | 5,222,677 -> 3,347,325 | -35.91% | 135,165 -> 60,186 | -55.47% | 5.67 -> 3.00 |
| Sensor Calibration | 10,255,365 -> 6,394,875 | -37.64% | 324,801 -> 125,377 | -61.40% | 9.33 -> 6.00 |

The geometric-mean reductions are 31.90% for allocated bytes and 53.43% for
allocation count. The prior `slice_bytes` and validation-owned `utf8_decode`
leaves disappear from all three exact candidate profiles. The remaining
`strings.genSplit` allocations are the native result slice required by the
observable fresh `Array String`.

## CPU profiles and next-owner selection

Ten verified candidate main-phase CPU profiles were merged per application.
After the split pipeline disappeared, there was no new legal semantic CPU
owner common to all three:

- Event Routing is dominated by allocator/collector work, nominal/channel
  conversions, and scheduler operations.
- Word Frequency is dominated by the runtime-backed generic Map operations
  and their String-key conversion.
- Sensor Calibration is dominated by integer parsing/arithmetic, successful
  pattern-result conversion, and allocator work.

`bridge.ToString` remains an exact leaf in all three, but its different direct
parents were already closed in
`2026-07-25-compiled-bridge-to-string-caller-closure.md`. Reopening that leaf
without a new shared semantic parent would violate the evidence gate.

Event Routing and Sensor Calibration do share avoidable-looking conversion of
successful pattern-assignment results that are immediately error-checked and
discarded. Word Frequency does not exercise that route. A general
liveness-aware pattern-expression rule is therefore only eligible after a
third unlike strict application independently reproduces the same owner.

## Verification

- Generated-source guards require the exact canonical signature and native
  validity/split path while requiring the canonical method body to retain its
  Able validation and slicing implementation.
- Executable guards cover ASCII, leading/trailing/repeated delimiters,
  multibyte values, empty-delimiter codepoint behavior, multibyte delimiters,
  independent result mutation, invalid receiver and delimiter UTF-8, and a
  bound-method compatibility call.
- All 180 timed processes and all 18 warmups passed public verification.
- All 30 candidate CPU-profile and nine exact-allocation-profile processes
  passed public verification.
- All three fresh `--no-fallbacks` dependency graphs omit `pkg/interpreter`.
- Focused compiler semantic/source tests pass within the one-minute per-test
  bound.
- The broader focused compiler group passes in 36.140 seconds.
- `go test ./cmd/ablec -count=1 -timeout 60s` passes in 6.217 seconds.
- All touched compiler source and test files remain below 1,000 lines.

## Next

Find a third unlike current strict application whose exact profiles reproduce
the successful pattern-assignment result materialization already present in
Event Routing and Sensor Calibration. Only then evaluate one general
liveness-aware pattern-expression lowering that avoids converting the success
value to `runtime.Value` when the enclosing use observes only mismatch/error
control.

This is next because the freshly lowered split pipeline no longer supplies a
shared owner, while the pattern-result route is the largest clearly
unnecessary compiled/runtime conversion repeated in two unlike applications.
The work entails profiling the current strict suite to select a third
independent guard, proving whether the assignment expression's success value
is live, preserving mismatch errors and observable success results, adding
generated-source/executable guards, and running repeated three-application A/B
measurements against equivalent Go implementations.

This is important because it could keep native Arrays and nominal structs in
their generated Go representation through destructuring instead of boxing
them solely to produce a discarded expression result. The rule must be based
on general expression-result liveness, never on `String`, a named container,
or an application. If no third unlike application reproduces the owner, retain
no code and record the closure.

Do not begin WASM work.
