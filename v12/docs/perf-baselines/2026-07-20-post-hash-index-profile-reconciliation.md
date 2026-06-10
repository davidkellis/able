# Post-hash-index profile reconciliation

Date: 2026-07-20

## Decision

Close the compiled and bytecode text/map frontier refresh with no new
performance candidate. The retained generic stored-hash index removed the
linear entry scan, but the next exact CPU and allocation owners split below
that former wall. No new removable operation is material in Inventory
Reconciliation, Word Frequency, and K-Nucleotide in either product mode.

Retain one benchmark-runner robustness fix: `bench_compare_external` now
recreates `v12/tmp` before its default `mktemp`. This is required because the
project cleanup command intentionally removes that directory. It changes no
runtime, compiler, VM, stdlib, workload, reference, or language behavior.

## Profile contract

Six ordinary verifier-backed application processes ran with `GOMEMLIMIT=1GiB`,
`GOGC=50`, the catalog's serial CPU contract, and a 59-second process ceiling.
CPU and cumulative-allocation profiles were collected separately for each
application/mode pair. The comparison reports retain the application,
verifier, input, stdout, and execution-contract fingerprints.

The two shortest compiled CPU rows were additionally repeated five times
against one preserved binary. All ten outputs verified. The five CPU profiles
for each application were merged so workstation scheduling could not make a
single 100-250 ms sample decide ownership. K-Nucleotide already supplied 3.32
seconds of compiled and 41.15 seconds of bytecode CPU samples. The ordinary
profile-process wall times below are diagnostic metadata, not benchmark
selection claims.

| Application | Compiled process | Bytecode process | Verification |
| --- | ---: | ---: | --- |
| Inventory Reconciliation | 0.46 s | 2.53 s | 1/1 in both modes, plus 5/5 repeated compiled CPU |
| Word Frequency | 0.26 s | 1.46 s | 1/1 in both modes, plus 5/5 repeated compiled CPU |
| K-Nucleotide | 3.44 s | 41.39 s | 1/1 in both modes |

The twelve canonical profile artifacts are under
`v12/interpreters/go/.profiles/20260720_post_hash_index_*`. The six
machine-readable comparison reports are
`v12/docs/perf-baselines/2026-07-20-post-index-*-profile.json` with matching
Markdown summaries.

## Compiled CPU ownership

The pre-index scan dominated Inventory at 94.12% flat, Word Frequency at
66.67% flat, and K-Nucleotide at 11.22% cumulative. Current profiles show:

| Exact operation | Inventory | Word Frequency | K-Nucleotide | Interpretation |
| --- | ---: | ---: | ---: | --- |
| generated `__able_hash_map_find_entry` | 2.88% flat / 24.04% cumulative | 2.22% flat / cumulative | 2.71% flat / 11.14% cumulative | The old linear scan is gone; remaining descendants are hashing, equality, conversion, and GC |
| primitive key equality | 10.58% flat / 15.38% cumulative | no material sample | 6.63% flat / 7.83% cumulative | Numeric Inventory and text-heavy K-Nucleotide do not establish three-way breadth |
| `runtime.mallocgc` | 1.92% flat / 34.62% cumulative | below the top exact leaves / 44.44% cumulative | 3.01% flat / 40.66% cumulative | Broad parent with different application allocation owners |

Inventory now exposes integer equality/conversion and GC. Word Frequency
exposes UTF-8/String conversion, formatting, and GC. K-Nucleotide exposes
boxing, String construction, integer conversion, and GC. The Go collector is
common only as a runtime parent; its allocating Able children are not the same
operation.

## Bytecode CPU ownership

Inventory's `hashMapFindEntryWithHash` falls from 57.75% flat / 58.46%
cumulative to 1.27% flat / 3.38% cumulative. Word Frequency falls from 4.98%
flat / 5.44% cumulative to 0.74% flat / 2.22% cumulative. It is not a material
K-Nucleotide leaf.

The residual VM work diverges:

- Inventory is distributed across call/member/cache/type machinery; no exact
  child exceeds the breadth rule across the other two applications.
- Word Frequency exposes slot stores, Array reads, parsing/String work, and
  GC.
- K-Nucleotide remains call/return/raw-integer heavy:
  `bytecodeRawIntegerValueInfo` is 2.82% flat,
  `finishInlineReturn` is 14.87% cumulative, and `execBinary` is 14.29%
  cumulative.

Go map hashing and lookup leaves recur at the machine level, but call graphs
assign them to different Able operations: member/type caches in Inventory,
Array/environment/lowering maps in Word Frequency, and type/raw-integer caches
in K-Nucleotide. Treating those as one language map operation would conflate
unrelated owners.

## Allocation ownership

Compiled allocations split into:

- Inventory: `bridge.ToInt` 61.11% of objects and nullable `i64` conversion
  8.33%;
- Word Frequency: String split/UTF-8/string conversion plus process startup;
- K-Nucleotide: `bridge.ToInt` 36.65%, `bridge.ToUint` 25.78%, and String
  construction/UTF-8 work about 30%.

Bytecode allocations split into raw integer results/interface coercion for
Inventory, parser/lowering/Array/String setup for Word Frequency, and String
result structs plus boxed/raw integers for K-Nucleotide. There is no
application-owned allocation leaf material in all three.

The process-global eager small-integer cache is visible in every allocation
profile. It is not a new candidate. Lazy initialization, pointer packing,
static interface literals, cache right-sizing, and call-site transport were
already measured and rejected for broad runtime regressions, binary growth, or
real high-value cache reuse. See:

- `2026-07-11-compiled-bootstrap-cache-gate.md`;
- `2026-07-17-packed-eager-raw-i32-cache-feasibility.md`;
- `2026-07-17-static-raw-i32-interface-table-feasibility.md`;
- `2026-07-17-raw-i32-cache-distribution-gate.md`; and
- `2026-07-17-raw-i32-cache-call-site-attribution-gate.md`.

Freshly observing a closed startup owner does not justify retrying a rejected
representation.

## Frontier disposition

Mark both `compiled-text-map` and `bytecode-text-map` as
`closed-no-shared-leaf`, with current-exact post-index profile freshness and
three unlike applications examined. The retained hash index remains valid;
this closure says only that the next descendant does not pass the generic
candidate admission rule.

## Verification

- all six ordinary profile processes completed and verified;
- all ten repeated compiled CPU-profile processes completed and verified;
- the repeated Inventory and Word Frequency profiles merge successfully with
  `go tool pprof`;
- `bash -n v12/bench_compare_external` passes;
- running a comparison after cleanup had removed `v12/tmp` succeeds, proving
  the runner recreates its scratch root;
- frontier, catalog, scorecard, and operation-depth checks pass after the
  evidence update; and
- no WASM work was performed.

## Next recommendation

Add a third unlike, portable Mutex application before reopening concurrency
optimization.

Why: the rebuilt performance frontier has no actionable shared-leaf group.
The operation-depth ledger identifies Mutex locking/`ensure` as one of only
three insufficient portable operations, currently represented by Mutex Ledger
and Mutex Await Journal. A third application can distinguish generic lock,
cleanup, cancellation, and scheduling cost from behavior specific to either
existing program. This improves language-feature coverage while preserving the
rule that profiles—not benchmark names—authorize optimizations.

What it entails: build one bounded real application with shared mutable state,
ordinary and awaited lock acquisition, guaranteed unlock through `ensure`, and
a workload shape unlike a ledger or journal; add equivalent Go, Python, and
Ruby references plus one shared verifier; register it in the catalog,
source-equivalence, feature-coverage, operation-depth, and selection manifests;
then collect repeated verifier-backed baselines. Profile the three Mutex
applications only if the new row is stable, and advance a candidate only when
the same exact generic leaf is material in all three. Any candidate must pass
unrelated numeric, text/map, Array, and current-target guards. Do not add
nominal-type compiler rules or begin WASM.
