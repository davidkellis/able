# Compiled dynamic-boundary reachability audit

Suite: `core`. This is a debug-only `dynamic-boundary` event audit, not a timing result.

| Benchmark | Status | Verification | Interpreter linked | Explicit dynamic | Residual polymorphic | Host/ABI | Runtime service |
| --- | --- | --- | --- | ---: | ---: | ---: | ---: |
| fib | ok | verified | no | 1 | 1 | 1 | 0 |
| binarytrees | ok | verified | no | 12 | 11 | 11 | 9 |
| matrixmultiply | ok | verified | no | 1 | 1 | 1 | 0 |
| quicksort | ok | verified | no | 20 | 20 | 20 | 0 |
| sudoku_masks | ok | verified | no | 100 | 100 | 100 | 0 |
| i_before_e | ok | verified | no | 1629 | 1629 | 1628 | 0 |

Totals count events in successful telemetry payloads only. Counters are reachability evidence, not CPU or allocation attribution; use them only to select a later verifier-backed profile set.
