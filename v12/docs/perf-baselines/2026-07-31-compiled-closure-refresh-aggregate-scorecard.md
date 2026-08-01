# External Application Scoreboard

- Source measurements through: `2026-07-31T22:45:38.198870Z`
- Scope: verifier-backed Able application launches only; tree-walker is intentionally excluded from this performance target.
- Guard: each source scorecard records its process count, CPU-affinity when used, runtime settings, and per-process timeout.
- Compiled: `9/66` selected rankable rows meet the 95%-of-Go target.
- Bytecode: `4/66` selected rankable rows meet both 95%-of-Python and 95%-of-Ruby targets.
- Canonical Able source fingerprints: `132` row fingerprints in JSON; `132` came from the measured source report and the remainder are current-source legacy fingerprints.
- Verifier/declared-input contracts: `132` row fingerprints in JSON; `132` were captured before the timed launch and the remainder are current-contract legacy reconstructions.
- Canonical stdlib runtime sources: `70` `.able` files, tree SHA-256 `6a412c872ee66752de7c4417a5eda99806a7631da3b880c55574c6a640b82d9b`; Git `219eff222c28406487231713753641bc49ee5b9a` (dirty).
- Strict candidate selection: `132` reviewed benchmark/mode rows, SHA-256 `17d7babe33c64c1f17eef97eaabf7bbfba156b0bba20062fd7617e32b259f7fb`; timeout rows remain in full status.
- Matched reference source fingerprints: `198` comparison fingerprints in JSON; `198` came from measured reference reports and the remainder are current-source legacy fingerprints.
- `unranked` means a partial, timed-out, failed, or unavailable matched run/reference; it is never counted as a pass or fail.
- `Unranked reason` identifies whether the Able launch or its required reference prevents ranking; reference-unavailable does not infer why that source has no valid ratio.

| Benchmark | Mode | Able status | Able (s) | Go / ratio | Python / ratio | Ruby / ratio | Target | Unranked reason |
| --- | --- | --- | ---: | --- | --- | --- | --- | --- |
| `array_slice_window` | `compiled` | `verified` | 0.0360 | 0.0050 / 7.20x | n/a | n/a | `miss` | — |
| `await_channel_mux` | `compiled` | `verified` | 0.1120 | 0.0047 / 23.83x | n/a | n/a | `miss` | — |
| `backup_dedup` | `compiled` | `verified` | 0.0420 | 0.0138 / 3.04x | n/a | n/a | `miss` | — |
| `base64` | `compiled` | `verified` | 2.3560 | 2.4163 / 0.98x | n/a | n/a | `meets` | — |
| `binary_event_log` | `compiled` | `verified` | 0.0720 | 0.0080 / 9.00x | n/a | n/a | `miss` | — |
| `binarytrees` | `compiled` | `verified` | 7.3740 | 11.9591 / 0.62x | n/a | n/a | `meets` | — |
| `channel_rollup` | `compiled` | `verified` | 0.0400 | 0.0061 / 6.56x | n/a | n/a | `miss` | — |
| `concurrent_audio_voices` | `compiled` | `verified` | 0.0300 | 0.0053 / 5.66x | n/a | n/a | `miss` | — |
| `concurrent_document_pipeline` | `compiled` | `verified` | 0.0260 | 0.0047 / 5.53x | n/a | n/a | `miss` | — |
| `concurrent_event_routing` | `compiled` | `verified` | 0.0340 | 0.0064 / 5.31x | n/a | n/a | `miss` | — |
| `concurrent_graph_visitors` | `compiled` | `verified` | 0.0200 | 0.0048 / 4.17x | n/a | n/a | `miss` | — |
| `concurrent_packet_codecs` | `compiled` | `verified` | 0.0300 | 0.0048 / 6.25x | n/a | n/a | `miss` | — |
| `concurrent_policy_callbacks` | `compiled` | `verified` | 0.0320 | 0.0070 / 4.57x | n/a | n/a | `miss` | — |
| `concurrent_scene_tiles` | `compiled` | `verified` | 0.0300 | 0.0052 / 5.77x | n/a | n/a | `miss` | — |
| `concurrent_signal_dispatch` | `compiled` | `verified` | 0.0360 | 0.0058 / 6.21x | n/a | n/a | `miss` | — |
| `concurrent_state_machines` | `compiled` | `verified` | 0.0320 | 0.0065 / 4.92x | n/a | n/a | `miss` | — |
| `concurrent_stateful_pipeline` | `compiled` | `verified` | 0.0640 | 0.0074 / 8.65x | n/a | n/a | `miss` | — |
| `concurrent_stencil_reduction` | `compiled` | `verified` | 0.0340 | 0.0061 / 5.57x | n/a | n/a | `miss` | — |
| `concurrent_text_index` | `compiled` | `verified` | 0.0400 | 0.0123 / 3.25x | n/a | n/a | `miss` | — |
| `concurrent_transform_chain` | `compiled` | `verified` | 0.0300 | 0.0068 / 4.41x | n/a | n/a | `miss` | — |
| `concurrent_tree_folds` | `compiled` | `verified` | 0.0280 | 0.0049 / 5.71x | n/a | n/a | `miss` | — |
| `config_validation_extraction` | `compiled` | `verified` | 0.0360 | 0.0057 / 6.32x | n/a | n/a | `miss` | — |
| `dependency_plan` | `compiled` | `verified` | 0.0360 | 0.0054 / 6.67x | n/a | n/a | `miss` | — |
| `dependency_wave_validation` | `compiled` | `verified` | 0.0420 | 0.0055 / 7.64x | n/a | n/a | `miss` | — |
| `discrete_event_simulation` | `compiled` | `verified` | 0.0400 | 0.0148 / 2.70x | n/a | n/a | `miss` | — |
| `distance_field` | `compiled` | `verified` | 0.0600 | 0.0156 / 3.85x | n/a | n/a | `miss` | — |
| `document_audit` | `compiled` | `verified` | 0.0200 | 0.0055 / 3.64x | n/a | n/a | `miss` | — |
| `fasta_generation` | `compiled` | `verified` | 0.0400 | 0.0173 / 2.31x | n/a | n/a | `miss` | — |
| `fib` | `compiled` | `verified` | 3.4160 | 3.5367 / 0.97x | n/a | n/a | `meets` | — |
| `fixed_width_128` | `compiled` | `verified` | 0.0840 | 0.0064 / 13.12x | n/a | n/a | `miss` | — |
| `future_await_race` | `compiled` | `verified` | 0.0300 | 0.0045 / 6.67x | n/a | n/a | `miss` | — |
| `future_pipeline` | `compiled` | `verified` | 0.0300 | 0.0070 / 4.29x | n/a | n/a | `miss` | — |
| `generic_slot_buffer` | `compiled` | `verified` | 0.0300 | 0.0058 / 5.17x | n/a | n/a | `miss` | — |
| `i_before_e` | `compiled` | `verified` | 0.0400 | 0.0718 / 0.56x | n/a | n/a | `meets` | — |
| `inventory_reconciliation` | `compiled` | `verified` | 0.1000 | 0.0092 / 10.87x | n/a | n/a | `miss` | — |
| `json` | `compiled` | `verified` | 0.5600 | 1.5717 / 0.36x | n/a | n/a | `meets` | — |
| `k_nucleotide` | `compiled` | `verified` | 1.3040 | 0.0620 / 21.03x | n/a | n/a | `miss` | — |
| `lexical_rollup` | `compiled` | `verified` | 0.0400 | 0.0041 / 9.76x | n/a | n/a | `miss` | — |
| `log_routing_redaction` | `compiled` | `verified` | 0.0400 | 0.0049 / 8.16x | n/a | n/a | `miss` | — |
| `mandelbrot` | `compiled` | `verified` | 0.0720 | 0.0620 / 1.16x | n/a | n/a | `miss` | — |
| `manifest_normalization` | `compiled` | `verified` | 0.0300 | 0.0043 / 6.98x | n/a | n/a | `miss` | — |
| `matrixmultiply` | `compiled` | `verified` | 1.0400 | 1.0344 / 1.01x | n/a | n/a | `meets` | — |
| `monte_carlo_pi` | `compiled` | `verified` | 0.1400 | 0.2537 / 0.55x | n/a | n/a | `meets` | — |
| `mutex_await_journal` | `compiled` | `verified` | 0.0220 | 0.0046 / 4.78x | n/a | n/a | `miss` | — |
| `mutex_ledger` | `compiled` | `verified` | 0.0300 | 0.0053 / 5.66x | n/a | n/a | `miss` | — |
| `mutex_work_queue` | `compiled` | `verified` | 0.0300 | 0.0050 / 6.00x | n/a | n/a | `miss` | — |
| `nbody` | `compiled` | `verified` | 0.0800 | 0.0387 / 2.07x | n/a | n/a | `miss` | — |
| `option_result_config` | `compiled` | `verified` | 0.0420 | 0.0038 / 11.05x | n/a | n/a | `miss` | — |
| `pidigits` | `compiled` | `verified` | 1.2000 | 1.2436 / 0.96x | n/a | n/a | `meets` | — |
| `policy_record_dispatch` | `compiled` | `verified` | 0.0680 | 0.0053 / 12.83x | n/a | n/a | `miss` | — |
| `quicksort` | `compiled` | `verified` | 1.7820 | 2.7145 / 0.66x | n/a | n/a | `meets` | — |
| `rational_series` | `compiled` | `verified` | 0.0500 | 0.0135 / 3.70x | n/a | n/a | `miss` | — |
| `regex_set_audit` | `compiled` | `verified` | 0.0460 | 0.0046 / 10.00x | n/a | n/a | `miss` | — |
| `regex_stream_audit` | `compiled` | `verified` | 0.0480 | 0.0046 / 10.43x | n/a | n/a | `miss` | — |
| `regex_suffix_audit` | `compiled` | `verified` | 0.0440 | 0.0047 / 9.36x | n/a | n/a | `miss` | — |
| `reverse_complement` | `compiled` | `verified` | 0.0400 | 0.0172 / 2.33x | n/a | n/a | `miss` | — |
| `rms_norm` | `compiled` | `verified` | 0.0400 | 0.0126 / 3.17x | n/a | n/a | `miss` | — |
| `sensor_calibration` | `compiled` | `verified` | 0.0340 | 0.0055 / 6.18x | n/a | n/a | `miss` | — |
| `sudoku_masks` | `compiled` | `verified` | 1.6020 | 0.7574 / 2.12x | n/a | n/a | `miss` | — |
| `tapelang_alphabet` | `compiled` | `verified` | 3.6680 | 3.2099 / 1.14x | n/a | n/a | `miss` | — |
| `transaction_ledger_audit` | `compiled` | `verified` | 0.0300 | 0.0071 / 4.23x | n/a | n/a | `miss` | — |
| `unicode_scalar_pipeline` | `compiled` | `verified` | 0.1020 | 0.0098 / 10.41x | n/a | n/a | `miss` | — |
| `validated_job_pipeline` | `compiled` | `verified` | 0.0600 | 0.0043 / 13.95x | n/a | n/a | `miss` | — |
| `versioned_telemetry_pipeline` | `compiled` | `verified` | 1.9880 | 0.2057 / 9.66x | n/a | n/a | `miss` | — |
| `wide_integer_records` | `compiled` | `verified` | 0.0600 | 0.0278 / 2.16x | n/a | n/a | `miss` | — |
| `word_frequency` | `compiled` | `verified` | 0.0320 | 0.0059 / 5.42x | n/a | n/a | `miss` | — |
| `fib` | `bytecode` | `verified` | 0.2360 | n/a | 5.4730 / 0.04x | 4.2566 / 0.06x | `meets` | — |
| `binarytrees` | `bytecode` | `verified` | 16.2000 | n/a | 0.5667 / 28.59x | 0.6081 / 26.64x | `miss` | — |
| `matrixmultiply` | `bytecode` | `verified` | 1.1460 | n/a | 3.1299 / 0.37x | 3.2610 / 0.35x | `meets` | — |
| `quicksort` | `bytecode` | `verified` | 13.3140 | n/a | 1.1138 / 11.95x | 1.2566 / 10.60x | `miss` | — |
| `sudoku_masks` | `bytecode` | `verified` | 28.0120 | n/a | 2.8424 / 9.86x | 2.5671 / 10.91x | `miss` | — |
| `monte_carlo_pi` | `bytecode` | `verified` | 3.1920 | n/a | 1.5028 / 2.12x | 1.6149 / 1.98x | `miss` | — |
| `pidigits` | `bytecode` | `verified` | 3.3560 | n/a | 4.2060 / 0.80x | 10.4238 / 0.32x | `meets` | — |
| `mandelbrot` | `bytecode` | `verified` | 10.5340 | n/a | 1.2522 / 8.41x | 1.9620 / 5.37x | `miss` | — |
| `reverse_complement` | `bytecode` | `verified` | 4.1200 | n/a | 0.0254 / 162.20x | 0.0747 / 55.15x | `miss` | — |
| `k_nucleotide` | `bytecode` | `verified` | 52.9380 | n/a | 1.3863 / 38.19x | 1.2642 / 41.87x | `miss` | — |
| `fasta_generation` | `bytecode` | `verified` | 1.9840 | n/a | 0.2002 / 9.91x | 0.3027 / 6.55x | `miss` | — |
| `nbody` | `bytecode` | `verified` | 9.9220 | n/a | 0.2174 / 45.64x | 0.3560 / 27.87x | `miss` | — |
| `tapelang_alphabet` | `bytecode` | `verified` | 25.2740 | n/a | 0.6164 / 41.00x | 0.7878 / 32.08x | `miss` | — |
| `distance_field` | `bytecode` | `verified` | 6.2600 | n/a | 0.5909 / 10.59x | 0.3238 / 19.33x | `miss` | — |
| `rms_norm` | `bytecode` | `verified` | 5.5480 | n/a | 0.8329 / 6.66x | 0.5018 / 11.06x | `miss` | — |
| `channel_rollup` | `bytecode` | `verified` | 0.5320 | n/a | 0.0417 / 12.76x | 0.0556 / 9.57x | `miss` | — |
| `future_pipeline` | `bytecode` | `verified` | 0.5980 | n/a | 0.0618 / 9.68x | 0.0797 / 7.50x | `miss` | — |
| `future_await_race` | `bytecode` | `verified` | 0.1840 | n/a | 0.0336 / 5.48x | 0.0617 / 2.98x | `miss` | — |
| `await_channel_mux` | `bytecode` | `verified` | 0.3640 | n/a | 0.1341 / 2.71x | 0.1066 / 3.41x | `miss` | — |
| `mutex_ledger` | `bytecode` | `verified` | 0.7760 | n/a | 0.0364 / 21.32x | 0.0640 / 12.12x | `miss` | — |
| `mutex_await_journal` | `bytecode` | `verified` | 0.3540 | n/a | 0.0241 / 14.69x | 0.0526 / 6.73x | `miss` | — |
| `mutex_work_queue` | `bytecode` | `verified` | 0.6240 | n/a | 0.0289 / 21.59x | 0.0512 / 12.19x | `miss` | — |
| `backup_dedup` | `bytecode` | `verified` | 3.3380 | n/a | 0.2664 / 12.53x | 0.1331 / 25.08x | `miss` | — |
| `fixed_width_128` | `bytecode` | `verified` | 9.1720 | n/a | 0.3484 / 26.33x | 0.6166 / 14.88x | `miss` | — |
| `rational_series` | `bytecode` | `verified` | 4.2440 | n/a | 0.0966 / 43.93x | 0.1342 / 31.62x | `miss` | — |
| `wide_integer_records` | `bytecode` | `verified` | 5.2980 | n/a | 0.0646 / 82.01x | 0.1490 / 35.56x | `miss` | — |
| `binary_event_log` | `bytecode` | `verified` | 5.9360 | n/a | 0.2003 / 29.64x | 0.2840 / 20.90x | `miss` | — |
| `word_frequency` | `bytecode` | `verified` | 1.4500 | n/a | 0.0244 / 59.43x | 0.0597 / 24.29x | `miss` | — |
| `document_audit` | `bytecode` | `verified` | 0.3140 | n/a | 0.0147 / 21.36x | 0.0439 / 7.15x | `miss` | — |
| `lexical_rollup` | `bytecode` | `verified` | 0.4000 | n/a | 0.0197 / 20.30x | 0.0575 / 6.96x | `miss` | — |
| `regex_suffix_audit` | `bytecode` | `verified` | 3.5420 | n/a | 0.0201 / 176.22x | 0.0461 / 76.83x | `miss` | — |
| `regex_set_audit` | `bytecode` | `verified` | 4.2560 | n/a | 0.0206 / 206.60x | 0.0440 / 96.73x | `miss` | — |
| `regex_stream_audit` | `bytecode` | `verified` | 3.5860 | n/a | 0.0184 / 194.89x | 0.0425 / 84.38x | `miss` | — |
| `log_routing_redaction` | `bytecode` | `verified` | 3.2960 | n/a | 0.0181 / 182.10x | 0.0462 / 71.34x | `miss` | — |
| `array_slice_window` | `bytecode` | `verified` | 0.6580 | n/a | 0.0281 / 23.42x | 0.0623 / 10.56x | `miss` | — |
| `dependency_plan` | `bytecode` | `verified` | 0.4700 | n/a | 0.0182 / 25.82x | 0.0580 / 8.10x | `miss` | — |
| `discrete_event_simulation` | `bytecode` | `verified` | 4.9920 | n/a | 0.2255 / 22.14x | 0.2424 / 20.59x | `miss` | — |
| `inventory_reconciliation` | `bytecode` | `verified` | 2.7020 | n/a | 0.0728 / 37.12x | 0.0972 / 27.80x | `miss` | — |
| `option_result_config` | `bytecode` | `verified` | 1.0760 | n/a | 0.0292 / 36.85x | 0.0697 / 15.44x | `miss` | — |
| `unicode_scalar_pipeline` | `bytecode` | `verified` | 3.9840 | n/a | 0.3700 / 10.77x | 0.4910 / 8.11x | `miss` | — |
| `config_validation_extraction` | `bytecode` | `verified` | 1.4240 | n/a | 0.0178 / 80.00x | 0.0482 / 29.54x | `miss` | — |
| `concurrent_text_index` | `bytecode` | `verified` | 0.7540 | n/a | 0.1017 / 7.41x | 0.0815 / 9.25x | `miss` | — |
| `validated_job_pipeline` | `bytecode` | `verified` | 0.6260 | n/a | 0.0255 / 24.55x | 0.0595 / 10.52x | `miss` | — |
| `dependency_wave_validation` | `bytecode` | `verified` | 0.8020 | n/a | 0.0323 / 24.83x | 0.0745 / 10.77x | `miss` | — |
| `concurrent_event_routing` | `bytecode` | `verified` | 4.6940 | n/a | 0.0516 / 90.97x | 0.0691 / 67.93x | `miss` | — |
| `concurrent_document_pipeline` | `bytecode` | `verified` | 0.2800 | n/a | 0.0235 / 11.91x | 0.0464 / 6.03x | `miss` | — |
| `manifest_normalization` | `bytecode` | `verified` | 1.4920 | n/a | 0.0209 / 71.39x | 0.0683 / 21.84x | `miss` | — |
| `policy_record_dispatch` | `bytecode` | `verified` | 7.8620 | n/a | 0.0276 / 284.86x | 0.0610 / 128.89x | `miss` | — |
| `sensor_calibration` | `bytecode` | `verified` | 2.8680 | n/a | 0.0322 / 89.07x | 0.0961 / 29.84x | `miss` | — |
| `transaction_ledger_audit` | `bytecode` | `verified` | 4.5880 | n/a | 0.0382 / 120.10x | 0.1235 / 37.15x | `miss` | — |
| `generic_slot_buffer` | `bytecode` | `verified` | 2.2400 | n/a | 0.1923 / 11.65x | 0.1149 / 19.50x | `miss` | — |
| `concurrent_stencil_reduction` | `bytecode` | `verified` | 1.8020 | n/a | 0.0846 / 21.30x | 0.1105 / 16.31x | `miss` | — |
| `concurrent_signal_dispatch` | `bytecode` | `verified` | 1.6100 | n/a | 0.0716 / 22.49x | 0.0887 / 18.15x | `miss` | — |
| `concurrent_transform_chain` | `bytecode` | `verified` | 2.6920 | n/a | 0.1297 / 20.76x | 0.1331 / 20.23x | `miss` | — |
| `concurrent_policy_callbacks` | `bytecode` | `verified` | 0.4280 | n/a | 0.0566 / 7.56x | 0.0544 / 7.87x | `miss` | — |
| `concurrent_graph_visitors` | `bytecode` | `verified` | 1.2980 | n/a | 0.0715 / 18.15x | 0.0698 / 18.60x | `miss` | — |
| `concurrent_audio_voices` | `bytecode` | `verified` | 1.4960 | n/a | 0.1374 / 10.89x | 0.1177 / 12.71x | `miss` | — |
| `concurrent_packet_codecs` | `bytecode` | `verified` | 0.8200 | n/a | 0.0923 / 8.88x | 0.1003 / 8.18x | `miss` | — |
| `concurrent_scene_tiles` | `bytecode` | `verified` | 0.6480 | n/a | 0.0773 / 8.38x | 0.0778 / 8.33x | `miss` | — |
| `concurrent_tree_folds` | `bytecode` | `verified` | 0.4540 | n/a | 0.0689 / 6.59x | 0.0691 / 6.57x | `miss` | — |
| `concurrent_state_machines` | `bytecode` | `verified` | 0.3660 | n/a | 0.0647 / 5.66x | 0.0631 / 5.80x | `miss` | — |
| `concurrent_stateful_pipeline` | `bytecode` | `verified` | 0.5000 | n/a | 0.0687 / 7.28x | 0.0573 / 8.73x | `miss` | — |
| `versioned_telemetry_pipeline` | `bytecode` | `verified` | 3.3040 | n/a | 0.2023 / 16.33x | 0.1258 / 26.26x | `miss` | — |
| `i_before_e` | `bytecode` | `verified` | 0.6300 | n/a | 0.0956 / 6.59x | 0.1638 / 3.85x | `miss` | — |
| `base64` | `bytecode` | `verified` | 2.9760 | n/a | 4.0818 / 0.73x | 2.7211 / 1.09x | `miss` | — |
| `json` | `bytecode` | `verified` | 0.8800 | n/a | 2.6764 / 0.33x | 1.7105 / 0.51x | `meets` | — |

## Source scorecards

- `v12/docs/perf-baselines/2026-07-31-compiled-closure-refresh-scorecard.json` — `custom` (`2026-07-31T22:45:38.198870Z`)
- `v12/docs/perf-baselines/2026-07-29-nullable-scalar-retained-generality-bytecode-01-selected.json` — `custom` (`2026-07-29T23:06:06.537052Z`)
- `v12/docs/perf-baselines/2026-07-29-nullable-scalar-retained-generality-bytecode-02-selected.json` — `custom` (`2026-07-29T23:09:39.636853Z`)
- `v12/docs/perf-baselines/2026-07-29-nullable-scalar-retained-generality-bytecode-04-selected.json` — `custom` (`2026-07-29T23:12:06.913697Z`)
- `v12/docs/perf-baselines/2026-07-29-nullable-scalar-retained-generality-bytecode-05-selected.json` — `custom` (`2026-07-29T23:17:11.386684Z`)
- `v12/docs/perf-baselines/2026-07-29-nullable-scalar-retained-generality-bytecode-06-selected.json` — `custom` (`2026-07-29T23:20:12.679332Z`)
- `v12/docs/perf-baselines/2026-07-29-nullable-scalar-retained-generality-bytecode-07-selected.json` — `custom` (`2026-07-29T23:21:17.999362Z`)
- `v12/docs/perf-baselines/2026-07-29-nullable-scalar-retained-async-bytecode-01-selected.json` — `custom` (`2026-07-29T23:22:50.078801Z`)
- `v12/docs/perf-baselines/2026-07-29-nullable-scalar-retained-async-bytecode-02-selected.json` — `custom` (`2026-07-29T23:23:15.366513Z`)
- `v12/docs/perf-baselines/2026-07-29-nullable-scalar-retained-coverage-extra-bytecode-01-selected.json` — `custom` (`2026-07-29T23:51:55.733619Z`)
- `v12/docs/perf-baselines/2026-07-29-nullable-scalar-retained-coverage-extra-bytecode-02-selected.json` — `custom` (`2026-07-29T23:52:24.838347Z`)
- `v12/docs/perf-baselines/2026-07-29-nullable-scalar-retained-coverage-extra-bytecode-03-selected.json` — `custom` (`2026-07-29T23:53:34.077089Z`)
- `v12/docs/perf-baselines/2026-07-29-nullable-scalar-retained-coverage-extra-bytecode-04-selected.json` — `custom` (`2026-07-29T23:54:53.654803Z`)
- `v12/docs/perf-baselines/2026-07-29-nullable-scalar-retained-coverage-extra-bytecode-05-selected.json` — `custom` (`2026-07-29T23:55:03.448526Z`)
- `v12/docs/perf-baselines/2026-07-29-nullable-scalar-retained-coverage-extra-bytecode-06-selected.json` — `custom` (`2026-07-29T23:59:12.829005Z`)
- `v12/docs/perf-baselines/2026-07-30-versioned-telemetry-pipeline-bytecode.json` — `custom` (`2026-07-30T13:40:26.758225Z`)
- `v12/docs/perf-baselines/2026-07-30-extern-plugin-toolchain-i-before-e-stable-row.json` — `custom` (`2026-07-29T23:10:31.909409Z`)
- `v12/docs/perf-baselines/2026-07-30-extern-plugin-toolchain-bytecode.json` — `custom` (`2026-07-30T18:18:40.955882Z`)

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
