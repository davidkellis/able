# Current-default three-application owner closure

Date: 2026-07-31

## Decision

Retain no production code.

Fresh interpreter-free CPU and allocation profiles for K-Nucleotide, Versioned
Telemetry Pipeline, and Inventory Reconciliation expose no exact general
compiler or generated-runtime leaf that is material in all three.

K-Nucleotide and Inventory Reconciliation share primitive boxing, hashing,
equality, and lookup work at the explicit runtime-backed HashMap boundary.
Versioned Telemetry Pipeline does not cross that boundary in its hot path:
98.89% of its exact allocation-shape objects are fresh `Sample` pointers
stored directly in a native specialized Array carrier.

Allocation and GC are material in all three, but their exact descendants are
different. Treating that aggregate parent as one candidate would hide a named
container rule for two rows and a non-primitive nominal-storage rule for the
third. Checked arithmetic is dominant only in Telemetry, is small in Inventory,
and is already closed.

No candidate therefore entered an A/B cohort. No compiler, generated runtime,
runtime package, interpreter, bytecode VM, canonical stdlib, language,
dependency, benchmark, fixture, frozen workspace, or WASM source changed.

## Current products and controls

All three applications were regenerated with the current default compiler and
`--no-fallbacks`. Each public verifier passed. Each final Go dependency graph
contains 96 dependencies and omits the exact package
`able/interpreter-go/pkg/interpreter`.

| Application | Current Able / Go | Ratio | Binary bytes | Binary SHA-256 |
| --- | ---: | ---: | ---: | --- |
| K-Nucleotide | 1.5000 / 0.0620 s | 24.194x | 15,513,168 | `827766a3...8e4a` |
| Versioned Telemetry Pipeline | 2.2500 / 0.2057 s | 10.938x | 12,970,680 | `a6ba381d...a3d9` |
| Inventory Reconciliation | 0.1400 / 0.0092 s | 15.217x | 10,578,648 | `612a7ca3...baa` |

The generated products used `ablec` SHA-256
`ff4dfcc2d98117f1f48e177649f3b5a2a4e593a343e034563de60d38d9c04d74`
and Go 1.26.5. Serial execution used CPU 12, `GOMAXPROCS=1`, `GOGC=50`,
`GOMEMLIMIT=1GiB`, and a 55-second process bound.

## Repeated CPU ownership

Only each generated launcher's registered main phase was profiled. The short
Inventory row required many independent launches to accumulate enough 100 Hz
samples.

| Application | Able profiles | Merged samples | Largest current owners |
| --- | ---: | ---: | --- |
| K-Nucleotide | 20 | 41.51 s | primitive HashMap equality 13.01% flat; hash 6.53%; find-entry 4.70%; `ToUint` 26.04% cumulative; `ToInt` 15.20%; allocation 30.57% |
| Versioned Telemetry Pipeline | 10 | 25.37 s | checked multiply 24.24% flat; divmod 16.91%; checked add 11.63%; allocation 19.31% cumulative; policy adapters 7.96% and 5.68% |
| Inventory Reconciliation | 100 | 10.43 s | primitive HashMap equality 13.71% flat; hash 8.53%; find-entry 6.04%; `ToDynamicI64` 21.86% cumulative; allocation 14.38% |

All 130 Able CPU processes passed their public verifiers and reproduced one
stable output hash per application.

Five Go reference profiles per application also passed the same verifier.
The profiling wrapper repeated only the unchanged reference main and discarded
intermediate output: 100 calls for K-Nucleotide, 10 for Telemetry, and 200 for
Inventory. Their merged samples were 27.20, 9.90, and 4.78 seconds.

The equivalent Go programs spend their sustained work in native
`mapassign`/`mapaccess` for the two map applications and direct value-struct
logic for Telemetry. There is no interpreter execution in any Able row.

## Exact main allocations

Three lightweight main-phase `MemStats` deltas per Able and Go product provide
totals without one-object-profile serialization overhead:

| Application | Able mean bytes / objects / GC | Go mean bytes / objects / GC |
| --- | ---: | ---: |
| K-Nucleotide | 598,183,514.67 / 12,232,486 / 274.67 | 4,959,208 / 85 / 2 |
| Versioned Telemetry Pipeline | 430,788,288 / 13,325,303 / 351.33 | 17,096 / 21 / 0 |
| Inventory Reconciliation | 14,875,450.67 / 282,727.67 / 11 | 296,000 / 41 / 0 |

One separate `runtime.MemProfileRate=1` shape process per Able application
passed its verifier. Start/end subtraction includes the known profile-writer
overhead, so the lightweight counters above determine totals and the exact
profiles determine shape:

- K-Nucleotide: `bridge.ToUint` accounts for 7,999,998 objects (65.30%)
  and `bridge.ToInt` for 3,961,373 (32.33%). The generated call sites are the
  runtime-backed HashMap `raw_get` and `raw_set` methods used by window
  counting.
- Telemetry: 13,208,878 fresh `Sample` pointers account for 98.89% of the
  exact object profile. They are stored in native `[]*Sample` storage; the hot
  path contains no `runtime.Value` conversion.
- Inventory: `bridge.ToDynamicI64` accounts for 270,336 objects (86.08%).
  Generated sites encode primitive receivers, keys, and values at the
  runtime-backed HashMap and Map-interface boundary.

## Owner matrix and admission

| Candidate parent | K-Nucleotide | Telemetry | Inventory | Admission |
| --- | --- | --- | --- | --- |
| Primitive runtime boxing | dominant | absent from hot Sample path | dominant | breadth two; explicit HashMap/runtime-service boundary |
| HashMap hashing/equality | dominant | absent | dominant | named-container/runtime-service route |
| Nominal pointer allocation | absent as shared leaf | dominant | absent as shared leaf | breadth one |
| Checked integer arithmetic | not dominant | dominant | small | closed route and insufficient material breadth |
| Native interface adapter | absent | material | material Map adapter | different interface paths and breadth below three |
| Allocation/GC | material | material | material | aggregate consequence, not one lowering rule |

The measured gap supports the project's representation goal, but it does not
authorize one current change:

- K-Nucleotide and Inventory demonstrate costly primitive boxing at an
  explicit runtime service; optimizing only HashMap would violate the
  named-container guardrail.
- Telemetry demonstrates a separate non-primitive identity/allocation cost;
  changing `Sample` pointers to Go values would require a sound general
  identity and alias-observability proof.
- General allocation/GC is not an exact generated-code or runtime leaf.

Accordingly, no benchmark, container, nominal type, or source family was used
as a compiler condition; no closed checked-arithmetic or execution-context
route was reopened; and no baseline/candidate timing cohort was manufactured.

The machine-readable companion is
`2026-07-31-current-default-three-application-owner-no-go.json`.

## Verification and cleanup

- 3/3 strict build smokes passed public verifiers.
- 130/130 Able main-only CPU processes passed.
- 9/9 Able lightweight allocation processes passed.
- 3/3 Able one-object allocation-shape processes passed.
- 15/15 repeated Go CPU processes passed.
- 9/9 Go lightweight allocation processes passed.
- All 169 measured processes stayed below one minute.
- Large modules, binaries, caches, and raw profiles lived under disk-backed
  `/var/tmp`, never RAM-backed `/tmp`.
- The exact 190,420 KiB task workspace was removed after compact evidence
  publication.

## Next

Run a current-default, whole-corpus primitive-boxing boundary census before
attempting another compiled optimization.

Why: this tranche finds millions of avoidable-looking primitive boxes in two
map-heavy programs, but the third application has a different owner. The next
question is whether the same conversion cost recurs through unrelated runtime
services or semantic boundaries, which could support a general rule rather
than a HashMap exception.

What it entails: generate all 66 strict applications, classify every emitted
primitive-to-`runtime.Value` conversion by semantic boundary and lowering
provenance, then use lightweight allocation counters and focused profiles on
at least three unlike applications for any category with broad static reach.
Exclude interpreter fallback, named-container grouping, host-only cold paths,
and already-closed arithmetic/execution-context routes. Admit a prototype only
if one exact non-named boundary category is materially shared by three unlike
programs.

Why it matters: a genuinely shared boxing boundary is the most direct
remaining route to preserving native Go primitives for longer and reducing
compiled/interpreted representation crossings without weakening Able
semantics or adding benchmark-specific lowering.
