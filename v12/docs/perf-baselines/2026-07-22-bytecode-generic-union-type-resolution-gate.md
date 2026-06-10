# Bytecode generic-union type-resolution gate

Date: 2026-07-22

## Decision

Keep no bytecode VM, interpreter, compiler, runtime, canonical-stdlib,
benchmark, fixture, language, or WASM change.

Fresh profiles do establish one exact shared bytecode leaf across three unlike
applications and two union definitions: static generic-union method matching
is 5.18% cumulative in Binary Event Log, 5.93% in Option/Result Config, and
2.56% in Manifest Normalization. Binary uses `Result`; the other applications
exercise both `Option` and `Result`. Policy Record Dispatch is the non-owner
guard.

Two generic implementations failed the broad gate:

1. A bounded, method-versioned map cache removed the matcher from the Option
   profile and reduced allocations, but its ten verifier-backed Option
   processes averaged 0.933 seconds against a 0.874-second midpoint of the
   surrounding controls (+6.75%). The lookup cost replaced more work than it
   saved.
2. Copy-on-write reuse of unchanged immutable generic, nullable, result,
   union, and function type-expression trees reduced owner allocations by
   1.86%-4.20% and improved Binary Event Log, but an order-balanced unrelated
   iterator guard regressed 2.48%. That is a real broad-workload cost, so the
   allocation win is not retained.

This closes global map caching and recursive identity-reuse as mechanisms for
this leaf. The result does not close a lower-overhead instruction-local plan,
but such a plan needs a fresh dynamic census before implementation.

## Protocol

- Used canonical external `../able-stdlib/src`, `GOMAXPROCS=1`, `GOGC=50`,
  `GOMEMLIMIT=1GiB`, `--source-root-only`, and a 55-second cap per process.
- Retained five independent verifier-backed direct-bytecode processes for each
  application cohort. All 70 direct processes verified; none failed or timed
  out.
- Used separately built warmed runtime harnesses for attribution and allocation
  measurements. Every successful process is retained and averaged.
- One initial Binary Event Log four-call profile calibration exceeded 55
  seconds. It produced no result and was replaced by the successful bounded
  two-call profile; the timeout remains recorded rather than silently dropped.
- Raw cleanup-eligible artifacts are under
  `v12/.profiles/20260722_bytecode_union_cohort/`; the checked-in JSON sibling
  is the durable aggregate.

## Baseline profiles

| Exact leaf | Binary Event Log | Option/Result Config | Manifest Normalization | Policy guard |
| --- | ---: | ---: | ---: | ---: |
| `resolveStaticGenericUnionMethodCallable` cumulative | 5.18% | 5.93% | 2.56% | not material |
| `staticGenericUnionMethodMatches` cumulative | 4.45% | 5.42% | 1.95% | not material |
| `genericUnionMethodMatchesStaticReceiver` cumulative | 4.25% | 5.27% | 1.95% | not material |
| `NewUnionTypeExpression` cumulative | 1.59% | 2.43% | 0.60% | not material |

The shared parent therefore passed the three-application/two-definition
admission rule. Its two plausible descendants required separate gates: repeat
matching and repeat construction of unchanged type-expression trees.

## Verifier-backed process results

| Application | Initial mean | Final copy-on-write mean | Raw change | Verification |
| --- | ---: | ---: | ---: | ---: |
| Binary Event Log | 6.792 s | 6.552 s | -3.53% | 10/10 |
| Option/Result Config | 0.824 s | 0.886 s | +7.52% | 10/10 |
| Manifest Normalization | 1.652 s | 1.706 s | +3.27% | 10/10 |
| Policy Record Dispatch | 7.152 s | 7.178 s | +0.36% | 10/10 |

These non-interleaved process blocks expose workstation drift: a later restored
Option control was 0.924 seconds, making the surrounding-control midpoint
0.874 seconds and the final candidate +1.37%. Selection therefore used the
warmed and order-balanced measurements below rather than claiming that every
raw before/after shift was caused by the patch.

## Warmed copy-on-write owner gate

| Application | Restored/control ns/op | Candidate ns/op | Time change | Bytes change | Allocation change |
| --- | ---: | ---: | ---: | ---: | ---: |
| Binary Event Log | 7,552,184,210 | 7,365,815,022 | -2.47% | -4.20% | -4.12% |
| Option/Result Config | 1,024,979,221 | 997,634,756 | -2.67% | -2.56% | -2.80% |
| Manifest Normalization | 1,674,600,130 | 1,681,365,684 | +0.40% | -1.86% | -2.32% |

Binary uses two measured calls per process, Option twenty, and Manifest eight.
Option's control is the mean of the initial and restored processes. Manifest
uses four candidate and three control processes. This confirms a broad owner
allocation reduction with no material Manifest wall change.

## Unrelated control gate

| Workload | Copy-on-write ns/op | Restored/current ns/op | Change | Decision |
| --- | ---: | ---: | ---: | --- |
| String split/join | 1,088,370,240 | 1,092,962,988 | -0.42% | neutral |
| Iterator collect | 445,385,779 | 434,601,933 | +2.48% | reject |
| Numeric Array map | 77,694,018 | 77,668,204 prior control | +0.03% | neutral |

Each row retains five warmed processes. The initial historical split and
iterator controls were 3%-4% faster than both adjacent current blocks, while
the candidate/restored comparison reverses that drift for split and isolates a
2.48% iterator cost. Iterator allocation changes were negligible (18 objects
per call), so there is no offsetting generic benefit. This guard decides the
gate.

## Verification

Focused generic-union/member-resolution behavior and the fixture active when
the broad package run reached its cap pass:

```text
go test ./pkg/interpreter -run 'TestOptionResultConfigGenericUnionMethodsRemainResolvedWhenWarm|TestBytecodeVM.*Member.*Cache|TestExecFixtures/06_11_truthiness_boolean_context' -count=1 -timeout 60s
ok able/interpreter-go/pkg/interpreter 3.981s
```

An attempted unfiltered `go test ./pkg/interpreter -count=1 -timeout 60s`
reached the repository cap while `TestExecFixtures/06_11_truthiness_boolean_context`
was active. The isolated fixture passes above; there was no assertion failure.

## Next recommendation

Run a stats-only census for an instruction-local monomorphic generic-union
method plan across Binary Event Log, Option/Result Config, and Manifest
Normalization, with Policy and the three unrelated shapes as guards.

Why: the exact matcher remains 2.56%-5.93% cumulative in three unlike owners,
and the rejected global cache proves that eliminating it is possible but a
generic map lookup is too expensive. An instruction-local identity/version
guard could retain the result without hashing or reflection, but only if the
same bytecode call site repeatedly sees one checked receiver expression, one
scope owner, and one method-cache version.

What it entails: add opt-in counters only, measure call-site stability,
binding-changing misses, lexical-owner changes, and potential monomorphic hit
rates in the three owners plus guards, then remove the counters or turn them
into a candidate only if one general rule is stable and material in all three.
Any implementation must preserve lexical shadowing and method invalidation,
name no nominal type, and clear the same verifier-backed broad gate. Continue
to defer WASM.
