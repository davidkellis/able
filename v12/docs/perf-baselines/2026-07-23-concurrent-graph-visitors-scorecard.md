# External Application Scoreboard

- Source measurements through: `2026-07-23T19:00:37.167936Z`
- Scope: verifier-backed Able application launches only; tree-walker is intentionally excluded from this performance target.
- Guard: each source scorecard records its process count, CPU-affinity when used, runtime settings, and per-process timeout.
- Compiled: `5/57` selected rankable rows meet the 95%-of-Go target.
- Bytecode: `3/50` selected rankable rows meet both 95%-of-Python and 95%-of-Ruby targets.
- Canonical Able source fingerprints: `114` row fingerprints in JSON; `114` came from the measured source report and the remainder are current-source legacy fingerprints.
- Verifier/declared-input contracts: `114` row fingerprints in JSON; `114` were captured before the timed launch and the remainder are current-contract legacy reconstructions.
- Canonical stdlib runtime sources: `70` `.able` files, tree SHA-256 `6a412c872ee66752de7c4417a5eda99806a7631da3b880c55574c6a640b82d9b`; Git `219eff222c28406487231713753641bc49ee5b9a` (dirty).
- Strict candidate selection: `107` reviewed benchmark/mode rows, SHA-256 `826345d3b6e77bd174273c73e2588b4e400b90d426d0f832082aaecee0606948`; timeout rows remain in full status.
- Matched reference source fingerprints: `165` comparison fingerprints in JSON; `165` came from measured reference reports and the remainder are current-source legacy fingerprints.
- `unranked` means a partial, timed-out, failed, or unavailable matched run/reference; it is never counted as a pass or fail.
- `Unranked reason` identifies whether the Able launch or its required reference prevents ranking; reference-unavailable does not infer why that source has no valid ratio.

| Benchmark | Mode | Able status | Able (s) | Go / ratio | Python / ratio | Ruby / ratio | Target | Unranked reason |
| --- | --- | --- | ---: | --- | --- | --- | --- | --- |
| `fib` | `compiled` | `verified` | 4.9340 | 3.4101 / 1.45x | n/a | n/a | `miss` | — |
| `binarytrees` | `compiled` | `verified` | 11.4440 | 11.9385 / 0.96x | n/a | n/a | `meets` | — |
| `matrixmultiply` | `compiled` | `verified` | 1.2140 | 1.0729 / 1.13x | n/a | n/a | `miss` | — |
| `quicksort` | `compiled` | `verified` | 1.9120 | 2.6813 / 0.71x | n/a | n/a | `meets` | — |
| `sudoku_masks` | `compiled` | `verified` | 1.9060 | 0.5782 / 3.30x | n/a | n/a | `miss` | — |
| `i_before_e` | `compiled` | `verified` | 0.1220 | 0.0637 / 1.92x | n/a | n/a | `miss` | — |
| `base64` | `compiled` | `verified` | 2.4480 | 2.6537 / 0.92x | n/a | n/a | `meets` | — |
| `binary_event_log` | `compiled` | `verified` | 0.5360 | 0.0090 / 59.56x | n/a | n/a | `miss` | — |
| `json` | `compiled` | `verified` | 0.7420 | 1.6387 / 0.45x | n/a | n/a | `meets` | — |
| `monte_carlo_pi` | `compiled` | `verified` | 0.1920 | 0.2826 / 0.68x | n/a | n/a | `meets` | — |
| `pidigits` | `compiled` | `verified` | 1.3200 | 1.2384 / 1.07x | n/a | n/a | `miss` | — |
| `mandelbrot` | `compiled` | `verified` | 0.1280 | 0.0527 / 2.43x | n/a | n/a | `miss` | — |
| `reverse_complement` | `compiled` | `verified` | 0.1140 | 0.0171 / 6.67x | n/a | n/a | `miss` | — |
| `k_nucleotide` | `compiled` | `verified` | 2.8980 | 0.0809 / 35.82x | n/a | n/a | `miss` | — |
| `nbody` | `compiled` | `verified` | 0.1720 | 0.0378 / 4.55x | n/a | n/a | `miss` | — |
| `tapelang_alphabet` | `compiled` | `verified` | 4.0460 | 2.1407 / 1.89x | n/a | n/a | `miss` | — |
| `distance_field` | `compiled` | `verified` | 0.0900 | 0.0133 / 6.77x | n/a | n/a | `miss` | — |
| `rms_norm` | `compiled` | `verified` | 0.0880 | 0.0119 / 7.39x | n/a | n/a | `miss` | — |
| `fasta_generation` | `compiled` | `verified` | 0.1100 | 0.0171 / 6.43x | n/a | n/a | `miss` | — |
| `fixed_width_128` | `compiled` | `verified` | 0.2060 | 0.0058 / 35.52x | n/a | n/a | `miss` | — |
| `rational_series` | `compiled` | `verified` | 0.1280 | 0.0145 / 8.83x | n/a | n/a | `miss` | — |
| `wide_integer_records` | `compiled` | `verified` | 0.1840 | 0.0267 / 6.89x | n/a | n/a | `miss` | — |
| `word_frequency` | `compiled` | `verified` | 0.1800 | 0.0059 / 30.51x | n/a | n/a | `miss` | — |
| `document_audit` | `compiled` | `verified` | 0.1020 | 0.0053 / 19.25x | n/a | n/a | `miss` | — |
| `lexical_rollup` | `compiled` | `verified` | 0.1200 | 0.0054 / 22.22x | n/a | n/a | `miss` | — |
| `channel_rollup` | `compiled` | `verified` | 0.5980 | 0.0063 / 94.92x | n/a | n/a | `miss` | — |
| `future_pipeline` | `compiled` | `verified` | 0.3960 | 0.0057 / 69.47x | n/a | n/a | `miss` | — |
| `future_await_race` | `compiled` | `verified` | 0.1060 | 0.0049 / 21.63x | n/a | n/a | `miss` | — |
| `await_channel_mux` | `compiled` | `verified` | 0.3600 | 0.0055 / 65.45x | n/a | n/a | `miss` | — |
| `mutex_ledger` | `compiled` | `verified` | 0.8200 | 0.0051 / 160.78x | n/a | n/a | `miss` | — |
| `mutex_await_journal` | `compiled` | `verified` | 0.8120 | 0.0045 / 180.44x | n/a | n/a | `miss` | — |
| `mutex_work_queue` | `compiled` | `verified` | 2.3300 | 0.0052 / 448.08x | n/a | n/a | `miss` | — |
| `regex_suffix_audit` | `compiled` | `verified` | 0.1680 | 0.0064 / 26.25x | n/a | n/a | `miss` | — |
| `regex_set_audit` | `compiled` | `verified` | 0.1440 | 0.0070 / 20.57x | n/a | n/a | `miss` | — |
| `regex_stream_audit` | `compiled` | `verified` | 0.1600 | 0.0062 / 25.81x | n/a | n/a | `miss` | — |
| `log_routing_redaction` | `compiled` | `verified` | 0.1440 | 0.0058 / 24.83x | n/a | n/a | `miss` | — |
| `config_validation_extraction` | `compiled` | `verified` | 0.1300 | 0.0047 / 27.66x | n/a | n/a | `miss` | — |
| `unicode_scalar_pipeline` | `compiled` | `verified` | 0.3060 | 0.0111 / 27.57x | n/a | n/a | `miss` | — |
| `array_slice_window` | `compiled` | `verified` | 0.0980 | 0.0047 / 20.85x | n/a | n/a | `miss` | — |
| `dependency_plan` | `compiled` | `verified` | 0.0940 | 0.0051 / 18.43x | n/a | n/a | `miss` | — |
| `inventory_reconciliation` | `compiled` | `verified` | 0.2120 | 0.0096 / 22.08x | n/a | n/a | `miss` | — |
| `option_result_config` | `compiled` | `verified` | 0.1760 | 0.0052 / 33.85x | n/a | n/a | `miss` | — |
| `concurrent_text_index` | `compiled` | `verified` | 0.9960 | 0.0064 / 155.62x | n/a | n/a | `miss` | — |
| `validated_job_pipeline` | `compiled` | `verified` | 1.0460 | 0.0038 / 275.26x | n/a | n/a | `miss` | — |
| `dependency_wave_validation` | `compiled` | `verified` | 1.4440 | 0.0043 / 335.81x | n/a | n/a | `miss` | — |
| `concurrent_event_routing` | `compiled` | `verified` | 2.9140 | 0.0056 / 520.36x | n/a | n/a | `miss` | — |
| `concurrent_document_pipeline` | `compiled` | `verified` | 0.2840 | 0.0051 / 55.69x | n/a | n/a | `miss` | — |
| `manifest_normalization` | `compiled` | `verified` | 0.2260 | 0.0047 / 48.09x | n/a | n/a | `miss` | — |
| `policy_record_dispatch` | `compiled` | `verified` | 0.2260 | 0.0087 / 25.98x | n/a | n/a | `miss` | — |
| `array_slice_window` | `bytecode` | `verified` | 0.7440 | n/a | 0.0610 / 12.20x | 0.1294 / 5.75x | `miss` | — |
| `await_channel_mux` | `bytecode` | `verified` | 0.2000 | n/a | 0.2202 / 0.91x | 0.2145 / 0.93x | `meets` | — |
| `base64` | `bytecode` | `verified` | 2.8400 | n/a | 6.6027 / 0.43x | 2.5509 / 1.11x | `miss` | — |
| `binary_event_log` | `bytecode` | `verified` | 7.0260 | n/a | 0.1779 / 39.49x | 0.2649 / 26.52x | `miss` | — |
| `channel_rollup` | `bytecode` | `verified` | 0.5380 | n/a | 0.0471 / 11.42x | 0.0574 / 9.37x | `miss` | — |
| `config_validation_extraction` | `bytecode` | `verified` | 1.3800 | n/a | 0.0218 / 63.30x | 0.0489 / 28.22x | `miss` | — |
| `concurrent_document_pipeline` | `bytecode` | `verified` | 0.2900 | n/a | 0.0229 / 12.66x | 0.0512 / 5.66x | `miss` | — |
| `concurrent_event_routing` | `bytecode` | `verified` | 3.2560 | n/a | 0.0336 / 96.90x | 0.0611 / 53.29x | `miss` | — |
| `concurrent_text_index` | `bytecode` | `verified` | 0.7680 | n/a | 0.0973 / 7.89x | 0.1070 / 7.18x | `miss` | — |
| `dependency_plan` | `bytecode` | `verified` | 0.5560 | n/a | 0.0193 / 28.81x | 0.0542 / 10.26x | `miss` | — |
| `dependency_wave_validation` | `bytecode` | `verified` | 0.5620 | n/a | 0.0353 / 15.92x | 0.0551 / 10.20x | `miss` | — |
| `distance_field` | `bytecode` | `verified` | 5.9140 | n/a | 0.5798 / 10.20x | 0.3877 / 15.25x | `miss` | — |
| `document_audit` | `bytecode` | `verified` | 0.2960 | n/a | 0.0149 / 19.87x | 0.0428 / 6.92x | `miss` | — |
| `fasta_generation` | `bytecode` | `verified` | 1.9000 | n/a | 0.2082 / 9.13x | 0.3160 / 6.01x | `miss` | — |
| `fixed_width_128` | `bytecode` | `verified` | 8.5220 | n/a | 0.3504 / 24.32x | 0.6762 / 12.60x | `miss` | — |
| `future_await_race` | `bytecode` | `verified` | 0.1520 | n/a | 0.0332 / 4.58x | 0.0598 / 2.54x | `miss` | — |
| `future_pipeline` | `bytecode` | `verified` | 0.4580 | n/a | 0.0626 / 7.32x | 0.0709 / 6.46x | `miss` | — |
| `i_before_e` | `bytecode` | `verified` | 0.5340 | n/a | 0.0841 / 6.35x | 0.1269 / 4.21x | `miss` | — |
| `inventory_reconciliation` | `bytecode` | `verified` | 2.6240 | n/a | 0.0750 / 34.99x | 0.0896 / 29.29x | `miss` | — |
| `json` | `bytecode` | `verified` | 0.8940 | n/a | 2.8646 / 0.31x | 1.9273 / 0.46x | `meets` | — |
| `k_nucleotide` | `bytecode` | `verified` | 46.5300 | n/a | 1.4742 / 31.56x | 1.4639 / 31.78x | `miss` | — |
| `lexical_rollup` | `bytecode` | `verified` | 0.4480 | n/a | 0.0269 / 16.65x | 0.0606 / 7.39x | `miss` | — |
| `log_routing_redaction` | `bytecode` | `verified` | 3.1280 | n/a | 0.0219 / 142.83x | 0.0574 / 54.49x | `miss` | — |
| `manifest_normalization` | `bytecode` | `verified` | 1.5820 | n/a | 0.0249 / 63.53x | 0.0781 / 20.26x | `miss` | — |
| `mandelbrot` | `bytecode` | `verified` | 6.8140 | n/a | 1.4879 / 4.58x | 2.0400 / 3.34x | `miss` | — |
| `monte_carlo_pi` | `bytecode` | `verified` | 2.6440 | n/a | 1.7064 / 1.55x | 1.8248 / 1.45x | `miss` | — |
| `mutex_await_journal` | `bytecode` | `verified` | 0.1960 | n/a | 0.0302 / 6.49x | 0.0513 / 3.82x | `miss` | — |
| `mutex_ledger` | `bytecode` | `verified` | 0.3760 | n/a | 0.0345 / 10.90x | 0.0567 / 6.63x | `miss` | — |
| `mutex_work_queue` | `bytecode` | `verified` | 0.3400 | n/a | 0.0290 / 11.72x | 0.0534 / 6.37x | `miss` | — |
| `option_result_config` | `bytecode` | `verified` | 0.8340 | n/a | 0.0191 / 43.66x | 0.0538 / 15.50x | `miss` | — |
| `pidigits` | `bytecode` | `verified` | 2.4540 | n/a | 4.3663 / 0.56x | 12.5234 / 0.20x | `meets` | — |
| `policy_record_dispatch` | `bytecode` | `verified` | 7.4520 | n/a | 0.0318 / 234.34x | 0.0675 / 110.40x | `miss` | — |
| `rational_series` | `bytecode` | `verified` | 4.3260 | n/a | 0.1895 / 22.83x | 0.2003 / 21.60x | `miss` | — |
| `regex_set_audit` | `bytecode` | `verified` | 4.5180 | n/a | 0.0258 / 175.12x | 0.0766 / 58.98x | `miss` | — |
| `regex_stream_audit` | `bytecode` | `verified` | 3.5960 | n/a | 0.0292 / 123.15x | 0.0719 / 50.01x | `miss` | — |
| `regex_suffix_audit` | `bytecode` | `verified` | 3.5520 | n/a | 0.0319 / 111.35x | 0.0744 / 47.74x | `miss` | — |
| `reverse_complement` | `bytecode` | `verified` | 3.7580 | n/a | 0.0379 / 99.16x | 0.1142 / 32.91x | `miss` | — |
| `rms_norm` | `bytecode` | `verified` | 4.7540 | n/a | 1.1947 / 3.98x | 0.7739 / 6.14x | `miss` | — |
| `unicode_scalar_pipeline` | `bytecode` | `verified` | 3.6500 | n/a | 0.3329 / 10.96x | 0.3499 / 10.43x | `miss` | — |
| `validated_job_pipeline` | `bytecode` | `verified` | 0.3800 | n/a | 0.0263 / 14.45x | 0.0479 / 7.93x | `miss` | — |
| `wide_integer_records` | `bytecode` | `verified` | 5.5860 | n/a | 0.0782 / 71.43x | 0.1611 / 34.67x | `miss` | — |
| `word_frequency` | `bytecode` | `verified` | 1.4140 | n/a | 0.0234 / 60.43x | 0.0517 / 27.35x | `miss` | — |
| `fib` | `bytecode` | `verified` | 0.1300 | n/a | n/a | 49.0330 / 0.00x | `unranked` | Python reference unavailable |
| `binarytrees` | `bytecode` | `timeout` | n/a | n/a | n/a | n/a | `unranked` | Able timed out |
| `matrixmultiply` | `bytecode` | `verified` | 5.1400 | n/a | n/a | 48.6021 / 0.11x | `unranked` | Python reference unavailable |
| `quicksort` | `bytecode` | `timeout` | n/a | n/a | 24.9782 / n/a | 15.7944 / n/a | `unranked` | Able timed out |
| `sudoku_masks` | `bytecode` | `timeout` | n/a | n/a | 18.0397 / n/a | 21.9767 / n/a | `unranked` | Able timed out |
| `nbody` | `bytecode` | `timeout` | n/a | n/a | 1.9757 / n/a | 3.3129 / n/a | `unranked` | Able timed out |
| `tapelang_alphabet` | `bytecode` | `timeout` | n/a | n/a | n/a | n/a | `unranked` | Able timed out |
| `sensor_calibration` | `compiled` | `verified` | 0.2560 | 0.0051 / 50.20x | n/a | n/a | `miss` | — |
| `sensor_calibration` | `bytecode` | `verified` | 3.7940 | n/a | 0.0331 / 114.62x | 0.0732 / 51.83x | `miss` | — |
| `concurrent_stencil_reduction` | `compiled` | `verified` | 0.2140 | 0.0051 / 41.96x | n/a | n/a | `miss` | — |
| `concurrent_stencil_reduction` | `bytecode` | `verified` | 1.8620 | n/a | 0.0963 / 19.34x | 0.1240 / 15.02x | `miss` | — |
| `concurrent_signal_dispatch` | `compiled` | `verified` | 0.2700 | 0.0052 / 51.92x | n/a | n/a | `miss` | — |
| `concurrent_signal_dispatch` | `bytecode` | `verified` | 1.6140 | n/a | 0.0612 / 26.37x | 0.0922 / 17.51x | `miss` | — |
| `concurrent_transform_chain` | `compiled` | `verified` | 8.1760 | 0.0065 / 1257.85x | n/a | n/a | `miss` | — |
| `concurrent_transform_chain` | `bytecode` | `verified` | 2.8300 | n/a | 0.1628 / 17.38x | 0.2095 / 13.51x | `miss` | — |
| `concurrent_policy_callbacks` | `compiled` | `verified` | 0.5420 | 0.0049 / 110.61x | n/a | n/a | `miss` | — |
| `concurrent_policy_callbacks` | `bytecode` | `verified` | 0.3820 | n/a | 0.0768 / 4.97x | 0.0574 / 6.66x | `miss` | — |
| `concurrent_stateful_pipeline` | `compiled` | `verified` | 0.8140 | 0.0044 / 185.00x | n/a | n/a | `miss` | — |
| `concurrent_stateful_pipeline` | `bytecode` | `verified` | 0.3500 | n/a | 0.0616 / 5.68x | 0.0492 / 7.11x | `miss` | — |
| `concurrent_state_machines` | `compiled` | `verified` | 0.2640 | 0.0037 / 71.35x | n/a | n/a | `miss` | — |
| `concurrent_state_machines` | `bytecode` | `verified` | 0.3020 | n/a | 0.0602 / 5.02x | 0.0548 / 5.51x | `miss` | — |
| `concurrent_graph_visitors` | `compiled` | `verified` | 0.2580 | 0.0040 / 64.50x | n/a | n/a | `miss` | — |
| `concurrent_graph_visitors` | `bytecode` | `verified` | 0.9540 | n/a | 0.0564 / 16.91x | 0.0539 / 17.70x | `miss` | — |

## Source scorecards

- `v12/docs/perf-baselines/2026-07-22-current-compiled-scorecard.json` — `coverage` (`2026-07-22T16:26:36.446220Z`)
- `v12/docs/perf-baselines/2026-07-22-current-bytecode-scorecard.json` — `custom` (`2026-07-22T17:11:45.971914Z`)
- `v12/docs/perf-baselines/2026-07-22-current-bytecode-status-scorecard.json` — `custom` (`2026-07-22T17:16:49.850539Z`)
- `v12/docs/perf-baselines/2026-07-22-sensor-calibration-promotion-compiled.json` — `custom` (`2026-07-23T00:18:21.468657Z`)
- `v12/docs/perf-baselines/2026-07-22-sensor-calibration-promotion-bytecode.json` — `custom` (`2026-07-23T00:18:54.891829Z`)
- `v12/docs/perf-baselines/2026-07-22-concurrent-stencil-reduction-application-comparison.json` — `custom` (`2026-07-23T04:55:27.757991Z`)
- `v12/docs/perf-baselines/2026-07-23-concurrent-signal-dispatch-comparison-a.json` — `custom` (`2026-07-23T13:13:58.403356Z`)
- `v12/docs/perf-baselines/2026-07-23-concurrent-transform-chain-comparison-a.json` — `custom` (`2026-07-23T14:25:00.143928Z`)
- `v12/docs/perf-baselines/2026-07-23-concurrent-policy-callbacks-comparison-a.json` — `custom` (`2026-07-23T15:41:47.349921Z`)
- `v12/docs/perf-baselines/2026-07-23-concurrent-stateful-pipeline-comparison-a.json` — `custom` (`2026-07-23T16:44:35.572183Z`)
- `v12/docs/perf-baselines/2026-07-23-concurrent-state-machines-comparison-a.json` — `custom` (`2026-07-23T17:43:21.352309Z`)
- `v12/docs/perf-baselines/2026-07-23-concurrent-graph-visitors-comparison-a.json` — `custom` (`2026-07-23T19:00:37.167936Z`)

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
