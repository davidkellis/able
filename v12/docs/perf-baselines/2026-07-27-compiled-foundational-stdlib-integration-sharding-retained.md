# Compiled foundational stdlib integration sharding retained

Date: 2026-07-27

## Decision

Retain a test-only split of the compiled foundational stdlib integration case
and a canonical-stdlib test organization correction. Retain no production
compiler, generated runtime, runtime, interpreter, bytecode VM, canonical
stdlib implementation, benchmark, language, dependency, or WASM change.

The ordinary diagnostics-off foundational case reproducibly violated the
project's one-minute individual-test rule. Three warm-cache runs had a
139.206-second mean and 138.924-second median. Silent temporary phase timing
showed that Able compilation, not generated program execution, owned the
duration.

The former three-file Go integration case is now three independent cases:

- `simple.test.able`;
- `assertions.test.able`; and
- the Array/Iterator portion of `enumerable.test.able`.

The HashSet eager-map assertion formerly in `enumerable.test.able` now lives in
the canonical `collections/hash_set.test.able` suite. The existing compiled
HashSet integration case asserts that the moved case executes. This preserves
tree-walker, bytecode, and compiled coverage while removing an unrelated
HashSet graph from the foundational Enumerable shard.

## Reproduction

The measurement protocol used:

- `ABLE_RUN_COMPILED_CLI_INTEGRATION=1`;
- `ABLE_COMPILER_REQUIRE_NO_FALLBACKS=true` through the existing test helper;
- no typed-boundary telemetry;
- no retained-workdir output;
- disk-backed `TMPDIR=/var/tmp`;
- warm shared `GOCACHE=/var/tmp/able-go-cache`; and
- a separate `go test -count=1` process for every baseline sample.

One uncounted warm-up took 140.144 seconds. The three counted inner compiled
test durations were:

| Run | Total (s) | Able compile (s) | Go build (s) | Execution (s) |
| --- | ---: | ---: | ---: | ---: |
| 1 | 140.507 | 119.128 | 19.969 | 0.070 |
| 2 | 138.924 | 117.347 | 20.241 | 0.063 |
| 3 | 138.187 | 117.519 | 19.408 | 0.077 |
| Mean | 139.206 | 117.998 | 19.873 | 0.070 |

The median total was 138.924 seconds. The mean outer process elapsed time was
140.323 seconds and mean peak RSS was 3,517,599 KB.

This proves that the earlier 138.88-second observer-enabled duration was not
primarily telemetry or retained-workdir overhead. Able compilation accounted
for 84.8% of the ordinary inner duration, generated-Go build for 14.3%, and
execution for 0.05%.

## Retained sharding result

Three diagnostics-off repetitions of every retained foundational shard passed:

| Shard | Run 1 (s) | Run 2 (s) | Run 3 (s) | Mean (s) | Median (s) |
| --- | ---: | ---: | ---: | ---: | ---: |
| Simple | 39.89 | 38.94 | 39.14 | 39.32 | 39.14 |
| Assertions | 51.47 | 53.62 | 52.76 | 52.62 | 52.76 |
| Enumerable Array/Iterator | 49.76 | 51.75 | 51.65 | 51.05 | 51.65 |

The worst observed individual duration was 53.62 seconds. Every sample stayed
below one minute and retained its original expected-output assertion.

The three shard durations totalled 141.12, 144.31, and 143.55 seconds per
repetition, a 142.99-second mean. That is 2.72% above the old combined mean,
which is a small aggregate cost for bounded individual tests and independent
failure attribution.

Single phase-attributed prototype samples corroborated the split:

| Shard | Total (s) | Able compile (s) | Go build (s) | Execution (s) |
| --- | ---: | ---: | ---: | ---: |
| Simple | 39.382 | 24.267 | 14.420 | 0.071 |
| Assertions | 53.683 | 36.057 | 16.883 | 0.065 |
| Enumerable Array/Iterator | 51.069 | 33.829 | 16.592 | 0.051 |

The temporary phase-timing observer was removed after measurement. Ordinary
and retained code has no timing environment variable or observer overhead.

## Semantic coverage

The affected canonical suites pass with the moved HashSet assertion:

- tree-walker: 3.50 seconds;
- bytecode: 3.01 seconds; and
- existing strict compiled HashMap/HashSet integration: 96.50 seconds.

The complete external canonical stdlib suite also passes:

- tree-walker: 21.07 seconds; and
- bytecode: 17.42 seconds.

The complete `./run_all_tests.sh` handoff passes every contract, all
non-compiler packages, all 33 compiler batches, and the final 85.483-second
bytecode fixture corpus in 736.28 seconds. The known heavy aggregate compiler
batches 19, 28, and 29 took 179.619, 60.523, and 75.191 seconds.

The compiled HashMap/HashSet duration is a pre-existing second individual-test
violation and is the next test-lane signal. The eager-map assertion adds only
execution work to the already compiled HashSet graph; it did not create a new
compiled integration case.

All files remain below 1,000 lines. Generated work directories were removed
after every run. The measurement directory occupied 44 KB under `/var/tmp`
before final cleanup; the shared Go build cache remains active and was not
treated as stale output.

## Why production performance remains paused

The measured owner is Able compilation of stdlib test graphs. Generated
program execution is about 70 milliseconds and supplies no new application
runtime owner. The prior strict-corpus concrete-to-lifted-interface census
therefore remains authoritative: there is still no production compiler/runtime
candidate material in three unlike applications.

No application lowering, carrier, boxing, interpreter boundary, or generated
runtime changed. No benchmark-specific or non-primitive nominal rule was
introduced. WASM remains deferred.

## Next recommendation

Keep production performance mutation paused. Audit the remaining compiled
stdlib integration cases for individual durations over one minute, starting
with the 96.50-second combined HashMap/HashSet case.

Why: the foundational case is now bounded, but the directly adjacent compiled
lane still contains at least one known over-limit test. The project cannot
claim a one-minute release gate until those cases are enumerated and addressed.

What it entails: collect warm diagnostics-off per-case timings across the
compiled stdlib lane; phase-attribute only cases over one minute; then prefer a
general generated-output/build reuse mechanism or narrower feature-package
shards that preserve every expected assertion. Reject a change that merely
moves a slow graph to another test, weakens compiled CLI coverage, or
materially inflates aggregate lane time.

Why it is important: bounded, reusable compiled-stdlib validation makes future
compiler correctness work faster and safer while keeping application runtime
optimization governed by the separate three-unlike-program evidence gate.
Do not begin WASM work.
