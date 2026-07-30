# Canonical stdlib scope release inventory

Date: 2026-07-30

## Decision

The post-canonical-stdlib-scope worktree has an exact, verified release
boundary:

- retain eight pre-record evidence-reconciliation, ledger-correction, and
  handoff paths;
- hold the unchanged 34-path deferred WASM boundary outside the release;
- add this Markdown record, its JSON companion, and the 42-row TSV manifest
  to form an exact 11-path retained candidate; and
- retain no generated-local artifact.

Nothing was staged, committed, pushed, reset, reverted, deleted from the
repository, or modified inside the deferred WASM boundary.

The maintainer subsequently authorized exact staging of the 11-path candidate
and creation of one local commit. The authorization does not include a push,
broad staging operation, reset, revert, repository deletion, or modification
of the deferred WASM boundary.

## Snapshot identity

The immutable inventory snapshot was captured before this record, its JSON
companion, or its TSV manifest existed:

- `HEAD`, `origin/master`, and remote `master`:
  `9a4eac717e1a46e10195e17c562b389fa452dcfd`;
- branch relation: zero commits ahead and zero behind;
- `HEAD` and index tree:
  `eed160e45beec4ac93292b7d8e708376811311f6`;
- staged paths: zero;
- fully expanded dirty files: 42;
- tracked modifications: 14;
- tracked deletions and renames: zero;
- untracked files: 28; and
- tracked diff: 786 additions and 69 deletions.

The authoritative snapshot uses `--untracked-files=all`. Its 1,874-byte
NUL-delimited porcelain SHA-256 is
`cf6790992e65eef4bd7203b24fa2fd4af636a7200b8e8656b4fc93e3ec36e198`.
The sorted 1,748-byte newline-delimited path list has SHA-256
`8ff6ac8fc8fa0c11005df6e5ff7aaba17f58c43ff77de2d93237c06ffa3eb324`.

The exact manifest is
`2026-07-30-canonical-stdlib-scope-release-inventory.tsv`. It contains 42
data rows, 7,556 bytes, and has SHA-256
`630c8f483f873cc08ae49f4884f9718ce6f8ebf03f7efd49f146d490f6c23478`.
The three inventory metadata files are intentionally not self-referential
snapshot rows.

## Dependency-ordered boundaries

| Order | Boundary | Tracked | Untracked | Paths | Lines | Bytes | Disposition |
| ---: | --- | ---: | ---: | ---: | ---: | ---: | --- |
| 1 | published-evidence reconciliation | 0 | 1 | 1 | 111 | 4,570 | retain |
| 2 | canonical-stdlib scope identity | 3 | 1 | 4 | 1,170 | 128,945 | retain |
| 3 | handoff and release boundary | 3 | 0 | 3 | 59,433 | 4,105,783 | retain |
| hold | deferred WASM | 8 | 26 | 34 | 3,266 | 103,984 | hold outside release |

The pre-record retained boundary contains exactly eight paths: six tracked and
two untracked. It has 60,714 lines and 4,239,298 bytes. Its sorted
newline-delimited path-list SHA-256 is
`de7ac3a98b3065edcee1f83eb26f5e4ef9ec09aceebbf46009390ca4e8eaead2`.

The deferred boundary contains 34 fully expanded files. Its path-list
SHA-256 is
`7f29de27f03da83138df19696be182a5139aa6536ba0bdb363a1a23ffafd1f4e`,
exactly matching the previous authoritative inventory. Every deferred state,
line, byte, and content identity also matches; no new WASM-like path exists.

## Record coverage

- The failed clean-tree proof is governed by
  `2026-07-30-post-relocatable-evidence-publication-reconciliation.md`.
- The ledger tool, tests, rebuilt baseline, and retained correction record are
  governed by
  `2026-07-30-canonical-stdlib-scope-identity-relocatable-retained.md`.
- The three root/v12 handoff files are governed by that same retained
  correction record.
- The unchanged 34-file WASM set remains governed by the deferred WASM hold.

Every pre-record path has exactly one disposition and governing record. There
are no unmatched, multiply classified, v10, v11, deprecated-stdlib,
external-stdlib, or other out-of-scope paths.

## Validation

- All 42 manifest state, line, byte, SHA-256, disposition, and governing-record
  identities reproduce.
- All dirty JSON files parse.
- All 11 dirty Go files are formatted.
- All ten dirty JavaScript modules pass syntax validation.
- Both dirty Python tools compile.
- Tracked and untracked whitespace and final-newline checks pass.
- No dirty maintained source reaches 1,000 lines. The ledger tool is largest
  at 756 lines.
- No secret-like filename or common private-key/service-token signature is
  present.
- The canonical external stdlib remains at Git
  `219eff222c28406487231713753641bc49ee5b9a`, with 70 `.able` files and
  source-tree SHA-256
  `6a412c872ee66752de7c4417a5eda99806a7631da3b880c55574c6a640b82d9b`.
  Its pre-existing dirty state is preserved.
- The 130-row scoreboard retains five successful Able/reference processes per
  row and 31 retained source/reference reports.
- All four path-relocation tests and five frontier tests pass.
- The frontier has zero actionable groups.
- All 23 evidence-ledger closures are current, invalidations are zero, and the
  selector is empty.
- The ten-test ledger suite passes with its intentionally conditional
  invalidated-Markdown test skipped.
- The index remains empty and its tree matches `HEAD`.
- Local `HEAD`, `origin/master`, and remote `master` agree with zero
  divergence.

No production, compiler, runtime, interpreter, VM, parser, canonical-stdlib,
language, dependency, benchmark measurement, fixture, frozen workspace, or
WASM behavior changed during this inventory.

## Post-record candidate

The exact retained candidate is the sorted union of:

- all eight pre-record paths marked `retain`; and
- this Markdown record, its JSON companion, and the 42-row TSV manifest.

The candidate contains exactly 11 paths. Its 574-byte sorted
newline-delimited path list has SHA-256
`b03361f845e6f64fb7c7332d776506617a1a3d7763d4304387984b8a9caddb21`;
the exact 574-byte NUL-delimited pathspec has SHA-256
`2b6aa507a583a598d2ca82353f5f7407c8fdfa64f8f70d0da40c78387335a204`.
Its complement must remain exactly the 34 deferred WASM files.

The immutable TSV records pre-record identities. Because final PLAN and log
edits occur after that snapshot, the eight non-self retained identities are
refreshed and validated separately:

- rows: 8;
- total lines: 60,765;
- total bytes: 4,242,205;
- identity-manifest bytes: 1,088; and
- identity-manifest SHA-256:
  `eba86e4e021544417287b2d386074a96ab461b8336ea25bbfaffd4b9b7e39bbd`.

The final expanded worktree contains exactly 45 paths: the 11-path retained
candidate plus the unchanged 34-path deferred complement.

## Cleanup

All inventory and Python-cache state lived under disk-backed `/var/tmp`. The
exact 1,960 KiB workspace was removed after preserving
the three readable inventory files. The guarded project cleanup dry run is
empty, and no task-owned workspace remains. No build state used RAM-backed
`/tmp`.

## Next

After the authorized local consolidation, obtain explicit maintainer
authorization before publishing that one commit.

Why: local consolidation is authorized, but remote mutation is not.

What it entails: verify the final one-commit divergence, exact remote
destination and parent, empty index, exact commit boundary, and unchanged
deferred complement, then push only that commit-to-branch refspec if
explicitly authorized.

Why it matters: the checkout-independent evidence correction can reach shared
history without publishing deferred WASM, another commit, or an unintended
ref.
