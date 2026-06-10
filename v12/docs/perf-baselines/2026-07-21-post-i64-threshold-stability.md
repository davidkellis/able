# Post-i64 compiled threshold reconciliation

Two independent, verifier-backed five-run cohorts were compared for the rows
near the compiled 95%-of-Go threshold. Each ratio is the arithmetic mean of
five Able process timings divided by the arithmetic mean of five freshly built
and timed Go-reference processes under the same catalog CPU contract.

| Benchmark | Cohort A | Cohort B | Pooled | Disposition |
| --- | ---: | ---: | ---: | --- |
| `fib` | 1.029x | 1.030x | 1.029x | Established meet: both cohorts are below the 1.0526x limit. |
| `base64` | 1.216x | 0.940x | 1.074x | Volatile crossing with a pooled miss; do not count as an established meet or infer a code regression. |
| `matrixmultiply` | 1.129x | 1.340x | 1.233x | Repeated miss. |
| `monte_carlo_pi` | 0.994x | 0.768x | 0.864x | Both current cohorts meet, but retain the existing volatile classification because prior current-source cohorts crossed materially and the workload is nondeterministic. |

All 20 Able executions and 20 Go executions succeeded and were verifier-backed.
The canonical stdlib source tree remained
`6a412c872ee66752de7c4417a5eda99806a7631da3b880c55574c6a640b82d9b`.
The machine-readable calculations are in
`2026-07-21-post-i64-threshold-stability.json`.

This reconciliation changes the stability manifest only where the fresh full
snapshot requires it: `fib/compiled` becomes an established guard and
`base64/compiled` is removed because the promoted snapshot itself misses.
Bytecode stability evidence is unchanged.
