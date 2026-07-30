# Canonical stdlib scope publication release inventory

Date: 2026-07-30

## Decision

The post-publication-reconciliation worktree has an exact, verified release
boundary:

- retain four pre-record publication-handoff and reconciliation paths;
- hold the unchanged 34-path deferred WASM boundary outside the release;
- add this Markdown record, its JSON companion, and the 38-row TSV manifest
  to form an exact seven-path retained candidate; and
- retain no generated-local artifact.

Nothing was staged, committed, pushed, reset, reverted, deleted from the
repository, or modified inside the deferred WASM boundary.

The maintainer subsequently authorized exact staging of the seven-path
candidate and creation of one local commit. The authorization does not
include a push, broad staging operation, reset, revert, repository deletion,
or modification of the deferred WASM boundary.

## Snapshot identity

The immutable snapshot was captured before this record, its JSON companion,
or its TSV manifest existed:

- `HEAD`, `origin/master`, and remote `master`:
  `e3a4e1e8000aeb09d80a3a1adc48548b9deeeb59`;
- branch relation: zero commits ahead and zero behind;
- `HEAD` and index tree:
  `877da26eea55749e82ccf8b3b400009ac211ce69`;
- staged paths: zero;
- fully expanded dirty files: 38;
- tracked modifications: 11;
- tracked deletions and renames: zero;
- untracked files: 27; and
- tracked diff: 712 additions and 63 deletions.

The snapshot uses `--untracked-files=all`. Its 1,642-byte NUL-delimited
porcelain SHA-256 is
`b19dfd6f52f802a94f8f07fa5f57f507f6bf93c4f967755fa6ca8ed08795cc56`.
The sorted 1,528-byte newline-delimited path list has SHA-256
`885eeaffe240affe0840b6a01122ec0f1433061f48b1e1a19316333d56970cb1`.

The exact manifest is
`2026-07-30-canonical-stdlib-scope-publication-release-inventory.tsv`. It
contains 38 data rows, 6,531 bytes, and has SHA-256
`0a22948d8c84ce9cbd2af611092960c028ddb85c0e8f41d3e8ff73e32ca67acd`.
The three inventory metadata files are intentionally not self-referential
snapshot rows.

## Dependency-ordered boundaries

| Order | Boundary | Tracked | Untracked | Paths | Lines | Bytes | Disposition |
| ---: | --- | ---: | ---: | ---: | ---: | ---: | --- |
| 1 | published-scope reconciliation | 0 | 1 | 1 | 109 | 4,199 | retain |
| 2 | handoff and release boundary | 3 | 0 | 3 | 59,637 | 4,117,082 | retain |
| hold | deferred WASM | 8 | 26 | 34 | 3,266 | 103,984 | hold outside release |

The pre-record retained boundary contains exactly four paths: three tracked
and one untracked. It has 59,746 lines and 4,121,281 bytes. Its sorted path
list has SHA-256
`04d50a63f07e90b7abcd76334ff30b0e1d74419801655fc8b0123e05abf9c1fb`.

The deferred boundary contains 34 fully expanded files. Its path-list
SHA-256 is
`7f29de27f03da83138df19696be182a5139aa6536ba0bdb363a1a23ffafd1f4e`,
exactly matching the previous authoritative inventories. Every deferred
state, line, byte, and content identity also matches; no new WASM-like path
exists.

## Record coverage

- The clean shared-history proof is governed by
  `2026-07-30-canonical-stdlib-scope-publication-reconciliation.md`.
- The three handoff files are governed by that same reconciliation record.
- The unchanged 34-file WASM set remains governed by the deferred WASM hold.

Every pre-record path has exactly one disposition and governing record. There
are no unmatched, multiply classified, v10, v11, deprecated-stdlib,
external-stdlib, or other out-of-scope paths.

## Validation

- All 38 manifest state, line, byte, SHA-256, disposition, and
  governing-record identities reproduce.
- All dirty JSON files parse.
- All 11 dirty Go files are formatted.
- All ten dirty JavaScript modules pass syntax validation.
- Tracked and untracked whitespace and final-newline checks pass.
- No dirty maintained source reaches 1,000 lines; `ast_adapter.mjs` is largest
  at 402 lines.
- No secret-like filename or common private-key/service-token signature is
  present.
- The canonical external stdlib remains at Git
  `219eff222c28406487231713753641bc49ee5b9a`, with 70 `.able` files and
  source-tree SHA-256
  `6a412c872ee66752de7c4417a5eda99806a7631da3b880c55574c6a640b82d9b`.
- The 130-row scoreboard retains five successful Able/reference processes per
  row and 31 retained source/reference reports.
- All four path tests and five frontier tests pass.
- The frontier has zero actionable groups.
- All 23 ledger closures are current, invalidations are zero, and the selector
  is empty.
- The ten-test ledger suite passes with one conditional skip.
- The index remains empty and its tree matches `HEAD`.
- Local `HEAD`, `origin/master`, and remote `master` agree with zero
  divergence.
- The guarded cleanup dry run is empty.

No production, compiler, runtime, interpreter, VM, parser, canonical-stdlib,
language, dependency, benchmark measurement, fixture, frozen workspace, or
WASM behavior changed during this inventory.

## Post-record candidate

The exact retained candidate is the sorted union of:

- all four pre-record paths marked `retain`; and
- this Markdown record, its JSON companion, and the 38-row TSV manifest.

The candidate contains exactly seven paths. Its 390-byte sorted
newline-delimited path list has SHA-256
`32b86dbd3f698ac4c878b84868ecd6985ba6fcc98e46a8136b6f2ca3ed118b0e`;
the exact 390-byte NUL-delimited pathspec has SHA-256
`726241625271f7564a6392cd8d982660dff9527a7add6ce10afd4a2272e7fd17`.
Its complement must remain exactly the 34 deferred WASM files.

Because final PLAN and log edits occur after the immutable snapshot, the four
non-self retained identities are refreshed separately:

- rows: 4;
- total lines: 59,790;
- total bytes: 4,123,637;
- identity-manifest bytes: 510; and
- identity-manifest SHA-256:
  `354ad8e464057c9a6c7ad927b7eeb27c1f658864939c19a235ebda8f6571763f`.

The final expanded worktree contains exactly 41 paths: the seven-path retained
candidate plus the unchanged 34-path deferred complement.

## Cleanup

All inventory and validation state lived under disk-backed `/var/tmp`. The
exact 696 KiB workspace was removed after preserving
the three readable inventory files. No build state used RAM-backed `/tmp`,
and no task-owned workspace remains.

## Next

After the authorized local consolidation, obtain explicit maintainer
authorization before publishing that one commit.

Why: local consolidation is authorized, but remote mutation is not.

What it entails: verify the one-commit divergence, exact remote destination
and parent, empty index, exact seven-path commit boundary, current evidence
gates, and unchanged deferred complement, then push only that commit-to-branch
refspec if explicitly authorized.

Why it matters: the publication reconciliation can reach shared history
without publishing deferred WASM, another commit, or an unintended ref.
