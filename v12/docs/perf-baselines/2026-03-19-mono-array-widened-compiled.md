# 2026-03-19 Mono-Array Widened Slice Compiled Snapshot

Scope:
- compare compiled-mode output with widened mono-array lowering enabled versus disabled
- target the staged array-heavy compiled fixtures: `bench/noop`, `bench/sieve_count`, `bench/sieve_full`
- use the checked-in `v12/bench_perf` helper

Environment:
- date: `2026-03-19`
- commit: `feebd884`
- tree: `dirty`
- host: `Linux 6.17.0-19-generic x86_64 GNU/Linux`
- runs: `5`
- timeout per run: `30s`

Harness details:
- compiled mode now builds through `cmd/ablec` directly
- mono-off comparison uses `--compiled-build-arg=--no-experimental-mono-arrays`

Commands:
```bash
./v12/bench_perf --runs 5 --timeout 30 --modes compiled v12/fixtures/bench/noop/main.able
./v12/bench_perf --runs 5 --timeout 30 --modes compiled v12/fixtures/bench/sieve_count/main.able
./v12/bench_perf --runs 5 --timeout 30 --modes compiled v12/fixtures/bench/sieve_full/main.able
./v12/bench_perf --runs 5 --timeout 30 --modes compiled --compiled-build-arg=--no-experimental-mono-arrays v12/fixtures/bench/noop/main.able
./v12/bench_perf --runs 5 --timeout 30 --modes compiled --compiled-build-arg=--no-experimental-mono-arrays v12/fixtures/bench/sieve_count/main.able
./v12/bench_perf --runs 5 --timeout 30 --modes compiled --compiled-build-arg=--no-experimental-mono-arrays v12/fixtures/bench/sieve_full/main.able
```

Results:

| benchmark | mono on real | mono on gc | mono off real | mono off gc | note |
| --- | ---: | ---: | ---: | ---: | --- |
| `bench/noop` | `0.0100s` | `0.00` | `0.0100s` | `0.00` | flat |
| `bench/sieve_count` | `0.0100s` | `0.00` | `0.0100s` | `0.00` | flat |
| `bench/sieve_full` | `0.0200s` | `1.00` | `0.0200s` | `3.00` | lower timed GC |

Takeaway:
- widened mono-array lowering is not yet producing a visible wall-clock win on this staged compiled trio
- it is reducing timed GC pressure on the heaviest staged array benchmark (`bench/sieve_full`)
- the next meaningful performance work should focus on residual generic `*Array` / `runtime.Value` paths, not on re-measuring the current explicit-typed slice again
