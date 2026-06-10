# Cross-mode performance frontier

Generated from `2026-07-22T05:59:20.973255Z` scorecard evidence. This ledger joins all 91 reviewed rows; it excludes unselected status-only rows.

## Outcome

- Selected rows: 49 compiled + 42 bytecode = 91.
- Product target: 8 meet, 83 miss.
- Established cross-cohort guards: 5 (3 compiled + 2 bytecode); 3 snapshot meets are not established.
- Aggregate time above the per-row 95%-of-reference budget: 151.191 seconds.
- Unclosed groups: 1.

## Recommended next gate

Refresh `compiled-iterator-control` first (1.289 aggregate excess seconds). Fifty verified CPU profiles place this exact helper at 11.83%-58.97% cumulative across four unlike applications. Exact line-level allocation profiles attribute 225,280, 147,456, and 20,480 escaping bound-method boxes to the first three, satisfying the generic admission bar without a nominal-type special case.

## Actionable groups

| Rank | Group | Action | Rows | Misses | Excess s | Max ratio | Freshness |
| ---: | --- | --- | ---: | ---: | ---: | ---: | --- |
| 1 | `compiled-iterator-control` | open-candidate | 6 | 6 | 1.289 | 72.941 | current-exact-nominal-result-allocation-shape-census |

## Complete selected-row ledger

`Excess s` is Able wall time beyond the fastest applicable reference multiplied by the allowed ratio (1 / 0.95). `Established` is a separate cross-cohort candidate-admission guard and never rewrites snapshot status.

| Benchmark | Mode | Snapshot | Established | Stability | Able s | Worst ratio | Excess s | Freshness | Disposition | Group |
| --- | --- | --- | --- | --- | ---: | ---: | ---: | --- | --- | --- |
| array_slice_window | bytecode | miss | not-applicable | - | 0.648 | 22.041 | 0.617 | current-exact-post-truthiness-cast-shared-semantics | closed-no-shared-leaf | `bytecode-iterator-control` |
| await_channel_mux | bytecode | miss | not-applicable | - | 0.208 | 1.726 | 0.081 | current-exact-validated-job-file-entry-refresh | closed-no-shared-leaf | `bytecode-concurrency` |
| base64 | bytecode | meets | not-established | variance-sensitive-miss | 2.806 | 1.048 | 0.000 | current-exact-post-truthiness-cast-refresh | closed-no-shared-leaf | `bytecode-byte-output` |
| binary_event_log | bytecode | miss | not-applicable | - | 6.896 | 28.914 | 6.645 | current-exact-binary-event-log-main-profile | closed-rejected-candidate | `bytecode-text-map` |
| channel_rollup | bytecode | miss | not-applicable | - | 0.446 | 10.825 | 0.403 | current-exact-validated-job-file-entry-refresh | closed-no-shared-leaf | `bytecode-concurrency` |
| concurrent_document_pipeline | bytecode | miss | not-applicable | - | 0.348 | 15.605 | 0.325 | current-exact-validated-job-file-entry-refresh | closed-no-shared-leaf | `bytecode-concurrency` |
| concurrent_event_routing | bytecode | miss | not-applicable | - | 3.032 | 102.088 | 3.001 | current-exact-validated-job-file-entry-refresh | closed-no-shared-leaf | `bytecode-concurrency` |
| concurrent_text_index | bytecode | miss | not-applicable | - | 0.634 | 7.512 | 0.545 | current-exact-validated-job-file-entry-refresh | closed-no-shared-leaf | `bytecode-concurrency` |
| config_validation_extraction | bytecode | miss | not-applicable | - | 1.386 | 56.341 | 1.360 | current-exact-post-truthiness-cast-refresh | closed-rejected-candidate | `bytecode-regex` |
| dependency_plan | bytecode | miss | not-applicable | - | 0.498 | 30.741 | 0.481 | current-exact-post-truthiness-cast-shared-semantics | closed-no-shared-leaf | `bytecode-iterator-control` |
| dependency_wave_validation | bytecode | miss | not-applicable | - | 0.420 | 8.750 | 0.369 | current-exact-validated-job-file-entry-refresh | closed-no-shared-leaf | `bytecode-concurrency` |
| distance_field | bytecode | miss | not-applicable | - | 5.748 | 13.718 | 5.307 | current-exact-post-truthiness-cast-near-path | closed-rejected-candidate | `bytecode-float-numeric` |
| document_audit | bytecode | miss | not-applicable | - | 0.274 | 19.571 | 0.259 | current-exact-post-truthiness-cast-shared-semantics | closed-no-shared-leaf | `bytecode-iterator-control` |
| fasta_generation | bytecode | miss | not-applicable | - | 1.828 | 9.036 | 1.615 | current-exact-post-truthiness-cast-refresh | closed-no-shared-leaf | `bytecode-byte-output` |
| fixed_width_128 | bytecode | miss | not-applicable | - | 7.988 | 21.873 | 7.604 | current-exact-post-truthiness-cast-numeric-next | closed-rejected-candidate | `bytecode-wide-numeric` |
| future_await_race | bytecode | miss | not-applicable | - | 0.136 | 4.012 | 0.100 | current-exact-validated-job-file-entry-refresh | closed-no-shared-leaf | `bytecode-concurrency` |
| future_pipeline | bytecode | miss | not-applicable | - | 0.406 | 6.612 | 0.341 | current-exact-validated-job-file-entry-refresh | closed-no-shared-leaf | `bytecode-concurrency` |
| i_before_e | bytecode | miss | not-applicable | - | 0.526 | 6.315 | 0.438 | current-exact-binary-event-log-main-profile | closed-rejected-candidate | `bytecode-text-map` |
| inventory_reconciliation | bytecode | miss | not-applicable | - | 2.536 | 37.294 | 2.464 | current-exact-binary-event-log-main-profile | closed-rejected-candidate | `bytecode-text-map` |
| json | bytecode | meets | established | established-meet | 0.816 | 0.474 | 0.000 | current-exact-post-truthiness-cast-shared-semantics | target-guard | `bytecode-target-guards` |
| k_nucleotide | bytecode | miss | not-applicable | - | 45.164 | 35.881 | 43.839 | current-exact-binary-event-log-main-profile | closed-rejected-candidate | `bytecode-text-map` |
| lexical_rollup | bytecode | miss | not-applicable | - | 0.398 | 23.275 | 0.380 | current-exact-post-truthiness-cast-shared-semantics | closed-no-shared-leaf | `bytecode-iterator-control` |
| log_routing_redaction | bytecode | miss | not-applicable | - | 3.016 | 124.628 | 2.991 | current-exact-post-truthiness-cast-refresh | closed-rejected-candidate | `bytecode-regex` |
| mandelbrot | bytecode | miss | not-applicable | - | 6.324 | 5.367 | 5.084 | current-exact-post-truthiness-cast-near-path | closed-rejected-candidate | `bytecode-float-numeric` |
| manifest_normalization | bytecode | miss | not-applicable | - | 1.784 | 78.590 | 1.760 | current-exact-binary-event-log-main-profile | closed-rejected-candidate | `bytecode-text-map` |
| monte_carlo_pi | bytecode | miss | not-applicable | - | 2.492 | 1.688 | 0.938 | current-exact-post-truthiness-cast-near-path | closed-rejected-candidate | `bytecode-float-numeric` |
| mutex_await_journal | bytecode | miss | not-applicable | - | 0.220 | 10.427 | 0.198 | current-exact-validated-job-file-entry-refresh | closed-no-shared-leaf | `bytecode-concurrency` |
| mutex_ledger | bytecode | miss | not-applicable | - | 0.368 | 8.700 | 0.323 | current-exact-validated-job-file-entry-refresh | closed-no-shared-leaf | `bytecode-concurrency` |
| mutex_work_queue | bytecode | miss | not-applicable | - | 0.346 | 12.628 | 0.317 | current-exact-validated-job-file-entry-refresh | closed-no-shared-leaf | `bytecode-concurrency` |
| option_result_config | bytecode | miss | not-applicable | - | 0.786 | 42.486 | 0.767 | current-exact-post-truthiness-cast-shared-semantics | closed-no-shared-leaf | `bytecode-iterator-control` |
| pidigits | bytecode | meets | established | established-meet | 2.356 | 0.563 | 0.000 | current-exact-post-truthiness-cast-shared-semantics | target-guard | `bytecode-target-guards` |
| policy_record_dispatch | bytecode | miss | not-applicable | - | 8.782 | 358.449 | 8.756 | current-exact-post-truthiness-cast-refresh | closed-rejected-candidate | `bytecode-regex` |
| rational_series | bytecode | miss | not-applicable | - | 4.046 | 35.152 | 3.925 | current-exact-post-truthiness-cast-numeric-next | closed-rejected-candidate | `bytecode-wide-numeric` |
| regex_set_audit | bytecode | miss | not-applicable | - | 4.020 | 222.099 | 4.001 | current-exact-post-truthiness-cast-refresh | closed-rejected-candidate | `bytecode-regex` |
| regex_stream_audit | bytecode | miss | not-applicable | - | 3.548 | 193.880 | 3.529 | current-exact-post-truthiness-cast-refresh | closed-rejected-candidate | `bytecode-regex` |
| regex_suffix_audit | bytecode | miss | not-applicable | - | 3.230 | 182.486 | 3.211 | current-exact-post-truthiness-cast-refresh | closed-rejected-candidate | `bytecode-regex` |
| reverse_complement | bytecode | miss | not-applicable | - | 3.290 | 123.684 | 3.262 | current-exact-post-truthiness-cast-refresh | closed-no-shared-leaf | `bytecode-byte-output` |
| rms_norm | bytecode | miss | not-applicable | - | 4.526 | 8.854 | 3.988 | current-exact-post-truthiness-cast-near-path | closed-rejected-candidate | `bytecode-float-numeric` |
| unicode_scalar_pipeline | bytecode | miss | not-applicable | - | 3.776 | 16.933 | 3.541 | current-exact-binary-event-log-main-profile | closed-rejected-candidate | `bytecode-text-map` |
| validated_job_pipeline | bytecode | miss | not-applicable | - | 0.414 | 11.250 | 0.375 | current-exact-validated-job-file-entry-refresh | closed-no-shared-leaf | `bytecode-concurrency` |
| wide_integer_records | bytecode | miss | not-applicable | - | 5.158 | 59.630 | 5.067 | current-exact-post-truthiness-cast-numeric-next | closed-rejected-candidate | `bytecode-wide-numeric` |
| word_frequency | bytecode | miss | not-applicable | - | 1.418 | 75.426 | 1.398 | current-exact-binary-event-log-main-profile | closed-rejected-candidate | `bytecode-text-map` |
| array_slice_window | compiled | miss | not-applicable | - | 0.088 | 17.600 | 0.083 | current-exact-nominal-result-allocation-shape-census | open-candidate | `compiled-iterator-control` |
| await_channel_mux | compiled | miss | not-applicable | - | 0.356 | 72.653 | 0.351 | current-exact-validated-job-file-entry-refresh | closed-rejected-candidate | `compiled-concurrency` |
| base64 | compiled | miss | not-applicable | - | 2.558 | 1.061 | 0.021 | current-exact-post-truthiness-cast-refresh | closed-no-shared-leaf | `compiled-byte-output` |
| binary_event_log | compiled | miss | not-applicable | - | 0.688 | 71.667 | 0.678 | current-exact-nominal-result-allocation-shape-census | open-candidate | `compiled-iterator-control` |
| binarytrees | compiled | meets | established | established-meet | 9.578 | 0.897 | 0.000 | current-exact-post-truthiness-cast-shared-semantics | target-guard | `compiled-target-guards` |
| channel_rollup | compiled | miss | not-applicable | - | 0.540 | 93.103 | 0.534 | current-exact-validated-job-file-entry-refresh | closed-rejected-candidate | `compiled-concurrency` |
| concurrent_document_pipeline | compiled | miss | not-applicable | - | 0.252 | 68.108 | 0.248 | current-exact-validated-job-file-entry-refresh | closed-rejected-candidate | `compiled-concurrency` |
| concurrent_event_routing | compiled | miss | not-applicable | - | 3.234 | 673.750 | 3.229 | current-exact-validated-job-file-entry-refresh | closed-rejected-candidate | `compiled-concurrency` |
| concurrent_text_index | compiled | miss | not-applicable | - | 0.876 | 141.290 | 0.869 | current-exact-validated-job-file-entry-refresh | closed-rejected-candidate | `compiled-concurrency` |
| config_validation_extraction | compiled | miss | not-applicable | - | 0.102 | 24.286 | 0.098 | current-exact-post-truthiness-cast-refresh | closed-rejected-candidate | `compiled-regex` |
| dependency_plan | compiled | miss | not-applicable | - | 0.094 | 22.381 | 0.090 | current-exact-nominal-result-allocation-shape-census | open-candidate | `compiled-iterator-control` |
| dependency_wave_validation | compiled | miss | not-applicable | - | 1.244 | 296.190 | 1.240 | current-exact-validated-job-file-entry-refresh | closed-rejected-candidate | `compiled-concurrency` |
| distance_field | compiled | miss | not-applicable | - | 0.104 | 7.939 | 0.090 | current-exact-post-truthiness-cast-numeric-next | closed-rejected-candidate | `compiled-float-numeric` |
| document_audit | compiled | miss | not-applicable | - | 0.088 | 22.564 | 0.084 | current-exact-nominal-result-allocation-shape-census | open-candidate | `compiled-iterator-control` |
| fasta_generation | compiled | miss | not-applicable | - | 0.126 | 9.545 | 0.112 | current-exact-post-truthiness-cast-refresh | closed-no-shared-leaf | `compiled-byte-output` |
| fib | compiled | miss | not-applicable | - | 3.846 | 1.210 | 0.501 | current-exact-post-truthiness-cast-shared-semantics | target-guard | `compiled-target-guards` |
| fixed_width_128 | compiled | miss | not-applicable | - | 0.240 | 42.105 | 0.234 | current-exact | closed-rejected-candidate | `compiled-wide-numeric` |
| future_await_race | compiled | miss | not-applicable | - | 0.090 | 21.951 | 0.086 | current-exact-validated-job-file-entry-refresh | closed-rejected-candidate | `compiled-concurrency` |
| future_pipeline | compiled | miss | not-applicable | - | 0.382 | 70.741 | 0.376 | current-exact-validated-job-file-entry-refresh | closed-rejected-candidate | `compiled-concurrency` |
| i_before_e | compiled | miss | not-applicable | - | 0.102 | 1.586 | 0.034 | current-exact-manifest-normalization-refresh | closed-no-shared-leaf | `compiled-text-map` |
| inventory_reconciliation | compiled | miss | not-applicable | - | 0.214 | 24.045 | 0.205 | current-exact-manifest-normalization-refresh | closed-no-shared-leaf | `compiled-text-map` |
| json | compiled | meets | established | established-meet | 0.792 | 0.535 | 0.000 | current-exact-post-truthiness-cast-shared-semantics | target-guard | `compiled-target-guards` |
| k_nucleotide | compiled | miss | not-applicable | - | 3.164 | 43.944 | 3.088 | current-exact-manifest-normalization-refresh | closed-no-shared-leaf | `compiled-text-map` |
| lexical_rollup | compiled | miss | not-applicable | - | 0.116 | 21.481 | 0.110 | current-exact-nominal-result-allocation-shape-census | open-candidate | `compiled-iterator-control` |
| log_routing_redaction | compiled | miss | not-applicable | - | 0.124 | 22.143 | 0.118 | current-exact-post-truthiness-cast-refresh | closed-rejected-candidate | `compiled-regex` |
| mandelbrot | compiled | miss | not-applicable | - | 0.166 | 3.018 | 0.108 | current-exact-post-truthiness-cast-numeric-next | closed-rejected-candidate | `compiled-float-numeric` |
| manifest_normalization | compiled | miss | not-applicable | - | 0.216 | 40.000 | 0.210 | current-exact-manifest-normalization-refresh | closed-no-shared-leaf | `compiled-text-map` |
| matrixmultiply | compiled | meets | not-established | variance-sensitive-miss | 1.138 | 1.050 | 0.000 | current-exact-post-truthiness-cast-shared-semantics | closed-no-shared-leaf | `compiled-current-control` |
| monte_carlo_pi | compiled | meets | not-established | volatile-crossing | 0.212 | 1.003 | 0.000 | current-exact-post-truthiness-cast-numeric-next | closed-rejected-candidate | `compiled-float-numeric` |
| mutex_await_journal | compiled | miss | not-applicable | - | 0.684 | 171.000 | 0.680 | current-exact-validated-job-file-entry-refresh | closed-rejected-candidate | `compiled-concurrency` |
| mutex_ledger | compiled | miss | not-applicable | - | 0.826 | 196.667 | 0.822 | current-exact-validated-job-file-entry-refresh | closed-rejected-candidate | `compiled-concurrency` |
| mutex_work_queue | compiled | miss | not-applicable | - | 1.366 | 310.455 | 1.361 | current-exact-validated-job-file-entry-refresh | closed-rejected-candidate | `compiled-concurrency` |
| nbody | compiled | miss | not-applicable | - | 0.184 | 5.750 | 0.150 | current-exact-post-truthiness-cast-numeric-next | closed-rejected-candidate | `compiled-float-numeric` |
| option_result_config | compiled | miss | not-applicable | - | 0.248 | 72.941 | 0.244 | current-exact-nominal-result-allocation-shape-census | open-candidate | `compiled-iterator-control` |
| pidigits | compiled | miss | not-applicable | - | 1.254 | 1.087 | 0.039 | current-exact-post-truthiness-cast-refresh | closed-no-shared-leaf | `compiled-byte-output` |
| policy_record_dispatch | compiled | miss | not-applicable | - | 0.226 | 40.357 | 0.220 | current-exact-post-truthiness-cast-refresh | closed-rejected-candidate | `compiled-regex` |
| quicksort | compiled | meets | established | established-meet | 1.846 | 0.729 | 0.000 | current-exact-post-truthiness-cast-shared-semantics | target-guard | `compiled-target-guards` |
| rational_series | compiled | miss | not-applicable | - | 0.138 | 10.615 | 0.124 | current-exact | closed-rejected-candidate | `compiled-wide-numeric` |
| regex_set_audit | compiled | miss | not-applicable | - | 0.152 | 30.400 | 0.147 | current-exact-post-truthiness-cast-refresh | closed-rejected-candidate | `compiled-regex` |
| regex_stream_audit | compiled | miss | not-applicable | - | 0.112 | 18.983 | 0.106 | current-exact-post-truthiness-cast-refresh | closed-rejected-candidate | `compiled-regex` |
| regex_suffix_audit | compiled | miss | not-applicable | - | 0.122 | 21.786 | 0.116 | current-exact-post-truthiness-cast-refresh | closed-rejected-candidate | `compiled-regex` |
| reverse_complement | compiled | miss | not-applicable | - | 0.154 | 8.800 | 0.136 | current-exact-post-truthiness-cast-refresh | closed-no-shared-leaf | `compiled-byte-output` |
| rms_norm | compiled | miss | not-applicable | - | 0.096 | 9.320 | 0.085 | current-exact-post-truthiness-cast-numeric-next | closed-rejected-candidate | `compiled-float-numeric` |
| sudoku_masks | compiled | miss | not-applicable | - | 1.792 | 3.196 | 1.202 | current-exact-post-truthiness-cast-shared-semantics | closed-insufficient-breadth | `compiled-sudoku-quotient` |
| tapelang_alphabet | compiled | miss | not-applicable | - | 3.570 | 1.900 | 1.592 | current-exact-post-truthiness-cast-shared-semantics | closed-no-shared-leaf | `compiled-current-control` |
| unicode_scalar_pipeline | compiled | miss | not-applicable | - | 0.294 | 28.544 | 0.283 | current-exact-manifest-normalization-refresh | closed-no-shared-leaf | `compiled-text-map` |
| validated_job_pipeline | compiled | miss | not-applicable | - | 1.130 | 235.417 | 1.125 | current-exact-validated-job-file-entry-refresh | closed-rejected-candidate | `compiled-concurrency` |
| wide_integer_records | compiled | miss | not-applicable | - | 0.200 | 7.874 | 0.173 | current-exact | closed-rejected-candidate | `compiled-wide-numeric` |
| word_frequency | compiled | miss | not-applicable | - | 0.188 | 32.982 | 0.182 | current-exact-manifest-normalization-refresh | closed-no-shared-leaf | `compiled-text-map` |

## Cross-cohort stability evidence

### `base64/bytecode`

- Classification: `variance-sensitive-miss`; established guard: `not-established`; pooled limiting ratio: 1.127x ruby; cohort ratios: 1.048, 1.206, 1.136.
- Samples: 15 Able and 15 limiting-reference; Able source: `b4676ab1b4392ed4433d7a2ce57c7388907e4719494e6edce32728b071750108`; evidence stdlib tree: `6a412c872ee66752de7c4417a5eda99806a7631da3b880c55574c6a640b82d9b`.
- Rationale: The current snapshot narrowly meets, but two of three Ruby cohorts and the pooled limiting ratio miss the bytecode product target.
- Evidence: [2026-07-20-threshold-stability-variance.json](../../docs/perf-baselines/2026-07-20-threshold-stability-variance.json), [2026-07-20-threshold-stability-reconciliation.md](../../docs/perf-baselines/2026-07-20-threshold-stability-reconciliation.md).

### `json/bytecode`

- Classification: `established-meet`; established guard: `established`; pooled limiting ratio: 0.533x ruby; cohort ratios: 0.474, 0.593.
- Samples: 10 Able and 10 limiting-reference; Able source: `84a895d80fee86e71b65c3614dc71088ef36d5a469b2756162007ee49d45b62a`; evidence stdlib tree: `6a412c872ee66752de7c4417a5eda99806a7631da3b880c55574c6a640b82d9b`.
- Rationale: Both independent current-source cohorts meet against both interpreters, including the faster Ruby reference.
- Evidence: [2026-07-20-source-exact-guards-variance.json](../../docs/perf-baselines/2026-07-20-source-exact-guards-variance.json), [2026-07-20-source-exact-established-guard-refresh.md](../../docs/perf-baselines/2026-07-20-source-exact-established-guard-refresh.md).

### `pidigits/bytecode`

- Classification: `established-meet`; established guard: `established`; pooled limiting ratio: 0.591x python; cohort ratios: 0.563, 0.620.
- Samples: 10 Able and 10 limiting-reference; Able source: `236a74ef456b4a5ca3e33a743a0b3b8e8767db9cfa07dd619df870bff876d5cb`; evidence stdlib tree: `6a412c872ee66752de7c4417a5eda99806a7631da3b880c55574c6a640b82d9b`.
- Rationale: Both independent current-source cohorts meet against both interpreters, including the faster Python reference.
- Evidence: [2026-07-20-source-exact-guards-variance.json](../../docs/perf-baselines/2026-07-20-source-exact-guards-variance.json), [2026-07-20-source-exact-established-guard-refresh.md](../../docs/perf-baselines/2026-07-20-source-exact-established-guard-refresh.md).

### `binarytrees/compiled`

- Classification: `established-meet`; established guard: `established`; pooled limiting ratio: 0.955x go; cohort ratios: 0.934, 0.975.
- Samples: 10 Able and 10 limiting-reference; Able source: `d973598c0ca4e88fcbe96dc852fa64a935cd651b3b85be10121222e9694293d8`; evidence stdlib tree: `6a412c872ee66752de7c4417a5eda99806a7631da3b880c55574c6a640b82d9b`.
- Rationale: Both independent current-source cohorts meet the compiled target.
- Evidence: [2026-07-20-source-exact-guards-variance.json](../../docs/perf-baselines/2026-07-20-source-exact-guards-variance.json), [2026-07-20-source-exact-established-guard-refresh.md](../../docs/perf-baselines/2026-07-20-source-exact-established-guard-refresh.md).

### `json/compiled`

- Classification: `established-meet`; established guard: `established`; pooled limiting ratio: 0.561x go; cohort ratios: 0.588, 0.533.
- Samples: 10 Able and 10 limiting-reference; Able source: `84a895d80fee86e71b65c3614dc71088ef36d5a469b2756162007ee49d45b62a`; evidence stdlib tree: `6a412c872ee66752de7c4417a5eda99806a7631da3b880c55574c6a640b82d9b`.
- Rationale: Both independent current-source cohorts meet the compiled target.
- Evidence: [2026-07-20-source-exact-guards-variance.json](../../docs/perf-baselines/2026-07-20-source-exact-guards-variance.json), [2026-07-20-source-exact-established-guard-refresh.md](../../docs/perf-baselines/2026-07-20-source-exact-established-guard-refresh.md).

### `matrixmultiply/compiled`

- Classification: `variance-sensitive-miss`; established guard: `not-established`; pooled limiting ratio: 1.155x go; cohort ratios: 1.050, 1.268.
- Samples: 10 Able and 10 limiting-reference; Able source: `f4e1ae5575094ab725db5ee0adb81eb6783221844aac9ef99b8f820f56c18246`; evidence stdlib tree: `6a412c872ee66752de7c4417a5eda99806a7631da3b880c55574c6a640b82d9b`.
- Rationale: The current snapshot narrowly meets, but the independent second cohort and the pooled ratio miss the compiled target.
- Evidence: [2026-07-21-dynamic-i64-lazy-slot-threshold-stability.json](../../docs/perf-baselines/2026-07-21-dynamic-i64-lazy-slot-threshold-stability.json), [2026-07-21-dynamic-i64-lazy-slot-threshold-stability.md](../../docs/perf-baselines/2026-07-21-dynamic-i64-lazy-slot-threshold-stability.md).

### `monte_carlo_pi/compiled`

- Classification: `volatile-crossing`; established guard: `not-established`; pooled limiting ratio: 1.005x go; cohort ratios: 0.927, 1.158, 0.950.
- Samples: 15 Able and 15 limiting-reference; Able source: `9afb1620c5f41eed0b519e225cf7a7cc32e837ccfccd2e047d56175c1805aa0c`; evidence stdlib tree: `6a412c872ee66752de7c4417a5eda99806a7631da3b880c55574c6a640b82d9b`.
- Rationale: The pooled ratio meets, but one of three independent cohorts misses materially, so this snapshot crossing is not a durable guard.
- Evidence: [2026-07-20-threshold-stability-variance.json](../../docs/perf-baselines/2026-07-20-threshold-stability-variance.json), [2026-07-20-threshold-stability-reconciliation.md](../../docs/perf-baselines/2026-07-20-threshold-stability-reconciliation.md).

### `quicksort/compiled`

- Classification: `established-meet`; established guard: `established`; pooled limiting ratio: 0.740x go; cohort ratios: 0.757, 0.722.
- Samples: 10 Able and 10 limiting-reference; Able source: `f54986a03bce5d3fa5bb8d1dd342299c1b6a434a682572d353c46d5f66c1af92`; evidence stdlib tree: `6a412c872ee66752de7c4417a5eda99806a7631da3b880c55574c6a640b82d9b`.
- Rationale: Both independent current-source cohorts meet the compiled target.
- Evidence: [2026-07-20-source-exact-guards-variance.json](../../docs/perf-baselines/2026-07-20-source-exact-guards-variance.json), [2026-07-20-source-exact-established-guard-refresh.md](../../docs/perf-baselines/2026-07-20-source-exact-established-guard-refresh.md).

## Ownership and disposition evidence

### `compiled-target-guards`

- Disposition: `target-guard`; exact unlike-application breadth: 0; profile freshness: `current-exact-post-truthiness-cast-shared-semantics`; artifact identity: `truthiness-cast-target-guard-current-cohort`.
- Owner: Different protected owners: Binary Trees node allocation/GC, Fib direct recursion, JSON host decoding, and QuickSort recursion.
- Rationale: Binary Trees, JSON, and QuickSort are established target guards. Fib currently misses after a noisy prior meet but remains a source-exact control for every admitted compiled candidate.
- Evidence: [2026-07-20-source-equivalence-scorecard.md](../../docs/perf-baselines/2026-07-20-source-equivalence-scorecard.md), [2026-07-20-compiled-control-current-invalidation-gate.md](../../docs/perf-baselines/2026-07-20-compiled-control-current-invalidation-gate.md), [2026-07-20-current-two-cohort-scorecard-reconciliation.md](../../docs/perf-baselines/2026-07-20-current-two-cohort-scorecard-reconciliation.md), [2026-07-20-threshold-stability-reconciliation.md](../../docs/perf-baselines/2026-07-20-threshold-stability-reconciliation.md), [2026-07-20-source-exact-established-guard-refresh.md](../../docs/perf-baselines/2026-07-20-source-exact-established-guard-refresh.md), [2026-07-21-post-i64-threshold-stability.md](../../docs/perf-baselines/2026-07-21-post-i64-threshold-stability.md), [2026-07-21-dynamic-i64-lazy-slot-threshold-stability.md](../../docs/perf-baselines/2026-07-21-dynamic-i64-lazy-slot-threshold-stability.md), [2026-07-21-truthiness-cast-target-guard-refresh.md](../../docs/perf-baselines/2026-07-21-truthiness-cast-target-guard-refresh.md), [2026-07-21-truthiness-cast-target-guards-compiled.json](../../docs/perf-baselines/2026-07-21-truthiness-cast-target-guards-compiled.json), [2026-07-21-truthiness-cast-target-guards-go-reference.json](../../docs/perf-baselines/2026-07-21-truthiness-cast-target-guards-go-reference.json), [2026-07-21-truthiness-cast-target-guards-fib-c2-compiled.json](../../docs/perf-baselines/2026-07-21-truthiness-cast-target-guards-fib-c2-compiled.json), [2026-07-21-truthiness-cast-target-guards-fib-c2-go-reference.json](../../docs/perf-baselines/2026-07-21-truthiness-cast-target-guards-fib-c2-go-reference.json).

### `compiled-current-control`

- Disposition: `closed-no-shared-leaf`; exact unlike-application breadth: 0; profile freshness: `current-exact-post-truthiness-cast-shared-semantics`; artifact identity: `truthiness-cast-control-current-cohorts-and-reach`.
- Owner: Matrix primitive nested arithmetic and TapeLang flat dispatch remain separate generated bodies.
- Rationale: Fresh current binaries expose no compiler/runtime child shared across these control-flow programs.
- Evidence: [2026-07-20-compiled-control-current-invalidation-gate.md](../../docs/perf-baselines/2026-07-20-compiled-control-current-invalidation-gate.md), [2026-07-21-truthiness-cast-control-closure-refresh.md](../../docs/perf-baselines/2026-07-21-truthiness-cast-control-closure-refresh.md), [2026-07-21-truthiness-cast-control-closure-reach.json](../../docs/perf-baselines/2026-07-21-truthiness-cast-control-closure-reach.json), [2026-07-21-truthiness-cast-control-closures-compiled.json](../../docs/perf-baselines/2026-07-21-truthiness-cast-control-closures-compiled.json), [2026-07-21-truthiness-cast-control-closures-go-reference.json](../../docs/perf-baselines/2026-07-21-truthiness-cast-control-closures-go-reference.json), [2026-07-21-truthiness-cast-control-closures-matrix-c2-compiled.json](../../docs/perf-baselines/2026-07-21-truthiness-cast-control-closures-matrix-c2-compiled.json), [2026-07-21-truthiness-cast-control-closures-matrix-c2-go-reference.json](../../docs/perf-baselines/2026-07-21-truthiness-cast-control-closures-matrix-c2-go-reference.json), [2026-07-21-truthiness-cast-control-closures-tape-c2-compiled.json](../../docs/perf-baselines/2026-07-21-truthiness-cast-control-closures-tape-c2-compiled.json), [2026-07-21-truthiness-cast-control-closures-tape-c2-go-reference.json](../../docs/perf-baselines/2026-07-21-truthiness-cast-control-closures-tape-c2-go-reference.json).

### `compiled-sudoku-quotient`

- Disposition: `closed-insufficient-breadth`; exact unlike-application breadth: 1; profile freshness: `current-exact-post-truthiness-cast-shared-semantics`; artifact identity: `truthiness-cast-sudoku-current-cohorts-generated-scan-and-two-profile-merge`.
- Owner: find_best_empty search plus square_index; signed Euclidean division is 12.53% cumulative in two current merged profiles and material only in Sudoku.
- Rationale: All nine generated application bodies avoid corrected truthiness/cast helpers. The exact quotient helper receives zero samples in three unlike controls, and even perfect removal leaves Sudoku 2.70x short of target.
- Evidence: [2026-07-21-truthiness-cast-sudoku-quotient-closure-refresh.md](../../docs/perf-baselines/2026-07-21-truthiness-cast-sudoku-quotient-closure-refresh.md), [2026-07-21-truthiness-cast-sudoku-quotient-closure-reach.json](../../docs/perf-baselines/2026-07-21-truthiness-cast-sudoku-quotient-closure-reach.json), [2026-07-21-truthiness-cast-sudoku-quotient-compiled.json](../../docs/perf-baselines/2026-07-21-truthiness-cast-sudoku-quotient-compiled.json), [2026-07-21-truthiness-cast-sudoku-quotient-c2-compiled.json](../../docs/perf-baselines/2026-07-21-truthiness-cast-sudoku-quotient-c2-compiled.json), [2026-07-21-truthiness-cast-sudoku-quotient-go-reference.json](../../docs/perf-baselines/2026-07-21-truthiness-cast-sudoku-quotient-go-reference.json), [2026-07-21-truthiness-cast-sudoku-quotient-c2-go-reference.json](../../docs/perf-baselines/2026-07-21-truthiness-cast-sudoku-quotient-c2-go-reference.json), [2026-07-20-compiled-quotient-only-ownership-census.md](../../docs/perf-baselines/2026-07-20-compiled-quotient-only-ownership-census.md).

### `compiled-float-numeric`

- Disposition: `closed-rejected-candidate`; exact unlike-application breadth: 3; profile freshness: `current-exact-post-truthiness-cast-numeric-next`; artifact identity: `frozen-current-repeated-timing-and-exact-reach-cohorts`.
- Owner: Float geometry/regions and normalized raw-float allocation recur, while Mandelbrot remains a distinct fused loop owner.
- Rationale: Fresh repeated timing and exact generated-runtime telemetry preserve the prior ownership result. All five rows have zero truthiness/cast bridge reach, so the closed raw-float lane/carrier evidence remains causal.
- Evidence: [2026-07-21-truthiness-cast-numeric-next-closure-refresh.md](../../docs/perf-baselines/2026-07-21-truthiness-cast-numeric-next-closure-refresh.md), [2026-07-21-truthiness-cast-numeric-next-closure-reach.json](../../docs/perf-baselines/2026-07-21-truthiness-cast-numeric-next-closure-reach.json), [2026-07-20-cross-mode-numeric-wide-profile-gate.md](../../docs/perf-baselines/2026-07-20-cross-mode-numeric-wide-profile-gate.md), [2026-07-20-threshold-stability-reconciliation.md](../../docs/perf-baselines/2026-07-20-threshold-stability-reconciliation.md).

### `compiled-wide-numeric`

- Disposition: `closed-rejected-candidate`; exact unlike-application breadth: 3; profile freshness: `current-exact`; artifact identity: `preserved-binary-matched`.
- Owner: Fresh post-semantic timing preserves the package-environment publication owner through SwapEnv/sync/atomic.StorePointer; exact compiled telemetry shows zero changed truthiness or cast-bridge reach in all three programs.
- Rationale: The semantic correction does not reach these mains, while general execution-context and package-linkage alternatives already regressed unrelated N-Body and K-Nucleotide wall time. Nominal specialization remains prohibited.
- Evidence: [2026-07-21-truthiness-cast-wide-text-closure-refresh.md](../../docs/perf-baselines/2026-07-21-truthiness-cast-wide-text-closure-refresh.md), [2026-07-21-truthiness-cast-wide-text-closure-reach.json](../../docs/perf-baselines/2026-07-21-truthiness-cast-wide-text-closure-reach.json), [2026-07-20-wide-integer-records-profile-gate.md](../../docs/perf-baselines/2026-07-20-wide-integer-records-profile-gate.md), [2026-07-20-cross-mode-numeric-wide-profile-gate.md](../../docs/perf-baselines/2026-07-20-cross-mode-numeric-wide-profile-gate.md), [2026-07-20-compiled-quotient-only-ownership-census.md](../../docs/perf-baselines/2026-07-20-compiled-quotient-only-ownership-census.md).

### `compiled-byte-output`

- Disposition: `closed-no-shared-leaf`; exact unlike-application breadth: 0; profile freshness: `current-exact-post-truthiness-cast-refresh`; artifact identity: `frozen-current-verifier-backed-timing-and-exact-reach`.
- Owner: Fresh post-semantic timing preserves the split among Base64 host codec/MD5 work, direct FASTA arithmetic, Go BigInt kernels, and Reverse transform/copy/GC work; exact telemetry shows zero changed truthiness or cast-bridge reach in all four programs.
- Rationale: The semantic correction does not reach these compiled mains. The retained write_all and generic u8 changes remain causal, and no compiler-controlled descendant is shared by three unlike applications.
- Evidence: [2026-07-21-truthiness-cast-byte-regex-closure-refresh.md](../../docs/perf-baselines/2026-07-21-truthiness-cast-byte-regex-closure-refresh.md), [2026-07-21-truthiness-cast-byte-regex-closure-reach.json](../../docs/perf-baselines/2026-07-21-truthiness-cast-byte-regex-closure-reach.json), [2026-07-20-compiled-byte-output-current-profile-gate.md](../../docs/perf-baselines/2026-07-20-compiled-byte-output-current-profile-gate.md), [2026-07-18-fasta-write-all-gate.md](../../docs/perf-baselines/2026-07-18-fasta-write-all-gate.md), [2026-07-18-cross-mode-array-capacity-growth-gate.md](../../docs/perf-baselines/2026-07-18-cross-mode-array-capacity-growth-gate.md).

### `compiled-text-map`

- Disposition: `closed-no-shared-leaf`; exact unlike-application breadth: 0; profile freshness: `current-exact-manifest-normalization-refresh`; artifact identity: `frozen-current-verifier-backed-timing-plus-manifest-main-profile`.
- Owner: Post-index text/map work still splits among integer conversion/equality, UTF-8/String, boxing, and allocation/GC descendants. Manifest Normalization adds a material String.to_builtin path shared cross-group with K-Nucleotide and Policy Record Dispatch.
- Rationale: The newly three-application indexed byte-conversion candidate was measured and reverted: it improved K-Nucleotide about 1.5% but regressed Policy about 2.9% and Manifest about 1%. No retained compiler-controlled leaf passes the broad guard.
- Evidence: [2026-07-21-truthiness-cast-text-byte-next-closure-refresh.md](../../docs/perf-baselines/2026-07-21-truthiness-cast-text-byte-next-closure-refresh.md), [2026-07-21-truthiness-cast-text-byte-next-closure-reach.json](../../docs/perf-baselines/2026-07-21-truthiness-cast-text-byte-next-closure-reach.json), [2026-07-20-post-hash-index-profile-reconciliation.md](../../docs/perf-baselines/2026-07-20-post-hash-index-profile-reconciliation.md), [2026-07-20-generic-hash-map-index-gate.md](../../docs/perf-baselines/2026-07-20-generic-hash-map-index-gate.md), [2026-07-21-manifest-normalization-application-gate.md](../../docs/perf-baselines/2026-07-21-manifest-normalization-application-gate.md), [2026-07-21-manifest-normalization-compiled-main-profile.txt](../../docs/perf-baselines/2026-07-21-manifest-normalization-compiled-main-profile.txt).

### `compiled-regex`

- Disposition: `closed-rejected-candidate`; exact unlike-application breadth: 6; profile freshness: `current-exact-post-truthiness-cast-refresh`; artifact identity: `frozen-current-verifier-backed-timing-and-exact-reach`.
- Owner: Fresh post-semantic timing preserves canonical NFA closure, move, and thread management across the related workload families. Exact telemetry finds generated truthiness only in Policy Record Dispatch, 2,048 times, with zero casts everywhere else.
- Rationale: The changed semantic path is single-application and cannot invalidate the prior six-row NFA ownership. Closure scratch, capture templates, and primitive thread carriers are retained while arenas, state indexes, character specialization, and carrier/call alternatives already failed broad gates.
- Evidence: [2026-07-21-truthiness-cast-regex-concurrency-closure-refresh.md](../../docs/perf-baselines/2026-07-21-truthiness-cast-regex-concurrency-closure-refresh.md), [2026-07-21-truthiness-cast-regex-concurrency-closure-reach.json](../../docs/perf-baselines/2026-07-21-truthiness-cast-regex-concurrency-closure-reach.json), [2026-07-20-regex-three-api-current-profile-gate.md](../../docs/perf-baselines/2026-07-20-regex-three-api-current-profile-gate.md), [2026-07-20-log-routing-redaction-profile-gate.md](../../docs/perf-baselines/2026-07-20-log-routing-redaction-profile-gate.md), [2026-07-20-config-validation-extraction-profile-gate.md](../../docs/perf-baselines/2026-07-20-config-validation-extraction-profile-gate.md).

### `compiled-concurrency`

- Disposition: `closed-rejected-candidate`; exact unlike-application breadth: 12; profile freshness: `current-exact-validated-job-file-entry-refresh`; artifact identity: `frozen-current-generated-binaries-plus-validated-job-main-profiles`.
- Owner: Goroutine identity through bridge.currentGID/runtime.Stack remains the exact repeated compiled wall; newly reached truthiness/cast bridges are CPU-immaterial.
- Rationale: The evolved file-driven pipeline is 235.417x Go in its promoted cohort. Three verified current main profiles put bridge.currentGID at 94.16% cumulative, reproducing the exact owner already seen at 74.07%-96.82% in unlike concurrency programs. The existing fixed-context candidate still fails broad guards and does not remove this owner outside a more tightly scoped design.
- Evidence: [2026-07-21-truthiness-cast-architecture-closure-refresh.md](../../docs/perf-baselines/2026-07-21-truthiness-cast-architecture-closure-refresh.md), [2026-07-21-truthiness-cast-architecture-closure-compiled-reach.json](../../docs/perf-baselines/2026-07-21-truthiness-cast-architecture-closure-compiled-reach.json), [2026-07-21-compiled-concurrency-fourteen-application-refresh.md](../../docs/perf-baselines/2026-07-21-compiled-concurrency-fourteen-application-refresh.md), [2026-07-20-mutex-work-queue-profile-gate.md](../../docs/perf-baselines/2026-07-20-mutex-work-queue-profile-gate.md), [2026-07-17-compiled-equal-cpu-concurrency-reconciliation.md](../../docs/perf-baselines/2026-07-17-compiled-equal-cpu-concurrency-reconciliation.md), [2026-07-18-post-string-full-scorecard-reconciliation.md](../../docs/perf-baselines/2026-07-18-post-string-full-scorecard-reconciliation.md), [2026-07-20-feature-interaction-application-gate.md](../../docs/perf-baselines/2026-07-20-feature-interaction-application-gate.md), [2026-07-20-dependency-wave-application-gate.md](../../docs/perf-baselines/2026-07-20-dependency-wave-application-gate.md), [2026-07-20-concurrent-event-routing-application-gate.md](../../docs/perf-baselines/2026-07-20-concurrent-event-routing-application-gate.md), [2026-07-21-concurrent-document-pipeline-application-gate.md](../../docs/perf-baselines/2026-07-21-concurrent-document-pipeline-application-gate.md), [2026-07-21-concurrent-document-pipeline-promotion-compiled.md](../../docs/perf-baselines/2026-07-21-concurrent-document-pipeline-promotion-compiled.md), [2026-07-21-concurrent-document-pipeline-compiled-main-profile.txt](../../docs/perf-baselines/2026-07-21-concurrent-document-pipeline-compiled-main-profile.txt), [2026-07-21-validated-job-file-entry-application-gate.md](../../docs/perf-baselines/2026-07-21-validated-job-file-entry-application-gate.md), [2026-07-21-validated-job-file-entry-comparison-d.json](../../docs/perf-baselines/2026-07-21-validated-job-file-entry-comparison-d.json), [2026-07-21-validated-job-file-entry-compiled-main-profile.txt](../../docs/perf-baselines/2026-07-21-validated-job-file-entry-compiled-main-profile.txt).

### `compiled-iterator-control`

- Disposition: `open-candidate`; exact unlike-application breadth: 4; profile freshness: `current-exact-nominal-result-allocation-shape-census`; artifact identity: `four-frozen-current-binaries-with-cpu-exact-allocation-and-repeated-phase-stats`.
- Owner: Static generic-union fast dispatch boxes a 64-byte NativeBoundMethodValue before invoking an already-known generated method entry. The exact allocation is material in Binary Event Log, Option/Result Config, and Manifest Normalization and has CPU reach in those three plus Policy Record Dispatch.
- Rationale: Fifty verified CPU profiles place this exact helper at 11.83%-58.97% cumulative across four unlike applications. Exact line-level allocation profiles attribute 225,280, 147,456, and 20,480 escaping bound-method boxes to the first three, satisfying the generic admission bar without a nominal-type special case.
- Evidence: [2026-07-22-compiled-nominal-result-allocation-shape-census.md](../../docs/perf-baselines/2026-07-22-compiled-nominal-result-allocation-shape-census.md), [2026-07-22-compiled-nominal-result-allocation-shape-census.json](../../docs/perf-baselines/2026-07-22-compiled-nominal-result-allocation-shape-census.json), [2026-07-22-binary-event-log-application-gate.md](../../docs/perf-baselines/2026-07-22-binary-event-log-application-gate.md), [2026-07-21-truthiness-cast-near-path-closure-refresh.md](../../docs/perf-baselines/2026-07-21-truthiness-cast-near-path-closure-refresh.md), [2026-07-21-truthiness-cast-near-path-closure-reach.json](../../docs/perf-baselines/2026-07-21-truthiness-cast-near-path-closure-reach.json), [2026-07-21-compiled-iterator-control-dynamic-i64-lazy-slot-gate.md](../../docs/perf-baselines/2026-07-21-compiled-iterator-control-dynamic-i64-lazy-slot-gate.md), [2026-07-20-current-iterator-control-profile-refresh.md](../../docs/perf-baselines/2026-07-20-current-iterator-control-profile-refresh.md), [2026-07-18-cross-mode-array-capacity-growth-gate.md](../../docs/perf-baselines/2026-07-18-cross-mode-array-capacity-growth-gate.md).

### `bytecode-target-guards`

- Disposition: `target-guard`; exact unlike-application breadth: 0; profile freshness: `current-exact-post-truthiness-cast-shared-semantics`; artifact identity: `truthiness-cast-target-guard-current-cohort`.
- Owner: JSON host decoding and PiDigits BigInt/native output are protected target-meeting rows.
- Rationale: Every admitted VM candidate must preserve these current target meets.
- Evidence: [2026-07-20-source-equivalence-scorecard.md](../../docs/perf-baselines/2026-07-20-source-equivalence-scorecard.md), [2026-07-20-source-exact-established-guard-refresh.md](../../docs/perf-baselines/2026-07-20-source-exact-established-guard-refresh.md), [2026-07-21-truthiness-cast-target-guard-refresh.md](../../docs/perf-baselines/2026-07-21-truthiness-cast-target-guard-refresh.md), [2026-07-21-truthiness-cast-target-guards-bytecode.json](../../docs/perf-baselines/2026-07-21-truthiness-cast-target-guards-bytecode.json), [2026-07-21-truthiness-cast-target-guards-interpreter-reference.json](../../docs/perf-baselines/2026-07-21-truthiness-cast-target-guards-interpreter-reference.json).

### `bytecode-float-numeric`

- Disposition: `closed-rejected-candidate`; exact unlike-application breadth: 3; profile freshness: `current-exact-post-truthiness-cast-near-path`; artifact identity: `frozen-current-repeated-timing-and-exact-reach-cohorts`.
- Owner: Validated float-slot reads and retained raw float arithmetic recur across three unlike scalar programs; Matrix dot-loop/Array, Monte Carlo integer recurrence/cast, and Reverse byte/Array work separate the concrete consumers.
- Rationale: Fresh repeated timing and two exact main-only censuses per selected row preserve the six-application ownership result. Every row has zero changed Error-truthiness fallback and zero explicit-cast reach, so the closed float carrier and dispatcher evidence remains causal.
- Evidence: [2026-07-21-truthiness-cast-near-path-closure-refresh.md](../../docs/perf-baselines/2026-07-21-truthiness-cast-near-path-closure-refresh.md), [2026-07-21-truthiness-cast-near-path-closure-reach.json](../../docs/perf-baselines/2026-07-21-truthiness-cast-near-path-closure-reach.json), [2026-07-21-bytecode-float-numeric-six-application-refresh.md](../../docs/perf-baselines/2026-07-21-bytecode-float-numeric-six-application-refresh.md), [2026-07-20-cross-mode-numeric-wide-profile-gate.md](../../docs/perf-baselines/2026-07-20-cross-mode-numeric-wide-profile-gate.md).

### `bytecode-wide-numeric`

- Disposition: `closed-rejected-candidate`; exact unlike-application breadth: 4; profile freshness: `current-exact-post-truthiness-cast-numeric-next`; artifact identity: `frozen-current-repeated-timing-reach-and-main-profile-cohorts`.
- Owner: Raw integer extraction recurs across the three wide programs and Reverse Complement, but its consumers split among checked wide operations, casts, comparisons, parsing, bitwise work, member calls, and byte loops.
- Rationale: Two exact censuses per row find zero changed Error fallback; Rational and Wide reach successful explicit casts, but four current profiles put zero flat CPU in the catchable wrapper. Their raw conversion and target-canonicalization descendants remain two-row owners whose general bypass already reversed sign across broad wall cohorts.
- Evidence: [2026-07-21-truthiness-cast-numeric-next-closure-refresh.md](../../docs/perf-baselines/2026-07-21-truthiness-cast-numeric-next-closure-refresh.md), [2026-07-21-truthiness-cast-numeric-next-closure-reach.json](../../docs/perf-baselines/2026-07-21-truthiness-cast-numeric-next-closure-reach.json), [2026-07-21-bytecode-wide-numeric-five-application-refresh.md](../../docs/perf-baselines/2026-07-21-bytecode-wide-numeric-five-application-refresh.md), [2026-07-20-wide-integer-records-profile-gate.md](../../docs/perf-baselines/2026-07-20-wide-integer-records-profile-gate.md), [2026-07-20-cross-mode-numeric-wide-profile-gate.md](../../docs/perf-baselines/2026-07-20-cross-mode-numeric-wide-profile-gate.md).

### `bytecode-byte-output`

- Disposition: `closed-no-shared-leaf`; exact unlike-application breadth: 3; profile freshness: `current-exact-post-truthiness-cast-refresh`; artifact identity: `frozen-current-verifier-backed-timing-and-exact-reach`.
- Owner: Fresh post-semantic timing preserves the residual split among Base64 host kernels, Reverse Array/extern work, and FASTA call/arithmetic work. Two exact censuses per row find only four primitive truthiness checks in Reverse and zero changed paths everywhere.
- Rationale: The retained generic monomorphic u8 path remains causal, while the semantic correction has zero material reach and exposes no second shared descendant across all three applications.
- Evidence: [2026-07-21-truthiness-cast-text-byte-next-closure-refresh.md](../../docs/perf-baselines/2026-07-21-truthiness-cast-text-byte-next-closure-refresh.md), [2026-07-21-truthiness-cast-text-byte-next-closure-reach.json](../../docs/perf-baselines/2026-07-21-truthiness-cast-text-byte-next-closure-reach.json), [2026-07-20-bytecode-byte-output-current-profile-gate.md](../../docs/perf-baselines/2026-07-20-bytecode-byte-output-current-profile-gate.md), [2026-07-18-fasta-write-all-gate.md](../../docs/perf-baselines/2026-07-18-fasta-write-all-gate.md), [2026-07-18-cross-mode-array-capacity-growth-gate.md](../../docs/perf-baselines/2026-07-18-cross-mode-array-capacity-growth-gate.md), [2026-07-20-threshold-stability-reconciliation.md](../../docs/perf-baselines/2026-07-20-threshold-stability-reconciliation.md).

### `bytecode-text-map`

- Disposition: `closed-rejected-candidate`; exact unlike-application breadth: 6; profile freshness: `current-exact-binary-event-log-main-profile`; artifact identity: `frozen-current-verifier-backed-timing-plus-binary-event-runtime-profile`.
- Owner: Binary Event Log repeats raw-integer extraction, call/member dispatch, type matching, Go map access, and allocation/GC ancestry already present in text/map workloads; its concrete record parsing and Result consumers remain distinct.
- Rationale: Three verified profiles put bytecodeRawIntegerValueInfo at 2.20% flat, extending an exact leaf already material in five unlike applications. That generic family has repeatedly failed broad guards; type-match and map descendants remain cumulative parents or workload-specific children, so no bytecode candidate is admitted.
- Evidence: [2026-07-22-binary-event-log-application-gate.md](../../docs/perf-baselines/2026-07-22-binary-event-log-application-gate.md), [2026-07-21-bytecode-cross-corpus-exact-leaf-selection.md](../../docs/perf-baselines/2026-07-21-bytecode-cross-corpus-exact-leaf-selection.md), [2026-07-21-truthiness-cast-wide-text-closure-refresh.md](../../docs/perf-baselines/2026-07-21-truthiness-cast-wide-text-closure-refresh.md), [2026-07-21-truthiness-cast-wide-text-closure-reach.json](../../docs/perf-baselines/2026-07-21-truthiness-cast-wide-text-closure-reach.json), [2026-07-21-bytecode-text-map-five-application-refresh.md](../../docs/perf-baselines/2026-07-21-bytecode-text-map-five-application-refresh.md), [2026-07-20-post-hash-index-profile-reconciliation.md](../../docs/perf-baselines/2026-07-20-post-hash-index-profile-reconciliation.md), [2026-07-20-generic-hash-map-index-gate.md](../../docs/perf-baselines/2026-07-20-generic-hash-map-index-gate.md), [2026-07-21-manifest-normalization-application-gate.md](../../docs/perf-baselines/2026-07-21-manifest-normalization-application-gate.md), [2026-07-21-manifest-normalization-bytecode-runtime-profile.json](../../docs/perf-baselines/2026-07-21-manifest-normalization-bytecode-runtime-profile.json).

### `bytecode-regex`

- Disposition: `closed-rejected-candidate`; exact unlike-application breadth: 6; profile freshness: `current-exact-post-truthiness-cast-refresh`; artifact identity: `frozen-current-verifier-backed-timing-and-exact-reach`.
- Owner: Fresh post-semantic timing and two exact censuses per row find 208,060-1,565,212 primitive truthiness checks but zero changed Error fallbacks or casts. Prior profiles still place the shared work in canonical NFA Array-slot cache/read and small named-struct field checks.
- Rationale: The semantic correction has zero changed-path reach. The current direct Array cache is allocation-free and its remaining version checks preserve invalidation, while general field-map and NFA carrier/index/arena alternatives already failed broad wall-time gates.
- Evidence: [2026-07-21-truthiness-cast-byte-regex-closure-refresh.md](../../docs/perf-baselines/2026-07-21-truthiness-cast-byte-regex-closure-refresh.md), [2026-07-21-truthiness-cast-byte-regex-closure-reach.json](../../docs/perf-baselines/2026-07-21-truthiness-cast-byte-regex-closure-reach.json), [2026-07-21-bytecode-regex-seven-application-refresh.md](../../docs/perf-baselines/2026-07-21-bytecode-regex-seven-application-refresh.md), [2026-07-21-bytecode-secondary-architecture-selection.md](../../docs/perf-baselines/2026-07-21-bytecode-secondary-architecture-selection.md), [2026-07-20-regex-three-api-current-profile-gate.md](../../docs/perf-baselines/2026-07-20-regex-three-api-current-profile-gate.md).

### `bytecode-concurrency`

- Disposition: `closed-no-shared-leaf`; exact unlike-application breadth: 12; profile freshness: `current-exact-validated-job-file-entry-refresh`; artifact identity: `frozen-current-verifier-backed-timing-plus-validated-job-profile`.
- Owner: Fresh post-semantic timing and two exact censuses per row find zero changed Error fallbacks or casts. Prior profiles still place recurring samples in RWMutex-backed environment/cache operations whose Able callers split across unrelated lookup, revision, runtime-data, parent, definition, alias, and method-cache stores.
- Rationale: The evolved pipeline is 11.250x Python and 6.656x Ruby in its promoted cohort. Three current profiles leave the largest named VM leaves at only 2.13% flat and split remaining work among parser/cgo, allocation/GC, member lookup, raw integers, and Go maps. No new concrete child is material and shared by three unlike families.
- Evidence: [2026-07-21-truthiness-cast-regex-concurrency-closure-refresh.md](../../docs/perf-baselines/2026-07-21-truthiness-cast-regex-concurrency-closure-refresh.md), [2026-07-21-truthiness-cast-regex-concurrency-closure-reach.json](../../docs/perf-baselines/2026-07-21-truthiness-cast-regex-concurrency-closure-reach.json), [2026-07-21-bytecode-concurrency-fourteen-application-refresh.md](../../docs/perf-baselines/2026-07-21-bytecode-concurrency-fourteen-application-refresh.md), [2026-07-20-mutex-work-queue-profile-gate.md](../../docs/perf-baselines/2026-07-20-mutex-work-queue-profile-gate.md), [2026-07-20-bytecode-concurrency-current-profile-gate.md](../../docs/perf-baselines/2026-07-20-bytecode-concurrency-current-profile-gate.md), [2026-07-21-concurrent-document-pipeline-application-gate.md](../../docs/perf-baselines/2026-07-21-concurrent-document-pipeline-application-gate.md), [2026-07-21-concurrent-document-pipeline-promotion-bytecode.md](../../docs/perf-baselines/2026-07-21-concurrent-document-pipeline-promotion-bytecode.md), [2026-07-21-validated-job-file-entry-application-gate.md](../../docs/perf-baselines/2026-07-21-validated-job-file-entry-application-gate.md), [2026-07-21-validated-job-file-entry-comparison-d.json](../../docs/perf-baselines/2026-07-21-validated-job-file-entry-comparison-d.json), [2026-07-21-validated-job-file-entry-bytecode-profile.txt](../../docs/perf-baselines/2026-07-21-validated-job-file-entry-bytecode-profile.txt).

### `bytecode-iterator-control`

- Disposition: `closed-no-shared-leaf`; exact unlike-application breadth: 5; profile freshness: `current-exact-post-truthiness-cast-shared-semantics`; artifact identity: `truthiness-cast-control-current-cohorts-and-reach`.
- Owner: Current Array copying, graph/Queue, iterator/text, lexical aggregation, and union/default profiles share only completed cached-member, active-lookup, return/frame, and raw-integer leaves; Go map/hash callers split across unrelated semantic stores.
- Rationale: Two high-sample CPU and two exact one-main allocation processes per selected row, plus Mandelbrot and Word Frequency controls, expose no new concrete VM operation shared by three unlike iterator/control families.
- Evidence: [2026-07-21-bytecode-iterator-control-seven-application-refresh.md](../../docs/perf-baselines/2026-07-21-bytecode-iterator-control-seven-application-refresh.md), [2026-07-20-current-iterator-control-profile-refresh.md](../../docs/perf-baselines/2026-07-20-current-iterator-control-profile-refresh.md), [2026-07-19-bytecode-text-map-graph-profile-gate.md](../../docs/perf-baselines/2026-07-19-bytecode-text-map-graph-profile-gate.md), [2026-07-19-bytecode-inline-return-profile-reconciliation.md](../../docs/perf-baselines/2026-07-19-bytecode-inline-return-profile-reconciliation.md), [2026-07-18-cross-mode-array-capacity-growth-gate.md](../../docs/perf-baselines/2026-07-18-cross-mode-array-capacity-growth-gate.md), [2026-07-21-truthiness-cast-control-closure-refresh.md](../../docs/perf-baselines/2026-07-21-truthiness-cast-control-closure-refresh.md), [2026-07-21-truthiness-cast-control-closure-reach.json](../../docs/perf-baselines/2026-07-21-truthiness-cast-control-closure-reach.json), [2026-07-21-truthiness-cast-control-closures-bytecode.json](../../docs/perf-baselines/2026-07-21-truthiness-cast-control-closures-bytecode.json), [2026-07-21-truthiness-cast-control-closures-interpreter-reference.json](../../docs/perf-baselines/2026-07-21-truthiness-cast-control-closures-interpreter-reference.json).
