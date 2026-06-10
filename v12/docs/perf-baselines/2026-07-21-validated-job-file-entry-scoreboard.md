# External Application Scoreboard

- Source measurements through: `2026-07-22T04:50:09.298732Z`
- Scope: verifier-backed Able application launches only; tree-walker is intentionally excluded from this performance target.
- Guard: each source scorecard records its process count, CPU-affinity when used, runtime settings, and per-process timeout.
- Compiled: `5/48` selected rankable rows meet the 95%-of-Go target.
- Bytecode: `3/41` selected rankable rows meet both 95%-of-Python and 95%-of-Ruby targets.
- Canonical Able source fingerprints: `96` row fingerprints in JSON; `96` came from the measured source report and the remainder are current-source legacy fingerprints.
- Verifier/declared-input contracts: `96` row fingerprints in JSON; `96` were captured before the timed launch and the remainder are current-contract legacy reconstructions.
- Canonical stdlib runtime sources: `70` `.able` files, tree SHA-256 `6a412c872ee66752de7c4417a5eda99806a7631da3b880c55574c6a640b82d9b`; Git `219eff222c28406487231713753641bc49ee5b9a` (dirty).
- Strict candidate selection: `89` reviewed benchmark/mode rows, SHA-256 `6d4ea2d230a1b285f1fed2eb68f4c81f442b2e684f76c4c494c751d27846d7f7`; timeout rows remain in full status.
- Matched reference source fingerprints: `139` comparison fingerprints in JSON; `139` came from measured reference reports and the remainder are current-source legacy fingerprints.
- `unranked` means a partial, timed-out, failed, or unavailable matched run/reference; it is never counted as a pass or fail.
- `Unranked reason` identifies whether the Able launch or its required reference prevents ranking; reference-unavailable does not infer why that source has no valid ratio.

| Benchmark | Mode | Able status | Able (s) | Go / ratio | Python / ratio | Ruby / ratio | Target | Unranked reason |
| --- | --- | --- | ---: | --- | --- | --- | --- | --- |
| `fib` | `compiled` | `verified` | 3.8460 | 3.1773 / 1.21x | n/a | n/a | `miss` | — |
| `binarytrees` | `compiled` | `verified` | 9.5780 | 10.6793 / 0.90x | n/a | n/a | `meets` | — |
| `matrixmultiply` | `compiled` | `verified` | 1.1380 | 1.0840 / 1.05x | n/a | n/a | `meets` | — |
| `quicksort` | `compiled` | `verified` | 1.8460 | 2.5326 / 0.73x | n/a | n/a | `meets` | — |
| `sudoku_masks` | `compiled` | `verified` | 1.7920 | 0.5607 / 3.20x | n/a | n/a | `miss` | — |
| `i_before_e` | `compiled` | `verified` | 0.1020 | 0.0643 / 1.59x | n/a | n/a | `miss` | — |
| `base64` | `compiled` | `verified` | 2.5580 | 2.4105 / 1.06x | n/a | n/a | `miss` | — |
| `json` | `compiled` | `verified` | 0.7920 | 1.4798 / 0.54x | n/a | n/a | `meets` | — |
| `monte_carlo_pi` | `compiled` | `verified` | 0.2120 | 0.2113 / 1.00x | n/a | n/a | `meets` | — |
| `pidigits` | `compiled` | `verified` | 1.2540 | 1.1541 / 1.09x | n/a | n/a | `miss` | — |
| `mandelbrot` | `compiled` | `verified` | 0.1660 | 0.0550 / 3.02x | n/a | n/a | `miss` | — |
| `reverse_complement` | `compiled` | `verified` | 0.1540 | 0.0175 / 8.80x | n/a | n/a | `miss` | — |
| `k_nucleotide` | `compiled` | `verified` | 3.1640 | 0.0720 / 43.94x | n/a | n/a | `miss` | — |
| `nbody` | `compiled` | `verified` | 0.1840 | 0.0320 / 5.75x | n/a | n/a | `miss` | — |
| `tapelang_alphabet` | `compiled` | `verified` | 3.5700 | 1.8789 / 1.90x | n/a | n/a | `miss` | — |
| `distance_field` | `compiled` | `verified` | 0.1040 | 0.0131 / 7.94x | n/a | n/a | `miss` | — |
| `rms_norm` | `compiled` | `verified` | 0.0960 | 0.0103 / 9.32x | n/a | n/a | `miss` | — |
| `fasta_generation` | `compiled` | `verified` | 0.1260 | 0.0132 / 9.55x | n/a | n/a | `miss` | — |
| `fixed_width_128` | `compiled` | `verified` | 0.2400 | 0.0057 / 42.11x | n/a | n/a | `miss` | — |
| `rational_series` | `compiled` | `verified` | 0.1380 | 0.0130 / 10.62x | n/a | n/a | `miss` | — |
| `wide_integer_records` | `compiled` | `verified` | 0.2000 | 0.0254 / 7.87x | n/a | n/a | `miss` | — |
| `word_frequency` | `compiled` | `verified` | 0.1880 | 0.0057 / 32.98x | n/a | n/a | `miss` | — |
| `document_audit` | `compiled` | `verified` | 0.0880 | 0.0039 / 22.56x | n/a | n/a | `miss` | — |
| `lexical_rollup` | `compiled` | `verified` | 0.1160 | 0.0054 / 21.48x | n/a | n/a | `miss` | — |
| `channel_rollup` | `compiled` | `verified` | 0.5400 | 0.0058 / 93.10x | n/a | n/a | `miss` | — |
| `future_pipeline` | `compiled` | `verified` | 0.3820 | 0.0054 / 70.74x | n/a | n/a | `miss` | — |
| `future_await_race` | `compiled` | `verified` | 0.0900 | 0.0041 / 21.95x | n/a | n/a | `miss` | — |
| `await_channel_mux` | `compiled` | `verified` | 0.3560 | 0.0049 / 72.65x | n/a | n/a | `miss` | — |
| `mutex_ledger` | `compiled` | `verified` | 0.8260 | 0.0042 / 196.67x | n/a | n/a | `miss` | — |
| `mutex_await_journal` | `compiled` | `verified` | 0.6840 | 0.0040 / 171.00x | n/a | n/a | `miss` | — |
| `mutex_work_queue` | `compiled` | `verified` | 1.3660 | 0.0044 / 310.45x | n/a | n/a | `miss` | — |
| `regex_suffix_audit` | `compiled` | `verified` | 0.1220 | 0.0056 / 21.79x | n/a | n/a | `miss` | — |
| `regex_set_audit` | `compiled` | `verified` | 0.1520 | 0.0050 / 30.40x | n/a | n/a | `miss` | — |
| `regex_stream_audit` | `compiled` | `verified` | 0.1120 | 0.0059 / 18.98x | n/a | n/a | `miss` | — |
| `log_routing_redaction` | `compiled` | `verified` | 0.1240 | 0.0056 / 22.14x | n/a | n/a | `miss` | — |
| `config_validation_extraction` | `compiled` | `verified` | 0.1020 | 0.0042 / 24.29x | n/a | n/a | `miss` | — |
| `unicode_scalar_pipeline` | `compiled` | `verified` | 0.2940 | 0.0103 / 28.54x | n/a | n/a | `miss` | — |
| `array_slice_window` | `compiled` | `verified` | 0.0880 | 0.0050 / 17.60x | n/a | n/a | `miss` | — |
| `dependency_plan` | `compiled` | `verified` | 0.0940 | 0.0042 / 22.38x | n/a | n/a | `miss` | — |
| `inventory_reconciliation` | `compiled` | `verified` | 0.2140 | 0.0089 / 24.04x | n/a | n/a | `miss` | — |
| `option_result_config` | `compiled` | `verified` | 0.2480 | 0.0034 / 72.94x | n/a | n/a | `miss` | — |
| `concurrent_text_index` | `compiled` | `verified` | 0.8760 | 0.0062 / 141.29x | n/a | n/a | `miss` | — |
| `dependency_wave_validation` | `compiled` | `verified` | 1.2440 | 0.0042 / 296.19x | n/a | n/a | `miss` | — |
| `concurrent_event_routing` | `compiled` | `verified` | 3.2340 | 0.0048 / 673.75x | n/a | n/a | `miss` | — |
| `fib` | `bytecode` | `verified` | 0.2200 | n/a | n/a | 45.8208 / 0.00x | `unranked` | Python reference unavailable |
| `binarytrees` | `bytecode` | `timeout` | n/a | n/a | n/a | n/a | `unranked` | Able timed out |
| `matrixmultiply` | `bytecode` | `verified` | 4.7200 | n/a | 52.4334 / 0.09x | 46.9037 / 0.10x | `meets` | — |
| `quicksort` | `bytecode` | `timeout` | n/a | n/a | 24.5204 / n/a | 15.7502 / n/a | `unranked` | Able timed out |
| `sudoku_masks` | `bytecode` | `timeout` | n/a | n/a | 18.6741 / n/a | 22.5619 / n/a | `unranked` | Able timed out |
| `i_before_e` | `bytecode` | `verified` | 0.5260 | n/a | 0.0833 / 6.31x | 0.1143 / 4.60x | `miss` | — |
| `base64` | `bytecode` | `verified` | 2.8060 | n/a | 4.0059 / 0.70x | 2.6783 / 1.05x | `meets` | — |
| `json` | `bytecode` | `verified` | 0.8160 | n/a | 2.6975 / 0.30x | 1.7209 / 0.47x | `meets` | — |
| `monte_carlo_pi` | `bytecode` | `verified` | 2.4920 | n/a | 1.4759 / 1.69x | 1.6183 / 1.54x | `miss` | — |
| `pidigits` | `bytecode` | `verified` | 2.3560 | n/a | 4.1866 / 0.56x | 10.1687 / 0.23x | `meets` | — |
| `mandelbrot` | `bytecode` | `verified` | 6.3240 | n/a | 1.1783 / 5.37x | 1.8131 / 3.49x | `miss` | — |
| `reverse_complement` | `bytecode` | `verified` | 3.2900 | n/a | 0.0266 / 123.68x | 0.0723 / 45.50x | `miss` | — |
| `k_nucleotide` | `bytecode` | `verified` | 45.1640 | n/a | 1.2587 / 35.88x | 1.2902 / 35.01x | `miss` | — |
| `fasta_generation` | `bytecode` | `verified` | 1.8280 | n/a | 0.2023 / 9.04x | 0.2986 / 6.12x | `miss` | — |
| `nbody` | `bytecode` | `timeout` | n/a | n/a | 2.1263 / n/a | 3.4073 / n/a | `unranked` | Able timed out |
| `tapelang_alphabet` | `bytecode` | `timeout` | n/a | n/a | n/a | n/a | `unranked` | Able timed out |
| `distance_field` | `bytecode` | `verified` | 5.7480 | n/a | 0.5347 / 10.75x | 0.4190 / 13.72x | `miss` | — |
| `rms_norm` | `bytecode` | `verified` | 4.5260 | n/a | 0.8428 / 5.37x | 0.5112 / 8.85x | `miss` | — |
| `channel_rollup` | `bytecode` | `verified` | 0.4460 | n/a | 0.0412 / 10.83x | 0.0504 / 8.85x | `miss` | — |
| `future_pipeline` | `bytecode` | `verified` | 0.4060 | n/a | 0.0614 / 6.61x | 0.0733 / 5.54x | `miss` | — |
| `future_await_race` | `bytecode` | `verified` | 0.1360 | n/a | 0.0339 / 4.01x | 0.0557 / 2.44x | `miss` | — |
| `await_channel_mux` | `bytecode` | `verified` | 0.2080 | n/a | 0.1286 / 1.62x | 0.1205 / 1.73x | `miss` | — |
| `mutex_ledger` | `bytecode` | `verified` | 0.3680 | n/a | 0.0423 / 8.70x | 0.0589 / 6.25x | `miss` | — |
| `mutex_await_journal` | `bytecode` | `verified` | 0.2200 | n/a | 0.0211 / 10.43x | 0.0446 / 4.93x | `miss` | — |
| `mutex_work_queue` | `bytecode` | `verified` | 0.3460 | n/a | 0.0274 / 12.63x | 0.0527 / 6.57x | `miss` | — |
| `fixed_width_128` | `bytecode` | `verified` | 7.9880 | n/a | 0.3652 / 21.87x | 0.7681 / 10.40x | `miss` | — |
| `rational_series` | `bytecode` | `verified` | 4.0460 | n/a | 0.1151 / 35.15x | 0.1518 / 26.65x | `miss` | — |
| `wide_integer_records` | `bytecode` | `verified` | 5.1580 | n/a | 0.0865 / 59.63x | 0.1471 / 35.06x | `miss` | — |
| `word_frequency` | `bytecode` | `verified` | 1.4180 | n/a | 0.0188 / 75.43x | 0.0485 / 29.24x | `miss` | — |
| `document_audit` | `bytecode` | `verified` | 0.2740 | n/a | 0.0140 / 19.57x | 0.0408 / 6.72x | `miss` | — |
| `lexical_rollup` | `bytecode` | `verified` | 0.3980 | n/a | 0.0171 / 23.27x | 0.0461 / 8.63x | `miss` | — |
| `regex_suffix_audit` | `bytecode` | `verified` | 3.2300 | n/a | 0.0177 / 182.49x | 0.0407 / 79.36x | `miss` | — |
| `regex_set_audit` | `bytecode` | `verified` | 4.0200 | n/a | 0.0181 / 222.10x | 0.0406 / 99.01x | `miss` | — |
| `regex_stream_audit` | `bytecode` | `verified` | 3.5480 | n/a | 0.0183 / 193.88x | 0.0473 / 75.01x | `miss` | — |
| `log_routing_redaction` | `bytecode` | `verified` | 3.0160 | n/a | 0.0242 / 124.63x | 0.0630 / 47.87x | `miss` | — |
| `array_slice_window` | `bytecode` | `verified` | 0.6480 | n/a | 0.0294 / 22.04x | 0.0638 / 10.16x | `miss` | — |
| `dependency_plan` | `bytecode` | `verified` | 0.4980 | n/a | 0.0162 / 30.74x | 0.0466 / 10.69x | `miss` | — |
| `inventory_reconciliation` | `bytecode` | `verified` | 2.5360 | n/a | 0.0680 / 37.29x | 0.1179 / 21.51x | `miss` | — |
| `option_result_config` | `bytecode` | `verified` | 0.7860 | n/a | 0.0185 / 42.49x | 0.0434 / 18.11x | `miss` | — |
| `unicode_scalar_pipeline` | `bytecode` | `verified` | 3.7760 | n/a | 0.2230 / 16.93x | 0.3237 / 11.67x | `miss` | — |
| `config_validation_extraction` | `bytecode` | `verified` | 1.3860 | n/a | 0.0246 / 56.34x | 0.0452 / 30.66x | `miss` | — |
| `concurrent_text_index` | `bytecode` | `verified` | 0.6340 | n/a | 0.0844 / 7.51x | 0.1024 / 6.19x | `miss` | — |
| `dependency_wave_validation` | `bytecode` | `verified` | 0.4200 | n/a | 0.0480 / 8.75x | 0.0514 / 8.17x | `miss` | — |
| `concurrent_event_routing` | `bytecode` | `verified` | 3.0320 | n/a | 0.0297 / 102.09x | 0.0488 / 62.13x | `miss` | — |
| `policy_record_dispatch` | `compiled` | `verified` | 0.2260 | 0.0056 / 40.36x | n/a | n/a | `miss` | — |
| `policy_record_dispatch` | `bytecode` | `verified` | 8.7820 | n/a | 0.0245 / 358.45x | 0.0519 / 169.21x | `miss` | — |
| `concurrent_document_pipeline` | `compiled` | `verified` | 0.2520 | 0.0037 / 68.11x | n/a | n/a | `miss` | — |
| `concurrent_document_pipeline` | `bytecode` | `verified` | 0.3480 | n/a | 0.0223 / 15.61x | 0.0419 / 8.31x | `miss` | — |
| `manifest_normalization` | `compiled` | `verified` | 0.2160 | 0.0054 / 40.00x | n/a | n/a | `miss` | — |
| `manifest_normalization` | `bytecode` | `verified` | 1.7840 | n/a | 0.0227 / 78.59x | 0.0536 / 33.28x | `miss` | — |
| `validated_job_pipeline` | `compiled` | `verified` | 1.1300 | 0.0048 / 235.42x | n/a | n/a | `miss` | — |
| `validated_job_pipeline` | `bytecode` | `verified` | 0.4140 | n/a | 0.0368 / 11.25x | 0.0622 / 6.66x | `miss` | — |

## Source scorecards

- `v12/docs/perf-baselines/2026-07-21-validated-job-file-entry-preserved-compiled.json` — `coverage` (`2026-07-21T13:34:14.191208Z`)
- `v12/docs/perf-baselines/2026-07-20-current-product-scorecard-generality-bytecode-01-status.json` — `custom` (`2026-07-21T02:00:51.129516Z`)
- `v12/docs/perf-baselines/2026-07-20-current-product-scorecard-generality-bytecode-02-status.json` — `custom` (`2026-07-21T02:02:45.268187Z`)
- `v12/docs/perf-baselines/2026-07-20-current-product-scorecard-generality-bytecode-03-selected.json` — `custom` (`2026-07-21T02:03:24.657931Z`)
- `v12/docs/perf-baselines/2026-07-20-current-product-scorecard-generality-bytecode-04-selected.json` — `custom` (`2026-07-21T02:04:27.347812Z`)
- `v12/docs/perf-baselines/2026-07-20-current-product-scorecard-generality-bytecode-05-selected.json` — `custom` (`2026-07-21T02:08:46.033830Z`)
- `v12/docs/perf-baselines/2026-07-20-current-product-scorecard-generality-bytecode-06-status.json` — `custom` (`2026-07-21T02:10:40.209835Z`)
- `v12/docs/perf-baselines/2026-07-20-current-product-scorecard-generality-bytecode-07-selected.json` — `custom` (`2026-07-21T02:11:36.565435Z`)
- `v12/docs/perf-baselines/2026-07-20-current-product-scorecard-async-bytecode-01-selected.json` — `custom` (`2026-07-21T02:14:11.159974Z`)
- `v12/docs/perf-baselines/2026-07-20-current-product-scorecard-async-bytecode-02-selected.json` — `custom` (`2026-07-21T02:14:26.499241Z`)
- `v12/docs/perf-baselines/2026-07-20-current-product-scorecard-coverage-extra-bytecode-01-selected.json` — `custom` (`2026-07-21T02:36:22.347335Z`)
- `v12/docs/perf-baselines/2026-07-20-current-product-scorecard-coverage-extra-bytecode-02-selected.json` — `custom` (`2026-07-21T02:36:48.856736Z`)
- `v12/docs/perf-baselines/2026-07-20-current-product-scorecard-coverage-extra-bytecode-03-selected.json` — `custom` (`2026-07-21T02:37:54.393923Z`)
- `v12/docs/perf-baselines/2026-07-20-current-product-scorecard-coverage-extra-bytecode-04-selected.json` — `custom` (`2026-07-21T02:38:41.645530Z`)
- `v12/docs/perf-baselines/2026-07-20-current-product-scorecard-coverage-extra-bytecode-05-selected.json` — `custom` (`2026-07-21T02:38:50.930291Z`)
- `v12/docs/perf-baselines/2026-07-21-validated-job-file-entry-preserved-bytecode.json` — `custom` (`2026-07-21T05:23:23.328393Z`)
- `v12/docs/perf-baselines/2026-07-21-policy-record-dispatch-promotion-compiled.json` — `custom` (`2026-07-21T16:49:21.617761Z`)
- `v12/docs/perf-baselines/2026-07-21-policy-record-dispatch-promotion-bytecode.json` — `custom` (`2026-07-21T16:50:18.320836Z`)
- `v12/docs/perf-baselines/2026-07-21-concurrent-document-pipeline-promotion-compiled.json` — `custom` (`2026-07-22T03:26:40.965451Z`)
- `v12/docs/perf-baselines/2026-07-21-concurrent-document-pipeline-promotion-bytecode.json` — `custom` (`2026-07-22T03:26:45.168787Z`)
- `v12/docs/perf-baselines/2026-07-21-manifest-normalization-promotion-compiled.json` — `custom` (`2026-07-22T03:51:03.537494Z`)
- `v12/docs/perf-baselines/2026-07-21-manifest-normalization-promotion-bytecode.json` — `custom` (`2026-07-22T03:51:15.051771Z`)
- `v12/docs/perf-baselines/2026-07-21-validated-job-file-entry-comparison-d.json` — `custom` (`2026-07-22T04:50:09.298732Z`)

Regenerate after a new verifier-backed source scorecard with:

```sh
just bench-scoreboard
```

To replace the selected sources, pass each new scorecard explicitly, for example
`just bench-scoreboard --input path/to/compiled.json --input path/to/bytecode.json`.

Validate the checked-in report without running performance workloads with:

```sh
just bench-scoreboard-check
```
