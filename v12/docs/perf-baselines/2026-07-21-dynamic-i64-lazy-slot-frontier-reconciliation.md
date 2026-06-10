# Dynamic-i64 lazy-slot frontier reconciliation

## Decision

Promote the complete compiled scorecard measured after the retained per-value
dynamic-`i64` cache change. All 45 compiled programs completed five verified
process executions with no failures or timeouts. The unchanged canonical
stdlib tree and same-day five-run Go references make this an Able-side refresh,
not a mixed-source comparison.

No additional runtime, compiler, VM, stdlib, language, workload, reference, or
WASM code changed in this reconciliation tranche.

## Full compiled cohort

The runner built the current compiler output separately for every catalog row,
used each row's serial or four-CPU goroutine execution contract, and retained
all workstation samples in the arithmetic means. The cohort contains 225/225
successful, verifier-backed Able executions.

Five of the 45 full-status compiled rows meet the snapshot limit:

| Benchmark | Able | Go | Able / Go | Snapshot |
| --- | ---: | ---: | ---: | --- |
| `binarytrees` | 9.5780 s | 10.6793 s | 0.8969x | meet |
| `matrixmultiply` | 1.1380 s | 1.0840 s | 1.0498x | meet |
| `quicksort` | 1.8460 s | 2.5326 s | 0.7289x | meet |
| `json` | 0.7920 s | 1.4798 s | 0.5352x | meet |
| `monte_carlo_pi` | 0.2120 s | 0.2113 s | 1.0033x | meet |

The 95%-of-Go limit is 1.0526x. Relative to the preceding promoted snapshot,
`fib` moved from meet to miss while `matrixmultiply` moved from miss to a
narrow meet. The total number of compiled snapshot meets therefore remains
five. Other large row-to-row moves occur in both directions and are treated as
workstation variance unless supported by focused causal A/B evidence.

## Threshold reconciliation

`matrixmultiply` received an independent second cohort with five freshly built
and timed Go processes and five freshly built and timed Able processes. All ten
executions verified.

| Cohort | Able / Go |
| --- | ---: |
| full snapshot A | 1.0498x |
| independent B | 1.2681x |
| pooled | 1.1554x |

The row is consequently a variance-sensitive pooled miss, not an established
guard. `fib` is absent from the stability manifest because the promoted
snapshot itself misses. The established compiled guards are now
`binarytrees`, `quicksort`, and `json`; `monte_carlo_pi` retains its historical
volatile-crossing classification.

## Promoted cross-mode frontier

The strict scorecard contains 90 full-status rows and passes the reviewed
83-row selection/evidence gate with five successful Able/reference samples per
row. The promoted frontier reports:

- 45 selected compiled rows and 38 selected bytecode rows;
- 8 snapshot meets and 75 misses;
- 5 established guards: 3 compiled and 2 bytecode;
- 3 unestablished snapshot meets;
- 134.951 seconds of target excess; and
- 0 currently actionable shared-hotspot groups.

The preceding frontier had the same 8/75 snapshot split but 6 established
guards and 135.228 excess seconds. The guard count falls only because current
`fib` timing no longer supports snapshot-meet membership; it is not evidence
that the retained lazy cache regressed primitive recursion, which does not use
that boundary.

Artifacts:

- `2026-07-21-dynamic-i64-lazy-slot-compiled-frontier-comparison.json`
- `2026-07-21-dynamic-i64-lazy-slot-scorecard.json`
- `2026-07-21-dynamic-i64-lazy-slot-threshold-stability.json`
- `2026-07-21-dynamic-i64-lazy-slot-performance-frontier.json`

## Verification

- External scorecard exact-current check passes.
- Strict scorecard evidence check passes: 83 selected rows, 90 full-status
  rows, and five successful Able/reference samples for every selected row.
- Performance frontier exact-current check passes with 83 rows and zero
  actionable groups.
- Both threshold-cohort artifacts and the stability manifest parse as valid
  JSON.

## Next recommendation

Refresh the bytecode text/map ownership gate across K-Nucleotide, Word
Frequency, Inventory Reconciliation, and two unlike text/container
discriminators such as Reverse Complement and JSON. Use bounded clean CPU and
allocation profiles, then admit a candidate only if one concrete VM-owned leaf
is material in at least three unlike applications.

Why: bytecode text/map owns 51.681 seconds of target excess, 38% of the entire
selected frontier, and K-Nucleotide alone owns 43.839 seconds. It is the
largest remaining performance wall by a wide margin. Its post-hash-index
closure predates the latest VM changes, while the recent broad six-program
selection included Word Frequency but not all three core text/map programs.
A focused current refresh can either expose a newly shared generic owner or
close the dominant wall with up-to-date evidence; it must not become a
HashMap-, benchmark-, or named-container special case.
