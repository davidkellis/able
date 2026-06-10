# Compiled dynamic-boundary reachability audit

Suite: `coverage`. This is a debug-only `dynamic-boundary` event audit, not a timing result.

| Benchmark | Status | Verification | Explicit dynamic | Residual polymorphic | Host/ABI | Runtime service | Truthy checks | Truthy fallback | Explicit casts |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| distance_field | ok | verified | 1 | 1 | 1 | 0 | 0 | 0 | 0 |
| mandelbrot | ok | verified | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| monte_carlo_pi | ok | verified | 5 | 5 | 5 | 0 | 0 | 0 | 0 |
| nbody | ok | verified | 2 | 2 | 2 | 0 | 0 | 0 | 0 |
| rms_norm | ok | verified | 2 | 2 | 2 | 0 | 0 | 0 | 0 |

Totals count events in successful telemetry payloads only. Counters are reachability evidence, not CPU or allocation attribution; use them only to select a later verifier-backed profile set.
