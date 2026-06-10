# Post-consolidation release inventory

Date: 2026-07-28

Decision: after retaining a non-mutating inventory of the work accumulated
after the July 27 consolidation commit, use the maintainer's explicit
authorization to commit the exact 340-path retained boundary. Do not reset,
rewrite, push, or modify the deferred WASM boundary.

## Snapshot identity

The inventory was captured before this record, its JSON companion, or its TSV
manifest existed:

- HEAD:
  `cbba634af6c5766e5b8aa162e8e66643654f620e`;
- consolidation commit:
  `0403ad8845e05eb3e136fa32f27db6ee4ea453a7`;
- `origin/master`:
  `237406eccdfb025a519d898daedadee1c8d13a7b`;
- branch relation: four commits ahead and zero behind;
- staged paths: zero;
- visible dirty paths: 386;
- tracked modifications: 93;
- untracked paths: 293;
- deleted or renamed paths: zero;
- tracked diff: 6,445 additions and 3,350 deletions.

The exact 386-row manifest is
`2026-07-28-post-consolidation-release-inventory.tsv`. It records worktree
state, review boundary, disposition, line count, byte count, SHA-256, and path.
Its SHA-256 is
`dd8c6c8ea7a5dc0ea6d7563dd20fc9410c9902d46646f16e18fc76cd582b71bc`.
The source porcelain snapshot has SHA-256
`2b9b0dccde5940fab3033ca2621e25501b1fe6902ac28410980fc882ed5cc141`.

The inventory metadata files are intentionally not self-referential manifest
rows. Any authorized exact-index operation must refresh identities after
including this record, its JSON companion, and its TSV manifest.

## Dependency-ordered review boundaries

| Order | Boundary | Tracked | Untracked | Disposition |
| ---: | --- | ---: | ---: | --- |
| 1 | compiler/AOT/generated-runtime implementation and guards | 36 | 12 | retain |
| 2 | benchmark applications | 0 | 2 | retain |
| 3 | benchmark tools and deterministic contracts | 25 | 16 | retain |
| 4 | performance evidence and dated records | 19 | 222 | retain |
| 5 | handoff and release tooling | 5 | 0 | retain |
| hold | deferred WASM | 8 | 26 | hold outside release boundary |
| exclude | generated local Python cache | 0 | 15 | exclude |

The pre-record retained boundary is therefore exactly 337 paths: 85 tracked
modifications and 252 untracked paths. Its newline-delimited path-list
SHA-256 is
`8eadc0b87174a6d0c09a2fd7be0bc7ab90189e950ffc3bbe719567b0d66d9026`.
The deferred WASM boundary remains the same 34 paths recorded by the July 27
index review; its path-list SHA-256 is
`fa6e81afbeb5701dd948b2c951f07607c417a5e4a94dfdda8a884e9333086fa8`.
The 15 generated-local exclusions have path-list SHA-256
`fc216a58a80061f03846caa0860fb38fb2283f7e6dcd7383cf6cdfbeed77a0f0`.
No path is unmatched.

The retained path families are governed chronologically by the 28 new entries
in both `LOG.md` and `v12/LOG.md`. The logs contain exactly 28 unique dated
record references, and every referenced record exists. They cover:

- the Backup Dedup and Discrete Event Simulation application gates plus the
  typed generic-nominal storage census;
- scorecard hermeticity and correctness;
- positional nominal boundaries, generated struct-field reads, and
  struct-field reach;
- compiled concurrency, callable context, direct dispatch, split receivers,
  and closure-owned native callables;
- native interface metadata and the complete retained/rejected Awaitable,
  reverse-context, task-state, scratch, and materialization sequence; and
- the post-materialization owner closure, current strict scorecard closure,
  semantic-work reconciliation, and final correctness/release-readiness
  closure.

## Validation

- all 60 dirty JSON evidence files parse;
- all nine dirty gzip profiles decompress completely;
- the 241 evidence paths total 3,174,250 bytes;
- `git diff --check` passes;
- all 59 dirty Go files pass `gofmt -l`;
- none of 96 dirty Go, Python, shell, JavaScript, or Able source files reaches
  1,000 lines; the largest is
  `pkg/compiler/generator_render_runtime.go` at 999 lines;
- no v10, v11, deprecated in-tree stdlib, or external `able-stdlib` path is
  present;
- no secret-like filename or common private-key/service-token signature is
  present;
- the index remains clean; and
- no production source, canonical stdlib, dependency, or WASM path was changed
  by this inventory.

The project cleanup dry run reports 12.35 GiB across four generated paths,
including the compiler Go cache and the 15 manifest-visible Python bytecode
files. Those paths were identified but not deleted because this tranche is
explicitly non-mutating. Separately, an exact 10,684,494-byte stale external-Go
workspace at `/tmp/able-v12-extern-go` was deleted after confirming it had no
open handles; it was outside the repository and does not affect the manifest.

## Authorized exact-index closure

The maintainer explicitly authorized the exact-index review and local commit
described by the inventory handoff. The candidate is the sorted union of:

- all 337 pre-record paths marked `retain`; and
- this Markdown record, its JSON companion, and its 386-row TSV manifest.

The newline-delimited sorted 340-path candidate has SHA-256
`d48d0a39dc528bc23bf7f954c15cd6295eb2bd4814b6f8783a78ad731c8cfaa0`.
The exact NUL-delimited pathspec supplied to Git has SHA-256
`ee5b5f59aa62af0c290cd21729fb5c7607242b9a42bda8c408d3ceaae922f4f0`.
No broad `git add` pathspec is used.

The 337 non-self metadata rows are refreshed after the final handoff edits and
contain 133,359 lines and 8,378,527 bytes. Their line/byte/content/path
identity manifest has SHA-256
`cd9b14be68b4d0237b3dfee1cb87c1d0286e841e3a4f8450acde5d18e09172ee`.
Those rows are revalidated immediately before committing. The three inventory
metadata files are included by exact path. Cached whitespace, scope, binary,
common-secret, and path-list checks must pass before the one authorized local
commit is created. The complement must remain exactly 34 deferred WASM paths
plus 15 generated-local exclusions before optional generated-cache cleanup.

## Next recommendation

Request explicit maintainer authorization before publishing the resulting
local branch.

Why: the exact retained boundary is now consolidated locally, but pushing
changes remote state and is not implied by authorization to create a local
commit. The branch was already four commits ahead of `origin/master` before
this consolidation.

What it entails: inspect the final local commit and branch divergence, confirm
the remote destination and expected commit range, then push only after the
maintainer explicitly authorizes that external mutation.

Why it is important: this preserves the verified native-carrier and
interpreter-free compiler state without silently publishing inactive WASM,
machine-local artifacts, or an unintended commit range. Until a concrete
three-unlike-application admission invalidation appears, keep production
performance mutation paused.
