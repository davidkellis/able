# Compiled dynamic-boundary reachability audit

Suite: `coverage`. This is a debug-only event audit, not a timing result.

| Benchmark | Status | Verification | Explicit dynamic | Residual polymorphic | Host/ABI | Runtime service |
| --- | --- | --- | ---: | ---: | ---: | ---: |
| fib | ok | verified | 1 | 1 | 1 | 0 |
| binarytrees | timeout | not-run | n/a | n/a | n/a | n/a |
| matrixmultiply | ok | verified | 1 | 1 | 1 | 0 |
| quicksort | ok | verified | 20 | 20 | 20 | 0 |
| sudoku | timeout | not-run | n/a | n/a | n/a | n/a |
| sudoku_masks | timeout | not-run | n/a | n/a | n/a | n/a |
| i_before_e | ok | verified | 1629 | 1629 | 1628 | 0 |
| base64 | ok | verified | 3 | 3 | 3 | 0 |
| json | ok | verified | 3 | 3 | 3 | 0 |
| monte_carlo_pi | ok | verified | 5 | 5 | 5 | 0 |
| pidigits | ok | verified | 1001 | 1001 | 1000 | 0 |
| mandelbrot | ok | verified | 0 | 0 | 0 | 0 |
| reverse_complement | ok | verified | 1 | 1 | 0 | 0 |
| k_nucleotide | ok | verified | 22 | 36 | 21 | 0 |
| nbody | ok | verified | 2 | 2 | 2 | 0 |
| tapelang_alphabet | ok | verified | 1 | 1 | 0 | 0 |
| fixed_width_128 | timeout | not-run | n/a | n/a | n/a | n/a |
| rational_series | ok | verified | 2 | 2 | 2 | 0 |
| word_frequency | ok | verified | 2 | 2 | 1 | 0 |
| document_audit | ok | verified | 2 | 2 | 1 | 0 |
| lexical_rollup | ok | verified | 2 | 2 | 1 | 0 |
| channel_rollup | ok | verified | 3 | 2 | 1 | 2 |

Totals count events in successful telemetry payloads only. Categories may overlap at one language boundary; for example, a dynamic call that reaches a host function contributes to both relevant counters.
