# External Scorecard Output Validation

## Decision

Keep the benchmark-harness change. `bench_compare_external` now validates each
successful Able output capture when its public external suite supplies
`verify.rb`; it records validation evidence even when a suite has no verifier.
No compiler, runtime, canonical-stdlib, or benchmark-program behavior changed.

## Behavior

`bench_compare_external` detects `verify.rb` in each selected suite and
forwards it to `bench_perf`. After the timed program process exits
successfully, `bench_perf` hashes its stdout and runs the verifier with that
capture on standard input. Verification is deliberately outside the timed
process, so it cannot affect the reported application timing.

- A passing verifier makes the mode `verified` and records its verified-run
  count and unique SHA-256 stdout hashes.
- A missing verifier is explicitly `unavailable`, but successful captures
  still receive SHA-256 evidence.
- A verifier failure increments the mode failure count, excludes that run from
  timing averages, and makes every external comparison ratio `n/a` for the
  mode. A fast invalid output is therefore never performance evidence.
- `bytecode-runtime` remains `unavailable`: its Go benchmark harness does not
  expose the application stdout contract to the verifier.

The comparison JSON nests this information at
`rows[].able.validation`; the Markdown table adds `Validation` and `Stdout
SHA-256` columns. `bench_perf` exposes the underlying opt-in
`--verify-ruby-script PATH` only so the external wrapper can use the same
capture and timing path.

## Validation

The following bounded real-suite comparison used one compiled run per
workload, CPU affinity `2`, and the normal 45-second guard:

```text
./v12/bench_compare_external \
  --benchmarks i_before_e,json --modes compiled --runs 1 --timeout 45 \
  --cpu-affinity 2 --keep \
  --workdir v12/tmp/external-validation-smoke \
  --output-json v12/tmp/external-validation-smoke/report.json \
  --output-md v12/tmp/external-validation-smoke/report.md
```

I-Before-E was `verified (1)` with stdout hash
`981f0d37a277be25f359c097e10df2ef68009fea2d3e322aeed3a65d0fbaca39`.
JSON, whose public suite has no verifier, was explicitly `unavailable` with
its stdout hash recorded. The retained JSON/Markdown report is under
`v12/tmp/external-validation-smoke/`.

A negative test ran I-Before-E against Base64's unrelated verifier. It
recorded `ok_runs=0`, `failures=1`, validation `failed`, one stdout hash, and
no timing average. The retained result is
`v12/tmp/bench-perf-validation-failure/result.json`.

Focused checks passed:

```text
bash -n v12/bench_perf v12/bench_compare_external
```

## Next recommendation

Refresh the complete external generality scorecard with this validation gate
before choosing another runtime/compiler optimization. Why: the existing
scorecards predate output validation, so the broad performance baseline must
now distinguish verified measurements from suites without a verifier. The work
entails one pinned, guarded compiled/bytecode pass over the `generality` suite,
recording status/hash evidence for every row, then using three-run follow-ups
only for verified misses that repeat a concrete generic hotspot. Do not infer
a performance target from an unavailable or timed-out row, and do not add
benchmark-specific lowering rules.
