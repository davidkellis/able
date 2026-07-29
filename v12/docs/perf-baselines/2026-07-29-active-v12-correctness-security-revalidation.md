# Active v12 correctness and security revalidation

Date: 2026-07-29

## Decision

Retain no production, dependency, benchmark, stdlib, language, or WASM
change.

The published active v12 module remains correct under Go 1.26.5 and reproduces
the accepted dependency-security boundary. `govulncheck` v1.6.0 reports zero
symbol-reachable vulnerabilities and zero vulnerabilities in imported
packages. The same eight module-only advisories documented by the July 28
security refresh remain unreachable by Able.

The complete bounded v12 handoff passes. A focused JSON timing audit of the
three compiler shards whose aggregate package times exceeded one minute found
zero individual tests over one minute. No correctness or dependency change
invalidates the checked performance frontier, so production performance
mutation remains paused.

## Frozen state

- Repository and `origin/master`:
  `b7702339ad41cceacefe41bcdbeefb879538fddf`.
- Go runtime/toolchain: `go version go1.26.5 linux/amd64`.
- Language compatibility floor: Go 1.25.
- `go.mod` SHA-256:
  `0dceec93c608f733816f5b926c00a73b05047abc446f7fca91bb664520ebaff3`.
- `go.sum` SHA-256:
  `10840efb4d3a0cabbef2e43d491d30bae0973faff3735dd85fb681fc60ff94b7`.
- Worktree before and after verification: the exact pre-existing 34-file
  deferred WASM hold, with an empty index.

No module metadata changed.

## Vulnerability result

The pinned command was:

```sh
GOTOOLCHAIN=go1.26.5 \
  go run golang.org/x/vuln/cmd/govulncheck@v1.6.0 ./...
```

The analyzer used the Go vulnerability database updated at
2026-07-27 20:14:16 UTC and scanned 29 root packages, 24 modules, and the
Go 1.26.5 standard library.

Result:

- symbol-reachable vulnerabilities: 0;
- vulnerabilities in imported packages: 0; and
- unreachable module-only advisories: 8.

Seven module-only advisories belong to `golang.org/x/net v0.54.0` packages
that Able does not import: `dns/dnsmessage`, `html`, and `idna`. The remaining
advisory states that `golang.org/x/crypto/openpgp` is unmaintained and has no
fixed version; Able does not import that package.

This exactly reproduces the accepted result in
`2026-07-28-active-go-dependency-security-refresh.md`. A dependency update
would therefore add compatibility and performance churn without reducing
active reachable risk. GitHub's repository-wide alert count also includes
frozen and non-active scopes and is not an active-v12 reachability result.

## Correctness and build gates

All commands used Go 1.26.5:

- `go mod verify` passes;
- `go vet ./...` passes;
- `go build ./...` passes; and
- `./run_all_tests.sh` passes completely.

The handoff cleared:

- exec coverage;
- external scorecard and five-sample evidence;
- feature coverage and selection contracts;
- threshold controls;
- generated-artifact cleanup policy;
- embedded-kernel synchronization;
- all non-compiler Go packages;
- all 34 bounded compiler short-mode shards; and
- the final bytecode fixture pass.

The ordinary interpreter package took 98.283 seconds and the final bytecode
fixture package took 87.219 seconds. These are aggregate package durations,
not individual-test durations.

## Individual compiler timing audit

Compiler shards 20, 29, and 30 took 103.173, 73.348, and 102.994 seconds
respectively. They contain 75 top-level canonical interface/stdlib,
strict-dispatch/String, and struct-pattern tests.

A single JSON event run over exactly those 75 names completed successfully:
66 tests passed, nine skipped under their normal short-mode guards, and zero
failed. No top-level test exceeded 60 seconds.

The slowest individual tests were:

| Test | Elapsed |
| --- | ---: |
| `TestCompilerCanonicalStdlibExpectationResultArgumentStaysConcrete` | 24.49s |
| `TestCompilerSpecializedImplCanonicalKeyPreventsDuplicateContainAllBodies` | 15.62s |
| `TestCompilerCanonicalSpecNullableExpectationLowers` | 15.56s |
| `TestCompilerContainAllStringMatcherUsesDirectStringSpecialization` | 15.55s |
| `TestCompilerCanonicalStdlibSpecImportKeepsIterableIteratorStatic` | 15.40s |
| `TestCompilerIterableDefaultMethodSelfSiblingStaysStaticForStdlibImports` | 15.20s |

The combined package elapsed time was 220.346 seconds. It is cumulative work,
not a slow-test regression, so no runner split or test weakening is justified.

## Cleanup

The audit removed:

- eight generated `v12/__pycache__/*.pyc` files;
- the exact temporary compiler JSON timing artifact; and
- an idle, unheld 37 MiB `/tmp/able-v12-extern-go` cache dated July 28.

No user process held the extern cache. No `/tmp/able-*` entry remains.
Reusable Go build/module caches remain under disk-backed storage.

## Next recommendation

Keep the evidence-triggered production pause.

Why: active correctness and security are green, the dependency result is the
already accepted unreachable-module boundary, and no performance closure is
invalidated.

What it entails: resume performance work only after a retained compiler,
runtime, language, stdlib, dependency, benchmark-source, or broad-application
change alters execution ownership. Then refresh only the invalidated profiles
and admit a candidate only if one exact material owner repeats in at least
three unlike applications and passes balanced verifier-backed A/B/reference
measurements.

Why it is important: this preserves native Go carriers and interpreter-free
strict graphs without using unrelated repository alerts or aggregate test
duration to manufacture a performance change. Do not begin WASM work.
