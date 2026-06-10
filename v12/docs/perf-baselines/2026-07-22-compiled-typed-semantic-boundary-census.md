# Compiled typed semantic-boundary census — 2026-07-22

## Decision

Retain no compiler execution, generated-runtime, bridge, canonical-stdlib,
bytecode, benchmark, language, or WASM performance change. Retain only the
opt-in `typed-boundary` diagnostic mode and its main-only audit support.

The five unlike applications carry 6.214 seconds of current target excess,
but no typed/runtime crossing is both CPU-material and present in three of
them. The high `control_from_error` count is a successful nil-error check that
Go almost entirely inlines away, not a hidden shared wall.

## Retained diagnostic contract

`ablec --typed-boundary-telemetry` adds twelve atomic counters to diagnostic
generated binaries. Normal generated output contains none of their variables,
markers, environment checks, branches, or atomics. The counters reset
immediately before the measured main phase and print one JSON object only when
`ABLE_COMPILER_TYPED_BOUNDARY_TELEMETRY` is set.

The categories cover conversion from arbitrary values, integer extraction,
struct/union/interface/callable conversion in both directions, and generated
control/error conversion in both directions. `bench_compiled_boundary_audit`
accepts `--telemetry typed-boundary` and preserves verifier status with the
counts.

These are execution-reach counters, not timers. Instrumented wall time is
never used as performance evidence because the atomics deliberately alter
inlining and cost.

## Current target cohort

The source is the current five-run compiled-versus-Go scorecard.

| Application | Able mean | Go mean | Ratio | Target excess |
| --- | ---: | ---: | ---: | ---: |
| Concurrent Event Routing | 2.914 s | 0.0056 s | 520.36x | 2.9081 s |
| Fixed Width 128 | 0.206 s | 0.0058 s | 35.52x | 0.1999 s |
| K-Nucleotide | 2.898 s | 0.0809 s | 35.82x | 2.8128 s |
| Distance Field | 0.090 s | 0.0133 s | 6.77x | 0.0760 s |
| Policy Record Dispatch | 0.226 s | 0.0087 s | 25.98x | 0.2168 s |

The cohort intentionally spans concurrency, wide nominal arithmetic, text/map
processing, primitive floating point, and regex/nominal dispatch. All five
telemetry processes completed under the 55-second cap and passed their public
verifiers.

## Main-only typed-boundary reach

| Category | Event | Fixed | K-Nucleotide | Distance | Policy | Breadth |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| any to runtime | 22,528 | 0 | 0 | 0 | 11,264 | 2 |
| integer from runtime | 152,591 | 0 | 24 | 0 | 114,048 | 2 material |
| struct from / to runtime | 27,660 / 40,968 | 0 | 0 | 0 | 7,680 / 14,336 | 2 |
| union from / to runtime | 11,269 / 15,364 | 0 | 14 / 0 | 0 | 3,585 / 5,632 | 2 material |
| interface from / to runtime | 0 | 0 | 0 | 0 | 0 | 0 |
| callable from / to runtime | 0 / 8,192 | 0 | 0 / 2 | 0 | 0 / 4,096 | 2 material |
| control from / to error | 170,408 / 0 | 0 | 12,333,418 / 0 | 0 | 62,988 / 0 | 3 |

All value, nominal, and callable categories that execute materially intersect
only Event Routing and Policy. Fixed Width and Distance Field execute none of
the measured boundary helpers. This rules out a universal runtime-value
crossing tax as their explanation.

The refreshed pre-existing views agree. Dynamic-boundary totals are 30
explicit calls, 9,259 residual polymorphic calls, 26 host calls, and four
runtime-service calls. Generic-union and fast-method counters both total
16,896 and occur only in Event Routing and Policy; fallback is zero.

## CPU materiality gate

`control_from_error` is the sole counter with apparent three-application
breadth, so it received a normal-build CPU gate. Five independent verified
main-only profiles per owner were pooled; no telemetry was present.

| Application | Pooled samples | Inline wrapper | with-node helper | Current dominant owner |
| --- | ---: | ---: | ---: | --- |
| K-Nucleotide | 15.26 s | 0.03 s / 0.20% | 0.01 s / 0.066% | primitive map equality/hash and GC/allocation |
| Policy Record Dispatch | 0.61 s | no samples | no samples | regex NFA bodies and GC/allocation |

K-Nucleotide's millions of calls are normally inlined success-path nil checks.
Their combined 0.262% sampled CPU is immaterial beside its 35.82x target gap.
Policy samples no control conversion at all. Because two of the three
count-reaching applications fail CPU materiality, profiling Event Routing
cannot restore the required three-program intersection; its known current
owner is the separately rejected goroutine-identity path.

## Why no candidate was admitted

- Counter volume alone cannot authorize an optimization. The only broad count
  is effectively free in normal binaries.
- Struct, union, callable, arbitrary-value, and integer crossings reach only
  two unlike applications here. Their exact descendants also reproduce the
  already-completed post-direct ABI and post-ABI profile gates.
- Fixed Width and Distance Field are important counterexamples: large gaps
  remain while every measured boundary count is zero. Another bridge or
  runtime-value fast path cannot address those rows.
- The surviving normal CPU owners divide into concurrency identity, wide
  nominal results, primitive map/text work, direct float code, and regex NFA
  storage. Naming those structures in compiler lowering would violate the
  shared nominal rule.

No A/B execution candidate was built, so there is no candidate timing claim
to average. The current five-run scorecard remains the wall-time authority.

## Verification

- Five typed-boundary, five dynamic-boundary, and five call-path processes all
  passed their public verifiers with zero timeout or failure.
- Ten normal CPU-profile processes passed their public verifiers.
- Focused normal/dynamic/call-path/typed telemetry compiler tests pass.
- The audit script passes shell syntax validation.
- Generated telemetry JSON is syntax-checked by the compiler test and by the
  five executed audit binaries.
- The malformed first diagnostic attempt and a later wrapper-only control
  count were excluded before selection; the companion JSON records only the
  corrected common-entry governing run.

## Next recommendation

Expand this new main-only typed-boundary census across the full 49-application
compiled scorecard before abandoning the boundary layer entirely.

Why: the representative high-excess cohort rejects a universal boundary tax,
but the new diagnostic has not yet tested every unlike application. A full
closed-world pass can either find a third independent user of one exact
boundary family or close this architecture level with stronger absence
evidence. This is new dynamic evidence, unlike the completed static
escape/bounds and emitted-helper censuses.

What it entails: run one bounded verifier-backed diagnostic process for each
portable compiled application; join counts to current target excess and prior
closed dispositions; profile only categories with material normalized reach
in at least three unlike applications; and build a candidate only when the
same primitive rule or shared nominal translation descendant is sampled in
all three. If no category passes, close the layer and return to the larger
bytecode architecture budget. Keep normal binaries telemetry-free, use five
independent complete A/B processes for any admitted candidate, and continue to
exclude named nominal/container/application rules and WASM.
