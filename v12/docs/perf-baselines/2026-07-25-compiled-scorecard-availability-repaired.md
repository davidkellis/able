# Compiled scorecard availability repaired

Date: 2026-07-25

## Decision

Retain the general compiled-benchmark preparation correction in `bench_perf`:
emit the generated Go module and build that module as two independently
bounded phases. Also honor an explicitly supplied `GOCACHE` so large benchmark
builds can use a disk-backed shared cache.

Retain the fast harness-contract regression test. Retain no compiler,
generated-code, runtime, canonical-stdlib, application, interpreter, bytecode,
language, dependency, or WASM change.

All eight rows from the July 25 repair report are now available. Two
independent five-run passes produced 80/80 public-verifier successes, with no
runtime failure or timeout. The strict modules inspected after the first pass
each contain 96 dependencies and omit
`able/interpreter-go/pkg/interpreter`.

The machine-readable companion is
`2026-07-25-compiled-scorecard-availability-repaired.json`.

## Failure classification

The initial current-harness smoke used each catalog execution contract,
`--no-fallbacks`, a 60-second runtime limit, and a 60-second build limit.
Mutex Ledger and Concurrent State Machines built and verified. The other six
rows exited 124 while `bench_perf` reported one combined "Preparing compiled
binary" operation.

That operation invoked `ablec -build`, so source emission and the generated Go
build shared one timeout. Direct phase attribution with the same frozen
compiler showed that every genuine phase independently fit the limit:

| Application | Emit wall | Emit peak RSS | Go bytes | Go-build wall | Go-build peak RSS |
| --- | ---: | ---: | ---: | ---: | ---: |
| Regex Suffix Audit | 25.17 s | 1,006,424 KB | 8,745,804 | 16.88 s | 2,438,956 KB |
| Regex Set Audit | 26.73 s | 983,376 KB | 8,733,142 | 16.26 s | 2,421,252 KB |
| Regex Stream Audit | 26.77 s | 964,204 KB | 8,743,473 | 15.94 s | 2,399,832 KB |
| Log Routing Redaction | 24.78 s | 947,380 KB | 8,764,671 | 15.26 s | 2,369,276 KB |
| Config Validation Extraction | 25.06 s | 1,013,456 KB | 8,771,260 | 15.64 s | 2,457,520 KB |
| Policy Record Dispatch | 31.36 s | 1,093,876 KB | 8,946,861 | 16.16 s | 2,407,932 KB |

This is a harness phase-accounting defect, not a compiler timeout, generated
Go build failure, deterministic runtime failure, verifier failure, or
concurrency flake. The correction keeps the one-minute guardrail intact for
both compiler emission and generated Go compilation.

## Retained harness rule

Compiled preparation now performs:

1. one bounded `ablec -main -pkg main -o ...` source-emission phase; then
2. one independently bounded `go build -C ... -mod=mod` phase.

Both use the existing `--build-timeout` value. The build uses
`${GOCACHE:-$GO_DIR/.gocache}`, so callers can keep the cache outside the
RAM-backed `/tmp` and share it across rows. The test
`bench_perf_compiled_phase_test.py` locks down those contracts and rejects a
hidden return to `ablec -build`.

The compiler binary used for phase attribution had SHA-256
`8a64cddbb3c20b341ea20205c75257b558ac05cbdfe4369c06157a00381cc30e`,
the same compiler as the preceding retained tranche.

## Repeated strict verification

The first post-correction pass built each row with `--no-fallbacks`, its
catalog directory, arguments, executor, CPU budget, public Ruby verifier, and
independent 60-second preparation limits. Each application ran in five fresh
processes:

| Application | Verified | Timeout | Failure | Mean | Stdout SHA-256 |
| --- | ---: | ---: | ---: | ---: | --- |
| Mutex Ledger | 5/5 | 0 | 0 | 0.080 s | `58f2e2fda882a7c1e5032a2f34c4d9d05f40ced51199c0be0672ecc4d4cf14e4` |
| Regex Suffix Audit | 5/5 | 0 | 0 | 0.040 s | `b5d5ccfabbfd4dc5952406cb1c42d62b807f75828661c4c3774b251abe38380f` |
| Regex Set Audit | 5/5 | 0 | 0 | 0.060 s | `3d8f861a312416f95b95d59be62614f0ffc7918e86fb984fe6035ca7b7b28da2` |
| Regex Stream Audit | 5/5 | 0 | 0 | 0.040 s | `dd7801f0b104d8bd47aaef64d685bc65a06263851f12c8df8cf75260d09e717b` |
| Log Routing Redaction | 5/5 | 0 | 0 | 0.050 s | `0d9585b01f83904fdf11d47b2902678c1718c8442ed1d84410d61d5d90f60bf4` |
| Config Validation Extraction | 5/5 | 0 | 0 | 0.042 s | `c1aa99b9a13bb6e0c7731cb2aea77e300cd3cecc695df7fd4af90036939341d1` |
| Concurrent State Machines | 5/5 | 0 | 0 | 0.020 s | `96296c1ea028df4cae0d4dde3e2f8a91533b7bb4daf1f19a611ea9b0ec2b0103` |
| Policy Record Dispatch | 5/5 | 0 | 0 | 0.060 s | `f15c66e53e4c89650ae12aff6c2f466f2cf0c7adb6ca3fe5be0127bfba217c05` |

The inspected strict binaries and graphs were:

| Application | Dependencies | Interpreter | Binary SHA-256 |
| --- | ---: | --- | --- |
| Mutex Ledger | 96 | no | `de027a906a8926db08cb334a344ef5a03c0762b061863b70b5a95ecdc345b9ed` |
| Regex Suffix Audit | 96 | no | `3d7e709361140921f604a057bd4b1d4397c09b84fd297555746f08e16122b29d` |
| Regex Set Audit | 96 | no | `084200275953cf19edca647d34af17ea554a57493d475c75cab0abdc4a4353eb` |
| Regex Stream Audit | 96 | no | `bcaf551c48f4e67461943d11f42ff86cd244121d71bd9f930dff734dc235a271` |
| Log Routing Redaction | 96 | no | `db324f5af740ca25585ee23668b2917363ee64c60214b7f0785e1d99e986bcb5` |
| Config Validation Extraction | 96 | no | `2802fccb901244ec51140f1a518403247de6c3f06c77b671f286bc12ab509c0d` |
| Concurrent State Machines | 96 | no | `da1866cc1e31c9a37ff7cd5f56ecbdc4143df9f36537f0347069906cbbb97d64` |
| Policy Record Dispatch | 96 | no | `19ff0c2f81a4da25e0ddc563e453394f75a6aa00cc7f1bebe2d5d515073aa2b3` |

Mutex Ledger and Concurrent State Machines had been described as flaky in the
earlier repair report. They each verified 10/10 across the two final passes,
so no concurrency correction is warranted.

## Current Go comparison slice

The second post-correction pass used `bench_compare_external`, repeated all
eight strict rows five times, and joined the fresh July 25 Go reference:

| Application | Able verified | Able mean | Go mean | Able/Go |
| --- | ---: | ---: | ---: | ---: |
| Mutex Ledger | 5/5 | 0.0800 s | 0.0048 s | 16.67x |
| Regex Suffix Audit | 5/5 | 0.0420 s | 0.0046 s | 9.13x |
| Regex Set Audit | 5/5 | 0.0480 s | 0.0051 s | 9.41x |
| Regex Stream Audit | 5/5 | 0.0660 s | 0.0051 s | 12.94x |
| Log Routing Redaction | 5/5 | 0.0420 s | 0.0045 s | 9.33x |
| Config Validation Extraction | 5/5 | 0.0420 s | 0.0041 s | 10.24x |
| Concurrent State Machines | 5/5 | 0.0320 s | 0.0043 s | 7.44x |
| Policy Record Dispatch | 5/5 | 0.1120 s | 0.0067 s | 16.72x |

These rows restore availability; they do not meet the 1.052632x compiled
target. No performance optimization was admitted from this tranche.

## Verification and artifact discipline

- Initial classification: eight strict rows attempted; two verified and six
  were correctly attributed to the combined preparation timeout.
- Final governing executions: 40/40 verified in the direct strict pass and
  40/40 verified in the comparison pass.
- Focused shell syntax, harness-contract, execution-contract, scorecard
  refresh, compiler-package, and whitespace checks passed.
- Generated modules, binaries, results, profiles, Go cache, and Go temporary
  work stayed under disk-backed `/var/tmp`.
- The raw 1,373-line comparison JSON was not checked in because repository
  files must remain below 1,000 lines. Its SHA-256 was
  `7b83fe779f0a82f2087542025489cc8d91bd337d0f4a61d382ea440641aeb13c`;
  the raw Markdown SHA-256 was
  `ec0dc982829a2ef232c660a49fb62246945e9dd0b1d759cf6a31372162789870`.
- The explicit benchmark workspace and shared Go cache were removed after the
  compact evidence and decision record were preserved.

## Next recommendation

Refresh the full current compiled scorecard using the corrected phase-split
harness, five public-verifier-backed runs per row, current Go references, and
strict dependency audits.

Why: this tranche proves that the eight unavailable rows were harness
accounting failures and exposes substantial 7.44x-16.72x gaps, but the
repository-wide current scorecard still contains stale failure rows and was
assembled before this correction. The next optimization owner must be chosen
from one complete, internally consistent corpus.

What it entails: run every current compiled catalog row with independent
one-minute emission and generated-Go-build limits; retain each catalog
working directory, arguments, executor, and verifier; require five successful
processes; join fresh equivalent-Go measurements; audit every successful
strict dependency graph for the interpreter; and publish a consolidated
scorecard before profiling or implementing another candidate.

Why it is important: Able's compiled goal is native Go performance through
native carriers and minimal runtime boundaries. A complete current scorecard
identifies the real performance frontier and the unlike applications that
must share the next lowering/runtime owner; it prevents optimization work
from being selected around stale timeouts or a narrow benchmark family. Do
not begin WASM work.
