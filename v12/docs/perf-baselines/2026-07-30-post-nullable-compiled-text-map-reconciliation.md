# Post-nullable compiled text/map reconciliation

## Decision

Reconcile `compiled-text-map` as causally current and retain no production
change.

The primitive nullable value carrier materially reaches seven of the nine
rows. Five reach it through a primitive-valued map result, Manifest
Normalization reaches `Option<String>`, and Sensor Calibration and Transaction
Ledger Audit reach `Result<i64>` matches. I Before E and Unicode Scalar
Pipeline have no application or material imported carrier reach.

No admissible post-carrier generated-code or generated-runtime owner repeats
in three unlike applications. No new CPU, allocation, or timing cohort was
warranted under the selective profile gate.

## Strict boundary and execution gate

Every application was rebuilt from the retained compiler with
`--no-fallbacks`. Each exact binary passed its public Ruby verifier.

| Application | Packages | Interpreter dependency | Verified smoke |
| --- | ---: | --- | --- |
| Backup Dedup | 96 | absent | 1/1 |
| I Before E | 96 | absent | 1/1 |
| Inventory Reconciliation | 96 | absent | 1/1 |
| K-Nucleotide | 96 | absent | 1/1 |
| Manifest Normalization | 96 | absent | 1/1 |
| Sensor Calibration | 96 | absent | 1/1 |
| Transaction Ledger Audit | 96 | absent | 1/1 |
| Unicode Scalar Pipeline | 96 | absent | 1/1 |
| Word Frequency | 96 | absent | 1/1 |

Smoke durations are execution checks, not timing evidence. The authoritative
scorecard retains five verifier-backed Able and Go processes per row.
Inventory Reconciliation and Transaction Ledger Audit also retain balanced
five-process nullable baseline/candidate/Go cohorts and three independent
exact allocation measurements per side.

## Causal carrier reach

Generated support definitions were excluded. Every generated module contains
primitive nullable conversion helpers; only references from application or
material imported paths count as causal reach.

| Application | Material reach | Exact reached path |
| --- | --- | --- |
| Backup Dedup | yes | `chunk_count -> HashMap<i64,i64>.raw_get -> __able_nullable[int64]` |
| I Before E | no | generated application callable has no primitive nullable reference |
| Inventory Reconciliation | yes | `lookup(Map<i64,i64>) -> Map.get -> HashMap.raw_get -> __able_nullable[int64]` |
| K-Nucleotide | yes | `increment_count -> HashMap<u64,i32>.raw_get -> __able_nullable[int32]` |
| Manifest Normalization | yes | `optional_owner -> Option<String> -> __able_nullable[string]` |
| Sensor Calibration | yes | three successful `Result<i64>` matches in `parse_reading` bind through `__able_nullable[int64]` |
| Transaction Ledger Audit | yes | `parse_transaction` binds a `Result<i64>` match; `lookup(Map<String,i64>)` reaches `HashMap.raw_get` |
| Unicode Scalar Pipeline | no | application and material imported paths contain no primitive nullable reference |
| Word Frequency | yes | `lookup(Map<String,i32>) -> Map.get -> HashMap.raw_get -> __able_nullable[int32]` |

I Before E initially appeared to reach `__able_nullable[int32]` because the
shared generated support file defines conversions for every primitive width.
Its generated package callable contains no such reference, so that dormant
support definition is not material reach.

## Retained direct A/B

The two closure rows in the nullable carrier's direct retained gate remain
the causal performance evidence:

| Application | Baseline | Candidate | Change | Equivalent Go |
| --- | ---: | ---: | ---: | ---: |
| Inventory Reconciliation | 0.1220 s | 0.1100 s | -9.84% | 0.0079 s |
| Transaction Ledger Audit | 0.0480 s | 0.0400 s | -16.67% | 0.0060 s |

All values are five-process verifier-backed means. Their three-run exact
allocation means are:

| Application | Bytes baseline | Bytes candidate | Objects baseline | Objects candidate | Object change |
| --- | ---: | ---: | ---: | ---: | ---: |
| Inventory Reconciliation | 17,037,123 | 14,874,320 | 553,060 | 282,723 | -48.88% |
| Transaction Ledger Audit | 5,784,275 | 5,702,339 | 115,269 | 105,029 | -8.88% |

Inventory's 135,172 nullable recovery allocations and Transaction's 5,122
nullable recovery allocations become zero. Inventory retains 270,336
`bridge.ToDynamicI64` allocations and Transaction retains 9,988 dynamic
conversions. Those surviving operations are explicit map-kernel dynamic
boundaries rather than missed primitive-nullable lowering.

## Residual-owner admission

The seven reached rows do not expose an admissible common successor:

- Backup Dedup, Inventory Reconciliation, K-Nucleotide, Transaction Ledger
  Audit, and Word Frequency share map-kernel handle/key/value conversions.
  Their common concrete caller is the named stdlib `HashMap` boundary. A
  compiler rule for that nominal type is forbidden, and the retained Generic
  Slot Buffer evidence already proves that user-defined generic nominal
  storage and interfaces can stay on direct native Go carriers.
- Sensor Calibration and Transaction Ledger Audit share parsing, splitting,
  integer conversion, and primitive `Result` match work, but that exact
  residual stops at two unlike applications.
- Manifest Normalization reaches a static `Option<String>` carrier. Its
  retained profile is led by String/nominal work, not the parsing-plus-`ToInt`
  path in the two record parsers.
- The broader generic-union/match, String conversion, checked-arithmetic,
  nominal-layout, call/member, allocation/GC, and execution-context routes
  have already completed broad gates. The carrier representation change does
  not create a new exact child shared by three unlike rows.

`runtime.mallocgc` is only a common consequence above distinct operations.
The same is true of aggregate bridge ancestry. Reprofiling map-heavy rows
would manufacture breadth from one named-container boundary, while
reprofiling the two parsers would still fail the required three-application
bar. The selective gate therefore correctly retained zero new profiles and no
speculative candidate.

## Current row state

The current full-scorecard snapshots remain:

| Application | Able compiled | Go | Able / Go |
| --- | ---: | ---: | ---: |
| Backup Dedup | 0.0720 s | 0.0110 s | 6.5455x |
| I Before E | 0.0680 s | 0.0634 s | 1.0726x |
| Inventory Reconciliation | 0.1400 s | 0.0095 s | 14.7368x |
| K-Nucleotide | 1.5920 s | 0.0669 s | 23.7967x |
| Manifest Normalization | 0.0400 s | 0.0050 s | 8.0000x |
| Sensor Calibration | 0.0580 s | 0.0056 s | 10.3571x |
| Transaction Ledger Audit | 0.0680 s | 0.0075 s | 9.0667x |
| Unicode Scalar Pipeline | 0.1200 s | 0.0107 s | 11.2150x |
| Word Frequency | 0.0480 s | 0.0052 s | 9.2308x |

The scorecard snapshots are not substituted for the direct nullable A/B.
The balanced retained A/B is the causal optimization gate; the full scorecard
remains the product-wide snapshot.

## Scope

No compiler, generated runtime, runtime package, interpreter, bytecode VM,
canonical stdlib, benchmark, language, dependency, or WASM source changed.
No named-container, non-primitive nominal, or benchmark-specific rule was
introduced.

The machine-readable record is
`2026-07-30-post-nullable-compiled-text-map-reconciliation.json`.
After verification, the exact 1,164 MiB disk-backed generated-module, binary,
compiler, and Go-cache workspace was removed. No matching tranche artifact
remains in `/var/tmp` or `/tmp`.

## Next

Reconcile `compiled-regex` against the retained nullable carrier.

Why: its six rows share the canonical NFA owner, whose thread, capture, and
state APIs plausibly return optional primitive positions. This is the
smallest remaining invalidated closure where carrier reach could affect one
already broad owner rather than several unrelated application bodies.

What it entails: strictly rebuild all six regex rows, verify that every graph
omits the interpreter, and distinguish actual NFA/capture carrier use from
dormant generated helpers. Attach the retained NFA profiles and reprofile only
if the changed carrier reaches a material owner whose exact residual remains
open. Preserve the previously rejected arena, state-index, character,
carrier/call, and application-specific alternatives.

Why it matters: if nullable positions still cross a dynamic or boxed boundary
inside the shared regex implementation, a general static carrier correction
could benefit six unlike applications. If not, causal reconciliation prevents
us from repeating closed NFA experiments and advances the evidence ledger
toward a trustworthy next optimization.
