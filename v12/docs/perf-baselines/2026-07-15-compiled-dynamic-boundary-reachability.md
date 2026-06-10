# Compiled dynamic-boundary reachability audit

Suite: `coverage`. This is a debug-only event audit, not a timing result.

| Benchmark | Status | Verification | Explicit dynamic | Residual polymorphic | Host/ABI | Runtime service |
| --- | --- | --- | ---: | ---: | ---: | ---: |
| fib | ok | verified | 1 | 1 | 1 | 0 |
| binarytrees | ok | verified | 12 | 11 | 11 | 9 |
| matrixmultiply | ok | verified | 1 | 1 | 1 | 0 |
| quicksort | ok | verified | 20 | 20 | 20 | 0 |
| sudoku | timeout | not-run | n/a | n/a | n/a | n/a |
| sudoku_masks | ok | verified | 100 | 100 | 100 | 0 |
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
| fixed_width_128 | ok | verified | 2 | 2 | 2 | 0 |
| rational_series | ok | verified | 2 | 2 | 2 | 0 |
| word_frequency | ok | verified | 2 | 2 | 1 | 0 |
| document_audit | ok | verified | 2 | 2 | 1 | 0 |
| lexical_rollup | ok | verified | 2 | 2 | 1 | 0 |
| channel_rollup | ok | verified | 3 | 2 | 1 | 2 |
| future_pipeline | ok | verified | 1155 | 1 | 1 | 7 |
| future_await_race | ok | verified | 1312 | 1 | 1 | 482 |
| await_channel_mux | ok | verified | 1538 | 1025 | 1 | 2560 |
| mutex_ledger | ok | verified | 2 | 1 | 1 | 4 |
| mutex_await_journal | ok | verified | 1 | 2049 | 1 | 2052 |
| regex_suffix_audit | ok | verified | 2 | 2 | 1 | 0 |
| regex_set_audit | ok | verified | 2 | 2 | 1 | 0 |
| regex_stream_audit | ok | verified | 2 | 2 | 1 | 0 |
| array_slice_window | ok | verified | 1 | 1 | 1 | 0 |
| dependency_plan | ok | verified | 1 | 1 | 1 | 0 |

Totals count events in successful telemetry payloads only. Categories may overlap at one language boundary; for example, a dynamic call that reaches a host function contributes to both relevant counters.
