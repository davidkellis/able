# Post-parser publication reconciliation

Date: 2026-07-30

## Decision

Retain no production, dependency-version, compiler, runtime, interpreter,
bytecode-VM, parser-semantic, canonical-stdlib, language, benchmark, fixture,
frozen-workspace, or WASM change.

The exact published commit independently reproduces the committed parser Go
binding contract from a clean exported tree:

- both Go modules retain their committed manifests and empty tidy diffs;
- both module caches verify;
- the standalone binding test passes under Go 1.26.5;
- test-inclusive `govulncheck v1.6.0` reports no vulnerabilities; and
- the raw vulnerability output exactly matches the committed SHA-256.

The developer worktree's 34 deferred WASM paths were absent from the exported
tree and did not influence any result.

## Published identity

- local `HEAD`, `origin/master`, and remote `refs/heads/master`:
  `da8cb15bc394b6528dfd6a6b0eb1de30e12fef51`;
- parent:
  `be9ecc505161085e1ec11f704571f589b3366c13`;
- tree:
  `27f296f7c3ced088157eabff41b1146ff2a502a0`;
- ahead/behind: zero/zero;
- index: empty; and
- exported regular files: 10,953.

The export was produced directly with `git archive` from the published commit
into disk-backed `/var/tmp`. It contains the root parser `go.sum`, omits the
deleted nested `bindings/go/go.mod`, and omits every untracked deferred WASM
file.

## Module integrity

Go 1.26.5 reproduced:

| Module | `go.mod` SHA-256 | `go.sum` SHA-256 | Listed modules | Tidy diff | Verify |
| --- | --- | --- | ---: | --- | --- |
| Main v12 Go runtime | `7a76c1bd85b76d079c9acc1148cc3dc6b6f03694aae7e3c29e650e1891aeb563` | `0eed327adef2c76281d55a5b3dfaa8a0be3df8346423667b0b2aee62f882e83f` | 65 | empty | pass |
| Parser Go binding | `d5c63a254931cd83e2b0cc415137cd076106facb235c0289437e9da2590bdab0` | `a7d2dd524ce7f8e47eb33d9aa480d5011628376107c15e3a415cfef44d22fc44` | 19 | empty | pass |

Main tidy completed in 10.82 seconds with 38,636 KB peak RSS. Parser tidy
completed in 0.63 seconds with 28,148 KB peak RSS. Both empty tidy outputs
have the empty-content SHA-256
`e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855`.
Both verification outputs have SHA-256
`b4537ed75f533f993f371954de47e42a793b8e5b0587577de7e27fb3e50696bd`.

## Parser test and security result

The standalone binding test passed:

- command: `go test -timeout 55s -count=1 ./bindings/go`;
- wall: 15.80 seconds;
- peak RSS: 196,324 KB; and
- output SHA-256:
  `2f2569e1d21054ea3603ff0ffe32c8e6d06a8ff6e82d12ff4b220acc23115f08`.

The scanner reproduced:

- Go: 1.26.5;
- scanner: `govulncheck v1.6.0`;
- database: `https://vuln.go.dev`;
- database update: 2026-07-27 20:14:16 UTC;
- test-inclusive root packages: 3;
- reachable modules: 3 plus the standard library;
- vulnerabilities: zero;
- wall: 1.21 seconds; and
- peak RSS: 435,752 KB.

The reachable modules are:

- `github.com/davidkellis/able`;
- `github.com/mattn/go-pointer v0.0.1`; and
- `github.com/tree-sitter/go-tree-sitter v0.24.0`.

Raw scanner SHA-256:
`ac2fbe8e62be0ddae77b8fc19198ec6daf9d00145e24d2fb03f32fd040a92811`.
This exactly matches
`2026-07-30-parser-go-binding-contract-retained.{md,json}`.

The main-module production and test-inclusive scans were not repeated: their
manifests and published production tree are unchanged, and the committed
post-publication security record already covers them. This tranche remained
bounded to the newly corrected standalone parser contract.

## Performance and scope

No current performance closure is invalidated. The clean export is the exact
published tree already covered by the complete v12 runner and current
scorecard/frontier/ledger evidence. Repeating the unchanged full performance
suite would add no new causal information.

GitHub's publication banner remains 74 repository-wide Dependabot alerts: 18
critical, 14 high, 34 moderate, and 8 low. The banner is not an active-v12
reachability result and does not contradict the clean published parser scan.

## Cleanup

All export, Go module, toolchain, scanner, build-cache, and raw-output state
lived under disk-backed `/var/tmp`. The exact 1,223,340 KiB (1,194.67 MiB)
workspace was removed after preserving this readable and machine-readable
record. The final guarded project cleanup dry run is empty, no task workspace
remains, and no build state was placed in RAM-backed `/tmp`.

## Next

Refresh the exact retained/deferred release inventory without staging,
committing, or pushing.

Why: after publication reconciliation, the worktree combines five retained
handoff/reconciliation paths with the unchanged 34 deferred WASM paths.

What it entails: expand all untracked files, record exact state/line/byte/hash
identities, validate the two reconciliation records plus PLAN/logs, reproduce
the 34-path deferred complement, and define the next retained candidate
without repository mutation.

Why it matters: a precise boundary closes the publication work cleanly and
prevents deferred WASM from entering any later handoff consolidation.
