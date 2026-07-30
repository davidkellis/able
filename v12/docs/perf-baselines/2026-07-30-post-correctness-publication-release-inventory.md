# Post-correctness publication release inventory

Date: 2026-07-30

## Decision

The post-correctness publication handoff has an exact, verified release
boundary:

- retain five pre-record publication-reconciliation and handoff paths;
- hold the unchanged 34-path deferred WASM boundary outside the release;
- add this Markdown record, its JSON companion, and the 39-row TSV manifest
  to form an exact eight-path retained candidate; and
- retain no generated-local artifact.

Nothing was staged, committed, pushed, reset, reverted, or modified inside
the deferred WASM boundary during this inventory.

The maintainer subsequently authorized exact staging of the eight-path
candidate and creation of one local consolidation commit. This authorization
does not include a push, broad staging operation, reset, revert, repository
deletion, or modification of the deferred WASM boundary.

## Snapshot identity

The immutable snapshot was captured before this record, its JSON companion,
or its TSV manifest existed:

- `HEAD`, `origin/master`, and remote `master`:
  `f3f3c92d512d313648563503904528e307b7c11e`;
- branch relation: zero commits ahead and zero behind;
- `HEAD` and index tree:
  `336258c8ac459e7412f3c1848da7a5272e0b1f7b`;
- staged paths: zero;
- fully expanded dirty files: 39;
- tracked modifications: 11;
- tracked deletions and renames: zero;
- untracked files: 28; and
- tracked diff: 618 additions and 69 deletions.

The snapshot uses `--untracked-files=all`. Its 1,723-byte NUL-delimited
porcelain SHA-256 is
`249155d4f7be5715071d1c571e602725234a86a935e84309abdf0761579e385a`.
The sorted 1,606-byte newline-delimited path list has SHA-256
`80f4a41e9091109a4241873cc16708b9aea318cb6bfc2b5a61dae5158e806225`.

The exact manifest is
`2026-07-30-post-correctness-publication-release-inventory.tsv`. It contains
39 data rows, 6,758 bytes, and has SHA-256
`fce6c4d72cb3bf987d7d89750be43e09410de3a838b448d8093d2705519f8b13`.
The three inventory metadata files are intentionally not self-referential
snapshot rows.

## Dependency-ordered boundaries

| Order | Boundary | Tracked | Untracked | Paths | Lines | Bytes | Disposition |
| ---: | --- | ---: | ---: | ---: | ---: | ---: | --- |
| 1 | publication reconciliation | 0 | 2 | 2 | 237 | 8,967 | retain |
| 2 | handoff and release boundary | 3 | 0 | 3 | 59,978 | 4,134,789 | retain |
| hold | deferred WASM | 8 | 26 | 34 | 3,266 | 103,984 | hold outside release |

The pre-record retained boundary contains exactly five paths: three tracked
and two untracked. It has 60,215 lines and 4,143,756 bytes. Its sorted path
list has SHA-256
`2b3bb72688645f140b9fa496f812dcb83b595277b8486a018db3d1b5b6ebf469`.

The deferred boundary contains 34 fully expanded files. Its path-list
SHA-256 is
`7f29de27f03da83138df19696be182a5139aa6536ba0bdb363a1a23ffafd1f4e`,
exactly matching the prior authoritative inventories. Every deferred state,
line, byte, and content identity also matches.

## Record coverage

- The two publication records are governed by
  `2026-07-30-post-correctness-publication-reconciliation.md`.
- The three handoff files are governed by that same reconciliation record.
- The unchanged 34-file WASM set remains governed by the deferred WASM hold.

Every pre-record path has exactly one disposition and governing record. There
are no unmatched, multiply classified, v10, v11, deprecated-stdlib,
external-stdlib, or other out-of-scope paths.

## Validation

- All 39 snapshot rows reproduced their state, line, byte, SHA-256,
  disposition, and governing-record identities.
- All six final dirty JSON files parse.
- All 11 dirty Go files are formatted.
- All ten dirty JavaScript modules pass syntax validation.
- Tracked and untracked whitespace and final-newline checks pass.
- No dirty maintained source reaches 1,000 lines; `v12/wasm/ast_adapter.mjs`
  is largest at 402 lines.
- No secret-like filename or common private-key/service-token signature is
  present.
- The live fixture inventory reports 176 manifests, zero active exclusions,
  one retired exclusion, and zero allowlist entries; all seven policy tests
  pass.
- The 130-row scoreboard retains five successful Able/reference processes per
  row and 31 retained source/reference reports.
- All four path-relocation tests and five frontier tests pass.
- The frontier has zero actionable groups.
- All 23 ledger closures are current, invalidations are zero, and the
  selector is empty.
- The ten-test ledger suite passes with one conditional skip.
- The exact runner source retains its complete full-suite pass.
- The index remains empty and its tree matches `HEAD`.
- Local `HEAD`, `origin/master`, and remote `master` agree with zero
  divergence.
- The final guarded cleanup dry run is empty.

No production, compiler, runtime, interpreter, VM, parser semantic,
canonical-stdlib, language, dependency, benchmark measurement, fixture
behavior, frozen-workspace, or WASM behavior changed during this inventory.

## Post-record candidate

The exact retained candidate is the sorted union of:

- all five pre-record paths marked `retain`; and
- this Markdown record, its JSON companion, and the 39-row TSV manifest.

The candidate contains exactly eight paths. Its 450-byte sorted
newline-delimited path list has SHA-256
`128a4a228f2540227da2a1d0465f88a63c7a314823a699940dc6924dbcac2867`;
the exact 450-byte NUL-delimited pathspec has SHA-256
`336d7fd44a8295c1f0a4d3fbe2aa7335aecb1164e7107b38edee2fafbc750d6d`.
Its complement must remain exactly the 34 deferred WASM files.

Because final PLAN and log edits occur after the immutable snapshot, the five
non-self retained identities are refreshed separately as a sorted
tab-separated `state`, `lines`, `bytes`, `sha256`, and `path` manifest with
one header row:

- rows: 5;
- total lines: 60,267;
- total bytes: 4,146,734;
- identity-manifest bytes: 671; and
- identity-manifest SHA-256:
  `c2deeef2e3f863df2ac1b2bb8afc43751ed15b1d96c462fb0008e80362d8a4a2`.

The final expanded worktree contains exactly 42 paths: the eight-path
retained candidate plus the unchanged 34-path deferred complement.

## Cleanup

All inventory and validation state lived under disk-backed `/var/tmp`. The
final 36 KiB workspace was removed after preserving the
three readable inventory files. Python bytecode generation was disabled, no
build state used RAM-backed `/tmp`, and no task-owned workspace remains.

## Next

After the authorized local consolidation, obtain explicit maintainer
authorization before publishing that one commit.

Why: exact local staging and one commit are authorized, but remote mutation
is not.

What it entails: verify the one-commit divergence, exact remote destination
and parent, empty index, exact eight-path commit boundary, current evidence
gates, and unchanged deferred complement, then push only that
commit-to-branch refspec if explicitly authorized.

Why it matters: the publication reconciliation can reach shared history
without publishing deferred WASM, another commit, or an unintended ref.
