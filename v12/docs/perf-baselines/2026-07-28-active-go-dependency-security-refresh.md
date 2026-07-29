# Active v12 Go dependency security refresh

Date: 2026-07-28

## Decision

Retain the active v12 Go dependency and toolchain refresh.

The change removes every symbol-reachable vulnerability reported for the
active v12 Go module while preserving the compiler's native lowering,
fallback-free generated programs, interpreter-free final dependency graphs,
and representative compiled performance.

No compiler production path, generated runtime, runtime semantic, interpreter,
bytecode VM, canonical stdlib, language, benchmark, fixture, reference
application, non-primitive nominal rule, archived v10/v11 workspace, or WASM
path changed.

## Triage baseline

The pre-change module used Go 1.26.4 at execution time, `go-git` v5.11.0,
`x/crypto` v0.16.0, `x/net` v0.19.0, `x/sys` v0.15.0, `go-billy` v5.5.0,
and CIRCL v1.3.3.

`govulncheck` v1.6.0, using the vulnerability database updated at
2026-07-27 20:14:16 UTC, found 26 symbol-reachable advisories:

| Module | Reachable advisories |
| --- | ---: |
| `github.com/go-git/go-git/v5` | 10 |
| `golang.org/x/crypto` | 10 |
| `github.com/go-git/go-billy/v5` | 2 |
| `github.com/cloudflare/circl` | 2 |
| `golang.org/x/net` | 1 |
| Go standard library | 1 |

Of the 25 reachable findings with GitHub advisory severity, 5 were critical,
7 high, 9 moderate, and 4 low. The remaining Go standard-library ECH
disclosure had no severity score in the consulted record.

The frozen v10 and v11 Go modules reproduce the same old dependency findings,
but remain outside the active v12 change boundary. The active parser's
disposable latest-compatible npm, Rust, and optional Python resolutions
reported no advisories. The tracked deferred-WASM npm lock also reported no
advisories.

## Retained versions

The module retains its Go 1.25 language compatibility floor and now declares
`toolchain go1.26.5`. Go 1.26.5 was installed locally for verification.

The direct security boundary is:

- `github.com/go-git/go-git/v5`: v5.11.0 to v5.19.1;
- `golang.org/x/crypto`: v0.16.0 to v0.52.0; and
- `golang.org/x/sys`: v0.15.0 to v0.45.0.

`go mod tidy` selected the corresponding safe transitive graph, including
`go-billy` v5.9.0, CIRCL v1.6.3, and `x/net` v0.54.0, and removed unused
historical `x/mod` and `x/tools` requirements.

A disposable staged experiment showed the reduction boundary before the
retained edit:

- raising only `go-git` to v5.19.1 reduced reachable findings from 26 to 8;
- adding `x/crypto` v0.52.0 reduced them from 8 to 1; and
- the remaining standard-library finding was fixed by Go 1.26.5.

The retained module scan reports:

```text
No vulnerabilities found.

Your code is affected by 0 vulnerabilities.
This scan also found 0 vulnerabilities in packages you import and 8
vulnerabilities in modules you require, but your code doesn't appear to call
these vulnerabilities.
```

## Correctness and build verification

All commands used Go 1.26.5 and disk-backed caches under
`/var/tmp/able-v12-security-upgrade-20260728`.

- `go mod verify`, `go vet ./...`, and `go build ./...` pass.
- `go test -json ./cmd/able` passes 105 tests; the package takes 48.956
  seconds and its longest individual test takes 10.6 seconds.
- The complete compiler package passes 1,084 reported tests with zero
  failures under a 30-minute package envelope.
- The initial unpartitioned `go test ./...` reached the default 10-minute
  compiler-package timeout after 3,845 passed tests and zero individual
  failures. This is an aggregate harness limit, not a dependency
  incompatibility.
- The ordinary `./run_all_tests.sh` handoff passes all preflights,
  non-compiler packages, 34 bounded compiler short-mode batches, and the
  bytecode fixture pass.

The monolithic compiler run is not individual-duration evidence. Its parallel
fixture children contend for the workstation, and the legacy
`TestCompilerExecFixtureFallbacks` wrapper serially loops over the full corpus.
The existing release gate remains the documented bounded/serial 128-partition
matrix; this security-only tranche did not redesign compiler test structure.

## Interpreter-free compiled boundary

Matrix Multiply, QuickSort, and Binary Trees were rebuilt with
`--no-fallbacks` and release-disabled dynamic-boundary telemetry. All three:

- completed within their process bounds;
- passed their public output verifier;
- retained strict fallback-free compilation; and
- produced final Go dependency graphs without
  `able/interpreter-go/pkg/interpreter`.

This confirms that the dependency refresh did not reopen the
compiler/interpreter boundary or disturb native-carrier lowering.

## Repeated performance smoke

Fresh Go 1.26.5 references and strict compiled Able each ran five
verifier-backed processes. Able and Go used the same per-row CPU budget and
executor policy.

| Application | Compiled Able samples (s) | Go samples (s) | Able mean | Go mean | Able/Go |
| --- | --- | --- | ---: | ---: | ---: |
| Matrix Multiply | 1.20, 1.06, 1.12, 1.12, 1.07 | 0.9524, 0.9832, 0.9660, 1.0093, 0.9747 | 1.1140 | 0.9771 | 1.1401x |
| QuickSort | 2.24, 1.88, 1.79, 2.75, 1.98 | 2.9251, 3.0700, 2.6796, 2.5290, 2.4332 | 2.1280 | 2.7274 | 0.7802x |
| Binary Trees | 8.87, 7.94, 7.62, 7.78, 7.97 | 11.1600, 11.0822, 12.5294, 11.2629, 12.1880 | 8.0360 | 11.6445 | 0.6901x |

Target disposition is unchanged: QuickSort and Binary Trees remain faster
than Go. Matrix Multiply remains the same existing target miss, and its
1.1140-second Able mean is effectively unchanged from the current scorecard's
1.1160 seconds. This is a retention smoke, not a new performance-candidate
admission.

## Recommendation

The maintainer authorized exact consolidation and publication of this
six-path active-v12 security tranche. The publication boundary includes only
the module metadata, this record, PLAN, and the two chronological logs; all
deferred WASM paths remain outside the index and commit.

After publication, next refresh the complete 63-row compiled Able/Go
scorecard with Go 1.26.5.

Why: the three-application retention smoke is green, but the authoritative
compiled frontier still records Go 1.26.4 reference processes.

What it entails: rebuild all 63 Go references with the patched toolchain, run
five current strict compiled Able and five matching Go processes per
application under the existing verifier and execution contracts, regenerate
the compiled scorecard/frontier evidence, and inspect target-status or
owner-rank changes.

Why it is important: the full refresh restores single-provenance evidence for
the 95%-of-Go goal. Production performance mutation remains inadmissible
unless that evidence exposes a general owner spanning at least three unlike
programs.
