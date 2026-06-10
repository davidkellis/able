# Compiled String/Byte Overlap Profiles

## Decision

Keep no compiler, runtime, canonical-stdlib, or benchmark-source performance
change. The Sudoku `String.bytes` materialization cost does not repeat in the
independent I-Before-E or Base64 compiled applications. JSON remains a healthy
control with a distinct numeric-parser profile. Therefore a `String.bytes`,
codec, environment-switch, or source-shape optimization is not authorized.

## Method

- Rebuilt current compiled I-Before-E, Base64, and JSON binaries with
  `ABLE_GO_PHASE_CPU_PROFILE_DIR`, which records a collector-free `main` CPU
  profile separately from bootstrap.
- Used the exact external-suite directories and inputs used by
  `bench_compare_external`: `../benchmarks/i-before-e` with `wordlist.txt`,
  `../benchmarks/base64`, and `../benchmarks/json`. JSON's existing suite
  setup was run before profiling.
- Used `taskset -c 2`, `GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, and a
  45-second guard. I-Before-E ran 60 times; Base64 3; JSON 8. Each workload
  had one stdout hash across all successful runs.
- Retained generated binaries, per-run profiles/hashes, and merged main
  profiles in `v12/tmp/compiled-string-byte-profiles/phase-cpu-external/`.

I-Before-E and Base64 also pass their external `verify.rb` contracts. JSON has
no suite verifier; its rebuilt binary emitted a stable three-line result with
SHA-256 `493f4be223297a435cd0d4af980f7dcaa397a8e913137c44cfd8df6d1c75e276`.

## Results

| Workload | Main CPU samples | Material descendants | Relation to Sudoku |
| --- | ---: | --- | --- |
| Sudoku, prior control | 2.52 s | `String.bytes`/`validated_bytes` 59.1% cumulative; per-byte runtime-value conversion 23.8% | The proposed generic string-to-byte boundary. |
| I-Before-E | 1.88 s | `read_lines` 33.5% cumulative; string validation 29.3%; `String.contains` 24.5%; `String.len_bytes` 15.4%; `SwapEnvIfNeeded` 14.9% | Uses string search and byte length, not `String.bytes` or per-byte value materialization. |
| Base64 | 7.06 s | Go base64 encode 39.1%; decode 31.6%; MD5 14.9% | Byte-array codec/MD5 computation; no material string-byte conversion path. |
| JSON control | 4.82 s | specialized JSON scan 80.9%; numeric parsing 50.2%; `ParseFloat` 38.8% | Independent parser and numeric-conversion cost; no material overlap. |

The I-Before-E generated code spends its direct string calls in the existing
environment bridge and `strings` search path. That bridge is not material in
Base64 or JSON, and the execution-context prototype has already failed its
broad fixture/external guard. Base64's sampled time is deliberately inside the
standard Go codec and MD5 implementations; changing it would not repair
Sudoku's iterator/value-allocation path. JSON's parser is a separate
purpose-built host helper and serves as the required control rather than an
authorization for a parser-specific lowering rule.

## Next recommendation

Add correctness validation to `bench_compare_external` before another
performance candidate. Why: the public benchmark suites already provide
`verify.rb` contracts for many workloads, yet the comparison runner currently
records timing without invoking them; deterministic hashes alone cannot prove
equivalence. The work entails running an available suite verifier against each
successful Able stdout capture, recording verified/unavailable status and a
stdout hash in the JSON/Markdown artifacts, and failing a measured row on
verification failure. Keep an explicit unavailable status for suites such as
JSON that have no verifier. This makes the broad benchmark suite a trustworthy
optimization gate without changing runtime behavior or introducing
benchmark-specific compiler paths.
