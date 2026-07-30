# Post-publication security reconciliation

Date: 2026-07-30

## Decision

Retain no production, dependency, compiler, runtime, interpreter, VM,
canonical-stdlib, language, benchmark, fixture, frozen-workspace, or WASM
change.

The published active-v12 module reproduces the retained dependency-security
boundary under Go 1.26.5 and `govulncheck v1.6.0`:

- zero symbol-reachable vulnerabilities;
- zero vulnerabilities in imported packages; and
- one required-module advisory for the unimported
  `golang.org/x/crypto/openpgp` package, which has no fixed version.

The test-inclusive main-module scan has the same result. The active parser
binding production package and its two declared parser dependency versions
have zero findings. All four canonical npm lockfiles report zero
vulnerabilities.

GitHub's publication banner now reports 74 repository-wide alerts: 18
critical, 14 high, 34 moderate, and 8 low. The previous attributed banner
reported 75, with the same severities except 35 moderate alerts. A
repository-wide count is not an active-v12 reachability result; exact
alert-to-manifest rows remain unavailable without authenticated
Dependabot-alert access.

## Published identity

- local `HEAD`, `origin/master`, and remote `refs/heads/master`:
  `be9ecc505161085e1ec11f704571f589b3366c13`;
- tree:
  `59a210723d6abe9677989b4d8440cf8b0b05e025`;
- ahead/behind: zero/zero;
- index: empty; and
- pre-record worktree: exactly the unchanged 34 deferred WASM paths.

The published commit contains exactly the verified 293-path retained
candidate. This reconciliation performed no commit or push.

## Active-v12 Go result

All scanner and build state lived under a disk-backed `/var/tmp` workspace.
The scanner reported:

- Go: 1.26.5;
- scanner: `govulncheck v1.6.0`;
- database: `https://vuln.go.dev`;
- database update: 2026-07-27 20:14:16 UTC; and
- main graph: 24 modules plus the Go 1.26.5 standard library.

| Scan | Root packages | Symbol | Imported package | Required module | Wall | Peak RSS |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Production | 29 | 0 | 0 | 1 | 14.09s | 1,580,644 KB |
| Tests included | 71 | 0 | 0 | 1 | 3.16s | 2,123,292 KB |

The sole module result is `GO-2026-5932` for
`golang.org/x/crypto/openpgp`. Able requires
`golang.org/x/crypto v0.53.0` transitively but does not import `openpgp`.
The advisory identifies no fixed version.

Raw text identities:

- production:
  `1e7de51a0d634468b73b6cccf59732c6b2f903abb1a259d73ac0db10fe0b8b57`;
- tests:
  `1a2b32f0e3d4cd9d1399bdcea58217ce72cf1685ae1e3b95243f1f3cc3473836`.

## Module integrity

The retained security-refresh versions remain:

- `golang.org/x/net v0.56.0`;
- `golang.org/x/crypto v0.53.0`; and
- `golang.org/x/sys v0.46.0`.

Identities remain:

- `go.mod`:
  `7a76c1bd85b76d079c9acc1148cc3dc6b6f03694aae7e3c29e650e1891aeb563`;
- `go.sum`:
  `0eed327adef2c76281d55a5b3dfaa8a0be3df8346423667b0b2aee62f882e83f`.

`go mod verify` passes and `go mod tidy -diff` is empty.

## Parser boundary

The standalone parser binding production package scans as one root package
and one module with zero findings. Its raw output SHA-256 is
`1fca5095e41c09a89b5fc1ce443fe1e5bc6c752c43806058f8d05a164f0e9ba6`.

Direct OSV queries also return zero advisories for:

- `github.com/tree-sitter/go-tree-sitter v0.24.0`, required by the parser
  root module; and
- `github.com/smacker/go-tree-sitter`
  `v0.0.0-20230720070738-0d0a9f78d8f8`, declared by the nested binding
  module.

The nested `bindings/go` test scan does not load. That generated scaffold:

- has no `go.sum`;
- declares module `github.com/tree-sitter/tree-sitter-able`;
- requires `github.com/smacker/go-tree-sitter`; but
- its test imports `github.com/tree-sitter/go-tree-sitter` and
  `github.com/davidkellis/able/bindings/go`.

No parser file was changed during this audit. The active parser used by the
Able interpreter is part of the clean 71-root main-module test scan. The
standalone scaffold mismatch is a correctness/tooling gap, not evidence of a
reachable vulnerability.

## npm result

`npm audit` 11.16.0 used an isolated disk-backed cache and changed no
lockfile:

| Lockfile | Dependencies | Vulnerabilities |
| --- | ---: | ---: |
| v10 parser | 28 | 0 |
| v11 parser | 28 | 0 |
| v12 parser | 28 | 0 |
| v12 deferred WASM | 1 | 0 |

The frozen v10/v11 Go findings in the prior attribution remain historical Go
dependency risk. Their clean npm lockfiles do not alter that disposition.

## Evidence and cleanup

- The authoritative 130-row scorecard remains valid with five successful
  Able/reference processes per row.
- The frontier remains at zero actionable groups.
- All 23 evidence-ledger closures remain current with zero invalidations and
  an empty selector.
- Three ignored, untracked benchmark `target` trees last written in April had
  zero tracked descendants and no active build process. Their exact removal
  reclaimed 452,344 KiB (441.74 MiB).
- The isolated scanner/module/npm workspace was removed after preserving
  this readable evidence.
- The final project cleanup dry run found no generated artifact.

No current performance closure was invalidated because no production,
dependency, benchmark, language, or semantic input changed.

## Next

Resolve the standalone parser Go binding test contract before resuming
unchanged performance work.

Why: security is green and the performance frontier has no actionable group,
but the audit exposed one concrete v12 tooling failure: the nested parser
binding's own tests cannot load reproducibly.

What it entails: determine whether `bindings/go` is an active supported module
or an obsolete generated scaffold. If active, align its module path and
tree-sitter dependency, add the required sum identity, and make both `go test`
and `govulncheck -test` pass in under one minute. If obsolete, remove it from
the active contract or mark it archival without changing parser semantics.
Then rerun the official active parser harness. Do not begin AST mapping or
WASM work.

Why it matters: every claimed active test and security surface should be
independently reproducible; resolving the ambiguity prevents a broken
standalone scaffold from silently overstating parser coverage.

## References

- Go vulnerability checking:
  <https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck>
- GitHub Dependabot alert API:
  <https://docs.github.com/en/rest/dependabot/alerts>
