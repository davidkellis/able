# Post-security release inventory

Date: 2026-07-29

## Decision

The maintainer explicitly authorized exact staging and one local commit for
the 12-path retained candidate. The authorization does not include a push,
reset, revert, repository deletion, broad staging operation, or modification
of the 34 deferred WASM files.

## Snapshot identity

The inventory was captured before this record, its JSON companion, or its TSV
manifest existed:

- `HEAD`: `6efad0a53120129510fdfbab7fbcc84dcd081768`;
- `origin/master`: `6efad0a53120129510fdfbab7fbcc84dcd081768`;
- branch relation: zero commits ahead and zero behind;
- `HEAD` tree: `00f1c924a464725b9f69903f7e03d2b1cf37b28f`;
- index tree: `00f1c924a464725b9f69903f7e03d2b1cf37b28f`;
- staged paths: zero;
- fully expanded dirty files: 43;
- tracked modifications: 13;
- untracked files: 30;
- deleted or renamed files: zero; and
- tracked diff: 715 additions and 75 deletions.

The authoritative snapshot uses `--untracked-files=all`. Its 1,891-byte
NUL-delimited porcelain SHA-256 is
`578f593e56565daa5c198272bf7993825985e14a49d175c7160254be4e9cdc06`.

The exact manifest is
`2026-07-29-post-security-release-inventory.tsv`. It contains 43 data rows,
7,480 bytes, and has SHA-256
`bfe9137e0bc61391edc75116e1a0acca8f2a448787729d75d64b9f7ab3f76b87`.
The three inventory metadata files are intentionally not self-referential
snapshot rows.

## Dependency-ordered boundaries

| Order | Boundary | Tracked | Untracked | Paths | Disposition |
| ---: | --- | ---: | ---: | ---: | --- |
| 1 | dependency-alert attribution | 0 | 2 | 2 | retain |
| 2 | v12 x/net security refresh | 2 | 2 | 4 | retain |
| 3 | handoff and release boundary | 3 | 0 | 3 | retain |
| hold | deferred WASM | 8 | 26 | 34 | hold outside release boundary |

The pre-record retained boundary is exactly nine paths: five tracked
modifications and four untracked files. It contains 57,975 lines and
4,009,191 bytes. Its sorted newline-delimited path-list SHA-256 is
`f676cad03e2e51efbae18aed027c993fe624614d20f7e28d2418c83d2c4b01dd`.

The deferred boundary contains 34 fully expanded files. Its path-list
SHA-256 is
`7f29de27f03da83138df19696be182a5139aa6536ba0bdb363a1a23ffafd1f4e`,
which exactly matches both previous authoritative inventories. No new
WASM-like path exists.

## Record coverage

- 2 paths are governed by
  `2026-07-29-dependency-alert-attribution.md`.
- 7 paths, including the three handoff files, are governed by
  `2026-07-29-v12-x-net-security-refresh.md`.
- 34 files remain governed by the deferred WASM hold.
- Every governing record exists; there are no unmatched or generated-local
  paths.

Review order reflects dependencies:

1. The attribution establishes zero reachable active-v12 vulnerabilities and
   separates frozen historical exposure.
2. The exact module update removes the seven fixable `x/net` advisories while
   preserving zero reachable findings and all correctness/performance gates.
3. The plan and logs describe the retained result and next release boundary.

## Validation

- All 43 manifest line, byte, and SHA-256 identities reproduce.
- All 6 dirty JSON files parse.
- All 11 dirty Go files are formatted.
- All 10 dirty JavaScript modules pass syntax validation.
- The tracked and untracked whitespace checks pass.
- None of 24 dirty maintained Go, JavaScript, or Able source files reaches
  1,000 lines. The largest is `v12/wasm/ast_adapter.mjs` at 402 lines.
- No v10, v11, deprecated in-tree stdlib, external `able-stdlib`, or
  out-of-scope repository path is present.
- No secret-like filename or common private-key/service-token signature is
  present.
- The retained `go.mod` and `go.sum` identities match the verified security
  refresh.
- The 126-row five-sample scorecard, zero-actionable frontier, and
  23-current-entry closure ledger reproduce.
- The index remains empty.

The cleanup dry run found no generated project artifact. The security scan
and refresh workspaces were already removed, no repository Go or Python cache
exists, and no `/tmp/able-*` path remains.

No compiler, generated-runtime, runtime, interpreter, VM, canonical stdlib,
language, benchmark, fixture, reference implementation, frozen workspace, or
WASM behavior changed during this inventory.

## Exact post-record candidate

If the maintainer later authorizes exact local consolidation, the candidate is
the sorted union of:

- all nine pre-record paths marked `retain`; and
- this Markdown record, its JSON companion, and the 43-row TSV manifest.

The candidate contains exactly 12 paths. Its sorted newline-delimited
path-list SHA-256 is
`1e668bf34401fab7c45c151051abf96eedbcae3217281e7ba934b5c4825e945f`;
the exact NUL-delimited pathspec SHA-256 is
`d798a1a728506d7ebce8c7bea0540b83830ec1974d54beb3fd2193f6773ec642`.
Its complement must remain exactly the 34 deferred WASM files.

The immutable manifest records pre-record identities. Because the final plan
and log edits occur after that snapshot, all nine non-self identities were
refreshed and revalidated before staging.

## Authorized consolidation identity

The refreshed nine-path non-self identity has:

- 9 rows;
- 58,053 total lines;
- 4,013,417 total bytes;
- a 1,023-byte sorted identity manifest; and
- identity-manifest SHA-256
  `49c2b8bc889b1a022c8763823efbc6adf5eedda8ba98b317490a1b48068b60c2`.

The three inventory metadata files remain excluded from this identity to
avoid self-reference. The exact 12-path candidate and its unchanged 34-path
deferred complement are revalidated immediately before staging.

## Next

Obtain explicit maintainer authorization before publishing the local
post-security consolidation.

Why: this tranche authorizes one exact local commit but does not authorize
remote mutation.

What it entails: verify the final one-commit divergence and remote
destination, then push only that exact commit-to-branch refspec if explicitly
authorized.

Why it matters: the verified security correction can reach shared history
without publishing deferred WASM or unintended local history.
