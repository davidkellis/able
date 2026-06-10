# Cross-mode performance frontier

Generated from `2026-07-21T12:04:45.151360Z` scorecard evidence. This ledger joins all 83 reviewed rows; it excludes unselected status-only rows.

## Outcome

- Selected rows: 45 compiled + 38 bytecode = 83.
- Product target: 8 meet, 75 miss.
- Established cross-cohort guards: 6 (4 compiled + 2 bytecode); 2 snapshot meets are not established.
- Aggregate time above the per-row 95%-of-reference budget: 135.228 seconds.
- Unclosed groups: 0.

## Actionable groups

| Rank | Group | Action | Rows | Misses | Excess s | Max ratio | Freshness |
| ---: | --- | --- | ---: | ---: | ---: | ---: | --- |
| - | None | - | - | - | - | - | - |

## Complete selected-row ledger

`Excess s` is Able wall time beyond the fastest applicable reference multiplied by the allowed ratio (1 / 0.95). `Established` is a separate cross-cohort candidate-admission guard and never rewrites snapshot status.

| Benchmark | Mode | Snapshot | Established | Stability | Able s | Worst ratio | Excess s | Freshness | Disposition | Group |
| --- | --- | --- | --- | --- | ---: | ---: | ---: | --- | --- | --- |
| array_slice_window | bytecode | miss | not-applicable | - | 0.648 | 22.041 | 0.617 | post-active-lookup-mixed | closed-no-shared-leaf | `bytecode-iterator-control` |
| await_channel_mux | bytecode | miss | not-applicable | - | 0.208 | 1.726 | 0.081 | current-exact-mixed-targeted | closed-no-shared-leaf | `bytecode-concurrency` |
| base64 | bytecode | meets | not-established | variance-sensitive-miss | 2.806 | 1.048 | 0.000 | current-exact | closed-no-shared-leaf | `bytecode-byte-output` |
| channel_rollup | bytecode | miss | not-applicable | - | 0.446 | 10.825 | 0.403 | current-exact-mixed-targeted | closed-no-shared-leaf | `bytecode-concurrency` |
| concurrent_event_routing | bytecode | miss | not-applicable | - | 3.032 | 102.088 | 3.001 | current-exact-mixed-targeted | closed-no-shared-leaf | `bytecode-concurrency` |
| concurrent_text_index | bytecode | miss | not-applicable | - | 0.634 | 7.512 | 0.545 | current-exact-mixed-targeted | closed-no-shared-leaf | `bytecode-concurrency` |
| config_validation_extraction | bytecode | miss | not-applicable | - | 1.386 | 56.341 | 1.360 | current-exact | closed-rejected-candidate | `bytecode-regex` |
| dependency_plan | bytecode | miss | not-applicable | - | 0.498 | 30.741 | 0.481 | post-active-lookup-mixed | closed-no-shared-leaf | `bytecode-iterator-control` |
| dependency_wave_validation | bytecode | miss | not-applicable | - | 0.420 | 8.750 | 0.369 | current-exact-mixed-targeted | closed-no-shared-leaf | `bytecode-concurrency` |
| distance_field | bytecode | miss | not-applicable | - | 5.748 | 13.718 | 5.307 | current-exact | closed-rejected-candidate | `bytecode-float-numeric` |
| document_audit | bytecode | miss | not-applicable | - | 0.274 | 19.571 | 0.259 | post-active-lookup-mixed | closed-no-shared-leaf | `bytecode-iterator-control` |
| fasta_generation | bytecode | miss | not-applicable | - | 1.828 | 9.036 | 1.615 | current-exact | closed-no-shared-leaf | `bytecode-byte-output` |
| fixed_width_128 | bytecode | miss | not-applicable | - | 7.988 | 21.873 | 7.604 | current-exact | closed-rejected-candidate | `bytecode-wide-numeric` |
| future_await_race | bytecode | miss | not-applicable | - | 0.136 | 4.012 | 0.100 | current-exact-mixed-targeted | closed-no-shared-leaf | `bytecode-concurrency` |
| future_pipeline | bytecode | miss | not-applicable | - | 0.406 | 6.612 | 0.341 | current-exact-mixed-targeted | closed-no-shared-leaf | `bytecode-concurrency` |
| i_before_e | bytecode | miss | not-applicable | - | 0.526 | 6.315 | 0.438 | current-exact-post-hash-index | closed-no-shared-leaf | `bytecode-text-map` |
| inventory_reconciliation | bytecode | miss | not-applicable | - | 2.536 | 37.294 | 2.464 | current-exact-post-hash-index | closed-no-shared-leaf | `bytecode-text-map` |
| json | bytecode | meets | established | established-meet | 0.816 | 0.474 | 0.000 | current-exact | target-guard | `bytecode-target-guards` |
| k_nucleotide | bytecode | miss | not-applicable | - | 45.164 | 35.881 | 43.839 | current-exact-post-hash-index | closed-no-shared-leaf | `bytecode-text-map` |
| lexical_rollup | bytecode | miss | not-applicable | - | 0.398 | 23.275 | 0.380 | post-active-lookup-mixed | closed-no-shared-leaf | `bytecode-iterator-control` |
| log_routing_redaction | bytecode | miss | not-applicable | - | 3.016 | 124.628 | 2.991 | current-exact | closed-rejected-candidate | `bytecode-regex` |
| mandelbrot | bytecode | miss | not-applicable | - | 6.324 | 5.367 | 5.084 | current-exact | closed-rejected-candidate | `bytecode-float-numeric` |
| monte_carlo_pi | bytecode | miss | not-applicable | - | 2.492 | 1.688 | 0.938 | current-exact | closed-rejected-candidate | `bytecode-float-numeric` |
| mutex_await_journal | bytecode | miss | not-applicable | - | 0.220 | 10.427 | 0.198 | current-exact-mixed-targeted | closed-no-shared-leaf | `bytecode-concurrency` |
| mutex_ledger | bytecode | miss | not-applicable | - | 0.368 | 8.700 | 0.323 | current-exact-mixed-targeted | closed-no-shared-leaf | `bytecode-concurrency` |
| mutex_work_queue | bytecode | miss | not-applicable | - | 0.346 | 12.628 | 0.317 | current-exact-mixed-targeted | closed-no-shared-leaf | `bytecode-concurrency` |
| option_result_config | bytecode | miss | not-applicable | - | 0.786 | 42.486 | 0.767 | post-active-lookup-mixed | closed-no-shared-leaf | `bytecode-iterator-control` |
| pidigits | bytecode | meets | established | established-meet | 2.356 | 0.563 | 0.000 | current-exact | target-guard | `bytecode-target-guards` |
| rational_series | bytecode | miss | not-applicable | - | 4.046 | 35.152 | 3.925 | current-exact | closed-rejected-candidate | `bytecode-wide-numeric` |
| regex_set_audit | bytecode | miss | not-applicable | - | 4.020 | 222.099 | 4.001 | current-exact | closed-rejected-candidate | `bytecode-regex` |
| regex_stream_audit | bytecode | miss | not-applicable | - | 3.548 | 193.880 | 3.529 | current-exact | closed-rejected-candidate | `bytecode-regex` |
| regex_suffix_audit | bytecode | miss | not-applicable | - | 3.230 | 182.486 | 3.211 | current-exact | closed-rejected-candidate | `bytecode-regex` |
| reverse_complement | bytecode | miss | not-applicable | - | 3.290 | 123.684 | 3.262 | current-exact | closed-no-shared-leaf | `bytecode-byte-output` |
| rms_norm | bytecode | miss | not-applicable | - | 4.526 | 8.854 | 3.988 | current-exact | closed-rejected-candidate | `bytecode-float-numeric` |
| unicode_scalar_pipeline | bytecode | miss | not-applicable | - | 3.776 | 16.933 | 3.541 | current-exact-post-hash-index | closed-no-shared-leaf | `bytecode-text-map` |
| validated_job_pipeline | bytecode | miss | not-applicable | - | 0.834 | 9.952 | 0.746 | current-exact-mixed-targeted | closed-no-shared-leaf | `bytecode-concurrency` |
| wide_integer_records | bytecode | miss | not-applicable | - | 5.158 | 59.630 | 5.067 | current-exact | closed-rejected-candidate | `bytecode-wide-numeric` |
| word_frequency | bytecode | miss | not-applicable | - | 1.418 | 75.426 | 1.398 | current-exact-post-hash-index | closed-no-shared-leaf | `bytecode-text-map` |
| array_slice_window | compiled | miss | not-applicable | - | 0.100 | 20.000 | 0.095 | pre-current-binary | closed-no-shared-leaf | `compiled-iterator-control` |
| await_channel_mux | compiled | miss | not-applicable | - | 0.416 | 84.898 | 0.411 | current-exact-mixed-targeted | closed-rejected-candidate | `compiled-concurrency` |
| base64 | compiled | miss | not-applicable | - | 2.930 | 1.216 | 0.393 | current-exact | closed-no-shared-leaf | `compiled-byte-output` |
| binarytrees | compiled | meets | established | established-meet | 9.708 | 0.909 | 0.000 | current-exact | target-guard | `compiled-target-guards` |
| channel_rollup | compiled | miss | not-applicable | - | 0.548 | 94.483 | 0.542 | current-exact-mixed-targeted | closed-rejected-candidate | `compiled-concurrency` |
| concurrent_event_routing | compiled | miss | not-applicable | - | 2.652 | 552.500 | 2.647 | current-exact-mixed-targeted | closed-rejected-candidate | `compiled-concurrency` |
| concurrent_text_index | compiled | miss | not-applicable | - | 1.144 | 184.516 | 1.137 | current-exact-mixed-targeted | closed-rejected-candidate | `compiled-concurrency` |
| config_validation_extraction | compiled | miss | not-applicable | - | 0.110 | 26.190 | 0.106 | current-exact | closed-rejected-candidate | `compiled-regex` |
| dependency_plan | compiled | miss | not-applicable | - | 0.142 | 33.810 | 0.138 | pre-current-binary | closed-no-shared-leaf | `compiled-iterator-control` |
| dependency_wave_validation | compiled | miss | not-applicable | - | 1.268 | 301.905 | 1.264 | current-exact-mixed-targeted | closed-rejected-candidate | `compiled-concurrency` |
| distance_field | compiled | miss | not-applicable | - | 0.118 | 9.008 | 0.104 | current-exact | closed-rejected-candidate | `compiled-float-numeric` |
| document_audit | compiled | miss | not-applicable | - | 0.102 | 26.154 | 0.098 | pre-current-binary | closed-no-shared-leaf | `compiled-iterator-control` |
| fasta_generation | compiled | miss | not-applicable | - | 0.110 | 8.333 | 0.096 | current-exact | closed-no-shared-leaf | `compiled-byte-output` |
| fib | compiled | meets | established | established-meet | 3.268 | 1.029 | 0.000 | current-exact | target-guard | `compiled-target-guards` |
| fixed_width_128 | compiled | miss | not-applicable | - | 0.284 | 49.825 | 0.278 | current-exact | closed-rejected-candidate | `compiled-wide-numeric` |
| future_await_race | compiled | miss | not-applicable | - | 0.124 | 30.244 | 0.120 | current-exact-mixed-targeted | closed-rejected-candidate | `compiled-concurrency` |
| future_pipeline | compiled | miss | not-applicable | - | 0.350 | 64.815 | 0.344 | current-exact-mixed-targeted | closed-rejected-candidate | `compiled-concurrency` |
| i_before_e | compiled | miss | not-applicable | - | 0.110 | 1.711 | 0.042 | current-exact-post-hash-index | closed-no-shared-leaf | `compiled-text-map` |
| inventory_reconciliation | compiled | miss | not-applicable | - | 0.270 | 30.337 | 0.261 | current-exact-post-hash-index | closed-no-shared-leaf | `compiled-text-map` |
| json | compiled | meets | established | established-meet | 0.796 | 0.538 | 0.000 | current-exact | target-guard | `compiled-target-guards` |
| k_nucleotide | compiled | miss | not-applicable | - | 2.906 | 40.361 | 2.830 | current-exact-post-hash-index | closed-no-shared-leaf | `compiled-text-map` |
| lexical_rollup | compiled | miss | not-applicable | - | 0.142 | 26.296 | 0.136 | pre-current-binary | closed-no-shared-leaf | `compiled-iterator-control` |
| log_routing_redaction | compiled | miss | not-applicable | - | 0.114 | 20.357 | 0.108 | current-exact | closed-rejected-candidate | `compiled-regex` |
| mandelbrot | compiled | miss | not-applicable | - | 0.138 | 2.509 | 0.080 | current-exact | closed-rejected-candidate | `compiled-float-numeric` |
| matrixmultiply | compiled | miss | not-applicable | - | 1.224 | 1.129 | 0.083 | current-exact | closed-no-shared-leaf | `compiled-current-control` |
| monte_carlo_pi | compiled | meets | not-established | volatile-crossing | 0.210 | 0.994 | 0.000 | current-exact | closed-rejected-candidate | `compiled-float-numeric` |
| mutex_await_journal | compiled | miss | not-applicable | - | 0.696 | 174.000 | 0.692 | current-exact-mixed-targeted | closed-rejected-candidate | `compiled-concurrency` |
| mutex_ledger | compiled | miss | not-applicable | - | 0.750 | 178.571 | 0.746 | current-exact-mixed-targeted | closed-rejected-candidate | `compiled-concurrency` |
| mutex_work_queue | compiled | miss | not-applicable | - | 1.830 | 415.909 | 1.825 | current-exact-mixed-targeted | closed-rejected-candidate | `compiled-concurrency` |
| nbody | compiled | miss | not-applicable | - | 0.172 | 5.375 | 0.138 | current-exact | closed-rejected-candidate | `compiled-float-numeric` |
| option_result_config | compiled | miss | not-applicable | - | 0.220 | 64.706 | 0.216 | pre-current-binary | closed-no-shared-leaf | `compiled-iterator-control` |
| pidigits | compiled | miss | not-applicable | - | 1.272 | 1.102 | 0.057 | current-exact | closed-no-shared-leaf | `compiled-byte-output` |
| quicksort | compiled | meets | established | established-meet | 1.788 | 0.706 | 0.000 | current-exact | target-guard | `compiled-target-guards` |
| rational_series | compiled | miss | not-applicable | - | 0.128 | 9.846 | 0.114 | current-exact | closed-rejected-candidate | `compiled-wide-numeric` |
| regex_set_audit | compiled | miss | not-applicable | - | 0.160 | 32.000 | 0.155 | current-exact | closed-rejected-candidate | `compiled-regex` |
| regex_stream_audit | compiled | miss | not-applicable | - | 0.118 | 20.000 | 0.112 | current-exact | closed-rejected-candidate | `compiled-regex` |
| regex_suffix_audit | compiled | miss | not-applicable | - | 0.146 | 26.071 | 0.140 | current-exact | closed-rejected-candidate | `compiled-regex` |
| reverse_complement | compiled | miss | not-applicable | - | 0.128 | 7.314 | 0.110 | current-exact | closed-no-shared-leaf | `compiled-byte-output` |
| rms_norm | compiled | miss | not-applicable | - | 0.098 | 9.515 | 0.087 | current-exact | closed-rejected-candidate | `compiled-float-numeric` |
| sudoku_masks | compiled | miss | not-applicable | - | 1.758 | 3.135 | 1.168 | current-exact | closed-insufficient-breadth | `compiled-sudoku-quotient` |
| tapelang_alphabet | compiled | miss | not-applicable | - | 4.268 | 2.272 | 2.290 | current-exact | closed-no-shared-leaf | `compiled-current-control` |
| unicode_scalar_pipeline | compiled | miss | not-applicable | - | 0.262 | 25.437 | 0.251 | current-exact-post-hash-index | closed-no-shared-leaf | `compiled-text-map` |
| validated_job_pipeline | compiled | miss | not-applicable | - | 2.988 | 543.273 | 2.982 | current-exact-mixed-targeted | closed-rejected-candidate | `compiled-concurrency` |
| wide_integer_records | compiled | miss | not-applicable | - | 0.194 | 7.638 | 0.167 | current-exact | closed-rejected-candidate | `compiled-wide-numeric` |
| word_frequency | compiled | miss | not-applicable | - | 0.180 | 31.579 | 0.174 | current-exact-post-hash-index | closed-no-shared-leaf | `compiled-text-map` |

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

### `fib/compiled`

- Classification: `established-meet`; established guard: `established`; pooled limiting ratio: 1.029x go; cohort ratios: 1.029, 1.030.
- Samples: 10 Able and 10 limiting-reference; Able source: `1613bf12ba84db9d59c04eb8a2d9b9912d27f51696b22210743e1e499f52ae02`; evidence stdlib tree: `6a412c872ee66752de7c4417a5eda99806a7631da3b880c55574c6a640b82d9b`.
- Rationale: Both independent current-source five-run cohorts meet the compiled target at approximately 1.03x Go.
- Evidence: [2026-07-21-post-i64-threshold-stability.json](../../docs/perf-baselines/2026-07-21-post-i64-threshold-stability.json), [2026-07-21-post-i64-threshold-stability.md](../../docs/perf-baselines/2026-07-21-post-i64-threshold-stability.md).

### `json/compiled`

- Classification: `established-meet`; established guard: `established`; pooled limiting ratio: 0.561x go; cohort ratios: 0.588, 0.533.
- Samples: 10 Able and 10 limiting-reference; Able source: `84a895d80fee86e71b65c3614dc71088ef36d5a469b2756162007ee49d45b62a`; evidence stdlib tree: `6a412c872ee66752de7c4417a5eda99806a7631da3b880c55574c6a640b82d9b`.
- Rationale: Both independent current-source cohorts meet the compiled target.
- Evidence: [2026-07-20-source-exact-guards-variance.json](../../docs/perf-baselines/2026-07-20-source-exact-guards-variance.json), [2026-07-20-source-exact-established-guard-refresh.md](../../docs/perf-baselines/2026-07-20-source-exact-established-guard-refresh.md).

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

- Disposition: `target-guard`; exact unlike-application breadth: 0; profile freshness: `current-exact`; artifact identity: `scorecard-source-current-cross-cohort`.
- Owner: Different protected owners: Binary Trees node allocation/GC, Fib direct recursion, JSON host decoding, and QuickSort recursion.
- Rationale: These rows currently meet the product target and guard every admitted compiled candidate.
- Evidence: [2026-07-20-source-equivalence-scorecard.md](../../docs/perf-baselines/2026-07-20-source-equivalence-scorecard.md), [2026-07-20-compiled-control-current-invalidation-gate.md](../../docs/perf-baselines/2026-07-20-compiled-control-current-invalidation-gate.md), [2026-07-20-current-two-cohort-scorecard-reconciliation.md](../../docs/perf-baselines/2026-07-20-current-two-cohort-scorecard-reconciliation.md), [2026-07-20-threshold-stability-reconciliation.md](../../docs/perf-baselines/2026-07-20-threshold-stability-reconciliation.md), [2026-07-20-source-exact-established-guard-refresh.md](../../docs/perf-baselines/2026-07-20-source-exact-established-guard-refresh.md), [2026-07-21-post-i64-threshold-stability.md](../../docs/perf-baselines/2026-07-21-post-i64-threshold-stability.md).

### `compiled-current-control`

- Disposition: `closed-no-shared-leaf`; exact unlike-application breadth: 0; profile freshness: `current-exact`; artifact identity: `preserved-binary-matched`.
- Owner: Matrix primitive nested arithmetic and TapeLang flat dispatch remain separate generated bodies.
- Rationale: Fresh current binaries expose no compiler/runtime child shared across these control-flow programs.
- Evidence: [2026-07-20-compiled-control-current-invalidation-gate.md](../../docs/perf-baselines/2026-07-20-compiled-control-current-invalidation-gate.md).

### `compiled-sudoku-quotient`

- Disposition: `closed-insufficient-breadth`; exact unlike-application breadth: 1; profile freshness: `current-exact`; artifact identity: `preserved-binary-matched`.
- Owner: find_best_empty allocation/search plus square_index; signed Euclidean division is 13.17% cumulative in this application only.
- Rationale: The exact quotient helper receives zero samples in three unlike quotient controls.
- Evidence: [2026-07-20-compiled-quotient-only-ownership-census.md](../../docs/perf-baselines/2026-07-20-compiled-quotient-only-ownership-census.md).

### `compiled-float-numeric`

- Disposition: `closed-rejected-candidate`; exact unlike-application breadth: 3; profile freshness: `current-exact`; artifact identity: `preserved-binary-matched`.
- Owner: Float geometry/regions and normalized raw-float allocation recur, while Mandelbrot remains a distinct fused loop owner.
- Rationale: Raw-float lane and carrier candidates improved local allocation but regressed all broad wall-time guards.
- Evidence: [2026-07-20-cross-mode-numeric-wide-profile-gate.md](../../docs/perf-baselines/2026-07-20-cross-mode-numeric-wide-profile-gate.md), [2026-07-20-threshold-stability-reconciliation.md](../../docs/perf-baselines/2026-07-20-threshold-stability-reconciliation.md).

### `compiled-wide-numeric`

- Disposition: `closed-rejected-candidate`; exact unlike-application breadth: 3; profile freshness: `current-exact`; artifact identity: `preserved-binary-matched`.
- Owner: Package-environment publication through SwapEnv/sync/atomic.StorePointer recurs in checked arithmetic, Rational division, and wide-record parsing; the dominant arithmetic descendants remain different.
- Rationale: General execution-context and package-linkage alternatives already regressed unrelated N-Body and K-Nucleotide wall time; recurrence supplies breadth but no invalidation, and nominal specialization is prohibited.
- Evidence: [2026-07-20-wide-integer-records-profile-gate.md](../../docs/perf-baselines/2026-07-20-wide-integer-records-profile-gate.md), [2026-07-20-cross-mode-numeric-wide-profile-gate.md](../../docs/perf-baselines/2026-07-20-cross-mode-numeric-wide-profile-gate.md), [2026-07-20-compiled-quotient-only-ownership-census.md](../../docs/perf-baselines/2026-07-20-compiled-quotient-only-ownership-census.md).

### `compiled-byte-output`

- Disposition: `closed-no-shared-leaf`; exact unlike-application breadth: 0; profile freshness: `current-exact`; artifact identity: `preserved-current-binaries-no-candidate`.
- Owner: Current generated-main profiles split into Base64 host codec/MD5 work, direct FASTA arithmetic, Go BigInt kernels, and Reverse transform/copy/GC work.
- Rationale: The retained write_all change already removed the shared copy; fresh CPU/allocation profiles expose no remaining exact compiler-controlled descendant shared by at least three unlike applications.
- Evidence: [2026-07-20-compiled-byte-output-current-profile-gate.md](../../docs/perf-baselines/2026-07-20-compiled-byte-output-current-profile-gate.md), [2026-07-18-fasta-write-all-gate.md](../../docs/perf-baselines/2026-07-18-fasta-write-all-gate.md), [2026-07-18-cross-mode-array-capacity-growth-gate.md](../../docs/perf-baselines/2026-07-18-cross-mode-array-capacity-growth-gate.md), [2026-07-21-post-i64-threshold-stability.md](../../docs/perf-baselines/2026-07-21-post-i64-threshold-stability.md).

### `compiled-text-map`

- Disposition: `closed-no-shared-leaf`; exact unlike-application breadth: 0; profile freshness: `current-exact-post-hash-index`; artifact identity: `current-profile-processes-merged-short-compiled`.
- Owner: Post-index Inventory, Word Frequency, and K-Nucleotide split into integer equality/conversion, UTF-8 and String work, boxing, and different allocation/GC descendants.
- Rationale: The retained index removes the common linear scan; fresh exact profiles expose no next compiler-controlled leaf material in all three unlike map applications.
- Evidence: [2026-07-20-post-hash-index-profile-reconciliation.md](../../docs/perf-baselines/2026-07-20-post-hash-index-profile-reconciliation.md), [2026-07-20-generic-hash-map-index-gate.md](../../docs/perf-baselines/2026-07-20-generic-hash-map-index-gate.md).

### `compiled-regex`

- Disposition: `closed-rejected-candidate`; exact unlike-application breadth: 5; profile freshness: `current-exact`; artifact identity: `current-profile-binaries`.
- Owner: Canonical NFA closure, move, and thread management recur across the related API audits, log routing/redaction, and configuration validation/extraction.
- Rationale: Three unlike workload families now establish breadth, but closure scratch, capture templates, and primitive thread carriers are already retained while arenas, state indexes, character specialization, and carrier/call alternatives failed broad gates; the new exact profile invalidates none of those decisions.
- Evidence: [2026-07-20-regex-three-api-current-profile-gate.md](../../docs/perf-baselines/2026-07-20-regex-three-api-current-profile-gate.md), [2026-07-20-log-routing-redaction-profile-gate.md](../../docs/perf-baselines/2026-07-20-log-routing-redaction-profile-gate.md), [2026-07-20-config-validation-extraction-profile-gate.md](../../docs/perf-baselines/2026-07-20-config-validation-extraction-profile-gate.md).

### `compiled-concurrency`

- Disposition: `closed-rejected-candidate`; exact unlike-application breadth: 11; profile freshness: `current-exact-mixed-targeted`; artifact identity: `current-source-generated-main-profiles`.
- Owner: Goroutine identity through bridge.currentGID/runtime.Stack is the exact repeated compiled wall.
- Rationale: Four unlike feature-interaction applications reproduce the owner at 95.48%-96.94% cumulative, but the fixed-context replacement improved concurrency rows while regressing unrelated NBody materially.
- Evidence: [2026-07-20-mutex-work-queue-profile-gate.md](../../docs/perf-baselines/2026-07-20-mutex-work-queue-profile-gate.md), [2026-07-17-compiled-equal-cpu-concurrency-reconciliation.md](../../docs/perf-baselines/2026-07-17-compiled-equal-cpu-concurrency-reconciliation.md), [2026-07-18-post-string-full-scorecard-reconciliation.md](../../docs/perf-baselines/2026-07-18-post-string-full-scorecard-reconciliation.md), [2026-07-20-feature-interaction-application-gate.md](../../docs/perf-baselines/2026-07-20-feature-interaction-application-gate.md), [2026-07-20-dependency-wave-application-gate.md](../../docs/perf-baselines/2026-07-20-dependency-wave-application-gate.md), [2026-07-20-concurrent-event-routing-application-gate.md](../../docs/perf-baselines/2026-07-20-concurrent-event-routing-application-gate.md).

### `compiled-iterator-control`

- Disposition: `closed-no-shared-leaf`; exact unlike-application breadth: 2; profile freshness: `pre-current-binary`; artifact identity: `source-compatible-binary-not-retained`.
- Owner: Required Array slice copies, Queue/graph work, iterator filtering, lexical aggregation, and union/default dispatch remain different descendants.
- Rationale: Array capacity, iterator, nullable, generic union, and map/cache candidates do not clear the unlike-program breadth and wall-time gates.
- Evidence: [2026-07-17-compiled-text-iterator-graph-main-profile-refresh.md](../../docs/perf-baselines/2026-07-17-compiled-text-iterator-graph-main-profile-refresh.md), [2026-07-18-cross-mode-array-capacity-growth-gate.md](../../docs/perf-baselines/2026-07-18-cross-mode-array-capacity-growth-gate.md), [2026-07-15-option-result-config-profile-gate.md](../../docs/perf-baselines/2026-07-15-option-result-config-profile-gate.md).

### `bytecode-target-guards`

- Disposition: `target-guard`; exact unlike-application breadth: 0; profile freshness: `current-exact`; artifact identity: `scorecard-source-current-cross-cohort`.
- Owner: JSON host decoding and PiDigits BigInt/native output are protected target-meeting rows.
- Rationale: Every admitted VM candidate must preserve these current target meets.
- Evidence: [2026-07-20-source-equivalence-scorecard.md](../../docs/perf-baselines/2026-07-20-source-equivalence-scorecard.md), [2026-07-20-source-exact-established-guard-refresh.md](../../docs/perf-baselines/2026-07-20-source-exact-established-guard-refresh.md).

### `bytecode-float-numeric`

- Disposition: `closed-rejected-candidate`; exact unlike-application breadth: 3; profile freshness: `current-exact`; artifact identity: `current-profile-binaries`.
- Owner: Typed float regions/raw-float normalization own Distance, RMS, and Mandelbrot; Monte Carlo is an integer recurrence discriminator.
- Rationale: Typed lanes and native scalar-result carriers reduced local work but regressed unlike guards.
- Evidence: [2026-07-20-cross-mode-numeric-wide-profile-gate.md](../../docs/perf-baselines/2026-07-20-cross-mode-numeric-wide-profile-gate.md).

### `bytecode-wide-numeric`

- Disposition: `closed-rejected-candidate`; exact unlike-application breadth: 3; profile freshness: `current-exact`; artifact identity: `current-profile-binaries`.
- Owner: Raw integer extraction recurs across checked UInt arithmetic, Rational division/casts, and wide-record parsing/comparison/bitwise work, with different callers and carriers below it.
- Rationale: General carrier, split-extractor, producer-fusion, and store variants already failed unlike-program wall-time guards; the third application supplies breadth but no invalidation, and named wide-type VM rules are prohibited.
- Evidence: [2026-07-20-wide-integer-records-profile-gate.md](../../docs/perf-baselines/2026-07-20-wide-integer-records-profile-gate.md), [2026-07-20-cross-mode-numeric-wide-profile-gate.md](../../docs/perf-baselines/2026-07-20-cross-mode-numeric-wide-profile-gate.md).

### `bytecode-byte-output`

- Disposition: `closed-no-shared-leaf`; exact unlike-application breadth: 3; profile freshness: `current-exact`; artifact identity: `preserved-current-binary-retained-u8-array-path`.
- Owner: The retained raw-u8 Array push/read path clears the shared primitive-array leaf; residual Base64 host kernels, Reverse Array/extern work, and FASTA call/arithmetic work diverge.
- Rationale: A generic monomorphic u8 write/read path improves all three applications and passes target and unlike guards; post-change profiles expose no second material exact descendant shared by all three.
- Evidence: [2026-07-20-bytecode-byte-output-current-profile-gate.md](../../docs/perf-baselines/2026-07-20-bytecode-byte-output-current-profile-gate.md), [2026-07-18-fasta-write-all-gate.md](../../docs/perf-baselines/2026-07-18-fasta-write-all-gate.md), [2026-07-18-cross-mode-array-capacity-growth-gate.md](../../docs/perf-baselines/2026-07-18-cross-mode-array-capacity-growth-gate.md), [2026-07-20-threshold-stability-reconciliation.md](../../docs/perf-baselines/2026-07-20-threshold-stability-reconciliation.md).

### `bytecode-text-map`

- Disposition: `closed-no-shared-leaf`; exact unlike-application breadth: 0; profile freshness: `current-exact-post-hash-index`; artifact identity: `current-profile-processes`.
- Owner: Post-index Inventory spreads across call/member/type caches, Word Frequency across slot/Array/String work, and K-Nucleotide across call/return/raw-integer operations.
- Rationale: Language HashMap search is no longer material across the three applications, and recurring Go map leaves belong to different VM caches and Able operations.
- Evidence: [2026-07-20-post-hash-index-profile-reconciliation.md](../../docs/perf-baselines/2026-07-20-post-hash-index-profile-reconciliation.md), [2026-07-20-generic-hash-map-index-gate.md](../../docs/perf-baselines/2026-07-20-generic-hash-map-index-gate.md).

### `bytecode-regex`

- Disposition: `closed-rejected-candidate`; exact unlike-application breadth: 5; profile freshness: `current-exact`; artifact identity: `current-profile-binaries`.
- Owner: NFA closure/move/thread work and generic Array-slot member traffic recur across the related API audits, log routing/redaction, and configuration validation/extraction.
- Rationale: Three unlike workload families now establish breadth, but generic Array-slot caches, raw-integer carriers, call/return changes, and NFA carrier alternatives are already retained or broadly rejected; the new exact profile exposes no distinct invalidation, and benchmark-shaped regex opcodes remain prohibited.
- Evidence: [2026-07-20-regex-three-api-current-profile-gate.md](../../docs/perf-baselines/2026-07-20-regex-three-api-current-profile-gate.md), [2026-07-20-log-routing-redaction-profile-gate.md](../../docs/perf-baselines/2026-07-20-log-routing-redaction-profile-gate.md), [2026-07-20-config-validation-extraction-profile-gate.md](../../docs/perf-baselines/2026-07-20-config-validation-extraction-profile-gate.md).

### `bytecode-concurrency`

- Disposition: `closed-no-shared-leaf`; exact unlike-application breadth: 11; profile freshness: `current-exact-mixed-targeted`; artifact identity: `current-source-warmed-main-profiles`.
- Owner: The retained lazy environment state combines synchronization and metadata allocation across the original concurrency applications; the four feature-interaction applications split across member lookup, call/return, scheduler atomic, type-match, allocation, and GC descendants.
- Rationale: The retained generic environment-state allocation passed the broad gate. Widened warmed profiles expose no second material child shared by at least three unlike applications whose generic family has not already completed a broad gate.
- Evidence: [2026-07-20-mutex-work-queue-profile-gate.md](../../docs/perf-baselines/2026-07-20-mutex-work-queue-profile-gate.md), [2026-07-20-bytecode-concurrency-current-profile-gate.md](../../docs/perf-baselines/2026-07-20-bytecode-concurrency-current-profile-gate.md), [2026-07-18-cross-mode-exact-leaf-selection-reconciliation.md](../../docs/perf-baselines/2026-07-18-cross-mode-exact-leaf-selection-reconciliation.md), [2026-07-20-feature-interaction-application-gate.md](../../docs/perf-baselines/2026-07-20-feature-interaction-application-gate.md), [2026-07-20-dependency-wave-application-gate.md](../../docs/perf-baselines/2026-07-20-dependency-wave-application-gate.md), [2026-07-20-concurrent-event-routing-application-gate.md](../../docs/perf-baselines/2026-07-20-concurrent-event-routing-application-gate.md).

### `bytecode-iterator-control`

- Disposition: `closed-no-shared-leaf`; exact unlike-application breadth: 2; profile freshness: `post-active-lookup-mixed`; artifact identity: `core-current-compatible-short-controls-older`.
- Owner: Array copying, graph/Queue, iterator filtering, lexical aggregation, and union/default dispatch split beneath cached member lookup.
- Rationale: Member validation, Array growth, return/frame, nullable, and union candidates are already retained or rejected.
- Evidence: [2026-07-19-bytecode-text-map-graph-profile-gate.md](../../docs/perf-baselines/2026-07-19-bytecode-text-map-graph-profile-gate.md), [2026-07-19-bytecode-inline-return-profile-reconciliation.md](../../docs/perf-baselines/2026-07-19-bytecode-inline-return-profile-reconciliation.md), [2026-07-18-cross-mode-array-capacity-growth-gate.md](../../docs/perf-baselines/2026-07-18-cross-mode-array-capacity-growth-gate.md).
