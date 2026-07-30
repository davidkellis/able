# Canonical stdlib scope publication reconciliation

Date: 2026-07-30

## Decision

Exact published commit
`e3a4e1e8000aeb09d80a3a1adc48548b9deeeb59` reproduces every checked
performance-evidence gate from a clean source-only export with the canonical
stdlib adjacent to it.

The ledger uses its default semantic root `../able-stdlib/src`; no
`--scope-override`, developer-worktree stdlib path, deferred WASM file, or
uncommitted Able path participates in the proof.

Retain this reconciliation record and no code. No compiler, runtime,
interpreter, VM, parser, stdlib, language, dependency, benchmark measurement,
fixture, frozen workspace, or WASM behavior changed.

## Exact clean export

The disk-backed export identity is:

- commit: `e3a4e1e8000aeb09d80a3a1adc48548b9deeeb59`;
- tree: `877da26eea55749e82ccf8b3b400009ac211ce69`;
- regular Able files: 10,971;
- deferred bytecode WASM proof files: zero; and
- filesystem: `/var/tmp` on `btrfs`.

The adjacent canonical stdlib reproduced:

- recorded Git identity:
  `219eff222c28406487231713753641bc49ee5b9a`;
- runtime `.able` files: 70;
- scorecard source-tree SHA-256:
  `6a412c872ee66752de7c4417a5eda99806a7631da3b880c55574c6a640b82d9b`;
- ledger file-set tree SHA-256:
  `382d256e2fb380220dcdd62a5cf83109fa72297f23d70bdd1ffe2d8daebed047`;
  and
- semantic ledger root: `../able-stdlib/src`.

The source-only stdlib export has no `.git` directory, so its diagnostic
`git_head` is null and `git_dirty` is false. Its selected runtime file set and
content identity exactly match the canonical evidence state.

## Gate results

| Gate | Result |
| --- | --- |
| External scoreboard | pass, 130 rows |
| Five-sample evidence | pass, 130 full-status rows and five Able/reference samples per row |
| Retained dependencies | 31 source reports and 31 reference reports |
| Path relocation tests | pass, four tests |
| Performance frontier | pass, 130 rows and zero actionable groups |
| Frontier tests | pass, five tests |
| Evidence ledger | pass, 23 current closures and zero invalidations |
| Closure selector | empty |
| Ledger tests | nine pass and one conditional skip |

The selected-row identity is
`d450605f6b271fbddbda7bf31e9f61c1d87cbf1a407f9047304d79bd64ff1684`.

## Published evidence identities

- ledger tool:
  `d6e64acc4217b445fc43d49aa8a4c5e572fb97fe7e0046758f79bc79109eedc7`;
- ledger tests:
  `3adec4b6b378178dfc40b4dac1a4614023c487d805ce5428d2c420c9173d1ad0`;
- closure baseline:
  `aaf867b3b78add52fbf5b36efae45a4c097b9ae9e142f7bf826c754c942d3171`;
- evaluated ledger JSON:
  `2d0f331971c938a7c086d8be95aa255bc9c34dfb6d7943686ce7d7debdc423b2`;
- evaluated ledger Markdown:
  `dcf133cbdd3fe987b9c541deca3eb3a6a84375c418145c9ba9021542f332a8ea`;
- current scoreboard JSON:
  `43c9a48d92ecaf02069655e6c1e78cc81bf1025997e578708da3c23acac8a4d8`;
- current frontier JSON:
  `f4fa5b26a9f6229d0cb61a27af0ff8b965ba424732f73998463fd8640e5d312b`;
  and
- selection manifest:
  `6bbe6579df9857a791a2f30d55792bba0827994766d4f27beebbf0dba24ec628`.

The clean proof closes the absolute canonical-stdlib-root failure recorded in
`2026-07-30-post-relocatable-evidence-publication-reconciliation.md`.
Closure validity now follows configured scope semantics, relative file paths,
and content rather than checkout location.

## Cleanup

The exact 141,004 KiB export, stdlib copy, and test workspace lived under
disk-backed `/var/tmp`. It contained no Python, Go-build, or project cache and
was removed after preserving this record. No task state used RAM-backed
`/tmp`, and no task-owned workspace remains.

## Next

Refresh the exact retained/deferred release inventory without staging,
committing, or pushing.

Why: the worktree now combines three publication handoff files and this
reconciliation record with the unchanged 34-path deferred WASM boundary.

What it entails: capture a fully expanded immutable snapshot, record exact
state, line, byte, and SHA-256 identities, reproduce the deferred complement,
run format/evidence/Git/cleanup gates, and define the retained candidate
without repository mutation.

Why it matters: the shared-history reconciliation can later be consolidated
without absorbing deferred WASM or unrelated dirty work.
