# Post-nullable compiled wide-numeric reconciliation

## Decision

Reconcile `compiled-wide-numeric` as causally current and retain no production
change.

The primitive nullable carrier does not reach Fixed Width 128, Rational
Series, or Wide Integer Records. All three generated modules emit dormant
nullable `runtime.Int128`/`runtime.Uint128` conversion support, but contain
zero construction sites and zero calls to those helpers.

Material primitive-wide arithmetic already uses direct two-word
`runtime.Int128` and `runtime.Uint128` Go carriers. The remaining
pointer-returning values are non-primitive nominal `UInt128`, `Int128`,
`Rational`, and result/record structures and must remain on the general
nominal lowering path.

## Strict boundary and execution gate

Every application was rebuilt from the retained compiler with
`--no-fallbacks`. Each exact binary passed its public Ruby verifier.

| Application | Packages | Interpreter dependency | Verified smoke |
| --- | ---: | --- | --- |
| Fixed Width 128 | 96 | absent | 1/1 |
| Rational Series | 96 | absent | 1/1 |
| Wide Integer Records | 96 | absent | 1/1 |

Smoke durations were 0.16, 0.06, and 0.15 seconds respectively. They are
execution checks, not timing evidence. The authoritative scorecard retains
five verifier-backed Able and Go processes per row.

## Whole-module nullable reach

Every complete generated module contains the same 12 textual references to
nullable wide support:

- type-switch cases for `__able_nullable[runtime.Int128]` and
  `__able_nullable[runtime.Uint128]`;
- `from_value`, `from_value_or_panic`, and `to_value` helper definitions for
  each type; and
- absent-value returns inside those helper definitions.

There are zero nullable-wide constructions outside those definitions and zero
helper call sites in every module. The support is dormant and cannot explain
any selected program's runtime or allocation profile.

Generated application bodies also contain zero primitive-wide nullable sites:

| Application | Material primitive-wide path |
| --- | --- |
| Fixed Width 128 | nominal `UInt128` methods convert fields to direct `runtime.Uint128`, perform native add/sub/compare, then fill general nominal results |
| Rational Series | Rational normalization and GCD pass direct `runtime.Int128` values through compare, multiply, and `DivMod`, then construct general `Rational` values |
| Wide Integer Records | nominal `UInt128`/`Int128` methods convert to direct primitive-wide carriers for multiply, add, remainder, min/max, compare, and stringification |

The selected paths do not call the stdlib `Result<T>` numeric-conversion
methods. Generated Result and DivMod declarations that are not reached by
these call chains remain dormant support, not evidence of a material carrier
boundary.

## Retained owner evidence

The retained three-application compiled profile gate used 30 separately
verified main-phase launches per row. It found
`sync/atomic.StorePointer` below package-environment `SwapEnv` at 16.85%,
10.40%, and 30.64% flat CPU for Fixed Width, Rational, and Wide Records.

That shared owner remains closed:

- a fixed execution-context design regressed unrelated N-Body wall time
  54.7%;
- a package-linkage variant regressed unrelated K-Nucleotide 16.6%; and
- the nullable representation change does not touch package publication.

The concrete numeric descendants remain distinct. Fixed Width is dominated by
loop-carried nominal UInt128 results and checked multiply. Rational is led by
primitive `Uint128.DivMod`, `Int128.DivMod`, GCD, and Rational construction.
Wide Records combines parsing, primitive-wide arithmetic, comparisons,
nominal wrapper results, and final stringification.

The quotient-only census likewise found Rational's wide `DivMod` work to be a
different primitive representation and operation mix from the material i32
helper in Sudoku. It does not establish a common new quotient-only lowering
candidate.

## Selective profile and candidate gate

Fresh profiling was not admitted:

- the changed nullable-wide carrier has zero material reach in all three
  applications;
- primitive-wide inputs and outputs already remain direct Go values across
  static calls;
- the repeated package-publication owner is unchanged and rejected by broad
  wall-time evidence; and
- optimizing nominal UInt128, Int128, Rational, or record results by name is
  forbidden.

No changed compiler or generated-runtime residual exists to measure.
Therefore no fresh CPU, allocation, timing, or A/B cohort was created.

## Current row state

The current five-process scorecard means remain:

| Application | Able compiled | Go | Able / Go |
| --- | ---: | ---: | ---: |
| Fixed Width 128 | 0.1060 s | 0.0059 s | 17.9661x |
| Rational Series | 0.0700 s | 0.0135 s | 5.1852x |
| Wide Integer Records | 0.0740 s | 0.0247 s | 2.9960x |

These are substantial product misses, but the nullable change supplies no
causal path to them. Primitive-wide boxing is not the residual owner.

## Scope and cleanup

No compiler, generated runtime, runtime package, interpreter, bytecode VM,
canonical stdlib, benchmark, language, dependency, or WASM source changed.
No UInt128, Int128, Rational, record, or other non-primitive nominal special
case was introduced.

`go test ./cmd/ablec` passed in 5.504 seconds. The machine-readable record is
`2026-07-30-post-nullable-compiled-wide-numeric-reconciliation.json`.

After retaining this evidence, the exact 472 MiB disk-backed generated
module, binary, audit, and Go-cache workspace was removed. No matching
tranche artifact remains in `/var/tmp` or `/tmp`.

## Next

Reconcile `compiled-byte-output` against the primitive nullable carrier.

Why: FASTA Generation and Reverse Complement are the smallest remaining
invalidated compiled family. Their material work uses native `u8` Arrays and
output paths directly covered by primitive and static-Array lowering, so
nullable-byte reach must be distinguished from ordinary bounds handling and
the explicit output boundary.

What it entails: strictly rebuild both rows, verify their graphs remain
interpreter-free, and audit generated Array reads/writes, byte transforms, and
`write_all` calls for nullable `u8` or runtime-value conversion. Reuse the
retained byte/output profiles. Because the closure contains only two rows,
advance no candidate unless the same exact general leaf also has current
evidence in a third unlike compiled application.

Why it matters: byte Arrays and bulk output should map closely to Go slices
and writes. Causal review can confirm that representation or identify a
general third-family question, while preventing a two-benchmark output rule
from bypassing the breadth requirement.
