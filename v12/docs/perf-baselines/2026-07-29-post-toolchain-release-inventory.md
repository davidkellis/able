# Post-toolchain release inventory

Date: 2026-07-29

## Decision

The release inventory for the compiled Go 1.26.5 scorecard and benchmark
toolchain-provenance tranches is complete. All current files are classified,
no path is unmatched, and no deferred WASM file entered the retained boundary.

The maintainer explicitly authorized exact staging and one local
consolidation commit for the 90-path retained candidate. This authorization
does not include pushing, resetting, reverting, deleting, or modifying the
34-file deferred WASM boundary.

## Snapshot identity

The inventory was captured before this record, its JSON companion, or its TSV
manifest existed:

- `HEAD`: `2243143aa449c9c764d7215496faf473b19fc73d`;
- `origin/master`: `2243143aa449c9c764d7215496faf473b19fc73d`;
- branch relation: zero commits ahead and zero behind;
- staged paths: zero;
- index SHA-256:
  `01ad931402aaa751ad129b6156f9cdef74d6dac1fd7709a26631d0ed71e5fb52`;
- fully expanded dirty files: 121;
- tracked modifications: 46;
- untracked files: 75;
- deleted or renamed files: zero; and
- tracked diff: 2,052 additions and 1,234 deletions.

Ordinary porcelain output reports 119 entries because it collapses
`v12/wasm/samples/modules/` into one untracked directory entry. The
authoritative snapshot uses `--untracked-files=all`, expands the directory to
three files, and therefore contains 121 file rows. Its NUL-delimited porcelain
SHA-256 is
`7ffeb04e770eac5d95f8c0545dc0b5cb925dca0c887bca2640ea4e089d44436e`.

The exact manifest is
`2026-07-29-post-toolchain-release-inventory.tsv`. It contains 121 rows,
27,076 bytes, and has SHA-256
`cd310513da21807d741c039109f42b32763020d2d86ba7e49fbb2ce91d9b795f`.
The three inventory metadata files are intentionally not self-referential
snapshot rows.

## Dependency-ordered boundaries

| Order | Boundary | Tracked | Untracked | Paths | Disposition |
| ---: | --- | ---: | ---: | ---: | --- |
| 1 | benchmark toolchain contract | 5 | 2 | 7 | retain |
| 2 | compiled Go 1.26.5 scorecard evidence | 3 | 43 | 46 | retain |
| 3 | derived frontier reconciliation | 27 | 2 | 29 | retain |
| 4 | handoff and release boundary | 3 | 2 | 5 | retain |
| hold | deferred WASM | 8 | 26 | 34 | hold outside release boundary |

The pre-record retained boundary is exactly 87 paths: 38 tracked
modifications and 49 untracked files. It contains 97,475 lines and 6,513,255
bytes. Its sorted newline-delimited path-list SHA-256 is
`ae573ff2d8824c196a679d39167beb0d082b5f0936b63e6c8d74111f2614c3b1`.

The deferred boundary contains 34 fully expanded files. Its path-list
SHA-256 is
`7f29de27f03da83138df19696be182a5139aa6536ba0bdb363a1a23ffafd1f4e`,
which exactly matches the deferred set in the 2026-07-28 post-frontier
inventory. The earlier 32 count came from collapsed porcelain entries, not a
change to the hold.

## Record coverage

- 76 paths are governed by
  `2026-07-29-compiled-go1265-scorecard-refresh.md`.
- 11 paths are governed by
  `2026-07-29-benchmark-go-toolchain-contract.md`.
- 34 files remain governed by the deferred WASM hold.
- Every governing dated record exists; there are no unmatched or
  generated-local paths.

Review order reflects dependencies:

1. The benchmark tooling establishes exact Go selector/version propagation and
   mismatch rejection.
2. The 630-process compiled cohort and stability evidence depend on that
   measurement contract.
3. The scorecard, frontier, closure ledger, and downstream architecture
   records derive from that cohort.
4. The plan, logs, and dated records describe the verified result and release
   boundary.

## Validation

- All 121 manifest line, byte, and SHA-256 identities reproduce.
- All 45 dirty JSON files parse.
- All 11 dirty Go files are formatted.
- Shell syntax, Python compilation, and `git diff --check` pass.
- All 31 benchmark contract test files pass with a 60-second ceiling.
- The authoritative 126-row scorecard, zero-actionable-group frontier,
  23-closure ledger, and five-sample evidence check reproduce.
- None of 36 dirty maintained Go, Python, shell, JavaScript, or Able source
  files reaches 1,000 lines. The largest is
  `v12/bench_compare_external` at 974 lines.
- No v10, v11, deprecated in-tree stdlib, external `able-stdlib`, or
  out-of-scope repository path is present.
- No secret-like filename or common private-key/service-token signature is
  present.
- The index remains empty.

The cleanup dry run found 141,144,046 bytes (0.131 GiB) across the ignored
`v12/tmp`, `v12/interpreters/go/.gocache`, and
`v12/interpreters/go/tmp` paths. Nothing was removed because this tranche is
inventory-only; the exact temporary workspaces created by the two preceding
tranches were already deleted.

No compiler, generated-runtime, runtime, interpreter, VM, canonical stdlib,
language, dependency, benchmark, fixture, reference implementation, or WASM
behavior changed during this inventory.

## Exact post-record candidate

If the maintainer later authorizes exact local consolidation, the candidate is
the sorted union of:

- all 87 pre-record paths marked `retain`; and
- this Markdown record, its JSON companion, and the 121-row TSV manifest.

The candidate contains exactly 90 paths. Its sorted newline-delimited path-list
SHA-256 is
`62825489bb4ac86248adf440705c2e76f6e3da4124ffae3981399fade79da4fb`;
the exact NUL-delimited pathspec SHA-256 is
`ef537311e2af6a320989f5fbe665ab7cc5f92483aa404f5771f2ac587c27b562`.
Its complement must remain exactly the 34 deferred WASM files.

The immutable manifest records pre-record identities. Because the final plan
and log edits occur after that snapshot, all 87 non-self identities must be
refreshed and revalidated before any separately authorized staging operation.

## Authorized consolidation identity

After the final authorization handoff edits, the refreshed 87 pre-record
identity rows contain 97,570 lines and 6,518,345 bytes. Their 12,508-byte
line/byte/content/path manifest has SHA-256
`fe73a8fa26b2b47657e3249313f8d4180a49184ade96c2a36c80f31015a0357d`.

The exact 90-path candidate and 34-file deferred complement must be
revalidated immediately before staging. Cached identity, JSON, formatting,
syntax, whitespace, source-size, scope, secret, evidence, and path-complement
checks must pass before the one authorized local commit. No broad `git add`
pathspec or push is authorized.

## Next

Obtain explicit maintainer authorization before publishing the local commit.

Why: exact local consolidation does not authorize mutation of the remote
repository.

What it entails: inspect the final commit and branch divergence, confirm the
remote destination and exact commit range, and push only after separate
explicit authorization.

Why it matters: this prevents publishing deferred WASM or an unintended
history range while preserving the verified Go 1.26.5 performance baseline and
toolchain contract.
