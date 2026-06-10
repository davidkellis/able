# Compiled call-path reachability audit

Suite: `coverage`. This is a debug-only `call-path` event audit, not a timing result.

| Benchmark | Status | Verification | Fast method-value | Generic union method | Generic union fallback |
| --- | --- | --- | ---: | ---: | ---: |
| fib | ok | verified | 0 | 0 | 0 |
| binarytrees | ok | verified | 9 | 0 | 0 |
| matrixmultiply | ok | verified | 0 | 0 | 0 |
| quicksort | ok | verified | 0 | 0 | 0 |
| sudoku | timeout | not-run | n/a | n/a | n/a |
| sudoku_masks | ok | verified | 0 | 0 | 0 |
| i_before_e | ok | verified | 0 | 0 | 0 |
| base64 | ok | verified | 0 | 0 | 0 |
| json | ok | verified | 0 | 0 | 0 |
| monte_carlo_pi | ok | verified | 0 | 0 | 0 |
| pidigits | ok | verified | 0 | 0 | 0 |
| mandelbrot | ok | verified | 0 | 0 | 0 |
| reverse_complement | ok | verified | 0 | 0 | 0 |
| k_nucleotide | ok | verified | 0 | 0 | 0 |
| nbody | ok | verified | 0 | 0 | 0 |
| tapelang_alphabet | ok | verified | 0 | 0 | 0 |
| fixed_width_128 | ok | verified | 0 | 0 | 0 |
| rational_series | ok | verified | 0 | 0 | 0 |
| word_frequency | ok | verified | 0 | 0 | 0 |
| document_audit | ok | verified | 0 | 0 | 0 |
| lexical_rollup | ok | verified | 0 | 0 | 0 |
| channel_rollup | ok | verified | 0 | 0 | 0 |
| future_pipeline | ok | verified | 8 | 0 | 0 |
| future_await_race | ok | verified | 1059 | 0 | 0 |
| await_channel_mux | ok | verified | 10752 | 0 | 0 |
| mutex_ledger | ok | verified | 4 | 0 | 0 |
| mutex_await_journal | ok | verified | 6196 | 0 | 0 |
| regex_suffix_audit | ok | verified | 0 | 0 | 0 |
| regex_set_audit | ok | verified | 0 | 0 | 0 |
| regex_stream_audit | ok | verified | 0 | 0 | 0 |
| array_slice_window | ok | verified | 0 | 0 | 0 |
| dependency_plan | ok | verified | 0 | 0 | 0 |
| option_result_config | ok | verified | 147456 | 147456 | 0 |

Totals count events in successful telemetry payloads only. Counters are reachability evidence, not CPU or allocation attribution; use them only to select a later verifier-backed profile set.
