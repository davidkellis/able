# Dependency-alert attribution

Date: 2026-07-29

## Decision

The GitHub push banner does not identify a reachable vulnerability in active
v12. Go 1.26.5 symbol-level production and test scans report zero reachable
vulnerabilities and zero vulnerable imported packages. Active parser bindings
and every npm lockfile also report zero vulnerabilities.

No production, dependency, runtime, compiler, interpreter, VM, stdlib,
language, benchmark, fixture, frozen v10/v11, deferred WASM, or external
repository file changed during the audit.

## GitHub banner and access boundary

Publishing `6efad0a53120129510fdfbab7fbcc84dcd081768` produced GitHub's
repository-wide banner:

- 75 total alerts;
- 18 critical;
- 14 high;
- 35 moderate; and
- 8 low.

The Dependabot alert API returned HTTP 401 because no authenticated GitHub
API client or token is available in this environment. GitHub documents that
listing repository Dependabot alerts requires a token with Dependabot-alert
read permission. The banner totals are therefore retained as reported, but
exact GitHub alert-to-manifest rows cannot be exported here.

This does not block active-v12 reachability attribution: current ecosystem
scanners evaluated every active dependency manifest and executable Go
package.

## Active v12 Go result

The official `govulncheck v1.6.0` scanner used the Go vulnerability database
last modified `2026-07-27T20:14:16Z`.

| Surface | Toolchain | Symbol | Imported package | Required module |
| --- | --- | ---: | ---: | ---: |
| v12 interpreter/compiler production | Go 1.26.5 | 0 | 0 | 8 |
| v12 interpreter/compiler with tests | Go 1.26.5 | 0 | 0 | 8 |
| v12 parser Go bindings | Go 1.26.5 | 0 | 0 | 0 |

The eight module-only advisories are:

- `GO-2026-5025` through `GO-2026-5030` in
  `golang.org/x/net@v0.54.0`;
- `GO-2026-5942` in `golang.org/x/net@v0.54.0`; and
- `GO-2026-5932` in `golang.org/x/crypto@v0.52.0`.

Able reaches `x/net` through `go-git` and the unaffected
`golang.org/x/net/context` compatibility package. It reaches `x/crypto`
through `golang.org/x/crypto/hkdf`. It does not import the affected
`x/net/html`, `x/net/idna`, `x/net/dns/dnsmessage`, or
`x/crypto/openpgp` packages.

The v12 parser root module contains no Go packages. Its sole
`github.com/tree-sitter/go-tree-sitter@v0.24.0` requirement has zero current
OSV advisories; the separately executable parser Go binding scan is also
clean.

## Toolchain control

An explicit Go 1.26.4 control found reachable `GO-2026-5856`, an Encrypted
Client Hello privacy leak in `crypto/tls`, plus imported-package
`GO-2026-4970` in `os`. Both are fixed in Go 1.26.5.

The required v12 module declares `toolchain go1.26.5`, `go env GOVERSION`
resolves to Go 1.26.5 in that module, and the explicit Go 1.26.5 production
and test scans contain neither advisory. The toolchain pin is therefore a
security boundary, not just benchmark provenance.

## Frozen workspace attribution

The v10 and v11 Go interpreter modules have identical historical dependency
graphs. Each scan reports:

- 25 reachable symbol advisories;
- 34 vulnerable imported-package advisories; and
- 50 required-module advisories.

Their reachable findings span old versions of `go-git`, `go-billy`,
`cloudflare/circl`, `x/crypto`, and `x/net`. These results help explain the
repository-wide GitHub banner but are not active-v12 findings. The frozen
workspaces were not edited.

Users should not ship or expose the archival v10/v11 executables as maintained
software. Their findings are retained historical risk unless a maintainer
explicitly authorizes a critical frozen-version correction.

## npm attribution

`npm audit` 11.16.0 used isolated disk-backed caches and changed no lockfile:

| Lockfile | Dependencies | Vulnerabilities |
| --- | ---: | ---: |
| v10 parser | 28 | 0 |
| v11 parser | 28 | 0 |
| v12 parser | 28 | 0 |
| v12 deferred WASM | 1 | 0 |

Fixture `package.json` files are Able package metadata rather than npm
dependency lockfiles.

## Proactive refresh probe

A temporary copy proved the bounded fixable update:

- `golang.org/x/net`: `v0.54.0` to `v0.56.0`;
- resolver-required `golang.org/x/crypto`: `v0.52.0` to `v0.53.0`; and
- resolver-required `golang.org/x/sys`: `v0.45.0` to `v0.46.0`.

The probe changes three `go.mod` lines and eight `go.sum` lines. Its Go 1.26.5
symbol scan remains at zero reachable and zero imported-package
vulnerabilities. All seven fixable `x/net` module advisories disappear; only
`GO-2026-5932` remains at module level because the upstream advisory marks
`x/crypto/openpgp` unsafe by design and supplies no fixed version. Able does
not import that package.

The probe modified only its `/var/tmp` copy and does not constitute retained
dependency work.

## Integrity

- All 197 `go.mod`, `go.sum`, `package.json`, and `package-lock.json` files
  remained byte-identical. Their identity-list SHA-256 is
  `5096554566b53df75d217d921c91262d5e70a3ab3e732eab169d5aa91d28fdb9`.
- `HEAD` and `origin/master` remained
  `6efad0a53120129510fdfbab7fbcc84dcd081768`.
- The index remained empty.
- The original 34 fully expanded deferred WASM paths remained byte-identical.
- Audit tools, Go caches, npm caches, copied sources, and raw scanner output
  stayed under disk-backed `/var/tmp`.

## Next

Perform the bounded proactive v12 `x/net` security refresh.

Why: active v12 has no reachable vulnerability, but seven fixable `x/net`
module advisories remain in its dependency graph and could become reachable
through future imports.

What it entails: update `x/net` to `v0.56.0` with the resolver-required
`x/crypto v0.53.0` and `x/sys v0.46.0` changes, rerun Go 1.26.5 production
and test `govulncheck` scans, execute bounded Go correctness tests, and
preserve every deferred WASM path.

Why it matters: this removes all fixable active-v12 module advisories without
conflating frozen v10/v11 exposure or the unfixable, unimported
`x/crypto/openpgp` advisory with current Able reachability.

## References

- Go vulnerability management: <https://go.dev/doc/security/vuln/>
- Go vulnerability database: <https://go.dev/doc/security/vuln/database>
- GitHub Dependabot alert API:
  <https://docs.github.com/en/rest/dependabot/alerts>
- npm audit: <https://docs.npmjs.com/cli/v11/commands/npm-audit>
