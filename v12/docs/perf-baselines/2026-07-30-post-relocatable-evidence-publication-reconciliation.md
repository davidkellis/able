# Post-relocatable-evidence publication reconciliation

Date: 2026-07-30

## Decision

The published path-relocation and native-WASM-scope corrections reproduce from
an exact clean export, but the complete evidence ledger is not yet
checkout-independent.

The scoreboard, five-sample evidence, focused relocation tests, and frontier
all pass. The ledger's canonical-stdlib scope definition still embeds the
absolute developer-checkout path. An adjacent stdlib with identical files and
content therefore invalidates all 23 closures solely because its absolute root
is different.

Retain no code from this reconciliation. Advance one general evidence-tooling
correction that gives a default repository-adjacent scope a stable semantic
root identity while preserving its complete content hash.

## Clean export

The source-only export was created from:

- commit: `9a4eac717e1a46e10195e17c562b389fa452dcfd`;
- tree: `eed160e45beec4ac93292b7d8e708376811311f6`;
- regular files: 10,966; and
- deferred bytecode proof files present: zero.

The export lived on disk-backed `/var/tmp` (`btrfs`), independently of the
developer worktree and its 34 deferred WASM paths.

The adjacent `able-stdlib` source reproduced the canonical evidence identity:

- recorded Git identity: `219eff222c28406487231713753641bc49ee5b9a`;
- runtime `.able` files: 70;
- scorecard source-tree SHA-256:
  `6a412c872ee66752de7c4417a5eda99806a7631da3b880c55574c6a640b82d9b`;
- ledger source-tree SHA-256:
  `382d256e2fb380220dcdd62a5cf83109fa72297f23d70bdd1ffe2d8daebed047`.

The scorecard and ledger hashes differ because they use distinct documented
serialization formats. Their file sets and contents are the same.

## Results

| Gate | Result |
| --- | --- |
| External scoreboard | pass, 130 rows |
| Five-sample evidence | pass, 130 full-status rows, five Able/reference samples per row, 31 retained source and reference reports |
| Relocation tests | pass, four tests |
| Performance frontier | pass, 130 rows, zero actionable groups |
| Frontier tests | pass, five tests |
| Evidence ledger | fail, zero current and 23 invalidated closures |
| Ledger tests | seven pass, one fail |

The selected-row identity is
`d450605f6b271fbddbda7bf31e9f61c1d87cbf1a407f9047304d79bd64ff1684`.
The checked scoreboard JSON has SHA-256
`43c9a48d92ecaf02069655e6c1e78cc81bf1025997e578708da3c23acac8a4d8`;
the checked frontier JSON has SHA-256
`f4fa5b26a9f6229d0cb61a27af0ff8b965ba424732f73998463fd8640e5d312b`.

The checked closure-ledger baseline has SHA-256
`15fcba70f930f002e7ef9e00084c4986973248db817e989266a693cc1c40f9c2`.
The clean export's evaluated invalidation report has SHA-256
`fd232d1f6c0613bcd68184d00b795d5b0049738ed3dd077124eff3df96d36ed4`.

## Exact cause

The canonical-stdlib scope records agree on every content property:

| Property | Checked baseline | Clean export |
| --- | --- | --- |
| files | 70 | 70 |
| include | `*.able`, `**/*.able` | `*.able`, `**/*.able` |
| exclude | none | none |
| tree SHA-256 | `382d256e…` | `382d256e…` |
| root | `/home/david/sync/projects/able-stdlib/src` | `/var/tmp/able-post-publication-cQO7mm/able-stdlib/src` |

The sole reason is
`scope-definition-drift:canonical-stdlib`. Pointing the clean ledger at the
original developer-checkout stdlib with `--scope-override` makes all 23
closures current. That diagnostic confirms content equivalence, but it is not
an acceptable clean-tree proof because it restores the hidden developer path.

The failure is in scope identity, not in the published performance evidence,
compiler, runtime, interpreter, VM, stdlib behavior, benchmark measurement, or
deferred WASM boundary.

## Cleanup

The reconciliation created no Python, Go-build, or project cache in the
export. Its 141,040 KiB task workspace was removed after preserving this
record. No RAM-backed `/tmp` state or task-owned workspace remains.

## Next

Make canonical-stdlib scope identity checkout-independent.

Why: a clean adjacent stdlib with byte-identical content currently invalidates
every closure because the ledger hashes its absolute location.

What it entails: retain the semantic configured root for default scopes,
continue hashing every selected file and its relative path, define explicit
override identity behavior, add relocated-adjacent and real-content-drift
regressions, rebuild the checked ledger, and rerun all gates in both the dirty
worktree and an exact clean export.

Why it matters: closure validity must follow semantic inputs and content, not
the workstation directory in which those inputs happen to be checked out.
