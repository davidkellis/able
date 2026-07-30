# Post-parser-publication shared-tree reconciliation

Date: 2026-07-30

## Decision

The published reconciliation commit is exact, but the clean source archive
exposes two general performance-evidence reproducibility gaps:

1. all 31 selected comparison reports retain absolute paths into the original
   Able workspace, so the scoreboard and five-sample validators cannot
   relocate the archive; and
2. the checked native bytecode scope includes three untracked deferred
   `*_wasm.go` proof files, so its committed ledger baseline differs from the
   published source tree.

Retain no production, dependency, benchmark-measurement, stdlib, language, or
WASM change in this reconciliation. The publication commit itself changes no
module or performance-evidence input and therefore does not causally
invalidate a measured result. The next tranche must correct the two evidence
gates before another performance optimization is admitted.

## Published identity

An exact Git archive was created under disk-backed `/var/tmp`, independently
of the developer worktree:

- commit: `e7a054032e189e9a606dad68c827be50af897417`;
- parent: `da8cb15bc394b6528dfd6a6b0eb1de30e12fef51`;
- tree: `6c62157b1ff3bb364995f2d11fd208a3f73d4cae`;
- local `HEAD`, `origin/master`, and remote `master`: exact commit identity;
- ahead/behind divergence: zero/zero;
- exported regular files: 10,958;
- exported symlinks: zero; and
- staged paths: zero.

The published commit changes exactly eight paths. Every path is one of the
authorized handoff/reconciliation records; no compiler, parser, runtime,
interpreter, VM, dependency, benchmark, fixture, stdlib, v10/v11, or deferred
WASM path entered the commit.

## Reconciliation and inventory identities

The committed eight-path boundary reproduces:

- 425-byte sorted path list with SHA-256
  `d65ec0bdc385996b9770c78d45e587da51ef00ade48377b2abaf9a713123a284`;
- zero difference between the recorded candidate and committed path set;
- five non-self rows, 59,234 lines, and 4,089,356 bytes;
- 561-byte non-self identity manifest with SHA-256
  `39a4c35428d9a243351375eeeab5cf20f9b59becfd0d8b8d2fb8e8805baf27bf`;
  and
- 39-row immutable TSV SHA-256
  `dc056e0a2ef32e804fb8d6ba8ed8fb0ceea33c612bef0583a483500d063919ad`.

The committed record hashes are:

| Record | SHA-256 |
| --- | --- |
| publication reconciliation JSON | `bf45aacd18dd5803c7ce49212376284c43b566a6dd4b6c0e90af38942576acd7` |
| publication reconciliation Markdown | `f4a9229b7fab30a0b344d04f35175cf9d47af53eaff0484e4b716d881d136e09` |
| release inventory JSON | `7f4ef8e157717657a9e06031abae1ae1cb9fd162071f893be2b5e79ad1023626` |
| release inventory Markdown | `ddb3acb4e7858efbb2148a8b28eaafce98a38757de160b9cf67058e5c7558e04` |
| release inventory TSV | `dc056e0a2ef32e804fb8d6ba8ed8fb0ceea33c612bef0583a483500d063919ad` |

## Unchanged module inputs

The publication commit has an empty diff across both module manifests and all
selected performance-evidence inputs. Clean-export module identities exactly
match the preceding tested reconciliation:

| Input | SHA-256 |
| --- | --- |
| main `go.mod` | `7a76c1bd85b76d079c9acc1148cc3dc6b6f03694aae7e3c29e650e1891aeb563` |
| main `go.sum` | `0eed327adef2c76281d55a5b3dfaa8a0be3df8346423667b0b2aee62f882e83f` |
| parser `go.mod` | `d5c63a254931cd83e2b0cc415137cd076106facb235c0289437e9da2590bdab0` |
| parser `go.sum` | `a7d2dd524ce7f8e47eb33d9aa480d5011628376107c15e3a415cfef44d22fc44` |

Module tidy/verify, parser tests, vulnerability scans, and the complete suite
were not repeated. The exact tested files and all parser/compiler/runtime
inputs are unchanged; repeating those long checks would add no causal
coverage.

## Performance evidence

The clean archive reproduces these committed inputs:

| Input | SHA-256 |
| --- | --- |
| external scoreboard | `43c9a48d92ecaf02069655e6c1e78cc81bf1025997e578708da3c23acac8a4d8` |
| selection manifest | `6bbe6579df9857a791a2f30d55792bba0827994766d4f27beebbf0dba24ec628` |
| frontier evidence | `6edd75cc356a4bdeb856d46d028e70cdb1d1774e8bad57b62f6b7c21fa7883c2` |
| frontier JSON | `f4fa5b26a9f6229d0cb61a27af0ff8b965ba424732f73998463fd8640e5d312b` |
| closure ledger | `67e0934be56885bf491ffbc2b53e2e850ff48290ce163b6dd178720e24327a7a` |
| ledger report | `2d0f331971c938a7c086d8be95aa255bc9c34dfb6d7943686ce7d7debdc423b2` |

The frontier independently passes with 130 rows and zero actionable groups.
In the developer worktree, the scoreboard, five-sample evidence checker,
frontier, and ledger all pass; the ledger reports 23 current closures and zero
invalidations.

The clean archive does not pass the other three gates, for two distinct
tooling reasons.

### Absolute selected-report paths

All 31 selected source reports contain absolute paths rooted at the original
Able workspace. Across them are 161 such occurrences:

- 130 fields named `path`;
- 16 `go_reference_json` fields; and
- 15 `reference_json` fields.

The scoreboard first fails at `fib/compiled` because its measured source path
does not equal the relocated canonical target. The five-sample checker first
fails because an absolute `go_reference_json` path is not under the relocated
`perf-baselines` directory. The committed hashes and sample counts are
unchanged; the validators are location-dependent.

### Native bytecode scope contamination

The checked `bytecode-production` baseline contains 206 files with tree
SHA-256
`b8472c5b17856a9eed8eecb4dcecd2e3030d6024e7f50ed050f5ae06902de197`.
The clean archive contains 203 matching committed files with SHA-256
`4c6b98eafd856f6bda13884e8b32004bf5af2842b4fcf6d2990e8c9321cd39a1`.

The three working-tree-only inputs are deferred WASM proof files:

- `bytecode_i32_frame_proof_wasm.go`;
- `bytecode_scalar_proof_wasm.go`; and
- `bytecode_type_proofs_wasm.go`.

All are untracked members of the authoritative deferred WASM boundary. Their
accidental inclusion in the native bytecode scope makes the clean export
select 12 closures for `scope-content-drift:bytecode-production`, leaving 11
current. This is evidence-scope contamination, not a native bytecode
production change or measured performance invalidation.

## Cleanup

The exact 139,996 KiB (136.71 MiB) archive and diagnostic workspace lived
under disk-backed `/var/tmp`. It was removed after preserving these two
records. No build state was placed in RAM-backed `/tmp`; the guarded project
cleanup dry run is empty, and no task workspace remains.

## Next

Repair the relocatable performance-evidence gates before admitting another
optimization tranche.

Why: the published source tree is exact, but its evidence gates are not
self-contained: absolute report paths prevent relocation and a native
bytecode scope includes untracked WASM-only files.

What it entails: normalize repository-owned report references against the
active repository, add relocated-export regressions, exclude `*_wasm.go` from
native bytecode production scope, rebuild the checked ledger from committed
inputs, and prove both the developer worktree and a clean relocated export
select zero closures.

Why it matters: performance evidence must reproduce from committed source
alone and must not depend on deferred WASM state before another optimization
tranche can be trusted.
