# Generic hash-map index gate — 2026-07-20

## Decision

Retain a generic stored-hash index for runtime hash maps with at least 16
entries. The index narrows candidate entries by their already-computed hash;
Able `Hash` and `Eq` semantics remain authoritative, including collision
handling. The implementation is shared by the tree-walker, bytecode VM, and
generated Go runtime helpers. It does not recognize `HashMap`, or any other
nominal container, in compiler lowering.

Also retain the numeric-key Inventory Reconciliation application and its
equivalent Go, Python, and Ruby references. It closes the operation-depth gap
for hash lookup/update with a third unlike workload.

No canonical-stdlib source changed in this tranche.

## Workload and correctness

Inventory Reconciliation builds 8,192 numeric-key records, applies 32 update
rounds, performs ordinary get/set operations, and emits deterministic counts
and checksums. All four implementations produce:

```text
8192:32:3364558277:215845142980
```

The application is registered in the external catalog, source-equivalence
checks, feature coverage, operation depth, and the mode-aware selection
manifest. The operation-depth contract now reports 15 sufficient portable
operations, three insufficient operations, and three local-only operations.

## Profile ownership

Pre-candidate bounded profiles distinguish the three workloads:

- Inventory Reconciliation compiled spends 94.12% flat and 95.22% cumulative
  in `__able_hash_map_find_entry`; its bytecode runtime spends 57.75% flat and
  58.46% cumulative in `hashMapFindEntryWithHash`.
- Word Frequency compiled spends 66.67% flat in
  `__able_hash_map_find_entry`; bytecode attributes 4.98% flat and 5.44%
  cumulative to `hashMapFindEntryWithHash`.
- K-Nucleotide compiled attributes 11.22% cumulative to
  `__able_hash_map_find_entry`. Its bytecode profile is dominated by call,
  return, raw-integer, and VM dispatch costs, so map scanning is not a shared
  material bytecode leaf across all three applications.

This evidence authorizes the compiler/runtime candidate and predicts a large
numeric-map win, a smaller text-map win, and a neutral K-Nucleotide bytecode
result. It does not authorize workload-specific key or container lowering.

## Implementation

`runtime.HashMapValue` lazily builds `map[uint64][]int` after 16 entries. The
value updates that index on append, rebuilds it after positional removal, and
clears it with the entries. Both interpreter lookup and emitted Go lookup use
the index only to choose stored-hash candidates, then invoke the existing key
equality path. Small maps retain the linear path and pay no index allocation.

The threshold is deliberately conservative: it avoids allocating an
auxiliary Go map for tiny Able maps while eliminating quadratic scans in the
large-map workloads. Clone remains safe because the private index is rebuilt
from cloned entries on first indexed lookup.

## Repeated result

The immediately preceding two complete cohorts provide the pre-index baseline
for Word Frequency and K-Nucleotide. Two fresh post-index cohorts provide ten
verifier-backed Able samples and ten fresh reference samples per selected row.

| Application / mode | Pre-index cohort means | Post-index cohort means | Pooled change |
| --- | --- | --- | ---: |
| Word Frequency compiled | 0.272 s, 0.244 s | 0.180 s, 0.158 s | -34.5% |
| Word Frequency bytecode | 1.508 s, 1.464 s | 1.516 s, 1.380 s | -2.6% |
| K-Nucleotide compiled | 3.580 s, 3.734 s | 3.538 s, 3.602 s | -2.4% |
| K-Nucleotide bytecode | 43.512 s, 42.428 s | 43.362 s, 39.738 s | -3.3% |

Inventory's five-run pre-index baseline was 2.824 s compiled and 4.638 s
bytecode. Its two post-index cohort means are 0.298 s and 0.258 s compiled,
and 2.650 s and 2.510 s bytecode: about 90.2% and 44.4% faster respectively
when compared with that baseline.

The two post-index complete cohorts contain 67 selected rows and 74 full-status
rows. Both classify the same six selected rows as meeting the product target:
compiled Binary Trees, QuickSort, Base64, and JSON, plus bytecode JSON and
PiDigits. The unselected full-status Matrix Multiply bytecode row also meets
in both. The second cohort's first K-Nucleotide bytecode group had
four successes plus one workstation-load timeout and was rejected by the
strict evidence check. A full three-application replacement group completed
5/5 at a 39.738 s K-Nucleotide mean under the 59-second process ceiling; the
reconciled cohort then passed exact evidence validation.

One pooled historical comparison suggested a Dependency Plan bytecode
regression, although that application contains Array and Queue work and no
hash map. An order-balanced isolation collected 20 samples for each build:
index-disabled means of 0.453 s and 0.563 s versus index-enabled means of
0.474 s and 0.473 s. The enabled pooled mean is 6.8% faster. The apparent
guard movement is therefore unowned workstation variance, not a map-index
regression.

The promoted current scorecard is cohort A. It passes the scoreboard check and
the strict five-run evidence check with selection SHA-256
`15fe3f6c76ba1fa495d565ff6dd79aac3f886c28a72a511714fedea671fdc4d8`.
Cohort B remains an independent reconciled control.

## Verification

- `go test ./pkg/runtime`
- focused interpreter HashMap and map-literal tests
- focused compiler HashMap tests
- canonical stdlib tree-walker suite, including collision and persistent-map
  tests
- canonical stdlib bytecode suite, including collision and persistent-map
  tests
- `just bench-catalog-check`
- `just bench-scoreboard-check`
- strict evidence validation for both complete post-index cohorts
- strict two-cohort variance aggregation with ten successful samples per row

## Next recommendation

Refresh exact post-index profiles for Inventory Reconciliation, Word
Frequency, and K-Nucleotide in both compiled and bytecode modes, then rebuild
the performance frontier from the promoted exact-source scorecard.

Why: the retained index deliberately removes one large shared wall. Fresh
profiles are needed to prove that scanning disappeared and to identify the
next shared descendant rather than optimizing a stale parent. The bytecode
profiles are especially important because K-Nucleotide was already dominated
by VM call/return and raw-integer work, while the two other applications may
now expose those same costs after their scans shrink.

This entails bounded one-process CPU/allocation profiles under the existing
memory and sub-minute limits, comparison of exact leaves across all three
unlike workloads, and a candidate only if one generic leaf is material in all
three. Any candidate must again pass repeated target and broad guards. Do not
begin WASM work.
