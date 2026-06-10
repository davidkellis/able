# Dynamic-i64 lazy-slot compiled threshold reconciliation

The refreshed 45-program compiled snapshot changed two rows at the
95%-of-Go boundary. Every timing below is the arithmetic mean of five
verifier-backed process executions under the catalog CPU contract.

| Benchmark | Cohort A | Cohort B | Pooled | Disposition |
| --- | ---: | ---: | ---: | --- |
| `fib` | 1.210x | — | — | Snapshot miss; remove the prior established guard from snapshot-meet coverage. |
| `matrixmultiply` | 1.050x | 1.268x | 1.155x | Variance-sensitive pooled miss; the snapshot meet is not an established guard. |

For `matrixmultiply`, cohort B rebuilt and timed the Go reference five times
and independently built and timed the current Able binary five times. All ten
executions succeeded and were verifier-backed. The first cohort narrowly met
the 1.0526x limit, the second missed, and their pooled ratio missed.

The `fib` snapshot itself misses, so it is not eligible for the frontier's
snapshot-meet stability set. This is a measurement-state change rather than an
optimization-specific conclusion: the dynamic-i64 lazy slots are not used by
that primitive recursive kernel.

The canonical stdlib source tree remained
`6a412c872ee66752de7c4417a5eda99806a7631da3b880c55574c6a640b82d9b`.
The machine-readable calculations are in the adjacent JSON artifact.
