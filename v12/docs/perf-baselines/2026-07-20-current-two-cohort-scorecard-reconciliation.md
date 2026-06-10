# Current two-cohort scorecard reconciliation — 2026-07-20

## Decision

Promote cohort A as the current complete scorecard, retain cohort B as its
independent control, and select no compiler, bytecode VM, runtime, stdlib, or
application optimization from this tranche.

Both cohorts evaluate the same current working-tree source, the same reviewed
selection, and the same canonical stdlib source tree. Every selected row has
five successful verifier-backed Able processes and five freshly measured
reference processes in each cohort. The seven excluded bytecode rows remain
visible as one-run bounded status probes.

No compiler, interpreter, runtime, canonical stdlib, benchmark, fixture,
reference, verifier, or language source changed.

## Protocol and evidence

- selected rows: 65 (36 compiled, 29 bytecode)
- full-status rows: 72
- timed processes: five per selected Able/reference row in each cohort
- excluded-row probes: one process
- per-process timeout: 55 seconds
- CPU pool: `0-3`, resolved to each row's cataloged logical-CPU budget
- memory guard: `GOMEMLIMIT=1GiB`, `GOGC=50`
- selection SHA-256:
  `e7b35985b05134e1619be193cbe21ddce846cc2392efe78560e629de048d97dc`
- canonical stdlib: 70 Able sources, tree SHA-256
  `64b66a5b49cf3779912010d288ea0bcd0256c291dd58fe1bda705ee22dee6863`,
  Git `219eff222c28406487231713753641bc49ee5b9a` (dirty state recorded)

The single-cohort evidence gate accepts both aggregates. The strict variance
reader also accepts their disjoint comparison sources, identical row coverage,
execution contracts, stdlib identity, and exact five-run evidence.

## Reconciled target status

Across the 65 selected application/mode contracts:

- 57 miss the 95% target in both cohorts;
- seven meet it in both cohorts; and
- one changes classification and is therefore volatile.

The established target meets are:

| Benchmark | Mode | Cohort A ratio | Cohort B ratio |
| --- | --- | ---: | ---: |
| Base64 | compiled | 1.0175x Go | 1.0376x Go |
| Binary Trees | compiled | 0.9031x Go | 0.9493x Go |
| JSON | compiled | 0.5146x Go | 0.5608x Go |
| Monte Carlo Pi | compiled | 0.7742x Go | 0.7223x Go |
| QuickSort | compiled | 0.7449x Go | 0.7020x Go |
| JSON | bytecode | 0.5316x faster interpreter | 0.4247x faster interpreter |
| PiDigits | bytecode | 0.6833x faster interpreter | 0.5953x faster interpreter |

Compiled Matrix Multiply is the only classification flip: cohort A is
1.0136x Go and meets the target, while cohort B is 1.1086x Go and misses it.
It is not an established meet. The promoted single-cohort scoreboard and
frontier necessarily retain cohort A's measured status, so their eight-meet
count must be read together with this two-cohort reconciliation.

The refresh also changes the current evidence boundary relative to the older
scorecard. Compiled Base64 and Monte Carlo Pi are now repeated target meets.
Compiled PiDigits is a repeated miss (1.0858x and 1.1437x Go), superseding its
older near-parity cohort. Bytecode Matrix Multiply remains a fast, verified
one-run status result, but it is excluded from the strict selection and is no
longer counted as an established five-run meet.

## Frontier result

The regenerated 65-row frontier has zero actionable ownership groups. Its
single-cohort summary contains 57 target misses, eight cohort-A target meets,
and 110.0977 seconds of target excess. All ownership groups remain protected
or closed by insufficient breadth, no shared exact leaf, rejected generic
candidates, or related-algorithm-only evidence.

Several short compiled coverage applications also spent visibly much longer
in AOT preparation than in their timed executable phase, especially the regex
applications. That qualitative build-latency observation is not part of the
application-runtime scorecard and does not authorize a runtime candidate.

## Next recommendation

Extend the existing feature-coverage contract from broad families to concrete
performance-relevant operations before adding another benchmark or profile
tranche. The current machine-checked manifest already covers all 15 families,
all 16 normative sections, all 36 portable applications, and three intentional
local-only families. Repeating that audit would add no information.

Instead, map specific semantic operations—such as dynamic interface dispatch,
nominal construction, closure capture/invocation, destructuring and union
matching, Result propagation, iterator advancement, and package initialization
—to the unlike portable applications and focused fixtures that actually
exercise them. Record whether each operation has zero, one or two, or at least
three unlike portable consumers, and connect the latter set to the current
frontier ownership evidence.

Why: the family-level matrix can group applications that reach different
compiler or VM leaves, while the frontier currently closes many groups for
exactly that reason. Operation-level depth can reveal already-existing
three-application breadth that the ownership ledger has missed, or prove that
an honest new application is needed. Add an Able/Go/Python/Ruby application
only for a real portable gap; do not manufacture work for local-only semantics.

This entails a versioned depth manifest (or schema extension), checker and
failure-mode tests, source/fixture evidence for every claimed operation, and a
reconciliation report. Timing should follow only if the map identifies one
concrete generic leaf in at least three unlike selected applications; that
candidate must then pass two fresh five-run cohorts and the broad guards. Do
not begin WASM work.

## Artifacts

- promoted cohort A:
  `2026-07-20-current-full-scorecard-cohort-a-refresh.json`
- independent cohort B:
  `2026-07-20-current-full-scorecard-cohort-b-refresh.json`
- strict variance:
  `2026-07-20-current-full-scorecard-variance.json`
- current scoreboard: `external-scoreboard-current.json`
- source-equivalent scorecard: `2026-07-20-source-equivalence-scorecard.json`
- regenerated frontier: `2026-07-20-cross-mode-performance-frontier.json`
