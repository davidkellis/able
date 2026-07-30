# Post-correctness release inventory

Date: 2026-07-30

## Decision

The completed post-publication correctness work has an exact, verified
release boundary:

- retain 18 pre-record non-WASM paths across five completed correctness
  tranches and the three handoff files;
- hold the unchanged 34-path deferred WASM boundary outside the release;
- add this Markdown record, its JSON companion, and the 52-row TSV manifest
  to form an exact 21-path retained candidate; and
- retain no generated-local artifact.

Nothing was staged, committed, pushed, reset, reverted, or modified inside
the deferred WASM boundary.

The maintainer subsequently authorized exact staging of the 21-path
candidate and creation of one local consolidation commit. This authorization
does not include a push, broad staging operation, reset, revert, repository
deletion, or modification of the deferred WASM boundary.

## Snapshot identity

The immutable snapshot was captured before this record, its JSON companion,
or its TSV manifest existed:

- `HEAD`, `origin/master`, and remote `master`:
  `9c32f2777536da2c948327720acc75187973a6d9`;
- branch relation: zero commits ahead and zero behind;
- `HEAD` and index tree:
  `355396727cd3f21b3ac48cd7cc182333855e16fc`;
- staged paths: zero;
- fully expanded dirty files: 52;
- tracked modifications: 13;
- tracked deletions and renames: zero;
- untracked files: 39; and
- tracked diff: 1,008 additions and 71 deletions.

The snapshot uses `--untracked-files=all`. Its 2,712-byte NUL-delimited
porcelain SHA-256 is
`5c1b9b0cf556807fbaca46b57def436727c77d0e949fcedb62388b17cdd5fb8a`.
The sorted 2,556-byte newline-delimited path list has SHA-256
`8c2fdc3fc293ce0dc9a8ea4f8009ca74fe21aa96a66c1d4823c9ebea37bd2d02`.

The exact manifest is
`2026-07-30-post-correctness-release-inventory.tsv`. It contains 52 data
rows, 10,173 bytes, and has SHA-256
`76734853cd739ef09511d21bd5deb8f12f3b0daa40e8051adbf006de30dfe2d6`.
The three inventory metadata files are intentionally not self-referential
snapshot rows.

## Dependency-ordered boundaries

| Order | Boundary | Tracked | Untracked | Paths | Lines | Bytes | Disposition |
| ---: | --- | ---: | ---: | ---: | ---: | ---: | --- |
| 1 | compiled profile admission closure | 0 | 2 | 2 | 275 | 11,013 | retain |
| 2 | implicit generic fixture coverage | 1 | 2 | 3 | 227 | 9,000 | retain |
| 3 | active target-exclusion guard | 0 | 5 | 5 | 724 | 23,512 | retain |
| 4 | parser fixture-corpus lane | 1 | 2 | 3 | 675 | 23,945 | retain |
| 5 | silent test-suppression closure | 0 | 2 | 2 | 257 | 9,817 | retain |
| 6 | handoff and release boundary | 3 | 0 | 3 | 60,111 | 4,144,623 | retain |
| hold | deferred WASM | 8 | 26 | 34 | 3,266 | 103,984 | hold outside release |

The pre-record retained boundary contains exactly 18 paths: five tracked and
13 untracked. It has 62,269 lines and 4,221,910 bytes. Its sorted path list
has SHA-256
`fe0075381ed6b0aa005bc106398c6c8f6d6a7ac1275b000a8a20d3ea6b5b83f8`.

The deferred boundary contains 34 fully expanded files. Its path-list
SHA-256 is
`7f29de27f03da83138df19696be182a5139aa6536ba0bdb363a1a23ffafd1f4e`,
exactly matching the previous authoritative inventories. Every deferred
state, line, byte, and content identity also matches; no new WASM-like path
exists.

## Record coverage

- The compiled-profile admission closure is governed by
  `2026-07-30-post-publication-compiled-profile-admission-closure.md`.
- The fixture correction is governed by
  `2026-07-30-implicit-generic-redeclaration-go-fixture-coverage-retained.md`.
- The fail-closed target policy is governed by
  `2026-07-30-active-go-fixture-target-exclusion-guard-retained.md`.
- The full-mode parser lane is governed by
  `2026-07-30-default-parser-fixture-corpus-lane-retained.md`.
- The audit-only closure and current handoff are governed by
  `2026-07-30-silent-go-test-suppression-audit-closure.md`.
- The unchanged 34-file WASM set remains governed by the deferred WASM hold.

Every pre-record path has exactly one disposition and governing record. There
are no unmatched, multiply classified, v10, v11, deprecated-stdlib,
external-stdlib, or other out-of-scope paths.

## Validation

- All 52 manifest state, line, byte, SHA-256, disposition, and
  governing-record identities reproduce.
- All 12 final dirty JSON files parse.
- All 11 dirty Go files are formatted.
- All 12 dirty JavaScript modules pass syntax validation.
- Tracked and untracked whitespace and final-newline checks pass.
- No dirty maintained source reaches 1,000 lines; `v12/run_all_tests.sh` is
  largest at 416 lines.
- No secret-like filename or common private-key/service-token signature is
  present.
- The canonical external stdlib remains at Git
  `219eff222c28406487231713753641bc49ee5b9a`, with 70 runtime `.able` files
  and source-tree SHA-256
  `6a412c872ee66752de7c4417a5eda99806a7631da3b880c55574c6a640b82d9b`.
- The 130-row scoreboard retains five successful Able/reference processes per
  row and 31 retained source/reference reports.
- All four path tests and five frontier tests pass.
- The frontier has zero actionable groups.
- All 23 ledger closures are current, invalidations are zero, and the
  selector is empty.
- The ten-test ledger suite passes with one conditional skip.
- The exact runner source in this boundary retains its complete full-suite
  pass: every preflight and non-compiler package, all 34 compiler batches, and
  the complete bytecode fixture corpus.
- The index remains empty and its tree matches `HEAD`.
- Local `HEAD`, `origin/master`, and remote `master` agree with zero
  divergence.
- The final guarded cleanup dry run is empty.

No new production, compiler, runtime, interpreter, VM, parser,
canonical-stdlib, language, dependency, benchmark measurement, fixture,
frozen-workspace, or WASM behavior changed during this inventory.

## Post-record candidate

The exact retained candidate is the sorted union of:

- all 18 pre-record paths marked `retain`; and
- this Markdown record, its JSON companion, and the 52-row TSV manifest.

The candidate contains exactly 21 paths. Its 1,364-byte sorted
newline-delimited path list has SHA-256
`b2325ad24655d9cdafcdfa1b3aeef66d0dbbd6ce2b7c6566e42aa6518084954b`;
the exact 1,364-byte NUL-delimited pathspec has SHA-256
`e463afc94f6a46e2c2ace7eb7099040ca2357669584957dd25456acd7ed23d69`.
Its complement must remain exactly the 34 deferred WASM files.

Because final PLAN and log edits occur after the immutable snapshot, the 18
non-self retained identities are refreshed separately as a sorted
tab-separated `state`, `lines`, `bytes`, `sha256`, and `path` manifest with
one header row:

- rows: 18;
- total lines: 62,083;
- total bytes: 4,208,985;
- identity-manifest bytes: 2,722; and
- identity-manifest SHA-256:
  `f7af0b486f9d7696080f6e2d3e9b01fcb12ba0857b6e491c3b683651e78ae004`.

The final expanded worktree contains exactly 55 paths: the 21-path retained
candidate plus the unchanged 34-path deferred complement.

## Cleanup

All inventory and validation state lived under disk-backed `/var/tmp`. The
guarded cleanup removed an empty generated Go workspace and 316 KiB of
reproducible Python bytecode cache. The final 48 KiB
inventory workspace was removed after preserving the three readable
inventory files. No build state used RAM-backed `/tmp`, and no task-owned
workspace remains.

## Next

After the authorized local consolidation, obtain explicit maintainer
authorization before publishing that one commit.

Why: exact local staging and one commit are authorized, but remote mutation
is not.

What it entails: verify the one-commit divergence, exact remote destination
and parent, empty index, exact 21-path commit boundary, current evidence
gates, and unchanged deferred complement, then push only that
commit-to-branch refspec if explicitly authorized.

Why it matters: the correctness consolidation can reach shared history
without publishing deferred WASM, another commit, or an unintended ref.
