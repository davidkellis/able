# Compiled dynamic-boundary reachability audit

Suite: `coverage`. This is a debug-only `dynamic-boundary` event audit, not a timing result.

| Benchmark | Status | Verification | Explicit dynamic | Residual polymorphic | Host/ABI | Runtime service | Truthy checks | Truthy fallback | Explicit casts |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| array_slice_window | ok | verified | 1 | 1 | 1 | 0 | 0 | 0 | 0 |
| dependency_plan | ok | verified | 1 | 1 | 1 | 0 | 0 | 0 | 0 |
| document_audit | ok | verified | 2 | 2 | 1 | 0 | 0 | 0 | 0 |
| lexical_rollup | ok | verified | 2 | 2 | 1 | 0 | 0 | 0 | 0 |
| option_result_config | ok | verified | 1 | 57649 | 1 | 0 | 24576 | 0 | 24384 |

Totals count events in successful telemetry payloads only. Counters are reachability evidence, not CPU or allocation attribution; use them only to select a later verifier-backed profile set.
