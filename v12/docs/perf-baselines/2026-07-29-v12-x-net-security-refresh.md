# V12 x/net security refresh

Date: 2026-07-29

## Decision

Retain the bounded active-v12 Go dependency refresh:

- `golang.org/x/net`: `v0.54.0` to `v0.56.0`;
- resolver-required `golang.org/x/crypto`: `v0.52.0` to `v0.53.0`; and
- resolver-required `golang.org/x/sys`: `v0.45.0` to `v0.46.0`.

The update removes all seven fixable module advisories identified by the
dependency-alert attribution. Production and test scans retain zero reachable
symbol and zero vulnerable imported-package findings.

No compiler, runtime, interpreter, VM, language, stdlib, benchmark, fixture,
frozen v10/v11, external repository, or deferred WASM file changed.

## Exact dependency delta

`go get golang.org/x/net@v0.56.0` produced exactly the three version changes
proven by the preceding temporary probe. `go mod tidy` then normalized only
the resolver-coupled sum entries:

| Module | Before | After | Manifest effect |
| --- | --- | --- | --- |
| `golang.org/x/net` | `v0.54.0` | `v0.56.0` | `go.mod`, `go.sum` |
| `golang.org/x/crypto` | `v0.52.0` | `v0.53.0` | `go.mod`, `go.sum` |
| `golang.org/x/sys` | `v0.45.0` | `v0.46.0` | `go.mod`, `go.sum` |
| `golang.org/x/term` | `v0.43.0` | `v0.44.0` | `go.sum` |
| `golang.org/x/text` | `v0.37.0` | `v0.38.0` | `go.sum` |

The final delta contains:

- three `go.mod` line replacements;
- ten old `go.sum` lines removed;
- ten new `go.sum` lines added; and
- zero new module families.

Final identities:

- `go.mod`:
  `7a76c1bd85b76d079c9acc1148cc3dc6b6f03694aae7e3c29e650e1891aeb563`;
- `go.sum`:
  `0eed327adef2c76281d55a5b3dfaa8a0be3df8346423667b0b2aee62f882e83f`.

`go mod verify` passes and a subsequent `go mod tidy -diff` is empty.

## Vulnerability result

`govulncheck v1.6.0` used the Go vulnerability database last modified
`2026-07-27T20:14:16Z` and explicit Go 1.26.5.

| Scan | Symbol | Imported package | Required module |
| --- | ---: | ---: | ---: |
| production | 0 | 0 | 1 |
| tests | 0 | 0 | 1 |

The following module advisories are gone:

- `GO-2026-5025`;
- `GO-2026-5026`;
- `GO-2026-5027`;
- `GO-2026-5028`;
- `GO-2026-5029`;
- `GO-2026-5030`; and
- `GO-2026-5942`.

The sole remaining module advisory is `GO-2026-5932`, which declares
`golang.org/x/crypto/openpgp` unsafe by design and provides no fixed version.
Able does not import that package; its production and test scans contain no
package or symbol finding for the advisory.

## Correctness and performance-evidence gates

- Every Go package compiles under Go 1.26.5 with
  `go test ./... -run '^$' -count=1 -timeout 60s`. The command took 38.52
  seconds and peaked at 1,583,472 KB RSS.
- Focused Git dependency-installer, URL-normalization, override, and deps CLI
  tests pass in `cmd/able`. Go test time is 0.158 seconds; the complete command
  took 1.50 seconds and peaked at 322,424 KB RSS.
- The authoritative scorecard check passes for all 126 rows with five
  Able/reference samples each.
- The performance frontier remains at zero actionable groups.
- All 23 closure-ledger entries remain current with zero invalidations.

No performance evidence was refreshed because the dependency-only update
changes neither generated source nor benchmark/runtime execution paths.

## Workspace integrity

- `HEAD` and `origin/master` remain
  `6efad0a53120129510fdfbab7fbcc84dcd081768`.
- The index remains empty.
- Every pre-existing deferred WASM path is preserved.
- Go toolchains, module caches, build caches, vulnerability output, and test
  output live only in the exact disk-backed `/var/tmp` refresh workspace and
  are removed at handoff.

## Next

Refresh the exact retained/deferred release inventory without staging,
committing, or pushing.

Why: the retained dependency correction and its attribution records now sit
beside the unchanged deferred WASM boundary, so the previous release manifest
no longer describes every dirty path.

What it entails: capture a fully expanded worktree snapshot, classify the nine
post-publication security and handoff paths against the 34 deferred WASM
files, validate identities, JSON, whitespace, scope, secrets, and the exact
path complement, and require separate authorization for staging or a commit.

Why it matters: an exact refreshed boundary preserves the security correction
without admitting deferred WASM or unrelated work and restores a safe basis
for later consolidation.
