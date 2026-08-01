# Post-publication invalidation-trigger audit

Date: 2026-07-30

## Decision

Retain no compiler, runtime, interpreter, bytecode VM, benchmark measurement,
canonical stdlib, language, dependency, fixture, frozen-workspace, or WASM
change. The post-publication audit found no changed performance-evidence
identity, no invalidated closure, and no selected candidate.

Retain one test-only maintenance correction:
`v12/bench_feature_interaction_triples_test.py` now expects the current
published frontier's 277.200421-second aggregate target excess instead of the
superseded 278.369263-second value. This changes neither the generator nor its
evidence.

## Audited identities

The audit started from published and local `master`
`c43c4e6825504cff6486ef6ce74aaed48aebc7e5`, with an empty index.

The current evidence identities reproduce:

- 132-row scorecard:
  `3652695dc7b1576ed4729ef30a7688b171114cda9b4ce269132fd868b37849f3`;
- combined frontier:
  `d819bb32c402d690b9e89926b5bb8a0fddd6f18d9b966f8c85b3d9f73baa0a76`;
- frontier evidence:
  `f9113788815c9000194573e3c7d7a2ceed70baa1c5db1aede1e32714aba718e7`;
- closure ledger:
  `e91bc0ea504ec7e79615b01265cba0359da1378790f596bde33af9345ba35c62`;
  and
- reviewed row selection:
  `17d7babe33c64c1f17eef97eaabf7bbfba156b0bba20062fd7617e32b259f7fb`.

All six production-scope identities remain current:

- bytecode production: 203 files,
  `4c6b98eafd856f6bda13884e8b32004bf5af2842b4fcf6d2990e8c9321cd39a1`;
- compiler production: 287 files,
  `8d46f0dc50ca92902be4574bacad9d8159c396b5b3411d6efff1220c4b76eab3`;
- runtime production: 40 files,
  `73b8d6fcd4ee148ca8f928e139b0c69f37f7fcecc7b778dfecfcc9561ec8209d`;
- shared interpreter semantics: 137 files,
  `29a8ffcc6236fa35b10dc33605017f2d7d71d410232068d1c02c8a1034f12326`;
- v12 specification: one file,
  `7083f1656a3452236a372c9b20e8efdcdf6f122681e04f7d6d8607099603e71f`;
  and
- canonical stdlib scope: 70 files,
  `382d256e2fb380220dcdd62a5cf83109fa72297f23d70bdd1ffe2d8daebed047`.

The independently recaptured canonical stdlib state exactly equals the
scorecard record: sibling HEAD
`219eff222c28406487231713753641bc49ee5b9a`, 70 source files, and source-tree
SHA-256
`6a412c872ee66752de7c4417a5eda99806a7631da3b880c55574c6a640b82d9b`.
The dirty sibling worktree was observed but not modified.

The external benchmark repository remains at
`f049e3f483cf5584681c4f0ef160eab773dea4dd`. Its dirty worktree was observed
but not modified. The source-aware scoreboard reconstruction proves that the
current Able sources, verifier/input contracts, and Go/Python/Ruby reference
sources match the retained measurement fingerprints.

## Read-only results

The current external scoreboard rebuild passes. Its complete evidence check
finds:

- 132 selected and 132 full-status rows;
- 66 compiled and 66 bytecode rows;
- five successful Able/reference samples per row;
- 34 retained source reports; and
- 34 retained reference reports.

The catalog and its protocol tests pass with:

- 66 portable applications;
- 67 canonical Able benchmark sources;
- one diagnostic benchmark source;
- 79 bounded local fixtures;
- 145 combined corpus programs;
- 15 feature families;
- 16 normative sections;
- 21 operation-depth categories, 18 sufficient, zero insufficient, and three
  intentional local-only categories; and
- zero actionable frontier-linked depth gaps.

The frontier check still reports ten guards, 122 misses, zero actionable
groups, and 277.200421 seconds of aggregate target excess. The evidence ledger
still reports 23 current closures and zero invalidations. Its selector emits
zero bytes and zero rows.

The evidence-ledger, scorecard-evidence, frontier, selection, execution
contract, preserved-report, refresh-protocol, feature-coverage,
feature-interaction, and operation-depth tests all pass. The only initial
failure was the stale aggregate-excess assertion corrected in this tranche;
the focused test and complete catalog protocol pass afterward.

## Worktree boundary and cleanup

Before adding this record, the fully expanded worktree contained 40 paths:

- the unchanged 34-path deferred WASM boundary;
- the five-path publication-reconciliation documentation handoff; and
- the corrected interaction-triple test.

The deferred path list remains 1,414 bytes with SHA-256
`7f29de27f03da83138df19696be182a5139aa6536ba0bdb363a1a23ffafd1f4e`.
All 34 deferred path contents match the release inventory.

The audit's Python tools recreated six bytecode cache files under
`v12/__pycache__`. The guarded cleanup script found only that 44 KiB
task-owned cache, removed it, and then reported no generated project
artifacts. All other audit state stayed under disk-backed `/var/tmp`.

Together with the prior reconciliation handoff, this record and its JSON
companion form an exact eight-path non-WASM candidate. Its sorted
newline-delimited path list is 382 bytes with SHA-256
`e5a7141fa1aac3693272b8409df162b0be7d67bad88abbe8cb8d5dea536ad850`;
the equivalent NUL pathspec is 382 bytes with SHA-256
`cfbc6b5737eaaf4af2e1fbcf38773938d1f05724b9ad452e16cb9cbe40790ec3`.
The final fully expanded worktree contains 42 paths: those eight paths plus
the unchanged 34-path deferred WASM complement. Nothing is staged.

## Next

Keep production performance mutation paused. Run a read-only v12
correctness-and-stdlib-completeness selection audit against
`spec/full_spec_v12.md`, `spec/TODO_v12.md`, the active non-WASM design plans,
current tests, and the broad benchmark catalog.

Why: no source, stdlib, application, observer, or evidence identity invalidated
the performance closures, so there is no admissible compiler or VM candidate
to implement now.

What it entails: identify one already-specified, reproducible non-WASM
correctness or canonical-stdlib completeness gap; map it across both Go
interpreters and relevant compiled lowering; and define focused semantic
guards before authorizing implementation. If no concrete gap exists, retain
no code and keep the performance frontier closed.

Why it matters: semantic completeness can expose legitimate new native
lowering work and broaden the applications Able can run, while preserving the
rule that performance changes begin from real cross-program evidence rather
than a named benchmark, container, or already-rejected mechanism.

Do not begin WASM work.
