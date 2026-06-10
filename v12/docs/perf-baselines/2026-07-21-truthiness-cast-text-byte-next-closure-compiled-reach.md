# Compiled dynamic-boundary reachability audit

Suite: `coverage`. This is a debug-only `dynamic-boundary` event audit, not a timing result.

| Benchmark | Status | Verification | Explicit dynamic | Residual polymorphic | Host/ABI | Runtime service | Truthy checks | Truthy Error fallback | Explicit casts | Cast failures |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| i_before_e | ok | verified | 1629 | 1629 | 1628 | 0 | 0 | 0 | 0 | 0 |
| inventory_reconciliation | ok | verified | 1 | 1 | 1 | 0 | 0 | 0 | 0 | 0 |
| k_nucleotide | ok | verified | 22 | 36 | 21 | 0 | 0 | 0 | 0 | 0 |
| unicode_scalar_pipeline | ok | verified | 2 | 2 | 2 | 0 | 0 | 0 | 0 | 0 |
| word_frequency | ok | verified | 2 | 2 | 1 | 0 | 0 | 0 | 0 | 0 |

Totals count events in successful telemetry payloads only. Counters are reachability evidence, not CPU or allocation attribution; use them only to select a later verifier-backed profile set.
