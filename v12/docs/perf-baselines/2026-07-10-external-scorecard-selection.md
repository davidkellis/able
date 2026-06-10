# 2026-07-10 External Scorecard Selection

This is a selection refresh, not a new timing run. It reconciles the latest
verified external publications and bounded Able readings already recorded in
the active plan. The shared `../benchmarks/results.json` is deliberately not
edited: it contains unrelated work. Rows from independently published suites
remain the authoritative comparison for those suites.

## Sources

- `../benchmarks/results.json` supplies the maintained Go/Python/Ruby
  reference rows for the established external suites.
- `2026-07-09-scorecard-tranche.md` and the active `PLAN.md` supply newer
  bounded Able rows and profile decisions for established suites.
- `../benchmarks-word-frequency-publish-20260710/results.json` supplies the
  six independently Docker-verified word-frequency rows.
- `../benchmarks-nominal-publish-20260710/results.json` supplies the twelve
  independently Docker-verified fixed-width-128 and rational-series rows.
- `../benchmarks-go-ref-publish-20260710/results.json` supplies the verified
  Go references for K-Nucleotide and N-body.
- `../benchmarks-lexical-rollup-publish-20260710/results.json` supplies the
  six independently Docker-verified lexical-rollup rows.

Times below are process-wall seconds from their respective verified source;
they are not combined into a ranking because input sizes and publication runs
differ. Ratios and target decisions are made only within a row.

## Current selection view

| Family | Current scorecard evidence | Selection decision |
| --- | --- | --- |
| Base64, JSON, Monte Carlo, MatrixMultiply, Sudoku | At/near the compiled-Go target or already faster than Python/Ruby; their profiles are direct host or primitive numeric kernels. | Controls only; do not perturb direct kernels. |
| Fib, BinaryTrees, Quicksort | Compiled gaps are modest or at parity; profiles are direct recurrence, allocation, or program recursion. | No shared generated-runtime leaf. |
| i-before-e, reverse-complement | Material gaps exist, but i-before-e is sub-tenth-second and reverse-complement's direct byte-array lane already failed its independent pairing. | Not reliable/new pair evidence. |
| Word-frequency | Docker: compiled `0.150955343s`, bytecode `6.623927764s`, Go `0.011430693s`, Python `0.044039261s`, Ruby `0.083066069s`. | Large gap, but its string/UTF-8 allocation source does not repeat K-Nucleotide. |
| K-Nucleotide, N-body | K compiled is `2.8900s` versus Go 1.26 `0.053543927s`; N-body compiled is `0.4400s` versus Go `0.054752445s`. K is map/string boxing; N-body is direct f64 plus bridge environment work. | No common language boundary; bytecode K exceeds the bounded steady-state cap. |
| Fixed-width-128 | Docker: compiled `5.428066437s`, bytecode `6.107921918s`, Go `0.004639621s`, Python `0.458474346s`, Ruby `0.661396768s`. | Explicit UInt128 nominal/BigInt representation work; no type-name lowering. |
| Rational-series | Docker: compiled `1.489584237s`, bytecode `3.704359985s`, Go `0.011639764s`, Python `0.117506846s`, Ruby `0.138126643s`. | Explicit Rational nominal work; paired profiles diverge from fixed-width. |
| Document-Audit, Lexical-Rollup | Docker lexical-rollup: compiled `0.059276659s`, bytecode `4.519517367s`, Go `0.003364905s`, Python `0.022383732s`, Ruby `0.055081010s`; both programs use ordinary public Iterable/filter/map/collect pipelines. | Paired bytecode profiles repeat only already-rejected member-cache and inline-return candidates; their remaining leaves diverge. |
| Pidigits | Compiled `1.3467s` versus GMP Go `0.7400s`; bytecode `2.5600s` remains faster than Ruby `9.1800s`. | BigIntRef host-library behavior lacks a second external program with the same material leaf. |
| Tapelang | Program-defined tape operations dominate and the external bytecode run exceeds the cap. | Not a runtime or stdlib specialization target. |

## Decision

No two currently verified, profile-long application programs are both
materially outside the appropriate target and led by the same concrete generic
descendant. The scorecard therefore authorizes no compiler, VM, or stdlib
candidate in this tranche.

The iterator/collection coverage gap is now closed by the independently
written Lexical-Rollup suite. The existing linked-list and array fixtures
remain guards; neither fixture nor a named collection is optimization evidence.

## Compiled follow-up

The bounded compiled CPU/allocation follow-up is complete at
`2026-07-10-compiled-pipeline-pair.md`. Both pipeline applications reach Go GC
scanning, but their concrete callers differ; MatrixMultiply remains 98% in its
direct f64 kernel. The allocation profiles are pre-main initialization
dominated. No lowering candidate is authorized.

The phase follow-up is complete at
`2026-07-10-compiled-phase-profiles.md`. Its generic split shows that generated
metadata JSON decoding recurs in package registration for Document-Audit,
Lexical-Rollup, and i-before-e. This is the first repeated compiler metadata
boundary in the current selection; the user-main profiles still diverge.

## Next work

Both generic AST metadata representation candidates are rejected: expanded Go
constructors made ordinary canonical-stdlib builds impractical, and a compact
tagged reflective codec was neutral in the two pipeline applications but
regressed i-before-e 2.49% across paired 50-launch runs. The allocation
snapshot follow-up now has exact phase-boundary attribution. After excluding
the opt-in collector's own `pprof`/`flate` allocations, all three programs
repeat eager interpreter method-cache backing allocation (687 KiB cumulative
per launch) and AST JSON decode allocation. JSON replacement is already
rejected. The generic lazy-first-store experiment for the four maps in those
three cache families produced only sub-percent/mixed startup movement and
regressed the persistent member-call guard 3.9%, so it was reverted. Refresh
the broad bytecode VM workload profiles next; no named-container, corpus, or
benchmark-specific path is authorized.
