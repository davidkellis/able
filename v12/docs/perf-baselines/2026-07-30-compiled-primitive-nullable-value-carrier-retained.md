# Compiled primitive nullable value carrier retained

## Decision

Retain the general compiled primitive-nullable value carrier. Primitive
nullable scalars now lower to `__able_nullable[T] { value, valid }` instead of
`*T`. This removes heap boxes while preserving absent versus present-zero
semantics. `?Error` and nullable non-primitive nominal values retain their
existing representations.

This is a primitive compiler rule. It adds no application, benchmark,
container, or non-primitive nominal special case and does not widen the
compiled/interpreter boundary.

## Repeated application gate

All rows are arithmetic means of five successful verifier-backed processes on
CPU 15. Baselines use a frozen pre-change compiler. Candidate builds use
`--no-fallbacks`.

| Application | Baseline | Candidate | Change | Equivalent Go |
| --- | ---: | ---: | ---: | ---: |
| Generic Slot Buffer | 0.0560 s | 0.0340 s | -39.29% | 0.0046 s |
| Inventory Reconciliation | 0.1220 s | 0.1100 s | -9.84% | 0.0079 s |
| Transaction Ledger Audit | 0.0480 s | 0.0400 s | -16.67% | 0.0060 s |

Every Able and Go process passed its public verifier. All three generated
strict dependency graphs omit `pkg/interpreter`.

Evidence:

- `2026-07-29-nullable-scalar-baseline-generic-slot-buffer.json`
- `2026-07-29-nullable-scalar-candidate-generic-slot-buffer.json`
- `2026-07-29-nullable-scalar-baseline-inventory-reconciliation.json`
- `2026-07-29-nullable-scalar-candidate-inventory-reconciliation.json`
- `2026-07-29-nullable-scalar-baseline-transaction-ledger-audit.json`
- `2026-07-29-nullable-scalar-candidate-transaction-ledger-audit.json`
- `2026-07-29-nullable-scalar-go-references.{json,md}`

## Exact main-phase allocations

Each value below is the arithmetic mean of three independent exact phase
measurements. Allocation counts were deterministic except for a two-object
baseline variation in Inventory Reconciliation.

| Application | Bytes baseline | Bytes candidate | Objects baseline | Objects candidate | Object change |
| --- | ---: | ---: | ---: | ---: | ---: |
| Generic Slot Buffer | 2,123,232 | 17,872 | 264,215 | 1,046 | -99.60% |
| Inventory Reconciliation | 17,037,123 | 14,874,320 | 553,060 | 282,723 | -48.88% |
| Transaction Ledger Audit | 5,784,275 | 5,702,339 | 115,269 | 105,029 | -8.88% |

The prior exact owners disappear:

- Generic Slot Buffer: 131,842 flat
  `VersionedBuffer_get_spec` nullable constructions become zero.
- Inventory Reconciliation: 135,172 flat
  `__able_nullable_i64_from_value` allocations become zero.
- Transaction Ledger Audit: 5,122 nullable recovery allocations become zero.

Inventory Reconciliation retains 270,336 `bridge.ToDynamicI64` allocations and
Transaction Ledger Audit retains 9,988 dynamic conversions. Those are
explicit dynamic-map boundaries and are outside this static carrier rule.

Generated candidate code contains `__able_nullable[int64]` and
`__able_some(slot.Value)` rather than `__able_ptr(slot.Value)`.

## Correctness and boundary verification

The retained tests cover:

- absent and present-zero distinction;
- primitive integer, wide integer, float, character, string, and Boolean
  carriers;
- nil and nullable equality;
- pattern matching and binding;
- `or`, propagation, and safe navigation;
- static Array methods;
- generic interfaces and Result/Option specialization;
- broad native-union and imported/shadowed alias paths;
- runtime/dynamic conversion;
- the unchanged `?Error` pointer carrier.

`go test ./cmd/ablec` passed. The complete `./run_all_tests.sh` pass completed
all preflights, non-compiler packages, 34 compiler batches, and bytecode
fixtures. Aggregate compiler batches over one minute were audited per test;
the slowest constituent was 26.60 seconds, below the one-minute rule.

After retaining the readable evidence, 442 MiB of exact disk-backed
`/var/tmp/able-nullable-scalar-*` build/profile workspaces and the generated
v12 Python cache were removed. No matching active artifact remains.

## Scorecard reconciliation

The full 65-application, two-mode scorecard completed 650/650 Able executions
and 650/650 reference executions with verifier success. One `pidigits`
compiled row reported an anomalous 2.224-second mean. Frozen A/B replays did
not reproduce it:

| Affinity | Frozen baseline | Candidate | Result |
| --- | ---: | ---: | --- |
| CPU 15 | 1.414 s | 1.272 s | candidate 10.04% faster |
| CPU 12 | 1.258 s | 1.276 s | neutral within workstation noise |

The reconciled row is a fresh five-run CPU-12 candidate mean of 1.234 seconds
versus Go at 1.2218 seconds, which meets the target. `i_before_e` repeated at
0.0680 seconds versus Go at 0.0634 seconds and was conservatively removed from
the established-guard set.

The current frontier therefore has:

- 130 selected rows;
- 6 compiled and 4 bytecode established guards;
- 120 misses and no unestablished snapshot meets;
- zero actionable frontier groups.

The compiler-production identity change intentionally invalidates 12 compiled
or cross-family closure records. They were not advanced wholesale because the
three-application nullable evidence does not re-prove unrelated ownership
families.

Reconciled evidence:

- `2026-07-30-nullable-scalar-retained-scorecard-reconciled.{json,md}`
- `2026-07-30-nullable-scalar-retained-frontier.{json,md}`
- `2026-07-30-nullable-scalar-retained-generality-compiled-03-replay.{json,md}`
- `2026-07-30-nullable-scalar-retained-generality-compiled-04-replay.{json,md}`
- `2026-07-29-nullable-scalar-pidigits-{baseline,candidate}.json`
- `2026-07-29-nullable-scalar-pidigits-cpu12-{baseline,candidate}.json`
- `2026-07-21-performance-evidence-invalidation-ledger.{json,md}`

## Scope

No canonical stdlib, tree-walker, bytecode VM, runtime package, language,
dependency, or WASM change was required. The generated nullable helper is
emitted by the compiler. The v12 spec is unchanged because the observable
nullable semantics did not change.

## Next

Refresh current exact CPU and allocation ownership across at least three
unlike compiled misses selected from the invalidated closures.

Why: the nullable carrier materially changed the compiler-production identity
and removed one shared allocation family. Old compiled closure profiles can no
longer establish the largest remaining owner.

What it entails: choose broad current misses from distinct feature families,
build them strict and interpreter-free, collect repeated main-only CPU and
exact allocation profiles, and select only the largest exact generated-code
or generated-runtime owner that repeats materially in all three. Any candidate
must again pass focused semantic guards and balanced verifier-backed A/B/Go
measurements.

Why it matters: this resumes the direct path toward native-Go compiled
performance from the post-carrier state without reopening closed routes or
mistaking dynamic boundaries for removable static boxing.
