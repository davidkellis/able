# Post-relocatable-evidence release inventory

Date: 2026-07-30

## Decision

The post-relocatable-evidence v12 worktree has an exact, verified release
boundary:

- retain 14 pre-record shared-tree reconciliation, evidence-tooling,
  evidence-record, and handoff paths;
- hold the unchanged 34-path deferred WASM boundary outside the release;
- add this Markdown record, its JSON companion, and the 48-row TSV manifest
  to form an exact 17-path post-record retained candidate; and
- retain no generated-local artifact.

The maintainer subsequently authorized exact staging and one local commit for
the 17-path retained candidate. The authorization does not include a push,
reset, revert, repository deletion, broad staging operation, or modification
of the deferred WASM boundary.

## Snapshot identity

The inventory was captured before this record, its JSON companion, or its TSV
manifest existed:

- `HEAD`, `origin/master`, and remote `master`:
  `e7a054032e189e9a606dad68c827be50af897417`;
- branch relation: zero commits ahead and zero behind;
- `HEAD` and index tree:
  `6c62157b1ff3bb364995f2d11fd208a3f73d4cae`;
- staged paths: zero;
- fully expanded dirty files: 48;
- tracked modifications: 17;
- tracked deletions and renames: zero;
- untracked files: 31; and
- tracked diff: 852 additions and 81 deletions.

The authoritative snapshot uses `--untracked-files=all`. Its 2,205-byte
NUL-delimited porcelain SHA-256 is
`dd11955ed045098f8e347f59d9f0d77eccea8d6bb8a8771c30f2d2bfa883b525`.
The sorted 2,061-byte newline-delimited path list has SHA-256
`9da23739ea10e47419bf36d68c42e44386ceaaa112fa95ce5bf5439109989698`.

The exact manifest is
`2026-07-30-post-relocatable-evidence-release-inventory.tsv`. It contains 48
data rows, 9,076 bytes, and has SHA-256
`ae0460cfa426b9a1e87f619e78ab462c07552afb2f0022015585e2ac2b1c8b8b`.
The three inventory metadata files are intentionally not self-referential
snapshot rows.

## Dependency-ordered boundaries

| Order | Boundary | Tracked | Untracked | Paths | Lines | Bytes | Disposition |
| ---: | --- | ---: | ---: | ---: | ---: | ---: | --- |
| 1 | post-parser-publication shared-tree reconciliation | 0 | 2 | 2 | 300 | 14,223 | retain |
| 2 | relocatable performance-evidence gates | 6 | 3 | 9 | 2,544 | 186,144 | retain |
| 3 | handoff and release boundary | 3 | 0 | 3 | 59,168 | 4,090,480 | retain |
| hold | deferred WASM | 8 | 26 | 34 | 3,266 | 103,984 | hold outside release |

The pre-record retained boundary contains exactly 14 paths: nine tracked and
five untracked. It has 62,012 lines and 4,290,847 bytes. Its sorted
newline-delimited path-list SHA-256 is
`94c795f4c82b7e38c413c73c9cac79d8c89b016b33aaa7cbc72ae4628b5399e3`.

The deferred boundary contains 34 fully expanded files. Its path-list
SHA-256 is
`7f29de27f03da83138df19696be182a5139aa6536ba0bdb363a1a23ffafd1f4e`,
exactly matching the previous authoritative inventory. Every deferred state,
line, byte, and content identity also matches; no new WASM-like path exists.

## Record coverage

- The two shared-tree reconciliation records are governed by
  `2026-07-30-post-parser-publication-shared-tree-reconciliation.md`.
- Seven evidence-tooling/ledger paths and the two retained correction records
  are governed by
  `2026-07-30-relocatable-performance-evidence-gates-retained.md`.
- The three root/v12 handoff files are governed by that same retained
  correction record.
- The unchanged 34-file WASM set remains governed by the deferred WASM hold.

Every pre-record path has exactly one disposition and one governing record.
There are no unmatched, multiply classified, v10, v11, deprecated-stdlib,
external-stdlib, or other out-of-scope paths.

## Validation

- All 48 manifest state, line, byte, SHA-256, disposition, and
  governing-record identities reproduce.
- All seven dirty JSON files parse.
- All 11 dirty Go files are formatted.
- All ten dirty JavaScript modules pass syntax validation.
- All six dirty Python tools/tests compile.
- Tracked and untracked whitespace checks pass.
- None of 30 dirty maintained Go, JavaScript, Python, or Able source files
  reaches 1,000 lines. `v12/bench_external_scoreboard` is largest at 995
  lines.
- No secret-like filename or common private-key/service-token signature is
  present.
- The canonical external stdlib remains at Git
  `219eff222c28406487231713753641bc49ee5b9a`, with 70 `.able` files and
  source-tree SHA-256
  `6a412c872ee66752de7c4417a5eda99806a7631da3b880c55574c6a640b82d9b`.
  Its pre-existing dirty state is preserved.
- The 130-row scoreboard retains five successful Able/reference processes per
  row and 31 retained source/reference reports.
- All four relocation tests and five frontier tests pass.
- The frontier has zero actionable groups.
- All 23 evidence-ledger closures are current, invalidations are zero, and the
  selector is empty.
- The eight-test ledger suite passes, with its intentionally conditional
  invalidated-Markdown test skipped.
- The index remains empty and its tree matches `HEAD`.
- Local `HEAD`, `origin/master`, and remote `master` agree with zero
  divergence.

The preceding retained tranche also passed these gates inside a relocated
10,959-file source-only export containing none of the three deferred WASM
proof files. This inventory changed no performance tool, checked ledger,
compiler, generated runtime, runtime package, interpreter, VM, parser,
canonical stdlib, language, dependency version, benchmark measurement,
fixture, reference implementation, frozen workspace, or WASM behavior.

## Cleanup

All inventory and Python-cache state lived under disk-backed `/var/tmp`. The
exact 240 KiB (0.23 MiB) workspace was removed after preserving the three
readable inventory files. The final guarded project cleanup dry run is empty,
and no task-owned workspace remains. No build state was placed in RAM-backed
`/tmp`.

## Authorized post-record candidate

The authorized local-consolidation candidate is the sorted union of:

- all 14 pre-record paths marked `retain`; and
- this Markdown record, its JSON companion, and the 48-row TSV manifest.

The candidate contains exactly 17 paths. Its 896-byte sorted
newline-delimited path list has SHA-256
`571fcde5b32dc6e68e4af21c735106e0373e0caa921cf1bf74777526e1a67932`;
the exact 896-byte NUL-delimited pathspec has SHA-256
`19c73c7b14ad8ac5cf6724907477c930e7f9e8ac10635bce43f482dbcceaf6dc`.
Its complement must remain exactly the 34 deferred WASM files.

The immutable TSV records pre-record identities. Because final PLAN and log
edits occur after that snapshot, the 14 non-self retained identities were
refreshed and revalidated separately. The three inventory metadata files
remain excluded from this identity to avoid self-reference.

The final non-self retained identity has:

- 14 rows;
- 62,114 total lines;
- 4,296,777 total bytes;
- a 1,692-byte sorted identity manifest; and
- identity-manifest SHA-256
  `36abc035e3fd3b318c0b2e2d71668e293214c289de9cb6241c4e6ff2b622bc59`.

The final expanded worktree contains exactly 51 paths: the 17-path retained
candidate plus the unchanged 34-path deferred complement, with no missing or
extra path.

## Next

Obtain explicit maintainer authorization before publishing the local
post-relocatable-evidence consolidation.

Why: this tranche authorizes one exact local commit but does not authorize
remote mutation.

What it entails: verify the final one-commit divergence, remote destination,
empty index, and unchanged deferred complement, then push only that exact
commit-to-branch refspec if explicitly authorized.

Why it matters: the evidence correction can reach shared history without
publishing deferred WASM, another local commit, or any unintended ref.
