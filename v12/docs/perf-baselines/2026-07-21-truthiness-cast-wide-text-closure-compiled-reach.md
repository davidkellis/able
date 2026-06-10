# Compiled dynamic-boundary reachability audit

Suite: `coverage`. This is a debug-only `dynamic-boundary` event audit, not a timing result.

| Benchmark | Status | Verification | Explicit dynamic | Residual polymorphic | Host/ABI | Runtime service | Truthy checks | Truthy Error fallback | Explicit casts | Cast failures |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| fixed_width_128 | ok | verified | 2 | 2 | 2 | 0 | 0 | 0 | 0 | 0 |
| rational_series | ok | verified | 2 | 2 | 2 | 0 | 0 | 0 | 0 | 0 |
| wide_integer_records | ok | verified | 3 | 3 | 3 | 0 | 0 | 0 | 0 | 0 |

Totals count events in successful telemetry payloads only. Counters are reachability evidence, not CPU or allocation attribution; use them only to select a later verifier-backed profile set.
