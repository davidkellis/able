# Compiled interface-dispatch registration capacity

Date: 2026-07-16

## Decision

Retain a generic generated-bootstrap improvement. The compiler now emits exact
outer interface-map and per-interface method-map capacities, plus a
registration-count capacity for each method's dispatch-entry slice. The
generator already knows these counts from the complete dispatch-group list.
Union target expansion can add entries at runtime, so the emitted entry count
is deliberately a safe lower bound rather than an assumption that every
registration produces one variant.

The change affects construction only. Runtime interface lookup, type-template
matching, overload specificity, alias expansion, constraint validation,
privacy, and dynamic fallback behavior are unchanged. It adds no named nominal
or benchmark-specific lowering rule. No bytecode VM, canonical-stdlib,
application, verifier, reference, benchmark source, or spec changed.

## Selection census

A temporary off-timing diagnostic counted calls and table growth during
generated `RegisterIn`. It was run against Document Audit, Dependency Plan,
Array Slice Window, and Base64, then removed before building the candidate.

| Application | Calls / variants | Interfaces | Method keys | Slice growth events | Alias expansions | Generic calls |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Document Audit | 253 / 253 | 17 | 58 | 132 | 7 | 31 |
| Dependency Plan | 214 / 214 | 16 | 58 | 140 | 7 | 73 |
| Array Slice Window | 156 / 156 | 15 | 40 | 82 | 7 | 15 |
| Base64 | 127 / 127 | 7 | 25 | 57 | 0 | 0 |

No call in this cohort produced an empty variant set. The important repeated
fact is not that every current call happens to produce one variant; it is that
all four unlike programs repeatedly grow the same generated interface and
method tables during registration. The compiler can pre-size those structures
without changing the dynamic registration algorithm.

## Implementation

`interfaceDispatchRegistrationCapacities` derives a deterministic
interface/method/count map from `interfaceDispatchGroups`. The generated
runtime uses it in two construction helpers:

- the outer map starts with the exact number of generated interface names;
- each generated interface map starts with the exact method-key count and an
  entry slice whose capacity is the number of emitted registration groups for
  that key.

Interface and method names are sorted before rendering so generated output is
stable. Unknown or runtime-added interface names retain the ordinary empty-map
fallback. A generated interface registered more than once retains its existing
table exactly as before. Runtime union expansion may append past the static
lower bound, preserving all variants and semantics.

## Allocation gate

Allocation-only phase profiles used `GOMEMLIMIT=1GiB`, `GOGC=50`, and
`GOMAXPROCS=1`. Each application was built once per side and ran from its
catalog working directory. Exact main-phase allocated-byte and allocation
counts were unchanged.

| Application | Baseline bytes / allocations | Candidate bytes / allocations | Byte change | Allocation change |
| --- | ---: | ---: | ---: | ---: |
| Document Audit | 626,936 / 7,752 | 583,784 / 7,669 | -43,152 (-6.88%) | -83 |
| Dependency Plan | 396,344 / 4,988 | 357,656 / 4,897 | -38,688 (-9.76%) | -91 |
| Array Slice Window | 283,928 / 3,625 | 257,816 / 3,576 | -26,112 (-9.20%) | -49 |
| Base64 | 225,056 / 2,797 | 203,312 / 2,761 | -21,744 (-9.66%) | -36 |

The generated executable-size changes were +0.003% to +0.035% across the
cohort. Every baseline/candidate pair reproduced the same canonical stdout
hash.

## Repeated timing gate

The timing baseline was rebuilt mechanically from each candidate generated
tree with only the two preallocated construction calls restored to their old
empty-map construction. This isolates the construction change from unrelated
compiler or source drift. Launch order alternated. Because this is a shared
workstation, two 60-launch batches and one 100-launch batch were combined for
each short application; Base64 used two five-launch batches and one ten-launch
batch.

| Application | Baseline combined mean | Candidate combined mean | Change | Runs per side |
| --- | ---: | ---: | ---: | ---: |
| Document Audit | 60.484 ms | 60.861 ms | +0.62% | 220 |
| Dependency Plan | 62.183 ms | 62.102 ms | -0.13% | 220 |
| Array Slice Window | 64.799 ms | 64.490 ms | -0.48% | 220 |
| Base64 | 2.24029 s | 2.23336 s | -0.31% | 20 |

Individual batches were noisy in both directions, but the longer third batch
favored the candidate in all four programs (-0.12%, -0.14%, -0.34%, and
-1.24%). The combined means show no broad wall-time regression. Unlike the
rejected shared type-node caches, this construction-only change adds no
synchronization or lookup to every metadata use, so the deterministic
allocation reduction clears the retention bar.

## Verification and cleanup

- Focused generated-capacity, generic constraint/revalidation, alias
  normalization, specificity, and imported-alias compiler tests pass.
- Generic interface existential execution passes.
- Focused compiled-method overload and dynamic/union/default interface
  dispatch tests pass in the interpreter package.
- `go build ./cmd/ablec` passes.
- Touched compiler files remain under 1,000 lines and diff hygiene passes.
- All four preserved application binaries completed with stable output.

Temporary diagnostics were removed before the retained build. Generated
trees, binaries, allocation profiles, stdout captures, and timing logs are
removed after this record.

## Next recommendation

Measure the generated bootstrap's internal wall-time budget before pursuing
another registration allocation micro-optimization. Add temporary off-timing
phase counters around package seeding, compiled-call registration,
compiled-method registration, interface registration, and package init across
the same short applications, then remove them and reconcile repeated-launch
means with the allocation stacks.

Why: this candidate removed 7-10% of bootstrap bytes but changed complete
process time by less than 1%. That is useful evidence that allocation count is
no longer a reliable proxy for the roughly 60 ms short-program wall. Continuing
to pre-size smaller maps without knowing their time share risks optimizing
noise rather than moving Able toward Go-equivalent application performance.

What it entails: emit diagnostic timestamps only in temporary generated
source, collect many independent launches because the workstation is noisy,
and require the same material phase in Document Audit, Dependency Plan, and
Array Slice Window with Base64 as a guard. If one registration phase is
material, profile only that phase and admit a generic construction candidate;
if none is material, close generated-bootstrap micro-allocation work and
return selection to the largest verifier-backed bytecode gap. Do not add
per-call production timing, a named-container rule, or a runtime-dispatch fast
path from startup evidence.
