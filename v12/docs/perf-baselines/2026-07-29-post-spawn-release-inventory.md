# Post-spawn release inventory

Date: 2026-07-29

## Decision

The release inventory for the retained compiler spawn-context, post-callable
scorecard, post-spawn scorecard, residual-owner closure, and correctness
tranches is complete. Every dirty file is classified, no path is unmatched,
and the deferred WASM boundary is unchanged.

The maintainer explicitly authorized exact staging and one local
consolidation commit for the 147-path retained candidate. This authorization
does not include pushing, resetting, reverting, deleting repository work, or
modifying the 34-file deferred WASM boundary.

## Snapshot identity

The inventory was captured before this record, its JSON companion, or its TSV
manifest existed:

- `HEAD`: `ada2a21c751baf51200149e9dac2d175e29aa222`;
- `origin/master`: `ada2a21c751baf51200149e9dac2d175e29aa222`;
- branch relation: zero commits ahead and zero behind;
- `HEAD` tree: `b1ee83f77176596c971432b4df02a85956526149`;
- index tree: `b1ee83f77176596c971432b4df02a85956526149`;
- staged paths: zero;
- fully expanded dirty files: 178;
- tracked modifications: 29;
- untracked files: 149;
- deleted or renamed files: zero; and
- tracked diff: 1,788 additions and 1,469 deletions.

The authoritative snapshot uses `--untracked-files=all`. Its NUL-delimited
porcelain SHA-256 is
`3d170cce99ff5931b1ac3fd0fa53d9fbff2ff82156b282d88c0db8abdd60e54e`.

The exact manifest is
`2026-07-29-post-spawn-release-inventory.tsv`. It contains 178 data rows,
43,507 bytes, and has SHA-256
`65e0882ebe95edabbba47ef1ece601de6666261beb03d16110b78b0509722b1c`.
The three inventory metadata files are intentionally not self-referential
snapshot rows.

## Dependency-ordered boundaries

| Order | Boundary | Tracked | Untracked | Paths | Disposition |
| ---: | --- | ---: | ---: | ---: | --- |
| 1 | compiler spawn-context | 12 | 7 | 19 | retain |
| 2 | comparison-row selection tool | 0 | 2 | 2 | retain |
| 3 | post-callable scorecard evidence | 0 | 70 | 70 | retain |
| 4 | post-spawn scorecard and correctness | 6 | 44 | 50 | retain |
| 5 | handoff and release boundary | 3 | 0 | 3 | retain |
| hold | deferred WASM | 8 | 26 | 34 | hold outside release boundary |

The pre-record retained boundary is exactly 144 paths: 21 tracked
modifications and 123 untracked files. It contains 118,103 lines and
7,865,388 bytes. Its sorted newline-delimited path-list SHA-256 is
`6e52c064cb81bc8efd347ec7939a4b7c3c5a62bf64ac3e26cb041fe6981dfdf7`.

The deferred boundary contains 34 fully expanded files. Its path-list
SHA-256 is
`7f29de27f03da83138df19696be182a5139aa6536ba0bdb363a1a23ffafd1f4e`,
which exactly matches the previous authoritative release inventory. No new
WASM-like path exists.

## Record coverage

- 19 paths are governed by
  `2026-07-29-compiled-spawn-gated-callable-context-retained.md`.
- 70 paths are governed by
  `2026-07-29-post-callable-context-scorecard-refresh.md`.
- 2 paths are governed by
  `2026-07-29-post-spawn-context-scorecard-and-owner-closure.md`.
- 53 paths are governed by
  `2026-07-29-post-spawn-correctness-release-gate.md`.
- 34 files remain governed by the deferred WASM hold.
- Every governing dated record exists; there are no unmatched or
  generated-local paths.

Review order reflects dependencies:

1. The compiler changes and semantic guards establish spawn-gated callable
   context without changing native data carriers or fallback boundaries.
2. The evidence utility establishes the exact scorecard-row selection
   contract.
3. The post-callable evidence records the broad A/B/Go admission case.
4. The post-spawn evidence, scorecard, frontier, closure ledger, and
   correctness gate derive from the retained compiler state.
5. The plan, logs, and this release boundary describe the verified handoff.

## Validation

- All 178 manifest line, byte, and SHA-256 identities reproduce.
- All 64 dirty JSON files parse; there are no dirty gzip files.
- All 22 dirty Go files are formatted.
- Python, shell, Node, and `git diff --check` validation pass.
- The 126-row scorecard, zero-actionable-group frontier, 23-current-entry
  closure ledger, and five-sample evidence checks reproduce.
- None of 39 dirty maintained Go, Python, shell, JavaScript, or Able source
  files reaches 1,000 lines. The largest is
  `v12/interpreters/go/pkg/compiler/compiler_dispatch_completeness_test.go`
  at 997 lines.
- No v10, v11, deprecated in-tree stdlib, external `able-stdlib`, or
  out-of-scope repository path is present.
- No secret-like filename or common private-key/service-token signature is
  present.
- The index remains empty.

The cleanup dry run found 8,492,171,264 allocated bytes (7.909 GiB) across
the ignored `v12/tmp`, `v12/interpreters/go/.gocache`, and
`v12/interpreters/go/tmp` paths. Nothing was removed because this tranche is
inventory-only. All disposable `/var/tmp` workspaces from the preceding
measurement and correctness tranches were already removed, and no
`/tmp/able-*` path remains.

No compiler, generated-runtime, runtime, interpreter, VM, canonical stdlib,
language, dependency, benchmark, fixture, reference implementation, or WASM
behavior changed during this inventory.

## Exact post-record candidate

If the maintainer later authorizes exact local consolidation, the candidate is
the sorted union of:

- all 144 pre-record paths marked `retain`; and
- this Markdown record, its JSON companion, and the 178-row TSV manifest.

The candidate contains exactly 147 paths. Its sorted newline-delimited
path-list SHA-256 is
`9aee340d93e8699b2501c5ec619ae3eead6dbf3e56c83a30cd7e3e471dfaf5d9`;
the exact NUL-delimited pathspec SHA-256 is
`38c9b5374272e97b50255caf214b837c2b44b0914ce411bf14ddbb5ce681e3b2`.
Its complement must remain exactly the 34 deferred WASM files.

The immutable manifest records pre-record identities. Because the final plan
and log edits occur after that snapshot, all 144 non-self identities were
refreshed and revalidated before exact staging.

## Authorized consolidation identity

After the final authorization handoff edits, the refreshed 144 pre-record
identity rows contain 118,183 lines and 7,869,977 bytes. Their 22,527-byte
line/byte/content/path manifest has SHA-256
`483364c319e0c4a6714e4f2f711c87c641f0b57b86ca76e442577eb5e42a8ebb`.

The exact 147-path candidate and 34-file deferred complement must reproduce
immediately before staging. Cached identity, JSON, formatting, syntax,
whitespace, source-size, scope, secret, evidence, and path-complement checks
must pass before the one authorized local commit. No broad `git add` pathspec
or push is authorized.

## Next

Obtain explicit maintainer authorization before publishing the local commit.

Why: exact local consolidation does not authorize mutation of the remote
repository.

What it entails: inspect the final commit and branch divergence, confirm the
remote destination and exact commit range, and push only after separate
explicit authorization.

Why it matters: this prevents publishing deferred WASM or an unintended
history range while preserving the verified native-carrier,
interpreter-free, scorecard, and correctness state.
