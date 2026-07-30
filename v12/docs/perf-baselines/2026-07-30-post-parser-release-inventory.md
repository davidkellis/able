# Post-parser release inventory

Date: 2026-07-30

## Decision

The post-parser v12 worktree has an exact, verified release boundary:

- retain 11 pre-record security, parser-contract, runner, and handoff paths;
- hold the unchanged 34-path deferred WASM boundary outside the release;
- add this Markdown record, its JSON companion, and the 45-row TSV manifest
  to form an exact 14-path post-record retained candidate; and
- retain no generated-local artifact.

The maintainer subsequently authorized exact staging and one local commit for
the 14-path retained candidate. The authorization does not include a push,
reset, revert, repository deletion, broad staging operation, or modification
of the deferred WASM boundary.

## Snapshot identity

The inventory was captured before this record, its JSON companion, or its TSV
manifest existed:

- `HEAD`: `be9ecc505161085e1ec11f704571f589b3366c13`;
- `origin/master`: `be9ecc505161085e1ec11f704571f589b3366c13`;
- branch relation: zero commits ahead and zero behind;
- `HEAD` tree: `59a210723d6abe9677989b4d8440cf8b0b05e025`;
- index tree: `59a210723d6abe9677989b4d8440cf8b0b05e025`;
- staged paths: zero;
- fully expanded dirty files: 45;
- tracked modifications: 13;
- tracked deletions: one;
- untracked files: 31;
- renamed files: zero; and
- tracked diff: 719 additions and 77 deletions.

The authoritative snapshot uses `--untracked-files=all`. Its 2,023-byte
NUL-delimited porcelain SHA-256 is
`31bf9ca588e5019b934b214ac3ff2b963ce61a879065d9d0f364d7f54f8c0c7e`.
The sorted 1,888-byte newline-delimited path list has SHA-256
`60504b9cb9c503299c216d3974c2e79a55fad15df3089c169690d4a8f187d7a8`.

The exact manifest is
`2026-07-30-post-parser-release-inventory.tsv`. It contains 45 data rows,
8,041 bytes, and has SHA-256
`93ee8db6a8f02461925d8b9d066c14979f17e65d8c5e98d2768480ebd7239a73`.
The three inventory metadata files are intentionally not self-referential
snapshot rows.

The deleted `bindings/go/go.mod` row records zero lines, zero bytes, and the
empty-content SHA-256. Reproduction requires that path to remain absent. This
makes the intentional module-boundary deletion part of the exact candidate
rather than silently omitting it from content validation.

## Dependency-ordered boundaries

| Order | Boundary | Tracked | Untracked | Paths | Lines | Bytes | Disposition |
| ---: | --- | ---: | ---: | ---: | ---: | ---: | --- |
| 1 | post-publication security | 0 | 2 | 2 | 328 | 11,584 | retain |
| 2 | parser Go binding contract | 3 | 3 | 6 | 709 | 28,532 | retain |
| 3 | handoff and release boundary | 3 | 0 | 3 | 58,706 | 4,063,793 | retain |
| hold | deferred WASM | 8 | 26 | 34 | 3,266 | 103,984 | hold outside release |

The pre-record retained boundary is exactly 11 paths: six tracked paths,
including the deletion, and five untracked files. It contains 59,743 lines
and 4,103,909 bytes. Its sorted newline-delimited path-list SHA-256 is
`d05511fa582b5871a2355e5115b1417003fcc190593a51aa65c1a0f3fd530381`.

The deferred boundary contains 34 fully expanded files. Its path-list
SHA-256 is
`7f29de27f03da83138df19696be182a5139aa6536ba0bdb363a1a23ffafd1f4e`,
exactly matching the previous authoritative inventory. Every deferred line,
byte, and content identity also matches; no new WASM-like path exists.

## Record coverage

- The two security records are governed by
  `2026-07-30-post-publication-security-reconciliation.md`.
- Six parser module, runner, and evidence paths are governed by
  `2026-07-30-parser-go-binding-contract-retained.md`.
- The three root/v12 handoff files are governed by the same retained parser
  record.
- The unchanged 34-file WASM set remains governed by the deferred WASM hold.

Every pre-record path has exactly one disposition and one governing record.
There are no unmatched, multiply classified, v10, v11, deprecated-stdlib,
external-stdlib, or other out-of-scope paths.

## Validation

- All 45 manifest state, absence, line, byte, SHA-256, disposition, and
  governing-record identities reproduce.
- All 6 dirty JSON files parse.
- All 11 dirty Go files are formatted.
- All 10 dirty JavaScript modules pass syntax validation.
- The one dirty shell script passes `bash -n`.
- Tracked and untracked whitespace checks pass.
- None of 25 dirty maintained Go, JavaScript, shell, or Able source files
  reaches 1,000 lines. `v12/run_all_tests.sh` and
  `v12/wasm/ast_adapter.mjs` tie for largest at 402 lines.
- No secret-like filename or common private-key/service-token signature is
  present.
- The parser root passes `go mod tidy -diff`, `go mod verify`, and the
  one-minute explicit binding test.
- The canonical external stdlib remains at Git
  `219eff222c28406487231713753641bc49ee5b9a`, with 70 `.able` files and
  source-tree SHA-256
  `6a412c872ee66752de7c4417a5eda99806a7631da3b880c55574c6a640b82d9b`.
- The 130-row scorecard retains five successful Able/reference processes per
  row.
- The frontier has zero actionable groups.
- All 23 evidence-ledger closures are current, invalidations are zero, and the
  selector is empty.
- The index remains empty and its tree matches `HEAD`.

The preceding parser tranche passed the complete default v12 runner, including
every preflight, the standalone binding gate, non-compiler packages, all 34
compiler batches, and the bytecode fixture pass. This inventory changed no
compiler, generated runtime, runtime package, interpreter, bytecode VM,
canonical stdlib, language, dependency version, benchmark, fixture, reference
implementation, frozen workspace, or WASM behavior.

## Cleanup

All inventory and parser-check state lived under disk-backed `/var/tmp`. The
exact 66,776 KiB (65.21 MiB) workspace was removed after preserving the three
readable inventory files. The final guarded project cleanup dry run is empty,
and no task-owned workspace remains. No build state was placed in RAM-backed
`/tmp`.

## Authorized post-record candidate

The authorized local-consolidation candidate is the sorted union of:

- all 11 pre-record paths marked `retain`; and
- this Markdown record, its JSON companion, and the 45-row TSV manifest.

The candidate contains exactly 14 paths. Its 681-byte sorted
newline-delimited path list has SHA-256
`2ed84bdc5e65ab3a9832ef499e50b2bd7878b7e1d9d1a6d8a06e170c9ca34c0a`;
the exact 681-byte NUL-delimited pathspec has SHA-256
`bf739c20a2ea2d85c23fde9d4a2f36e3f3e4ee553b85c69434c961112799a4fd`.
Its complement must remain exactly the 34 deferred WASM files.

The immutable TSV records pre-record identities. Because final PLAN and log
edits occur after that snapshot, the 11 non-self retained identities were
refreshed and revalidated separately. The three inventory metadata files
remain excluded from this identity to avoid self-reference.

The final non-self retained identity has:

- 11 rows;
- 59,849 total lines;
- 4,109,810 total bytes;
- a 1,289-byte sorted identity manifest; and
- identity-manifest SHA-256
  `d6fa31632c79d1e592a3b441d7559c7dc7d533e173038ad6a6495dc5aaa71211`.

The final expanded worktree contains exactly 48 paths: the 14-path retained
candidate plus the unchanged 34-path deferred complement, with no missing or
extra path.

## Next

Obtain explicit maintainer authorization before publishing the local
post-parser consolidation.

Why: this tranche authorizes one exact local commit but does not authorize
remote mutation.

What it entails: verify the final one-commit divergence and remote
destination, then push only that exact commit-to-branch refspec if explicitly
authorized.

Why it matters: the retained security evidence and parser binding correction
can reach shared history without publishing deferred WASM or unintended local
commits.
