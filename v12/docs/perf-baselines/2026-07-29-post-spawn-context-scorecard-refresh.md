# External Application Scoreboard

- Source measurements through: `2026-07-29T16:00:48.159470Z`
- Scope: verifier-backed Able application launches only; tree-walker is intentionally excluded from this performance target.
- Guard: each source scorecard records its process count, CPU-affinity when used, runtime settings, and per-process timeout.
- Compiled: `7/63` selected rankable rows meet the 95%-of-Go target.
- Bytecode: `4/63` selected rankable rows meet both 95%-of-Python and 95%-of-Ruby targets.
- Canonical Able source fingerprints: `126` row fingerprints in JSON; `126` came from the measured source report and the remainder are current-source legacy fingerprints.
- Verifier/declared-input contracts: `126` row fingerprints in JSON; `126` were captured before the timed launch and the remainder are current-contract legacy reconstructions.
- Canonical stdlib runtime sources: `70` `.able` files, tree SHA-256 `6a412c872ee66752de7c4417a5eda99806a7631da3b880c55574c6a640b82d9b`; Git `219eff222c28406487231713753641bc49ee5b9a` (dirty).
- Strict candidate selection: `126` reviewed benchmark/mode rows, SHA-256 `0c72eaf2a1b12d3a5a2f88d00b3382a706f5c5c16977c24b55fb64214f8d429e`; timeout rows remain in full status.
- Matched reference source fingerprints: `189` comparison fingerprints in JSON; `189` came from measured reference reports and the remainder are current-source legacy fingerprints.
- `unranked` means a partial, timed-out, failed, or unavailable matched run/reference; it is never counted as a pass or fail.
- `Unranked reason` identifies whether the Able launch or its required reference prevents ranking; reference-unavailable does not infer why that source has no valid ratio.

| Benchmark | Mode | Able status | Able (s) | Go / ratio | Python / ratio | Ruby / ratio | Target | Unranked reason |
| --- | --- | --- | ---: | --- | --- | --- | --- | --- |
| `backup_dedup` | `compiled` | `verified` | 0.0540 | 0.0094 / 5.74x | n/a | n/a | `miss` | — |
| `fixed_width_128` | `compiled` | `verified` | 0.1180 | 0.0076 / 15.53x | n/a | n/a | `miss` | — |
| `rational_series` | `compiled` | `verified` | 0.0520 | 0.0158 / 3.29x | n/a | n/a | `miss` | — |
| `wide_integer_records` | `compiled` | `verified` | 0.0640 | 0.0504 / 1.27x | n/a | n/a | `miss` | — |
| `binary_event_log` | `compiled` | `verified` | 0.1480 | 0.0167 / 8.86x | n/a | n/a | `miss` | — |
| `word_frequency` | `compiled` | `verified` | 0.0380 | 0.0139 / 2.73x | n/a | n/a | `miss` | — |
| `document_audit` | `compiled` | `verified` | 0.0360 | 0.0096 / 3.75x | n/a | n/a | `miss` | — |
| `lexical_rollup` | `compiled` | `verified` | 0.0400 | 0.0102 / 3.92x | n/a | n/a | `miss` | — |
| `regex_suffix_audit` | `compiled` | `verified` | 0.0520 | 0.0110 / 4.73x | n/a | n/a | `miss` | — |
| `regex_set_audit` | `compiled` | `verified` | 0.0540 | 0.0117 / 4.62x | n/a | n/a | `miss` | — |
| `regex_stream_audit` | `compiled` | `verified` | 0.0420 | 0.0109 / 3.85x | n/a | n/a | `miss` | — |
| `log_routing_redaction` | `compiled` | `verified` | 0.0520 | 0.0100 / 5.20x | n/a | n/a | `miss` | — |
| `array_slice_window` | `compiled` | `verified` | 0.0300 | 0.0098 / 3.06x | n/a | n/a | `miss` | — |
| `dependency_plan` | `compiled` | `verified` | 0.0260 | 0.0076 / 3.42x | n/a | n/a | `miss` | — |
| `discrete_event_simulation` | `compiled` | `verified` | 0.0420 | 0.0189 / 2.22x | n/a | n/a | `miss` | — |
| `inventory_reconciliation` | `compiled` | `verified` | 0.1140 | 0.0134 / 8.51x | n/a | n/a | `miss` | — |
| `option_result_config` | `compiled` | `verified` | 0.0420 | 0.0064 / 6.56x | n/a | n/a | `miss` | — |
| `unicode_scalar_pipeline` | `compiled` | `verified` | 0.1380 | 0.0138 / 10.00x | n/a | n/a | `miss` | — |
| `config_validation_extraction` | `compiled` | `verified` | 0.0440 | 0.0086 / 5.12x | n/a | n/a | `miss` | — |
| `quicksort` | `compiled` | `verified` | 1.6260 | 2.3153 / 0.70x | n/a | n/a | `meets` | — |
| `sudoku_masks` | `compiled` | `verified` | 1.4940 | 0.5939 / 2.52x | n/a | n/a | `miss` | — |
| `i_before_e` | `compiled` | `verified` | 0.0620 | 0.0591 / 1.05x | n/a | n/a | `meets` | — |
| `base64` | `compiled` | `verified` | 1.9980 | 2.2859 / 0.87x | n/a | n/a | `meets` | — |
| `json` | `compiled` | `verified` | 0.6200 | 1.3121 / 0.47x | n/a | n/a | `meets` | — |
| `monte_carlo_pi` | `compiled` | `verified` | 0.1360 | 0.1876 / 0.72x | n/a | n/a | `meets` | — |
| `pidigits` | `compiled` | `verified` | 1.0700 | 1.1088 / 0.97x | n/a | n/a | `meets` | — |
| `mandelbrot` | `compiled` | `verified` | 0.0700 | 0.0461 / 1.52x | n/a | n/a | `miss` | — |
| `reverse_complement` | `compiled` | `verified` | 0.0440 | 0.0147 / 2.99x | n/a | n/a | `miss` | — |
| `k_nucleotide` | `compiled` | `verified` | 1.3460 | 0.0504 / 26.71x | n/a | n/a | `miss` | — |
| `fasta_generation` | `compiled` | `verified` | 0.0400 | 0.0130 / 3.08x | n/a | n/a | `miss` | — |
| `nbody` | `compiled` | `verified` | 0.0740 | 0.0311 / 2.38x | n/a | n/a | `miss` | — |
| `tapelang_alphabet` | `compiled` | `verified` | 3.2760 | 2.7348 / 1.20x | n/a | n/a | `miss` | — |
| `distance_field` | `compiled` | `verified` | 0.0300 | 0.0114 / 2.63x | n/a | n/a | `miss` | — |
| `rms_norm` | `compiled` | `verified` | 0.0300 | 0.0108 / 2.78x | n/a | n/a | `miss` | — |
| `fib` | `bytecode` | `verified` | 0.2480 | n/a | 5.2074 / 0.05x | 4.3390 / 0.06x | `meets` | — |
| `binarytrees` | `bytecode` | `verified` | 12.2980 | n/a | 0.5610 / 21.92x | 0.5744 / 21.41x | `miss` | — |
| `matrixmultiply` | `bytecode` | `verified` | 1.0160 | n/a | 3.3338 / 0.30x | 3.1064 / 0.33x | `meets` | — |
| `quicksort` | `bytecode` | `verified` | 12.1780 | n/a | 0.6732 / 18.09x | 0.6615 / 18.41x | `miss` | — |
| `sudoku_masks` | `bytecode` | `verified` | 24.3940 | n/a | 1.7962 / 13.58x | 2.1343 / 11.43x | `miss` | — |
| `i_before_e` | `bytecode` | `verified` | 0.5180 | n/a | 0.0843 / 6.14x | 0.1191 / 4.35x | `miss` | — |
| `base64` | `bytecode` | `verified` | 3.0420 | n/a | 3.8207 / 0.80x | 2.4968 / 1.22x | `miss` | — |
| `json` | `bytecode` | `verified` | 0.8720 | n/a | 2.5929 / 0.34x | 1.6712 / 0.52x | `meets` | — |
| `monte_carlo_pi` | `bytecode` | `verified` | 2.7140 | n/a | 1.4459 / 1.88x | 1.5447 / 1.76x | `miss` | — |
| `pidigits` | `bytecode` | `verified` | 2.4320 | n/a | 4.0342 / 0.60x | 10.0133 / 0.24x | `meets` | — |
| `mandelbrot` | `bytecode` | `verified` | 6.5320 | n/a | 1.3706 / 4.77x | 1.9661 / 3.32x | `miss` | — |
| `reverse_complement` | `bytecode` | `verified` | 3.2920 | n/a | 0.0263 / 125.17x | 0.0748 / 44.01x | `miss` | — |
| `k_nucleotide` | `bytecode` | `verified` | 42.6140 | n/a | 1.3037 / 32.69x | 1.2331 / 34.56x | `miss` | — |
| `fasta_generation` | `bytecode` | `verified` | 1.9020 | n/a | 0.2215 / 8.59x | 0.2976 / 6.39x | `miss` | — |
| `nbody` | `bytecode` | `verified` | 9.1580 | n/a | 0.2002 / 45.74x | 0.3351 / 27.33x | `miss` | — |
| `tapelang_alphabet` | `bytecode` | `verified` | 20.5300 | n/a | 0.5897 / 34.81x | 0.7746 / 26.50x | `miss` | — |
| `distance_field` | `bytecode` | `verified` | 5.7600 | n/a | 0.5807 / 9.92x | 0.3396 / 16.96x | `miss` | — |
| `rms_norm` | `bytecode` | `verified` | 4.6820 | n/a | 0.7947 / 5.89x | 0.5263 / 8.90x | `miss` | — |
| `channel_rollup` | `bytecode` | `verified` | 0.4440 | n/a | 0.0388 / 11.44x | 0.0518 / 8.57x | `miss` | — |
| `future_pipeline` | `bytecode` | `verified` | 0.6640 | n/a | 0.0570 / 11.65x | 0.0673 / 9.87x | `miss` | — |
| `future_await_race` | `bytecode` | `verified` | 0.1420 | n/a | 0.0310 / 4.58x | 0.0534 / 2.66x | `miss` | — |
| `await_channel_mux` | `bytecode` | `verified` | 0.2280 | n/a | 0.1149 / 1.98x | 0.0909 / 2.51x | `miss` | — |
| `mutex_ledger` | `bytecode` | `verified` | 0.4120 | n/a | 0.0364 / 11.32x | 0.0580 / 7.10x | `miss` | — |
| `mutex_await_journal` | `bytecode` | `verified` | 0.2320 | n/a | 0.0223 / 10.40x | 0.0512 / 4.53x | `miss` | — |
| `mutex_work_queue` | `bytecode` | `verified` | 0.3740 | n/a | 0.0278 / 13.45x | 0.0508 / 7.36x | `miss` | — |
| `backup_dedup` | `bytecode` | `verified` | 1.9140 | n/a | 0.2661 / 7.19x | 0.1320 / 14.50x | `miss` | — |
| `fixed_width_128` | `bytecode` | `verified` | 8.0300 | n/a | 0.3702 / 21.69x | 0.6500 / 12.35x | `miss` | — |
| `rational_series` | `bytecode` | `verified` | 4.1460 | n/a | 0.1082 / 38.32x | 0.1302 / 31.84x | `miss` | — |
| `wide_integer_records` | `bytecode` | `verified` | 5.5700 | n/a | 0.0650 / 85.69x | 0.1352 / 41.20x | `miss` | — |
| `binary_event_log` | `bytecode` | `verified` | 5.8360 | n/a | 0.1818 / 32.10x | 0.2483 / 23.50x | `miss` | — |
| `word_frequency` | `bytecode` | `verified` | 1.5020 | n/a | 0.0204 / 73.63x | 0.0551 / 27.26x | `miss` | — |
| `document_audit` | `bytecode` | `verified` | 0.3000 | n/a | 0.0138 / 21.74x | 0.0429 / 6.99x | `miss` | — |
| `lexical_rollup` | `bytecode` | `verified` | 0.3860 | n/a | 0.0166 / 23.25x | 0.0463 / 8.34x | `miss` | — |
| `regex_suffix_audit` | `bytecode` | `verified` | 3.4480 | n/a | 0.0172 / 200.47x | 0.0400 / 86.20x | `miss` | — |
| `regex_set_audit` | `bytecode` | `verified` | 4.1300 | n/a | 0.0186 / 222.04x | 0.0469 / 88.06x | `miss` | — |
| `regex_stream_audit` | `bytecode` | `verified` | 3.6160 | n/a | 0.0179 / 202.01x | 0.0417 / 86.71x | `miss` | — |
| `log_routing_redaction` | `bytecode` | `verified` | 3.1140 | n/a | 0.0177 / 175.93x | 0.0426 / 73.10x | `miss` | — |
| `array_slice_window` | `bytecode` | `verified` | 0.6860 | n/a | 0.0272 / 25.22x | 0.0611 / 11.23x | `miss` | — |
| `dependency_plan` | `bytecode` | `verified` | 0.4940 | n/a | 0.0180 / 27.44x | 0.0492 / 10.04x | `miss` | — |
| `discrete_event_simulation` | `bytecode` | `verified` | 4.7040 | n/a | 0.1737 / 27.08x | 0.2210 / 21.29x | `miss` | — |
| `inventory_reconciliation` | `bytecode` | `verified` | 2.4900 | n/a | 0.0698 / 35.67x | 0.0898 / 27.73x | `miss` | — |
| `option_result_config` | `bytecode` | `verified` | 0.7400 | n/a | 0.0179 / 41.34x | 0.0495 / 14.95x | `miss` | — |
| `unicode_scalar_pipeline` | `bytecode` | `verified` | 3.8120 | n/a | 0.2359 / 16.16x | 0.3176 / 12.00x | `miss` | — |
| `config_validation_extraction` | `bytecode` | `verified` | 1.3640 | n/a | 0.0187 / 72.94x | 0.0467 / 29.21x | `miss` | — |
| `concurrent_text_index` | `bytecode` | `verified` | 0.6020 | n/a | 0.0584 / 10.31x | 0.0783 / 7.69x | `miss` | — |
| `validated_job_pipeline` | `bytecode` | `verified` | 0.3700 | n/a | 0.0232 / 15.95x | 0.0485 / 7.63x | `miss` | — |
| `dependency_wave_validation` | `bytecode` | `verified` | 0.5080 | n/a | 0.0312 / 16.28x | 0.0494 / 10.28x | `miss` | — |
| `concurrent_event_routing` | `bytecode` | `verified` | 2.9540 | n/a | 0.0312 / 94.68x | 0.0578 / 51.11x | `miss` | — |
| `concurrent_document_pipeline` | `bytecode` | `verified` | 0.2860 | n/a | 0.0246 / 11.63x | 0.0509 / 5.62x | `miss` | — |
| `manifest_normalization` | `bytecode` | `verified` | 1.5360 | n/a | 0.0172 / 89.30x | 0.0522 / 29.43x | `miss` | — |
| `policy_record_dispatch` | `bytecode` | `verified` | 7.5820 | n/a | 0.0202 / 375.35x | 0.0450 / 168.49x | `miss` | — |
| `sensor_calibration` | `bytecode` | `verified` | 2.6240 | n/a | 0.0277 / 94.73x | 0.0715 / 36.70x | `miss` | — |
| `concurrent_stencil_reduction` | `bytecode` | `verified` | 1.8100 | n/a | 0.0751 / 24.10x | 0.0956 / 18.93x | `miss` | — |
| `concurrent_signal_dispatch` | `bytecode` | `verified` | 1.5600 | n/a | 0.0631 / 24.72x | 0.1054 / 14.80x | `miss` | — |
| `concurrent_transform_chain` | `bytecode` | `verified` | 2.7360 | n/a | 0.1521 / 17.99x | 0.1392 / 19.66x | `miss` | — |
| `concurrent_policy_callbacks` | `bytecode` | `verified` | 0.3860 | n/a | 0.0493 / 7.83x | 0.0657 / 5.88x | `miss` | — |
| `concurrent_graph_visitors` | `bytecode` | `verified` | 1.2900 | n/a | 0.0762 / 16.93x | 0.0640 / 20.16x | `miss` | — |
| `concurrent_audio_voices` | `bytecode` | `verified` | 1.3540 | n/a | 0.1256 / 10.78x | 0.1269 / 10.67x | `miss` | — |
| `concurrent_packet_codecs` | `bytecode` | `verified` | 0.8380 | n/a | 0.0779 / 10.76x | 0.0822 / 10.19x | `miss` | — |
| `concurrent_scene_tiles` | `bytecode` | `verified` | 0.6400 | n/a | 0.0717 / 8.93x | 0.0711 / 9.00x | `miss` | — |
| `concurrent_tree_folds` | `bytecode` | `verified` | 0.4260 | n/a | 0.0658 / 6.47x | 0.0607 / 7.02x | `miss` | — |
| `concurrent_state_machines` | `bytecode` | `verified` | 0.3900 | n/a | 0.0609 / 6.40x | 0.0613 / 6.36x | `miss` | — |
| `concurrent_stateful_pipeline` | `bytecode` | `verified` | 0.4560 | n/a | 0.0679 / 6.72x | 0.0561 / 8.13x | `miss` | — |
| `binarytrees` | `compiled` | `verified` | 10.2700 | 11.0316 / 0.93x | n/a | n/a | `meets` | — |
| `fib` | `compiled` | `verified` | 3.1180 | 2.8996 / 1.08x | n/a | n/a | `miss` | — |
| `matrixmultiply` | `compiled` | `verified` | 0.9680 | 0.9040 / 1.07x | n/a | n/a | `miss` | — |
| `channel_rollup` | `compiled` | `verified` | 0.0480 | 0.0057 / 8.42x | n/a | n/a | `miss` | — |
| `future_pipeline` | `compiled` | `verified` | 0.0300 | 0.0061 / 4.92x | n/a | n/a | `miss` | — |
| `future_await_race` | `compiled` | `verified` | 0.0340 | 0.0049 / 6.94x | n/a | n/a | `miss` | — |
| `mutex_ledger` | `compiled` | `verified` | 0.0300 | 0.0054 / 5.56x | n/a | n/a | `miss` | — |
| `await_channel_mux` | `compiled` | `verified` | 0.1280 | 0.0055 / 23.27x | n/a | n/a | `miss` | — |
| `mutex_await_journal` | `compiled` | `verified` | 0.0340 | 0.0052 / 6.54x | n/a | n/a | `miss` | — |
| `mutex_work_queue` | `compiled` | `verified` | 0.0400 | 0.0055 / 7.27x | n/a | n/a | `miss` | — |
| `concurrent_text_index` | `compiled` | `verified` | 0.0520 | 0.0072 / 7.22x | n/a | n/a | `miss` | — |
| `validated_job_pipeline` | `compiled` | `verified` | 0.0740 | 0.0049 / 15.10x | n/a | n/a | `miss` | — |
| `dependency_wave_validation` | `compiled` | `verified` | 0.0360 | 0.0049 / 7.35x | n/a | n/a | `miss` | — |
| `concurrent_event_routing` | `compiled` | `verified` | 0.0420 | 0.0048 / 8.75x | n/a | n/a | `miss` | — |
| `concurrent_document_pipeline` | `compiled` | `verified` | 0.0320 | 0.0046 / 6.96x | n/a | n/a | `miss` | — |
| `concurrent_stencil_reduction` | `compiled` | `verified` | 0.0400 | 0.0055 / 7.27x | n/a | n/a | `miss` | — |
| `concurrent_signal_dispatch` | `compiled` | `verified` | 0.0440 | 0.0055 / 8.00x | n/a | n/a | `miss` | — |
| `concurrent_transform_chain` | `compiled` | `verified` | 0.0440 | 0.0057 / 7.72x | n/a | n/a | `miss` | — |
| `concurrent_policy_callbacks` | `compiled` | `verified` | 0.0340 | 0.0047 / 7.23x | n/a | n/a | `miss` | — |
| `concurrent_graph_visitors` | `compiled` | `verified` | 0.0420 | 0.0045 / 9.33x | n/a | n/a | `miss` | — |
| `concurrent_audio_voices` | `compiled` | `verified` | 0.0400 | 0.0047 / 8.51x | n/a | n/a | `miss` | — |
| `concurrent_packet_codecs` | `compiled` | `verified` | 0.0320 | 0.0047 / 6.81x | n/a | n/a | `miss` | — |
| `concurrent_scene_tiles` | `compiled` | `verified` | 0.0360 | 0.0043 / 8.37x | n/a | n/a | `miss` | — |
| `concurrent_tree_folds` | `compiled` | `verified` | 0.0300 | 0.0040 / 7.50x | n/a | n/a | `miss` | — |
| `concurrent_state_machines` | `compiled` | `verified` | 0.0340 | 0.0041 / 8.29x | n/a | n/a | `miss` | — |
| `concurrent_stateful_pipeline` | `compiled` | `verified` | 0.0600 | 0.0052 / 11.54x | n/a | n/a | `miss` | — |
| `manifest_normalization` | `compiled` | `verified` | 0.0420 | 0.0060 / 7.00x | n/a | n/a | `miss` | — |
| `policy_record_dispatch` | `compiled` | `verified` | 0.0820 | 0.0057 / 14.39x | n/a | n/a | `miss` | — |
| `sensor_calibration` | `compiled` | `verified` | 0.0420 | 0.0061 / 6.89x | n/a | n/a | `miss` | — |

## Source scorecards

- `v12/docs/perf-baselines/2026-07-29-compiled-go1265-coverage-extra-compiled-01-selected.json` — `custom` (`2026-07-29T05:56:24.045791Z`)
- `v12/docs/perf-baselines/2026-07-29-compiled-go1265-coverage-extra-compiled-02-selected.json` — `custom` (`2026-07-29T05:59:05.828946Z`)
- `v12/docs/perf-baselines/2026-07-29-compiled-go1265-coverage-extra-compiled-03-selected.json` — `custom` (`2026-07-29T06:03:20.568558Z`)
- `v12/docs/perf-baselines/2026-07-29-compiled-go1265-coverage-extra-compiled-04-selected.json` — `custom` (`2026-07-29T06:04:56.432109Z`)
- `v12/docs/perf-baselines/2026-07-29-compiled-go1265-coverage-extra-compiled-05-selected.json` — `custom` (`2026-07-29T06:06:15.116812Z`)
- `v12/docs/perf-baselines/2026-07-29-compiled-go1265-generality-compiled-02-selected.json` — `custom` (`2026-07-29T05:45:49.774683Z`)
- `v12/docs/perf-baselines/2026-07-29-compiled-go1265-generality-compiled-03-selected.json` — `custom` (`2026-07-29T05:47:33.637330Z`)
- `v12/docs/perf-baselines/2026-07-29-compiled-go1265-generality-compiled-04-selected.json` — `custom` (`2026-07-29T05:48:45.368007Z`)
- `v12/docs/perf-baselines/2026-07-29-compiled-go1265-generality-compiled-05-selected.json` — `custom` (`2026-07-29T05:50:23.880559Z`)
- `v12/docs/perf-baselines/2026-07-29-compiled-go1265-generality-compiled-06-selected.json` — `custom` (`2026-07-29T05:51:40.820532Z`)
- `v12/docs/perf-baselines/2026-07-29-compiled-go1265-generality-compiled-07-selected.json` — `custom` (`2026-07-29T05:52:11.817179Z`)
- `v12/docs/perf-baselines/2026-07-28-mode-aware-benchmark-contract-generality-bytecode-01-selected.json` — `custom` (`2026-07-28T19:41:52.326251Z`)
- `v12/docs/perf-baselines/2026-07-28-mode-aware-benchmark-contract-generality-bytecode-02-selected.json` — `custom` (`2026-07-28T19:45:00.133335Z`)
- `v12/docs/perf-baselines/2026-07-28-mode-aware-benchmark-contract-generality-bytecode-03-selected.json` — `custom` (`2026-07-28T19:45:41.588280Z`)
- `v12/docs/perf-baselines/2026-07-28-mode-aware-benchmark-contract-generality-bytecode-04-selected.json` — `custom` (`2026-07-28T19:46:47.239929Z`)
- `v12/docs/perf-baselines/2026-07-28-mode-aware-benchmark-contract-generality-bytecode-05-selected.json` — `custom` (`2026-07-28T19:50:53.692222Z`)
- `v12/docs/perf-baselines/2026-07-28-mode-aware-benchmark-contract-generality-bytecode-06-selected.json` — `custom` (`2026-07-28T19:53:26.887713Z`)
- `v12/docs/perf-baselines/2026-07-28-mode-aware-benchmark-contract-generality-bytecode-07-selected.json` — `custom` (`2026-07-28T19:54:23.937604Z`)
- `v12/docs/perf-baselines/2026-07-28-mode-aware-benchmark-contract-async-bytecode-01-selected.json` — `custom` (`2026-07-28T19:55:48.412153Z`)
- `v12/docs/perf-baselines/2026-07-28-mode-aware-benchmark-contract-async-bytecode-02-selected.json` — `custom` (`2026-07-28T19:56:04.829950Z`)
- `v12/docs/perf-baselines/2026-07-28-mode-aware-benchmark-contract-coverage-extra-bytecode-01-selected.json` — `custom` (`2026-07-28T20:20:04.924888Z`)
- `v12/docs/perf-baselines/2026-07-28-mode-aware-benchmark-contract-coverage-extra-bytecode-02-selected.json` — `custom` (`2026-07-28T20:20:32.778538Z`)
- `v12/docs/perf-baselines/2026-07-28-mode-aware-benchmark-contract-coverage-extra-bytecode-03-selected.json` — `custom` (`2026-07-28T20:21:39.948693Z`)
- `v12/docs/perf-baselines/2026-07-28-mode-aware-benchmark-contract-coverage-extra-bytecode-04-selected.json` — `custom` (`2026-07-28T20:22:53.054504Z`)
- `v12/docs/perf-baselines/2026-07-28-mode-aware-benchmark-contract-coverage-extra-bytecode-05-selected.json` — `custom` (`2026-07-28T20:23:02.540091Z`)
- `v12/docs/perf-baselines/2026-07-28-mode-aware-benchmark-contract-coverage-extra-bytecode-06-selected.json` — `custom` (`2026-07-28T20:26:09.989405Z`)
- `v12/docs/perf-baselines/2026-07-29-post-spawn-context-generality-01-compiled-selected.json` — `custom` (`2026-07-29T15:50:43.463377Z`)
- `v12/docs/perf-baselines/2026-07-29-post-spawn-context-generality-01-compiled-preserved-controls.json` — `custom` (`2026-07-29T05:44:31.578614Z`)
- `v12/docs/perf-baselines/2026-07-29-post-spawn-context-async-01-compiled-selected.json` — `custom` (`2026-07-29T15:51:26.928101Z`)
- `v12/docs/perf-baselines/2026-07-29-post-spawn-context-async-01-compiled-preserved-controls.json` — `custom` (`2026-07-29T12:07:28.179055Z`)
- `v12/docs/perf-baselines/2026-07-29-post-spawn-context-async-02-compiled-selected.json` — `custom` (`2026-07-29T15:52:09.112118Z`)
- `v12/docs/perf-baselines/2026-07-29-post-spawn-context-async-02-compiled-preserved-controls.json` — `custom` (`2026-07-29T12:08:10.565700Z`)
- `v12/docs/perf-baselines/2026-07-29-post-spawn-context-coverage-extra-06-compiled-selected.json` — `custom` (`2026-07-29T16:00:48.159470Z`)
- `v12/docs/perf-baselines/2026-07-29-post-spawn-context-coverage-extra-06-compiled-preserved-controls.json` — `custom` (`2026-07-29T06:13:35.851464Z`)

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
