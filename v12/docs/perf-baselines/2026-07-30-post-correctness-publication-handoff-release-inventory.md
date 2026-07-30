# Post-correctness publication handoff release inventory

Date: 2026-07-30

## Decision

The published correctness handoff has an exact release boundary:

- retain five pre-record reconciliation and handoff paths;
- hold the unchanged 34-path deferred WASM boundary outside the release;
- add this Markdown record, its JSON companion, and the 39-row TSV manifest
  to form an exact eight-path retained candidate; and
- retain no generated-local artifact.

Nothing was staged, committed, pushed, reset, reverted, or modified inside
the deferred boundary during this inventory.

The maintainer subsequently authorized exact staging of the eight-path
candidate and creation of one local consolidation commit. This authorization
does not include a push, broad staging operation, reset, revert, repository
deletion, or modification of the deferred WASM boundary.

## Snapshot identity

The immutable snapshot was captured before the three inventory files existed:

- `HEAD`, `origin/master`, and remote `master`:
  `8be1779aeac40e7590b852626141d01ad57deba0`;
- branch relation: zero ahead and zero behind;
- `HEAD` and index tree:
  `08c23ead09731365782982ea706cdcc64989804c`;
- staged paths: zero;
- fully expanded dirty files: 39;
- tracked modifications: 11;
- untracked files: 28;
- tracked deletions and renames: zero; and
- tracked diff: 611 additions and 67 deletions.

The 1,739-byte NUL-delimited porcelain SHA-256 is
`a6e086302ec3daf6677594a096b6378eb76d60fe1b77b657554f71bb14408f36`.
The sorted 1,622-byte newline-delimited path list has SHA-256
`85741aedb8d48582a3a33d72bcb01ca23f240acb1793f877b10454ae453ebe13`.

The exact
`2026-07-30-post-correctness-publication-handoff-release-inventory.tsv`
manifest contains 39 data rows, 6,831 bytes, and has SHA-256
`0bf0b803d6269dfcbd700a8ec897c0858b915bf568eb8b65fb248437deaff103`.
The three metadata files are intentionally not self-referential rows.

## Boundaries

| Order | Boundary | Tracked | Untracked | Paths | Lines | Bytes | Disposition |
| ---: | --- | ---: | ---: | ---: | ---: | ---: | --- |
| 1 | publication handoff reconciliation | 0 | 2 | 2 | 220 | 8,167 | retain |
| 2 | handoff and release boundary | 3 | 0 | 3 | 60,078 | 4,140,345 | retain |
| hold | deferred WASM | 8 | 26 | 34 | 3,266 | 103,984 | hold outside release |

The five-path pre-record retained boundary has 60,298 lines and 4,148,512
bytes. Its 208-byte sorted path list has SHA-256
`183bed0c9c32cace6fe29059a2a1f7e9ff35f04e650f705eac458f128a712bc7`.

The deferred boundary's 1,414-byte path list has SHA-256
`7f29de27f03da83138df19696be182a5139aa6536ba0bdb363a1a23ffafd1f4e`.
Every deferred state, line, byte, and SHA-256 identity matches the prior
authoritative inventory.

## Validation

- All 39 snapshot identities and classifications reproduce.
- All final dirty JSON files parse.
- All 11 dirty Go files are formatted, and all ten dirty JavaScript modules
  pass syntax validation.
- Whitespace, final-newline, source-size, scope, and secret checks pass.
- The 176-manifest fixture inventory and all seven policy tests pass.
- The 130-row scoreboard retains five successful Able/reference samples per
  row and 31 retained source/reference reports.
- All four path-relocation tests and five frontier tests pass.
- The frontier has zero actionable groups.
- All 23 ledger closures are current with zero invalidations.
- The selector is empty, and all ten ledger tests pass with one conditional
  skip.
- The exact runner source retains its complete full-suite pass.
- The index is empty; local, tracking, and remote `master` agree with zero
  divergence.
- The guarded cleanup dry run is empty.

No production, compiler, runtime, interpreter, VM, parser semantic, stdlib,
language, dependency, benchmark measurement, fixture behavior, frozen
workspace, or WASM behavior changed.

## Post-record candidate

The exact retained candidate is the five retained rows plus these three
inventory files. It contains eight paths.

Its 490-byte sorted newline-delimited path list has SHA-256
`d1623af5a251af848a9862ff3d0a00591bd360e1726d4e91df3083f3ca7679fe`;
its exact 490-byte NUL-delimited pathspec has SHA-256
`6de0f227aa3c54e8416086e127c1c78da6118effd98ece09b3f81af061bc5fe6`.
Its complement must remain exactly the 34 deferred WASM files.

Because final PLAN and log edits occur after the snapshot, the five non-self
retained identities are refreshed separately as a sorted tab-separated
`state`, `lines`, `bytes`, `sha256`, and `path` manifest:

- rows: 5;
- total lines: 60,342;
- total bytes: 4,150,870;
- manifest bytes: 688; and
- manifest SHA-256:
  `550813e38d9d47a2eb8255a25577c2193d4b603dc3ca2110cb2f0146d4f8a62b`.

The final expanded worktree contains exactly 42 paths: the eight-path
candidate and the unchanged 34-path deferred complement.

## Cleanup

All inventory state lived under disk-backed `/var/tmp`. The final
32 KiB workspace was removed after preserving these
records. Python bytecode generation was disabled, no build state used
RAM-backed `/tmp`, and no task-owned workspace remains.

## Next

After the authorized local consolidation, obtain explicit maintainer
authorization before publishing that one commit.

Why: exact local staging and one commit are authorized, but remote mutation
is not.

What it entails: verify the one-commit divergence, exact remote destination
and parent, empty index, exact eight-path commit boundary, current evidence
gates, and unchanged deferred complement, then push only that
commit-to-branch refspec if explicitly authorized.

Why it matters: the final handoff can reach shared history without publishing
deferred WASM, another commit, or an unintended ref.
