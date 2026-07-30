# Post-parser-publication release inventory

Date: 2026-07-30

## Decision

The post-parser-publication v12 worktree has an exact, verified release
boundary:

- retain five pre-record publication-reconciliation and handoff paths;
- hold the unchanged 34-path deferred WASM boundary outside the release;
- add this Markdown record, its JSON companion, and the 39-row TSV manifest
  to form an exact eight-path post-record retained candidate; and
- retain no generated-local artifact.

The maintainer subsequently authorized exact staging and one local commit for
the eight-path retained candidate. The authorization does not include a push,
reset, revert, repository deletion, broad staging operation, or modification
of the deferred WASM boundary.

## Snapshot identity

The inventory was captured before this record, its JSON companion, or its TSV
manifest existed:

- `HEAD`, `origin/master`, and remote `master`:
  `da8cb15bc394b6528dfd6a6b0eb1de30e12fef51`;
- branch relation: zero commits ahead and zero behind;
- `HEAD` and index tree:
  `27f296f7c3ced088157eabff41b1146ff2a502a0`;
- staged paths: zero;
- fully expanded dirty files: 39;
- tracked modifications: 11;
- tracked deletions and renames: zero;
- untracked files: 28; and
- tracked diff: 649 additions and 65 deletions.

The authoritative snapshot uses `--untracked-files=all`. Its 1,713-byte
NUL-delimited porcelain SHA-256 is
`b946363e65c252994d7d45390f37f467e1339dddeb8130d8d2a74010333b2ef6`.
The sorted 1,596-byte newline-delimited path list has SHA-256
`e65c77a23a72ccb773f5661d98ab6fdaf28e368565222050f3d540cb080124bd`.

The exact manifest is
`2026-07-30-post-parser-publication-release-inventory.tsv`. It contains 39
data rows, 6,748 bytes, and has SHA-256
`dc056e0a2ef32e804fb8d6ba8ed8fb0ceea33c612bef0583a483500d063919ad`.
The three inventory metadata files are intentionally not self-referential
snapshot rows.

## Dependency-ordered boundaries

| Order | Boundary | Tracked | Untracked | Paths | Lines | Bytes | Disposition |
| ---: | --- | ---: | ---: | ---: | ---: | ---: | --- |
| 1 | post-parser publication reconciliation | 0 | 2 | 2 | 237 | 8,933 | retain |
| 2 | handoff and release boundary | 3 | 0 | 3 | 58,900 | 4,074,724 | retain |
| hold | deferred WASM | 8 | 26 | 34 | 3,266 | 103,984 | hold outside release |

The pre-record retained boundary contains exactly five paths: three tracked
and two untracked. It has 59,137 lines and 4,083,657 bytes. Its sorted
newline-delimited path-list SHA-256 is
`5dcf7f8572cb1c17940972214410e697bdcb69b320443189f3b1e529972b593b`.

The deferred boundary contains 34 fully expanded files. Its path-list
SHA-256 is
`7f29de27f03da83138df19696be182a5139aa6536ba0bdb363a1a23ffafd1f4e`,
exactly matching the previous authoritative inventory. Every deferred line,
byte, and content identity also matches; no new WASM-like path exists.

## Record coverage

- The two clean-tree reconciliation records are governed by
  `2026-07-30-post-parser-publication-reconciliation.md`.
- The three root/v12 handoff files are governed by that same reconciliation
  record.
- The unchanged 34-file WASM set remains governed by the deferred WASM hold.

Every pre-record path has exactly one disposition and one governing record.
There are no unmatched, multiply classified, v10, v11, deprecated-stdlib,
external-stdlib, or other out-of-scope paths.

## Validation

- All 39 manifest state, line, byte, SHA-256, disposition, and
  governing-record identities reproduce.
- All five dirty JSON files parse.
- All 11 dirty Go files are formatted.
- All ten dirty JavaScript modules pass syntax validation.
- Tracked and untracked whitespace checks pass.
- None of 24 dirty maintained Go, JavaScript, or Able source files reaches
  1,000 lines. `v12/wasm/ast_adapter.mjs` is largest at 402 lines.
- No secret-like filename or common private-key/service-token signature is
  present.
- The canonical external stdlib remains at Git
  `219eff222c28406487231713753641bc49ee5b9a`, with 70 `.able` files and
  source-tree SHA-256
  `6a412c872ee66752de7c4417a5eda99806a7631da3b880c55574c6a640b82d9b`.
  Its pre-existing dirty state is preserved.
- The 130-row scorecard retains five successful Able/reference processes per
  row.
- The frontier has zero actionable groups.
- All 23 evidence-ledger closures are current, invalidations are zero, and
  the selector is empty.
- The index remains empty and its tree matches `HEAD`.
- Local `HEAD`, `origin/master`, and remote `master` agree with zero
  divergence.

The preceding clean-tree reconciliation independently reproduced the exact
published commit's main and parser module identities, empty tidy diffs,
module verification, standalone parser binding test, and test-inclusive
vulnerability result. This inventory changed no compiler, generated runtime,
runtime package, interpreter, bytecode VM, canonical stdlib, language,
dependency version, benchmark, fixture, reference implementation, frozen
workspace, or WASM behavior.

## Cleanup

All inventory state lived under disk-backed `/var/tmp`. The exact 84 KiB
(0.08 MiB) workspace was removed after preserving the three readable
inventory files. The final guarded project cleanup dry run is empty, and no
task-owned workspace remains. No build state was placed in RAM-backed
`/tmp`.

## Authorized post-record candidate

The authorized local-consolidation candidate is the sorted union of:

- all five pre-record paths marked `retain`; and
- this Markdown record, its JSON companion, and the 39-row TSV manifest.

The candidate contains exactly eight paths. Its 425-byte sorted
newline-delimited path list has SHA-256
`d65ec0bdc385996b9770c78d45e587da51ef00ade48377b2abaf9a713123a284`;
the exact 425-byte NUL-delimited pathspec has SHA-256
`d61873f4361dbc4881025e2a0218ca946871fc3c08af1af46bc9eaac55946a3c`.
Its complement must remain exactly the 34 deferred WASM files.

The immutable TSV records pre-record identities. Because final PLAN and log
edits occur after that snapshot, the five non-self retained identities were
refreshed and revalidated separately. The three inventory metadata files
remain excluded from this identity to avoid self-reference.

The final non-self retained identity has:

- five rows;
- 59,234 total lines;
- 4,089,356 total bytes;
- a 561-byte sorted identity manifest; and
- identity-manifest SHA-256
  `39a4c35428d9a243351375eeeab5cf20f9b59becfd0d8b8d2fb8e8805baf27bf`.

The final expanded worktree contains exactly 42 paths: the eight-path retained
candidate plus the unchanged 34-path deferred complement, with no missing or
extra path.

## Next

Obtain explicit maintainer authorization before publishing the local
post-parser-publication consolidation.

Why: this tranche authorizes one exact local commit but does not authorize
remote mutation.

What it entails: verify the final one-commit divergence, remote destination,
empty index, and unchanged deferred complement, then push only that exact
commit-to-branch refspec if explicitly authorized.

Why it matters: the reconciliation can reach shared history without
publishing deferred WASM, another local commit, or any unintended ref.
