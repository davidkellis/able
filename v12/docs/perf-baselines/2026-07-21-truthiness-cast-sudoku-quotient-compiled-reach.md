# Compiled dynamic-boundary reachability audit

Suite: `coverage`. This is a debug-only `dynamic-boundary` event audit, not a timing result.

| Benchmark | Status | Verification | Explicit dynamic | Residual polymorphic | Host/ABI | Runtime service |
| --- | --- | --- | ---: | ---: | ---: | ---: |
| sudoku_masks | ok | verified | 100 | 100 | 100 | 0 |

Totals count events in successful telemetry payloads only. Counters are reachability evidence, not CPU or allocation attribution; use them only to select a later verifier-backed profile set.
