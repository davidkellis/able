# Exact publication reconciliation

Date: 2026-07-30

## Decision

The verified v12 performance consolidation is published and locally
reconciled. Live `refs/heads/master`, local `HEAD`, and local
`origin/master` all equal
`c43c4e6825504cff6486ef6ce74aaed48aebc7e5`, with zero divergence and an
empty index.

Retain this documentation-only reconciliation and no compiler, runtime,
interpreter, bytecode VM, benchmark measurement, canonical stdlib, language,
dependency, fixture, frozen-workspace, or WASM change.

## Published commit

The published commit has:

- subject: `perf(v12): consolidate verified native lowering work`;
- parent: `418886c70aee64b92b5bb3266ee5fe6453ac4320`;
- tree: `0c9eadbe760ce51f2843cd011dcf637ad005e899`;
- changed paths: 112;
- additions: 34,817;
- deletions: 1,680;
- added paths: 57; and
- modified paths: 55.

Its exact 7,868-byte NUL-delimited path set has SHA-256
`6ab8db8d4ede1872794ff79deca560dbdf901d782b498acca5c6a11a71549790`.
It matches the reviewed candidate and has zero intersection with the deferred
WASM boundary.

The publication used one non-force commit-to-branch refspec:

```text
c43c4e6825504cff6486ef6ce74aaed48aebc7e5:refs/heads/master
```

No tag, wildcard, secondary ref, force update, reset, revert, or history
rewrite was used.

## Pre-record worktree

Before this record, the fully expanded worktree contained exactly 34 paths:

- eight tracked modifications;
- 26 untracked files; and
- zero staged or generated-local paths.

The 1,516-byte NUL-delimited porcelain snapshot has SHA-256
`6b8e529a5469bc7993cdd9d134df22ac08581647b7ce792e990b401a6eab0e6f`.

Every path is a member of the deferred WASM boundary. All state, line, byte,
and content hashes reproduce the authoritative release inventory. The
1,414-byte sorted deferred path list retains SHA-256
`7f29de27f03da83138df19696be182a5139aa6536ba0bdb363a1a23ffafd1f4e`.

## Current performance state

The published evidence remains unchanged:

- extern builder SHA-256:
  `8f388d679a079093a6e48ec6f0d1938ba8a12d475ed9565f6832b952b2e3b276`;
- 132-row scorecard SHA-256:
  `3652695dc7b1576ed4729ef30a7688b171114cda9b4ce269132fd868b37849f3`;
- combined frontier SHA-256:
  `d819bb32c402d690b9e89926b5bb8a0fddd6f18d9b966f8c85b3d9f73baa0a76`;
- 23-entry closure-ledger SHA-256:
  `e91bc0ea504ec7e79615b01265cba0359da1378790f596bde33af9345ba35c62`;
- compiled rows: 66;
- bytecode rows: 66;
- frontier guards: ten;
- frontier misses: 122;
- actionable groups: zero;
- aggregate target excess: 277.200421 seconds; and
- current closures: 23, with zero invalidations.

The complete release-readiness suite remains authoritative for the published
source. The cleanup preview is empty. No build was repeated merely to record
remote convergence.

## Documentation-only handoff

This record, its JSON companion, `PLAN.md`, `LOG.md`, and `v12/LOG.md` form an
exact five-path documentation candidate.

Its path identities are:

- newline-delimited list: 170 bytes, SHA-256
  `5a82e862fddbcbbca884260cc9aeefd19abc4ad66e996faf45bb89da1bff978b`;
  and
- NUL-delimited pathspec: 170 bytes, SHA-256
  `6c1b2dcd876f587044d8e9ff80f9edaf5b608564cada444f238a9c127ef53097`.

The three non-self handoff identities are refreshed after their final edits:

- rows: 3;
- total lines: 60,730;
- total bytes: 4,179,497;
- manifest bytes: 307; and
- manifest SHA-256:
  `fd572408e9998f034b05a8787de2b96b5a7637e776079cea5618783b528b594d`.

The final expanded worktree contains exactly 39 paths: the five-path
documentation candidate and unchanged 34-path deferred WASM complement.
Nothing was staged, committed, or pushed in this reconciliation tranche.
All audit state lived under disk-backed `/var/tmp`; the final 24 KiB task
workspace was removed after preserving this record.

## Next

Keep production performance mutation paused and run a non-mutating
post-publication invalidation-trigger audit before opening another compiler or
VM tranche.

Why: the published frontier has zero actionable groups and all 23 closures
are current. A new production candidate is admissible only if a new
application, correctness repair, source change, canonical-stdlib change, or
observer result invalidates a checked identity and exposes one exact material
owner in three unlike programs.

What it entails: compare the published source and evidence identities with
the current external benchmark catalog, canonical stdlib state, correctness
results, and any new observer output. Recompute only affected evidence if an
identity changed; otherwise retain no code and record that the selector stays
empty.

Why it matters: this resumes work from evidence rather than release momentum,
protecting native carriers and interpreter-free compiled graphs from
speculative, benchmark-specific, or already-closed optimization routes.

Do not begin WASM work.
