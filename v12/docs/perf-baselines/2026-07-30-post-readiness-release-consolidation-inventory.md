# Post-readiness release consolidation inventory

Date: 2026-07-30

## Decision

The verified post-publication v12 worktree has an exact release boundary:

- retain 107 pre-record paths across eight dependency-ordered review stages;
- add this Markdown record, its JSON companion, and the 148-row TSV manifest
  to form an exact 110-path release candidate;
- hold the unchanged 34-path deferred WASM boundary outside the candidate;
- exclude seven Python bytecode files as generated-local state; and
- retain no unmatched path.

Nothing was staged, committed, pushed, reset, reverted, deleted, or modified
inside the deferred WASM boundary. No production performance candidate was
opened.

## Snapshot identity

The immutable snapshot was captured before the three inventory files existed:

- `HEAD` and local `origin/master`:
  `418886c70aee64b92b5bb3266ee5fe6453ac4320`;
- branch relation: zero ahead and zero behind;
- `HEAD` and index tree:
  `b497c7064a80e0992da3f00969fb4a1cbc7c6e64`;
- staged paths: zero;
- fully expanded dirty files: 148;
- tracked modifications: 63;
- untracked files: 85;
- tracked deletions and renames: zero; and
- tracked diff: 5,040 additions and 1,733 deletions.

The 9,703-byte newline-delimited porcelain snapshot has SHA-256
`6e6f19f07fe75d95256d396a565f38f03cc2c7afd4b54503a6674c990c324feb`.
Its 9,703-byte NUL-delimited form has SHA-256
`5a7206a07a5c1264a37cce602021767a2249ea52c9f5d871854133c35161a829`.
The sorted 9,259-byte newline-delimited path list has SHA-256
`1c11a96fd3ba557967710c0eb8c61545a77bd1d72590508310326828cbe9cbe9`.

The exact
`2026-07-30-post-readiness-release-consolidation-inventory.tsv` manifest
contains 148 data rows, 34,362 bytes, and has SHA-256
`fb980f00c87c6862ebbea7db7ba7ba688e9379a72924d703169c1f7276853352`.
The three metadata files are intentionally not self-referential rows.

## Review boundaries

| Order | Boundary | Tracked | Untracked | Paths | Lines | Bytes | Disposition |
| ---: | --- | ---: | ---: | ---: | ---: | ---: | --- |
| 1 | performance direction model | 32 | 2 | 34 | 8,436 | 385,989 | retain |
| 2 | sustained workload depth | 0 | 4 | 4 | 2,094 | 77,935 | retain |
| 3 | telemetry application gate | 7 | 19 | 26 | 14,399 | 1,040,702 | retain |
| 4 | captured-callable native carrier | 3 | 8 | 11 | 7,274 | 484,912 | retain |
| 5 | post-captured reconciliation | 0 | 2 | 2 | 323 | 16,365 | retain |
| 6 | compiled static-boundary census | 0 | 6 | 6 | 8,797 | 306,433 | retain |
| 7 | current performance state | 9 | 0 | 9 | 10,369 | 1,015,624 | retain |
| 7 | extern toolchain readiness | 1 | 11 | 12 | 2,051 | 305,710 | retain |
| 8 | handoff and release state | 3 | 0 | 3 | 60,541 | 4,168,530 | retain |
| hold | deferred WASM | 8 | 26 | 34 | 3,266 | 103,984 | hold |
| exclude | generated local | 0 | 7 | 7 | 196 | 34,476 | exclude |

The 107-path pre-record retained boundary has 114,284 lines and 7,802,200
bytes. Its 7,450-byte sorted path list has SHA-256
`6c63053b86669b5b51e7f2fc12ee01acd2689ddade069dc318c89c98aff8355a`.

The deferred WASM boundary has the same 34 paths, line counts, byte counts,
and content hashes as the prior authoritative publication-handoff inventory.
Its 1,414-byte path list retains SHA-256
`7f29de27f03da83138df19696be182a5139aa6536ba0bdb363a1a23ffafd1f4e`.

The seven excluded files are all under `v12/__pycache__/`. Their 395-byte
path list has SHA-256
`74355a2b6b5d6d0db5867ca2acdae9e4926317771333ec4afcf4b7d21617b7ac`.

## Validation

- All 148 snapshot identities and classifications reproduce.
- All 54 dirty JSON files parse.
- All 18 dirty Go files are formatted.
- All nine dirty Python files parse without writing bytecode.
- All four dirty shell programs and ten dirty JavaScript modules pass syntax
  validation.
- Whitespace, final-newline, source-size, scope, and secret-pattern checks
  pass.
- All 46 maintained source files are below 1,000 lines; the largest is
  `v12/bench_external_catalog.sh` at 883 lines.
- The current extern builder, 132-row scoreboard, combined frontier, and
  23-entry closure ledger reproduce the exact identities in the retained
  release-readiness record.
- The frontier retains ten guards, 122 misses, zero actionable groups, and
  277.200421 seconds of aggregate target excess. All 23 closures are current.
- The release-readiness record retains the complete canonical stdlib,
  ordinary runner, architecture, static-check, 640-partition compiler, and
  generated-Go CLI passes for the exact production source.
- The index is empty and local `HEAD` equals local `origin/master`.

No compiler, generated runtime, runtime, interpreter, bytecode VM, parser
semantic, canonical stdlib, language, dependency, benchmark measurement,
fixture behavior, frozen workspace, or WASM behavior changed in this
inventory.

## Exact post-record candidate

The exact retained candidate is the 107 retained rows plus the three inventory
files. It contains 110 paths.

Its 7,708-byte sorted newline-delimited path list has SHA-256
`c1e7dc8f5ee052e52251ce62273f2770429d3c322acb3208750f1c306129dc21`;
its exact 7,708-byte NUL-delimited pathspec has SHA-256
`08b25c1b1ca46a0721994ff256cd12be732155da186af2f60bffd236a9dde771`.
Its complement is exactly the 34 deferred WASM paths and seven generated-local
paths.

Because final PLAN and log edits occur after the snapshot, the 107 non-self
retained identities are refreshed separately:

- rows: 107;
- total lines: 114,335;
- total bytes: 7,805,332;
- manifest bytes: 16,861; and
- manifest SHA-256:
  `4de1bc465797cdfeb1e34e6248c407dcfa5e05a49d5749c2a7a2481540449411`.

The final expanded worktree contains exactly 151 paths: the 110-path
candidate, 34 deferred WASM paths, and seven generated-local paths.

## Cleanup preview

The guarded read-only cleanup preview found four generated project paths:

| Path | Reclaimable |
| --- | ---: |
| `v12/tmp` | 0 KiB |
| `v12/interpreters/go/.gocache` | 1,742,064 KiB |
| `v12/interpreters/go/tmp` | 0 KiB |
| `v12/__pycache__` | 52 KiB |
| **Total** | **1,742,116 KiB** |

The cleanup was deliberately not applied because this tranche forbids
project deletion. All inventory work used disk-backed `/var/tmp`; the final
128 KiB task workspace was removed after preserving these records.

## Next

Run the guarded generated-local cleanup, then verify the exact 110-path
candidate and unchanged 34-path deferred complement without staging.

Why: the release inventory is complete, but seven visible Python bytecode
files and 1.66 GiB of ignored generated caches remain outside the candidate.

What it entails: first confirm the selected caches have no active process or
open-handle owner, run the repository's guarded cleanup in apply mode, and
recapture Git status. Then reproduce the candidate pathspec, the deferred
WASM identity, formatting and evidence gates, and obtain explicit maintainer
authorization before any exact staging or commit.

Why it matters: removing only reproducible local state leaves one auditable
110-path v12 candidate and one protected deferred boundary, preventing cache
artifacts or WASM work from entering a future release operation.

Do not begin WASM work.
