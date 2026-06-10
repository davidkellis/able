# Truthiness/cast target-guard refresh

Date: 2026-07-21

## Decision

The compiled and bytecode target-guard closures are current against the
post-truthiness/cast shared interpreter semantics. No performance candidate is
admitted by this refresh.

Three of four compiled guards meet the project target. Fib remains a measured
miss after an order-balanced second cohort: all 10 retained Able processes
average 4.084 seconds and all 10 fresh Go processes average 3.780 seconds, or
1.080x Go. This exceeds the 1.052632x ceiling implied by the 95% target.

Both bytecode guards meet the target with wide margins. PiDigits is 0.479x its
faster Python reference and JSON is 0.405x its faster Ruby reference.

## Frozen contract

- One current Able source/toolchain and canonical external stdlib were used
  throughout. The v12 spec SHA-256 is
  `4f0405b86c122993723e8617abd6f825d9a8ff858d4c72acaf4e33469452f080`;
  the sorted canonical stdlib-source tree SHA-256 is
  `43ff2e68e59c8be7fb1024c86a1f61a0eea84596279b4f0e146511d66c5308d8`.
- Executables were built before timing. Every successful process passed the
  benchmark repository's public verifier.
- Each ordinary row retains five independent processes and uses the arithmetic
  mean without removing outliers. Fib retains two oppositely ordered cohorts,
  for 10 Able and 10 Go processes.
- Every process was bounded by 55 seconds. The CPU-affinity pool was `0-15`;
  the catalog resolved Binary Trees to CPUs `0,1,2,3` with the goroutine
  executor policy and all other rows to CPU `0` with the serial policy.
- The runner used its guarded defaults, including `GOMEMLIMIT=1GiB` and
  `GOGC=50` for Able executions.

## Results

| Benchmark | Mode | Able samples | Able mean (s) | Limiting reference | Reference samples | Reference mean (s) | Ratio | Target |
| --- | --- | ---: | ---: | --- | ---: | ---: | ---: | --- |
| Binary Trees | compiled | 5 | 8.254 | Go | 5 | 12.281 | 0.672x | meet |
| Fib | compiled | 10 | 4.084 | Go | 10 | 3.780 | 1.080x | miss |
| JSON | compiled | 5 | 0.844 | Go | 5 | 1.987 | 0.425x | meet |
| QuickSort | compiled | 5 | 1.998 | Go | 5 | 3.303 | 0.605x | meet |
| PiDigits | bytecode | 5 | 2.436 | Python | 5 | 5.082 | 0.479x | meet |
| JSON | bytecode | 5 | 0.964 | Ruby | 5 | 2.381 | 0.405x | meet |

All 80 timed processes represented in the decision verified successfully.
The largest Able coefficient of variation was JSON bytecode at 11.84%, below
the 15% instability boundary; its margin to the target is also large. Fib's
extra cohort was collected because its first five-process result was close to
the target and had crossed it in prior evidence. The combined result remains
a miss, so no favorable sample selection is involved.

## Exact artifacts

- Compiled Able cohort:
  `2026-07-21-truthiness-cast-target-guards-compiled.json`
  (`ced7f885aac3364870086f9af07372d84424e2c035188b00fab6a5ffa2c68737`)
- Initial Go references:
  `2026-07-21-truthiness-cast-target-guards-go-reference.json`
  (`d645e7f02ff1348c3f4f456c30ebaeda38ca50989065927f7b2050b48b3e8dc5`)
- Fib second Able cohort:
  `2026-07-21-truthiness-cast-target-guards-fib-c2-compiled.json`
  (`a8d3f77eb7c1c90f4f869b260802d97b3114cb6368796035c23004d6e57dff3d`)
- Fib second Go cohort:
  `2026-07-21-truthiness-cast-target-guards-fib-c2-go-reference.json`
  (`66a82a7da50bdd5a23d04fba352f916107ac8ec89a63e056b031e00dfd81f5d6`)
- Bytecode Able cohort:
  `2026-07-21-truthiness-cast-target-guards-bytecode.json`
  (`9307ed901c24daaa90f880a68fc71740f243a2f8b0315b7962967030f09bc9e4`)
- Python/Ruby references:
  `2026-07-21-truthiness-cast-target-guards-interpreter-reference.json`
  (`1859c2fee74e1d28a811c2adb4c68f2dda7faa7cbdd5bb043094c21d0a29e5c3`)

The JSON artifacts retain every process sample, stdout identity, source
identity, verifier identity, execution contract, failure count, and timeout
count. Their adjacent Markdown renderings are human-readable views only.

## Reproduction

```sh
./v12/bench_refresh_go_refs \
  --benchmarks quicksort,json,fib,binarytrees \
  --runs 5 --timeout 55 --cpu-affinity 0-15 \
  --output-json v12/docs/perf-baselines/2026-07-21-truthiness-cast-target-guards-go-reference.json \
  --output-md v12/docs/perf-baselines/2026-07-21-truthiness-cast-target-guards-go-reference.md

./v12/bench_compare_external \
  --benchmarks binarytrees,fib,json,quicksort --modes compiled --languages go \
  --runs 5 --timeout 55 --cpu-affinity 0-15 \
  --go-reference-json v12/docs/perf-baselines/2026-07-21-truthiness-cast-target-guards-go-reference.json \
  --output-json v12/docs/perf-baselines/2026-07-21-truthiness-cast-target-guards-compiled.json \
  --output-md v12/docs/perf-baselines/2026-07-21-truthiness-cast-target-guards-compiled.md

./v12/bench_refresh_interpreter_refs \
  --benchmarks json,pidigits --languages python,ruby \
  --runs 5 --timeout 55 --cpu-affinity 0-15 \
  --output-json v12/docs/perf-baselines/2026-07-21-truthiness-cast-target-guards-interpreter-reference.json \
  --output-md v12/docs/perf-baselines/2026-07-21-truthiness-cast-target-guards-interpreter-reference.md

./v12/bench_compare_external \
  --benchmarks pidigits,json --modes bytecode --languages python,ruby \
  --runs 5 --timeout 55 --cpu-affinity 0-15 \
  --reference-json v12/docs/perf-baselines/2026-07-21-truthiness-cast-target-guards-interpreter-reference.json \
  --output-json v12/docs/perf-baselines/2026-07-21-truthiness-cast-target-guards-bytecode.json \
  --output-md v12/docs/perf-baselines/2026-07-21-truthiness-cast-target-guards-bytecode.md
```

The second Fib cohort deliberately ran Able before its fresh Go reference to
balance the first cohort's reference-before-Able order. It uses the same flags
and is retained in the two `fib-c2` artifacts above.

## Next recommendation

Refresh the compiled current-control and bytecode iterator/control closures
next, reusing this frozen source state and reference evidence.

Why: the correctness change was specifically in shared truthiness and explicit
cast control flow. Those closures are the nearest performance dependents and
are therefore the most informative remaining invalidations; refreshing an
unrelated ownership family first would consume benchmark time without testing
the path that changed.

What it entails: run their source-exact applications in repeated verified
processes, collect bounded current CPU/allocation profiles where the shared
truthiness/cast boundary is reached, retain every sample, and update only those
two closure evidence records. A candidate should be built only if one concrete
leaf is material in at least three unlike applications and survives the target
guards recorded here.
