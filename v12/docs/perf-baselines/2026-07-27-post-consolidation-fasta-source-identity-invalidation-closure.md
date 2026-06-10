# Post-consolidation Fasta source-identity invalidation closure

Date: 2026-07-27

Decision: refresh only the invalidated Fasta Generation scorecard and
byte-output evidence, retain no production change, and return the performance
frontier to its evidence-gated paused state.

## Invalidation

The first read-only audit after the atomic v12 consolidation found one
authoritative trigger. The release-index whitespace review removed an extra
blank line at the end of
`v12/examples/benchmarks/fasta_generation/fasta_generation.able`, changing its
exact SHA-256 from
`f8c67c9ab16e29d92904db2d58091f2512f83319a3bca5caf8ee90c37a2a96d7`
to
`7b30ce2139b20f4b30495a44e5afc99bb4ab664c2ec9e0817512784adcb06c0e`.

The program semantics and verifier output did not change, but exact evidence
provenance did. The scoreboard rejected the stale compiled row immediately.
The evidence ledger independently selected exactly two closures:

- `compiled-byte-output`;
- `bytecode-byte-output`.

The compiled closure also detected the retained trailing-space correction in
its governing Markdown evidence. No compiler, runtime, VM, language, canonical
stdlib, dependency, benchmark algorithm, reference implementation, or WASM
source changed after the consolidated tree.

## Bounded refresh

The portable coverage catalog still exactly contains 61 applications. Only
Fasta Generation was measured.

A quiet-host preflight accepted the five-run Go reference on CPU 4. Ongoing
unrelated MarketLab ingestion made subsequent quiet preflights intermittent;
seven rejected Able preflights produced no measurements. The retained refresh
therefore used one fixed CPU, five verifier-backed samples for every
Able/reference row, and a second independent five-sample Able cohort to expose
workstation spread.

Fresh reference means on CPU 4:

- Go: 0.0166 seconds;
- Python: 0.2031 seconds;
- Ruby: 0.3046 seconds.

Able cohort A:

- compiled: 0.0580 seconds, 3.49 times Go;
- bytecode: 1.8980 seconds, 9.35 times Python and 6.23 times Ruby.

Able cohort B, selected for the current scorecard:

- compiled: 0.0420 seconds, 2.53 times Go;
- bytecode: 1.9580 seconds, 9.64 times Python and 6.43 times Ruby.

All twenty Able executions passed the unchanged public verifier with stdout
SHA-256
`2907f3fb66fea247549c0f26b5b5d5cd1940a055574b72dad344283e1eb0fd10`.
The compiled builds used `--no-fallbacks`.

Across both cohorts, bytecode's ten-sample mean is 1.9280 seconds with 9.54%
coefficient of variation. Compiled launch timing is below the precision where
`time -p` is stable: its ten-sample mean is 0.0500 seconds with 49.89%
coefficient of variation. Both independent compiled cohort ratios remain far
outside the target, so the noise cannot change disposition.

## Exact roll-forward

Two derived base reports exclude only the stale Fasta rows from their original
compiled and bytecode source reports. The current exact-source Fasta report
replaces those rows. Every other current scorecard row continues to reference
its original measured artifact.

The resulting current scorecard has:

- 122 full-status rows;
- 115 selected rows with five successful Able/reference samples each;
- eight target meets and 107 target misses;
- selection SHA-256
  `1dc70106786a7e668982f070428fed3a81f77ba2abb4adf72d97848265f9dead`.

The refreshed frontier has zero actionable groups. Total target excess changes
from 182.011579 to 182.007474 seconds; no group disposition changes.

Exact source identities were propagated through the cross-engine architecture
budget, structural strategy, bytecode semantic-region feasibility, native
hot-tier budget, portable backend ADR, shared runtime semantic-ABI
feasibility, and closed-region cutover. Every decision remains unchanged and
no implementation candidate is admitted.

Only the two selected closure snapshots were advanced. The evidence ledger is
again 21 current closures and zero invalidations.

## Verification

- feature coverage: 15 families, 16 normative sections, 61 portable
  applications, and three intentional local-only families;
- current scoreboard synchronization: pass;
- selected five-run evidence completeness: 115 of 115 rows;
- current frontier: 115 rows and zero actionable groups;
- evidence ledger: 21 current and zero invalidated closures;
- complete deterministic architecture gate: pass in 8.36 seconds at 164,376
  KB peak RSS;
- affected generator and architecture contract tests: pass;
- `git diff --check`: pass.

Disposable `/var/tmp` measurement and audit files plus generated Python cache
were removed. Reusable disk-backed Go/compiler caches were retained.

## Next recommendation

Keep production performance mutation paused until another authoritative
invalidation exists.

Why: the only post-consolidation trigger was semantically inert source
identity drift. Its exact refresh preserves the same zero-candidate frontier.

What it entails: wait for a new broad application, retained execution-semantic
change, correctness failure, or report-only observation of one non-closed
owner material in at least three unlike applications. Refresh only affected
evidence before considering one general implementation rule.

Why it is important: this preserves native primitive/static-Array carriers and
minimal compiled/interpreted boundary crossings without reopening closed
routes, manufacturing benchmark-specific work, or weakening provenance. Do
not begin WASM work.
