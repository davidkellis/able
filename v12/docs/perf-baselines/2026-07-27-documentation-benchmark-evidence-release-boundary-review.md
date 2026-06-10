# Documentation, benchmark, and evidence release-boundary review

Date: 2026-07-27

Decision: the fourth and final dependency-ordered release-consolidation
boundary is reviewed. Retain the general integrity corrections described
below, retain no production performance candidate, perform no Git history
operation, and request maintainer review before staging any boundary.

## Boundary

The pre-record worktree contained 5,080 changed or untracked paths. The first
three release reviews already own 1,086 paths. Thirty-four paths are deferred
WASM work: eight are held in the runtime manifest and the remaining 26 are
excluded here. The final boundary therefore contains 3,968 paths:

- 52 tracked modifications and 3,916 untracked paths;
- 356 reviewed source/tool/contract paths;
- 81 reviewed documentation and handoff paths;
- 258 governing decision records;
- two current scoreboards;
- 1,484 supporting reports, 1,685 raw evidence artifacts, and 102 archived
  profiles.

The large manifest is split so every file remains below 1,000 lines. The
[manifest index](2026-07-27-documentation-benchmark-evidence-release-boundary-manifest-index.tsv)
records the row count, range, and SHA-256 identity of all five parts. The
conceptual combined 3,968-row manifest SHA-256 is
`8bc063622617d82a04e644a62d21b4ec56a4b24faed163a3d0a7319670c1c0cc`.
The [machine-readable review](2026-07-27-documentation-benchmark-evidence-release-boundary-review.json)
contains the complete family and disposition counts.

No manifest part is self-referential: this review, its JSON companion, the
index, and the five manifest parts are release metadata outside the 3,968
reviewed rows.

## Integrity review

The final boundary passed these non-timing checks:

- all 1,798 JSON artifacts parse;
- both gzip archives pass integrity testing;
- all 285 pre-record governing Markdown references in `PLAN.md`, `LOG.md`,
  and `v12/LOG.md` exist;
- 1,828 Markdown files expose 3,029 explicit local links, with no missing
  target after normalizing optional line suffixes;
- the catalog contains 61 portable applications, 62 canonical Able benchmark
  sources, one diagnostic source, and 79 bounded fixtures;
- the 115-row selection has SHA-256
  `1dc70106786a7e668982f070428fed3a81f77ba2abb4adf72d97848265f9dead`;
- threshold controls, 15-family feature coverage, and the 21-operation depth
  contract pass.

One real documentation defect was corrected:
`v12/design/parser-ast-coverage.md` now links to the root v12 spec rather than
the nonexistent `v12/spec` path.

## General integrity corrections

### Generated-artifact cleanup

The default cleanup selector previously missed the generated
`v12/interpreters/go/ablec` binary and persistent sorted-set benchmark state.
Both are now covered. An isolated repository test proves dry-run selection,
tracked-file preservation, profile preservation, and apply behavior.
`v12/run_all_tests.sh` executes this guard.

Post-verification cleanup removed 214 MiB of reproducible project-local Go and
Python cache. It did not remove profiles or reusable disk-backed caches, and a
second dry run found no generated project artifact.

### Kernel synchronization

The canonical kernel lacked the two TypeScript/Go `__able_f64_sqrt`
declarations already present in the compiler-embedded kernel and required by
the canonical stdlib math surface. The declarations are restored through the
general kernel ABI. A new guard checks both `package.yml` and
`src/kernel.able` for byte equality on every full handoff.

Canonical and embedded kernel source now share SHA-256
`e126225c3064348d49229baab2a6987bc4c77e85b4a67ad6778f726d734d50e0`.
No stdlib implementation changed.

## Derived evidence reconciliation

The cross-mode frontier had not been regenerated after the current July 26
scoreboard. Regeneration changes only the current Sudoku Masks and Tapelang
compiled rows and their aggregates:

- 115 rows, eight target meets, and 107 target misses;
- 182.011579 seconds total target excess;
- 34.684632 seconds compiled and 147.326947 seconds bytecode excess;
- zero actionable groups.

The update was propagated through semantic-region feasibility, native-hot-tier
budget, cross-engine budgets, structural strategy, portable-backend ADR,
shared semantic-ABI feasibility, and closed-region cutover. All candidate
decisions remain unchanged. The current result still admits no general
compiler, runtime, VM, or stdlib implementation.

The closure ledger’s July 21 baseline was reconciled to the already-reviewed
July 26 corpus frontier and July 27 final source boundaries. This avoids
treating the complete intervening project history as a new invalidation. All
21 closures are current and zero are selected. Future production, semantic,
stdlib, workload, or frontier drift will again select only from this reviewed
state.

All 28 top-level Python contract programs pass, and every individual program
finishes in under one minute.

## Behavioral verification

The complete `./run_all_tests.sh` handoff passed in 13:58.84 at 4,679,628 KB
peak RSS:

- every coverage, selection, threshold, cleanup, and kernel check;
- all non-compiler Go packages;
- all 33 bounded compiler batches; and
- the final bytecode fixture aggregate in 87.323 seconds.

The full-suite log SHA-256 before cleanup was
`d55d14c29fdfdfb0886929af9329acaf376994ed904bdf479bc7f0d9c8770be9`.

Canonical stdlib verification passed in 57.22 seconds at 852,252 KB peak RSS.
Tree-walker and bytecode modes reported 19 and 16 seconds. Its log SHA-256 was
`dd5641b844048fa63ec60040a41be502eecdcb2583ad0e89a336b79825ff5153`.

Direct checked generators, shell syntax, cleanup isolation, kernel equality,
JSON/gzip/link validation, and `git diff --check` also pass.

## Scope decision

This tranche advances no performance optimization. It adds no named-container
or non-primitive nominal lowering, does not widen the compiled/interpreted
boundary, changes no benchmark workload or reference implementation, and
performs no WASM work.

The retained corrections are general release integrity rules: cleanup
completeness, kernel ABI synchronization, one valid documentation target, and
current evidence identities. No file was staged, committed, reset, reverted,
or rewritten through a history operation.

## Next recommendation

Request maintainer review of the four dependency-ordered manifests and
explicit authorization for an exact staging boundary.

Why: the language/parser/typechecker, runtime/tree-walker/bytecode,
compiler/AOT/semantic-ABI, and evidence/documentation final states are now
independently audited, but the extremely dirty worktree must not be staged or
committed by assumption.

What it entails: review the four manifests in order, confirm the deferred WASM
and generated-local exclusions, decide whether the reviewed final states
belong in one release consolidation or smaller subsystem commits, and name the
exact authorized path set before any Git mutation.

Why it is important: this preserves honest provenance and prevents unrelated,
generated, deferred, or stale artifacts from entering history. Performance
mutation remains paused until authoritative new evidence identifies one
non-closed production owner that is material in at least three unlike
applications. Do not begin WASM work.
