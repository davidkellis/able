# Post-census release inventory

Date: 2026-07-29

## Decision

The maintainer explicitly authorized exact staging and one local commit for
the eight-path retained candidate. The authorization does not include a push,
reset, revert, repository deletion, broad staging operation, or modification
of the 34 deferred WASM files.

## Snapshot identity

The inventory was captured before this record, its JSON companion, or its TSV
manifest existed:

- `HEAD`: `5f84e42c43d3d514c231c7887e835895a9ba557a`;
- `origin/master`: `5f84e42c43d3d514c231c7887e835895a9ba557a`;
- branch relation: zero commits ahead and zero behind;
- `HEAD` tree: `41539a64d2ecf89f9de3cb8a90510472d62e4387`;
- index tree: `41539a64d2ecf89f9de3cb8a90510472d62e4387`;
- staged paths: zero;
- fully expanded dirty files: 39;
- tracked modifications: 11;
- untracked files: 28;
- deleted or renamed files: zero; and
- tracked diff: 658 additions and 65 deletions.

The authoritative snapshot uses `--untracked-files=all`. Its 1,727-byte
NUL-delimited porcelain SHA-256 is
`79ebe4d68010863e20a025188e87c431dd970991b02091caea84e6f1b5800d13`.

The exact manifest is `2026-07-29-post-census-release-inventory.tsv`. It
contains 39 data rows, 6,810 bytes, and has SHA-256
`f60fbd1bb26fa160a7e0feca4fda6d2c8e0000be8def5fd51c8cbeddafc353b2`.
The three inventory metadata files are intentionally not self-referential
snapshot rows.

## Dependency-ordered boundaries

| Order | Boundary | Tracked | Untracked | Paths | Disposition |
| ---: | --- | ---: | ---: | ---: | --- |
| 1 | compiled native-carrier census reconciliation | 0 | 2 | 2 | retain |
| 2 | handoff and release boundary | 3 | 0 | 3 | retain |
| hold | deferred WASM | 8 | 26 | 34 | hold outside release boundary |

The pre-record retained boundary is exactly five paths: three tracked
modifications and two untracked files. It contains 57,602 lines and
3,993,986 bytes. Its sorted newline-delimited path-list SHA-256 is
`4900de0d696d2771abb014587346bf2c01bd2170719d09db1dd1617e8b34a652`.

The deferred boundary contains 34 fully expanded files. Its path-list
SHA-256 is
`7f29de27f03da83138df19696be182a5139aa6536ba0bdb363a1a23ffafd1f4e`,
which exactly matches the post-spawn and post-security inventories. No new
WASM-like path exists.

The census-record path-list SHA-256 is
`edfc3ad7e804ff993d20112f9da22053690966701968b5b333fe039f19a7a450`.
The handoff path-list SHA-256 is
`f9a85fb7af9ed46f62707946d78bc8d9840068d7105f2f1a4fde6e33ea3620e4`.

## Record coverage

- All five retained paths are governed by
  `2026-07-29-compiled-native-carrier-census-reconciliation.md`.
- The 34 deferred files remain governed by the deferred WASM hold.
- Every governing record exists; there are no unmatched or generated-local
  paths.

Review order reflects dependencies:

1. The reconciliation proves that the completed strict-corpus census and
   current 23-entry closure ledger admit no repeated performance work.
2. The plan and logs preserve the no-code decision and next release boundary.

## Validation

- All 39 manifest line, byte, and SHA-256 identities reproduce.
- All snapshot JSON files parse.
- All dirty Go files are formatted.
- All dirty JavaScript modules pass syntax validation.
- Tracked and untracked whitespace checks pass.
- No dirty maintained Go, JavaScript, or Able source reaches 1,000 lines.
- No v10, v11, deprecated in-tree stdlib, external `able-stdlib`, or
  out-of-scope repository path is present.
- No secret-like filename or common private-key/service-token signature is
  present.
- The 126-row five-sample scoreboard, zero-actionable frontier, and
  23-current-entry closure ledger reproduce.
- The index remains empty.

No compiler, generated-runtime, runtime, interpreter, VM, canonical stdlib,
language, benchmark, fixture, reference implementation, dependency, frozen
workspace, or WASM behavior changed during this inventory.

## Exact post-record candidate

If the maintainer later authorizes exact local consolidation, the candidate is
the sorted union of:

- all five pre-record paths marked `retain`; and
- this Markdown record, its JSON companion, and the 39-row TSV manifest.

The candidate contains exactly eight paths. Its sorted newline-delimited
path-list SHA-256 is
`f5fca8d15cd6c3b1ba37379e73f0c52a65c970fe0b25f8bf6c156124e2b366d8`;
the exact NUL-delimited pathspec SHA-256 is
`9a3d36a00f6beba53cedb261e3ef54229c61545e758594db432caf2adb46b96e`.
Its complement must remain exactly the 34 deferred WASM files.

The immutable manifest records pre-record identities. Because final plan and
log edits occur after that snapshot, all five non-self identities were
refreshed and revalidated before staging.

## Authorized consolidation identity

The refreshed five-path non-self identity has:

- 5 rows;
- 57,671 total lines;
- 3,997,574 total bytes;
- a 574-byte sorted identity manifest; and
- identity-manifest SHA-256
  `26a2cad521fc7248ee193b7bb76cbe1f7efc750ca4b07567aa4ba53cf79a3b76`.

The three inventory metadata files remain excluded from this identity to
avoid self-reference. The exact eight-path candidate and unchanged 34-path
deferred complement are revalidated immediately before staging.

## Next

Obtain explicit maintainer authorization before publishing the local
post-census consolidation.

Why: this tranche authorizes one exact local commit but does not authorize
remote mutation.

What it entails: verify the final one-commit divergence and remote
destination, then push only that exact commit-to-branch refspec if explicitly
authorized.

Why it matters: the evidence-backed stopping decision can reach shared history
without publishing deferred WASM or unintended local history.
