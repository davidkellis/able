# External Application Scoreboard

- Source measurements through: `2026-07-21T12:04:45.151360Z`
- Scope: verifier-backed Able application launches only; tree-walker is intentionally excluded from this performance target.
- Guard: each source scorecard records its process count, CPU-affinity when used, runtime settings, and per-process timeout.
- Compiled: `5/45` selected rankable rows meet the 95%-of-Go target.
- Bytecode: `3/38` selected rankable rows meet both 95%-of-Python and 95%-of-Ruby targets.
- Canonical Able source fingerprints: `90` row fingerprints in JSON; `90` came from the measured source report and the remainder are current-source legacy fingerprints.
- Verifier/declared-input contracts: `90` row fingerprints in JSON; `90` were captured before the timed launch and the remainder are current-contract legacy reconstructions.
- Canonical stdlib runtime sources: `70` `.able` files, tree SHA-256 `6a412c872ee66752de7c4417a5eda99806a7631da3b880c55574c6a640b82d9b`; Git `219eff222c28406487231713753641bc49ee5b9a` (dirty).
- Strict candidate selection: `83` reviewed benchmark/mode rows, SHA-256 `1ad518fdb927dde805960ec549673016f669830d47a4f033e4830d4970bad66f`; timeout rows remain in full status.
- Matched reference source fingerprints: `130` comparison fingerprints in JSON; `130` came from measured reference reports and the remainder are current-source legacy fingerprints.
- `unranked` means a partial, timed-out, failed, or unavailable matched run/reference; it is never counted as a pass or fail.
- `Unranked reason` identifies whether the Able launch or its required reference prevents ranking; reference-unavailable does not infer why that source has no valid ratio.

| Benchmark | Mode | Able status | Able (s) | Go / ratio | Python / ratio | Ruby / ratio | Target | Unranked reason |
| --- | --- | --- | ---: | --- | --- | --- | --- | --- |
| `fib` | `compiled` | `verified` | 3.2680 | 3.1773 / 1.03x | n/a | n/a | `meets` | — |
| `binarytrees` | `compiled` | `verified` | 9.7080 | 10.6793 / 0.91x | n/a | n/a | `meets` | — |
| `matrixmultiply` | `compiled` | `verified` | 1.2240 | 1.0840 / 1.13x | n/a | n/a | `miss` | — |
| `quicksort` | `compiled` | `verified` | 1.7880 | 2.5326 / 0.71x | n/a | n/a | `meets` | — |
| `sudoku_masks` | `compiled` | `verified` | 1.7580 | 0.5607 / 3.14x | n/a | n/a | `miss` | — |
| `i_before_e` | `compiled` | `verified` | 0.1100 | 0.0643 / 1.71x | n/a | n/a | `miss` | — |
| `base64` | `compiled` | `verified` | 2.9300 | 2.4105 / 1.22x | n/a | n/a | `miss` | — |
| `json` | `compiled` | `verified` | 0.7960 | 1.4798 / 0.54x | n/a | n/a | `meets` | — |
| `monte_carlo_pi` | `compiled` | `verified` | 0.2100 | 0.2113 / 0.99x | n/a | n/a | `meets` | — |
| `pidigits` | `compiled` | `verified` | 1.2720 | 1.1541 / 1.10x | n/a | n/a | `miss` | — |
| `mandelbrot` | `compiled` | `verified` | 0.1380 | 0.0550 / 2.51x | n/a | n/a | `miss` | — |
| `reverse_complement` | `compiled` | `verified` | 0.1280 | 0.0175 / 7.31x | n/a | n/a | `miss` | — |
| `k_nucleotide` | `compiled` | `verified` | 2.9060 | 0.0720 / 40.36x | n/a | n/a | `miss` | — |
| `nbody` | `compiled` | `verified` | 0.1720 | 0.0320 / 5.37x | n/a | n/a | `miss` | — |
| `tapelang_alphabet` | `compiled` | `verified` | 4.2680 | 1.8789 / 2.27x | n/a | n/a | `miss` | — |
| `distance_field` | `compiled` | `verified` | 0.1180 | 0.0131 / 9.01x | n/a | n/a | `miss` | — |
| `rms_norm` | `compiled` | `verified` | 0.0980 | 0.0103 / 9.51x | n/a | n/a | `miss` | — |
| `fasta_generation` | `compiled` | `verified` | 0.1100 | 0.0132 / 8.33x | n/a | n/a | `miss` | — |
| `fixed_width_128` | `compiled` | `verified` | 0.2840 | 0.0057 / 49.82x | n/a | n/a | `miss` | — |
| `rational_series` | `compiled` | `verified` | 0.1280 | 0.0130 / 9.85x | n/a | n/a | `miss` | — |
| `wide_integer_records` | `compiled` | `verified` | 0.1940 | 0.0254 / 7.64x | n/a | n/a | `miss` | — |
| `word_frequency` | `compiled` | `verified` | 0.1800 | 0.0057 / 31.58x | n/a | n/a | `miss` | — |
| `document_audit` | `compiled` | `verified` | 0.1020 | 0.0039 / 26.15x | n/a | n/a | `miss` | — |
| `lexical_rollup` | `compiled` | `verified` | 0.1420 | 0.0054 / 26.30x | n/a | n/a | `miss` | — |
| `channel_rollup` | `compiled` | `verified` | 0.5480 | 0.0058 / 94.48x | n/a | n/a | `miss` | — |
| `future_pipeline` | `compiled` | `verified` | 0.3500 | 0.0054 / 64.81x | n/a | n/a | `miss` | — |
| `future_await_race` | `compiled` | `verified` | 0.1240 | 0.0041 / 30.24x | n/a | n/a | `miss` | — |
| `await_channel_mux` | `compiled` | `verified` | 0.4160 | 0.0049 / 84.90x | n/a | n/a | `miss` | — |
| `mutex_ledger` | `compiled` | `verified` | 0.7500 | 0.0042 / 178.57x | n/a | n/a | `miss` | — |
| `mutex_await_journal` | `compiled` | `verified` | 0.6960 | 0.0040 / 174.00x | n/a | n/a | `miss` | — |
| `mutex_work_queue` | `compiled` | `verified` | 1.8300 | 0.0044 / 415.91x | n/a | n/a | `miss` | — |
| `regex_suffix_audit` | `compiled` | `verified` | 0.1460 | 0.0056 / 26.07x | n/a | n/a | `miss` | — |
| `regex_set_audit` | `compiled` | `verified` | 0.1600 | 0.0050 / 32.00x | n/a | n/a | `miss` | — |
| `regex_stream_audit` | `compiled` | `verified` | 0.1180 | 0.0059 / 20.00x | n/a | n/a | `miss` | — |
| `log_routing_redaction` | `compiled` | `verified` | 0.1140 | 0.0056 / 20.36x | n/a | n/a | `miss` | — |
| `config_validation_extraction` | `compiled` | `verified` | 0.1100 | 0.0042 / 26.19x | n/a | n/a | `miss` | — |
| `unicode_scalar_pipeline` | `compiled` | `verified` | 0.2620 | 0.0103 / 25.44x | n/a | n/a | `miss` | — |
| `array_slice_window` | `compiled` | `verified` | 0.1000 | 0.0050 / 20.00x | n/a | n/a | `miss` | — |
| `dependency_plan` | `compiled` | `verified` | 0.1420 | 0.0042 / 33.81x | n/a | n/a | `miss` | — |
| `inventory_reconciliation` | `compiled` | `verified` | 0.2700 | 0.0089 / 30.34x | n/a | n/a | `miss` | — |
| `option_result_config` | `compiled` | `verified` | 0.2200 | 0.0034 / 64.71x | n/a | n/a | `miss` | — |
| `concurrent_text_index` | `compiled` | `verified` | 1.1440 | 0.0062 / 184.52x | n/a | n/a | `miss` | — |
| `validated_job_pipeline` | `compiled` | `verified` | 2.9880 | 0.0055 / 543.27x | n/a | n/a | `miss` | — |
| `dependency_wave_validation` | `compiled` | `verified` | 1.2680 | 0.0042 / 301.90x | n/a | n/a | `miss` | — |
| `concurrent_event_routing` | `compiled` | `verified` | 2.6520 | 0.0048 / 552.50x | n/a | n/a | `miss` | — |
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
| `validated_job_pipeline` | `bytecode` | `verified` | 0.8340 | n/a | 0.1056 / 7.90x | 0.0838 / 9.95x | `miss` | — |
| `dependency_wave_validation` | `bytecode` | `verified` | 0.4200 | n/a | 0.0480 / 8.75x | 0.0514 / 8.17x | `miss` | — |
| `concurrent_event_routing` | `bytecode` | `verified` | 3.0320 | n/a | 0.0297 / 102.09x | 0.0488 / 62.13x | `miss` | — |

## Source scorecards

- `v12/docs/perf-baselines/2026-07-21-post-i64-compiled-frontier-comparison.json` — `coverage` (`2026-07-21T12:04:45.151360Z`)
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
- `v12/docs/perf-baselines/2026-07-21-interaction-promotion-bytecode.json` — `custom` (`2026-07-21T05:23:23.328393Z`)

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
