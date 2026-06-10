# JSON external verifier and bytecode refresh (2026-07-11)

## Purpose

Make the external JSON benchmark eligible performance evidence before using it
to select a bytecode optimization. The previous generality scorecard retained
only an stdout hash for JSON, so its timing was not a correctness-checked
comparison.

## Verifier

`../benchmarks/json/verify.rb` now independently parses that suite's exact
`sample.json`, computes the `x`, `y`, and `z` coordinate means, and accepts
only three finite numeric output lines within `1e-8` relative tolerance. The
tolerance admits the checked-in Go implementation's eight-decimal formatting
without accepting a different mean. It runs after the timed Able process, like
the other external-suite verifiers.

The focused verifier checks passed:

- It accepts the output of `go run ./go-1.26/app.go` against the current
  generated sample.
- It rejects the deliberately incorrect `0`, `0`, `0` means.
- `bench_compare_external --benchmarks json` discovers the new verifier
  automatically and records verified status rather than an unavailable
  validation row.

## Current verified measurement

Under CPU `2`, `GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, and the normal
45-second process guard, three bytecode processes all verified with the same
stdout SHA-256:

| Mode | Verified runs | Able real time | Go | Ruby | Python |
| --- | ---: | ---: | ---: | ---: | ---: |
| Bytecode | 3/3 | 0.8533 s | 0.63x | 0.55x | 0.30x |

The prior one-run generality snapshot listed JSON at 3.7300 s and had no
verifier. The new verified result is therefore a baseline correction, not
evidence that a new VM change produced a speedup. JSON is currently faster
than all three stored reference rows and is not a bytecode-miss candidate.

## Profile and decision

The normal-process profile
`.profiles/20260711_json_external_bytecode_process.cpu.pprof` has 720 ms of
samples. `ableJsonF64FieldMeansFast` accounts for 66.67% cumulative, with
`strconv` float parsing and JSON number/string/value scanning beneath it.
`execAndFinishExactNativeCall` is its broad 69.44% caller, not a reusable
optimization leaf. This differs from the prior Base64 native codec/MD5 path;
there is no shared host-operation descendant and JSON is a verified neutral
control rather than a miss.

Keep no VM, compiler, or `able-stdlib` performance change.

## Next recommendation

Refresh the completed verifier-backed external bytecode rows before selecting
another implementation candidate. A one-run, unverified JSON row differed by
more than fourfold from the current three-run verified result, so the existing
scorecard cannot safely rank remaining misses. The refresh should use the same
CPU/OOM guardrails and three normal processes per eligible workload, retain
stdout verification, and then choose two *current* misses from independent
families plus a current neutral control. This prevents optimizing an obsolete
or unchecked result.
