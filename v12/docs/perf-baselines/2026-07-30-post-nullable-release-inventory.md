# Post-nullable release inventory

Date: 2026-07-30

## Decision

The post-nullable v12 worktree has an exact, verified release boundary:

- retain 290 pre-record paths;
- hold the unchanged 34-path deferred WASM boundary outside the release;
- add this Markdown record, its JSON companion, and the 324-row TSV manifest
  to form an exact 293-path post-record retained candidate; and
- retain no generated-local artifact.

The maintainer subsequently authorized exact staging and one local commit for
the 293-path retained candidate. The authorization does not include a push,
reset, revert, repository deletion, broad staging operation, or modification
of the deferred WASM boundary.

## Snapshot identity

The inventory was captured before this record, its JSON companion, or its TSV
manifest existed:

- `HEAD`: `358368c24f84367ee8f39c5df6b07ec488b477ef`;
- `origin/master`: `358368c24f84367ee8f39c5df6b07ec488b477ef`;
- branch relation: zero commits ahead and zero behind;
- `HEAD` tree: `97febd6f56f68719fbd1867fd686ad71f3d742b7`;
- index tree: `97febd6f56f68719fbd1867fd686ad71f3d742b7`;
- staged paths: zero;
- fully expanded dirty files: 324;
- tracked modifications: 86;
- untracked files: 238;
- deleted or renamed files: zero; and
- tracked diff: 7,996 additions and 4,656 deletions.

The authoritative snapshot uses `--untracked-files=all`. Its 28,121-byte
NUL-delimited porcelain SHA-256 is
`f4f23b6f5c2ec976a801a2e6333a70ab42c580a99e4ea04c462a0187fed0c70b`.
The sorted 27,149-byte newline-delimited path list has SHA-256
`0333ea0edf227f128bb1d05a1c0646da90b7185f3ed6276494ea8c5b019b9c30`.

The exact manifest is
`2026-07-30-post-nullable-release-inventory.tsv`. It contains 324 data rows,
83,705 bytes, and has SHA-256
`5134feddffad1f401488db693dc649eb41142b879b6cf8ff4f2951421313d6c7`.
The three inventory metadata files are intentionally not self-referential
snapshot rows.

## Dependency-ordered boundaries

| Order | Boundary | Tracked | Untracked | Paths | Disposition |
| ---: | --- | ---: | ---: | ---: | --- |
| 1 | nullable scorecard and retained evidence | 47 | 142 | 189 | retain |
| 2 | primitive-nullable compiler and design | 28 | 1 | 29 | retain |
| 3 | post-nullable closure evidence | 0 | 67 | 67 | retain |
| 4 | correctness release gate | 0 | 2 | 2 | retain |
| 5 | handoff and release boundary | 3 | 0 | 3 | retain |
| hold | deferred WASM | 8 | 26 | 34 | hold outside release |

The pre-record retained boundary is exactly 290 paths: 78 tracked
modifications and 212 untracked files. It contains 170,671 lines and
10,321,629 bytes. Its sorted newline-delimited path-list SHA-256 is
`c0fd6477a10ce7f0680d7fcfed294d021964229e011c20cac0e472adc2785823`.

The deferred boundary contains 34 fully expanded files. Its path-list
SHA-256 is
`7f29de27f03da83138df19696be182a5139aa6536ba0bdb363a1a23ffafd1f4e`.
That identity exactly matches the post-spawn, post-security, and post-census
inventories. All 34 prior deferred paths remain present, and no new
WASM-like path exists.

## Record coverage

- 189 scorecard, application, profile, and retained evidence paths are
  governed by `2026-07-30-nullable-scalar-retained-frontier.md`.
- 29 compiler/design paths are governed by
  `2026-07-30-compiled-primitive-nullable-value-carrier-retained.md`.
- 67 post-nullable closure paths are governed by
  `2026-07-30-post-nullable-cross-family-architecture-ownership-reconciliation.md`.
- The two correctness records and three handoff files are governed by
  `2026-07-30-post-nullable-correctness-release-gate.md`.
- The unchanged 34-file WASM set remains governed by the deferred WASM hold.

Every pre-record path has exactly one disposition and one governing record.
There are no unmatched, multiply classified, v10, v11, deprecated-stdlib, or
external-stdlib paths.

## Validation

- All 324 manifest line, byte, and SHA-256 identities reproduced.
- All 135 dirty JSON files parse.
- No dirty gzip archive exists; the archive check is vacuously clean.
- All 39 dirty Go files are formatted.
- All 10 dirty JavaScript modules pass syntax validation.
- All 7 dirty Python files parse without creating bytecode.
- The one dirty shell script passes `bash -n`.
- Tracked and untracked whitespace checks pass.
- None of 63 dirty maintained source files reaches 1,000 lines. The largest
  is `v12/interpreters/go/pkg/compiler/generator_exprs.go` at 987 lines.
- No secret-like filename or common private-key/service-token signature is
  present.
- No path exists outside v12 plus the root handoff files.
- The canonical external stdlib source identity is unchanged.
- The 130-row five-sample scorecard, zero-actionable frontier, and
  23-current-entry closure ledger reproduce; the selector is empty.
- The index remains empty.

The preceding correctness gate already passed the complete default v12 suite
and both canonical stdlib execution modes. This inventory changed no compiler,
generated runtime, runtime package, interpreter, bytecode VM, canonical
stdlib, language, dependency, benchmark, fixture, reference implementation,
or WASM behavior.

## Generated-artifact cleanup

The first guarded cleanup dry run found three ignored project-local paths:

- `v12/interpreters/go/.gocache`, 9.70 GiB;
- empty `v12/tmp`; and
- empty `v12/interpreters/go/tmp`.

The separate `/tmp/able-v12-extern-go` generated cache used 60,520 KiB, had
zero open handles, and had not been written since the previous day. The
repository cleanup command removed its three approved targets, and an exact
depth-first deletion removed the inactive external-host cache. Total reclaimed
space was 10,230,888 KiB, or 9.76 GiB. The final guarded cleanup dry run found
no generated project artifact.

## Authorized post-record candidate

The authorized local-consolidation candidate is the sorted union of:

- all 290 pre-record paths marked `retain`; and
- this Markdown record, its JSON companion, and the 324-row TSV manifest.

The candidate contains exactly 293 paths. Its sorted newline-delimited
path-list SHA-256 is
`c61b5f0e44c430785d6f3eb8ac6a2742d66c21c35e03b329a07cc1b5798e1976`;
the exact NUL-delimited pathspec SHA-256 is
`bd83d0389537589932c07dc82a52a8cb86f4a7837658735ccc774c4e151939e6`.
Its complement must remain exactly the 34 deferred WASM files.

The immutable manifest records pre-record identities. Because final PLAN and
log edits occur after that snapshot, the 290 non-self retained identities are
refreshed and revalidated separately after those edits. The three inventory
metadata files remain excluded from that identity to avoid self-reference.

The final non-self retained identity has:

- 290 rows;
- 170,764 total lines;
- 10,326,710 total bytes;
- a 47,175-byte sorted identity manifest; and
- identity-manifest SHA-256
  `7958bfcd02b0eee139af541acf3ebd85f0ec431b1bdcbda57d8e7dcb25eef18b`.

The final expanded worktree contains exactly 327 paths: the 293-path retained
candidate plus the 34-path deferred complement, with no missing or extra
path.

## Next

Obtain explicit maintainer authorization before publishing the local
post-nullable consolidation.

Why: this tranche authorizes exact staging and one local commit, but it does
not authorize remote mutation.

What it entails: verify the final one-commit divergence and remote
destination, then push only that exact commit-to-branch refspec if explicitly
authorized.

Why it matters: the verified native-carrier, interpreter-free, scorecard, and
correctness work can reach shared history without publishing deferred WASM or
unintended local commits.
