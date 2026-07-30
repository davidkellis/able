# Post-correctness publication reconciliation

Date: 2026-07-30

## Decision

Exact commit
`f3f3c92d512d313648563503904528e307b7c11e` is published to
`origin/master`.

The commit contains exactly the authorized 21-path non-WASM candidate. Its
developer-worktree residual remains exactly the unchanged 34-path deferred
WASM boundary. A clean archive of the published commit independently
reproduces the fixture-policy and performance-evidence gates.

Retain this reconciliation record and no production change. No compiler,
runtime, interpreter, VM, parser semantic, canonical-stdlib, language,
dependency, benchmark measurement, fixture behavior, frozen workspace, or
WASM behavior changed during publication.

## Published identity

- local `HEAD`, `origin/master`, and remote `refs/heads/master`:
  `f3f3c92d512d313648563503904528e307b7c11e`;
- parent:
  `9c32f2777536da2c948327720acc75187973a6d9`;
- tree:
  `336258c8ac459e7412f3c1848da7a5272e0b1f7b`;
- ahead/behind: zero/zero;
- index: empty; and
- explicit published refspec:
  `f3f3c92d512d313648563503904528e307b7c11e:refs/heads/master`.

The commit has exactly 21 changed paths, 2,681 additions, and 228 deletions.
Its 1,364-byte sorted newline-delimited path list has SHA-256
`b2325ad24655d9cdafcdfa1b3aeef66d0dbbd6ce2b7c6566e42aa6518084954b`,
matching the authorized release inventory.

## Deferred-boundary proof

The developer worktree retains exactly 34 deferred WASM paths. Their sorted
1,414-byte path list has SHA-256
`7f29de27f03da83138df19696be182a5139aa6536ba0bdb363a1a23ffafd1f4e`.
Every deferred state, line, byte, and content identity matches the
authoritative inventory.

The clean published archive proves the distinction between tracked base
files and deferred modifications:

- all 26 deferred untracked paths are absent;
- the eight tracked paths contain their committed clean versions;
- none of the eight committed versions matches its deferred dirty SHA-256;
  and
- no deferred dirty identity participates in the published gates.

## Clean-export identity

The source-only Git archive lived under disk-backed `/var/tmp` and contains:

- 10,991 regular files;
- 114,792,673 regular-file bytes; and
- tree-content SHA-256:
  `0ed1f35a55fe277c39b39738a31f5736c55851bdee0af0db7a2c6e97e88ce2b9`.

The adjacent source-only canonical stdlib copy contains 70 runtime `.able`
files with source-tree SHA-256
`6a412c872ee66752de7c4417a5eda99806a7631da3b880c55574c6a640b82d9b`.
Because it is a source-only copy, its diagnostic Git head is null and its
diagnostic dirty flag is false.

## Gate results

The following gates passed both immediately before publication and from the
clean published archive:

- the live fixture-target inventory: 176 manifests, zero active exclusions,
  one retired exclusion, and zero allowlist entries;
- all seven fixture-target policy tests;
- runner shell syntax;
- the external scoreboard: 130 rows;
- repeated evidence: 130 full-status rows with five successful
  Able/reference samples each and 31 retained source/reference reports;
- all four path-relocation tests;
- the performance frontier: 130 rows and zero actionable groups;
- all five frontier tests;
- the evidence ledger: 23 current closures and zero invalidations;
- the closure selector: empty; and
- all ten ledger tests, with one conditional skip.

The exact runner source in the published commit retains its earlier complete
suite pass: every preflight and non-compiler package, all 34 compiler
batches, and the complete bytecode fixture corpus. Repeating that unchanged
multi-hour suite would add no new causal evidence to the publication proof.

GitHub displayed its existing repository-wide 74-alert Dependabot banner
during the push. That banner is not an active-v12 reachability result and
does not change the previously retained security reconciliation.

## Cleanup

The exact 140,860 KiB (137.56 MiB) publication workspace held the clean
archive, canonical stdlib source copy, and raw verification output under
disk-backed `/var/tmp`. It was removed after preserving this readable and
machine-readable record. Python bytecode generation was disabled, the final
guarded project cleanup dry run is empty, and no task-owned workspace
remains.

## Next

Refresh the exact retained/deferred release inventory without staging,
committing, or pushing.

Why: after publication reconciliation, the worktree combines the two new
reconciliation records and three handoff files with the unchanged 34-path
deferred WASM boundary.

What it entails: capture a fully expanded immutable snapshot, record exact
state, line, byte, and SHA-256 identities, reproduce the 34-path deferred
complement, run format/evidence/Git/cleanup gates, and define the retained
candidate without repository mutation.

Why it matters: an exact five-path retained boundary closes publication
cleanly and prevents deferred WASM from entering a later consolidation.
