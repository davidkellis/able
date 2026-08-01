# External Application Variance Report

This is a report-only sample spread analysis. It does not set or enforce a performance threshold.
When one report embeds repeated runs, timing columns use those runs directly; ratio samples remain report-level because independent Able and reference runs are not paired.
Every Able and fresh-reference component retained exactly 5 successful verifier-backed runs.

| Benchmark | Mode | Verified samples | Able seconds | Reference seconds | Source-level reference ratios |
| --- | --- | ---: | --- | --- | --- |
| `array_slice_window` | `compiled` | 5 | median=0.0400, mean=0.0360, range=0.0100, CV=15.21% | go: median=0.0050, mean=0.0050, range=0.0003, CV=2.53% | go: median=7.2000, mean=7.2000, range=0.0000, CV=0.00% |
| `await_channel_mux` | `compiled` | 5 | median=0.1200, mean=0.1120, range=0.0300, CV=11.64% | go: median=0.0047, mean=0.0047, range=0.0003, CV=2.72% | go: median=23.8298, mean=23.8298, range=0.0000, CV=0.00% |
| `backup_dedup` | `compiled` | 5 | median=0.0400, mean=0.0420, range=0.0100, CV=10.65% | go: median=0.0138, mean=0.0138, range=0.0005, CV=1.43% | go: median=3.0435, mean=3.0435, range=0.0000, CV=0.00% |
| `base64` | `compiled` | 5 | median=2.2900, mean=2.3560, range=0.2700, CV=4.80% | go: median=2.4056, mean=2.4163, range=0.2152, CV=3.28% | go: median=0.9750, mean=0.9750, range=0.0000, CV=0.00% |
| `binary_event_log` | `compiled` | 5 | median=0.0700, mean=0.0720, range=0.0100, CV=6.21% | go: median=0.0078, mean=0.0080, range=0.0019, CV=9.08% | go: median=9.0000, mean=9.0000, range=0.0000, CV=0.00% |
| `binarytrees` | `compiled` | 5 | median=7.4400, mean=7.3740, range=0.4200, CV=2.74% | go: median=10.9040, mean=11.9591, range=6.0126, CV=21.91% | go: median=0.6166, mean=0.6166, range=0.0000, CV=0.00% |
| `channel_rollup` | `compiled` | 5 | median=0.0400, mean=0.0400, range=0.0200, CV=25.00% | go: median=0.0061, mean=0.0061, range=0.0002, CV=1.17% | go: median=6.5574, mean=6.5574, range=0.0000, CV=0.00% |
| `concurrent_audio_voices` | `compiled` | 5 | median=0.0300, mean=0.0300, range=0.0000, CV=0.00% | go: median=0.0053, mean=0.0053, range=0.0004, CV=3.17% | go: median=5.6604, mean=5.6604, range=0.0000, CV=0.00% |
| `concurrent_document_pipeline` | `compiled` | 5 | median=0.0300, mean=0.0260, range=0.0100, CV=21.07% | go: median=0.0047, mean=0.0047, range=0.0011, CV=9.66% | go: median=5.5319, mean=5.5319, range=0.0000, CV=0.00% |
| `concurrent_event_routing` | `compiled` | 5 | median=0.0300, mean=0.0340, range=0.0100, CV=16.11% | go: median=0.0062, mean=0.0064, range=0.0031, CV=19.81% | go: median=5.3125, mean=5.3125, range=0.0000, CV=0.00% |
| `concurrent_graph_visitors` | `compiled` | 5 | median=0.0200, mean=0.0200, range=0.0000, CV=0.00% | go: median=0.0048, mean=0.0048, range=0.0002, CV=1.93% | go: median=4.1667, mean=4.1667, range=0.0000, CV=0.00% |
| `concurrent_packet_codecs` | `compiled` | 5 | median=0.0300, mean=0.0300, range=0.0000, CV=0.00% | go: median=0.0048, mean=0.0048, range=0.0006, CV=4.84% | go: median=6.2500, mean=6.2500, range=0.0000, CV=0.00% |
| `concurrent_policy_callbacks` | `compiled` | 5 | median=0.0300, mean=0.0320, range=0.0100, CV=13.98% | go: median=0.0063, mean=0.0070, range=0.0053, CV=31.35% | go: median=4.5714, mean=4.5714, range=0.0000, CV=0.00% |
| `concurrent_scene_tiles` | `compiled` | 5 | median=0.0300, mean=0.0300, range=0.0200, CV=23.57% | go: median=0.0051, mean=0.0052, range=0.0008, CV=6.53% | go: median=5.7692, mean=5.7692, range=0.0000, CV=0.00% |
| `concurrent_signal_dispatch` | `compiled` | 5 | median=0.0400, mean=0.0360, range=0.0100, CV=15.21% | go: median=0.0061, mean=0.0058, range=0.0010, CV=7.29% | go: median=6.2069, mean=6.2069, range=0.0000, CV=0.00% |
| `concurrent_state_machines` | `compiled` | 5 | median=0.0300, mean=0.0320, range=0.0100, CV=13.98% | go: median=0.0059, mean=0.0065, range=0.0041, CV=26.34% | go: median=4.9231, mean=4.9231, range=0.0000, CV=0.00% |
| `concurrent_stateful_pipeline` | `compiled` | 5 | median=0.0600, mean=0.0640, range=0.0300, CV=17.82% | go: median=0.0068, mean=0.0074, range=0.0032, CV=18.07% | go: median=8.6486, mean=8.6486, range=0.0000, CV=0.00% |
| `concurrent_stencil_reduction` | `compiled` | 5 | median=0.0300, mean=0.0340, range=0.0100, CV=16.11% | go: median=0.0061, mean=0.0061, range=0.0003, CV=1.86% | go: median=5.5738, mean=5.5738, range=0.0000, CV=0.00% |
| `concurrent_text_index` | `compiled` | 5 | median=0.0400, mean=0.0400, range=0.0000, CV=0.00% | go: median=0.0085, mean=0.0123, range=0.0117, CV=45.59% | go: median=3.2520, mean=3.2520, range=0.0000, CV=0.00% |
| `concurrent_transform_chain` | `compiled` | 5 | median=0.0300, mean=0.0300, range=0.0000, CV=0.00% | go: median=0.0065, mean=0.0068, range=0.0017, CV=10.48% | go: median=4.4118, mean=4.4118, range=0.0000, CV=0.00% |
| `concurrent_tree_folds` | `compiled` | 5 | median=0.0300, mean=0.0280, range=0.0100, CV=15.97% | go: median=0.0050, mean=0.0049, range=0.0005, CV=3.97% | go: median=5.7143, mean=5.7143, range=0.0000, CV=0.00% |
| `config_validation_extraction` | `compiled` | 5 | median=0.0300, mean=0.0360, range=0.0200, CV=24.85% | go: median=0.0057, mean=0.0057, range=0.0009, CV=6.27% | go: median=6.3158, mean=6.3158, range=0.0000, CV=0.00% |
| `dependency_plan` | `compiled` | 5 | median=0.0400, mean=0.0360, range=0.0100, CV=15.21% | go: median=0.0055, mean=0.0054, range=0.0009, CV=6.47% | go: median=6.6667, mean=6.6667, range=0.0000, CV=0.00% |
| `dependency_wave_validation` | `compiled` | 5 | median=0.0400, mean=0.0420, range=0.0100, CV=10.65% | go: median=0.0056, mean=0.0055, range=0.0009, CV=6.63% | go: median=7.6364, mean=7.6364, range=0.0000, CV=0.00% |
| `discrete_event_simulation` | `compiled` | 5 | median=0.0400, mean=0.0400, range=0.0200, CV=17.68% | go: median=0.0148, mean=0.0148, range=0.0013, CV=3.45% | go: median=2.7027, mean=2.7027, range=0.0000, CV=0.00% |
| `distance_field` | `compiled` | 5 | median=0.0600, mean=0.0600, range=0.0000, CV=0.00% | go: median=0.0138, mean=0.0156, range=0.0104, CV=28.54% | go: median=3.8462, mean=3.8462, range=0.0000, CV=0.00% |
| `document_audit` | `compiled` | 5 | median=0.0200, mean=0.0200, range=0.0000, CV=0.00% | go: median=0.0055, mean=0.0055, range=0.0004, CV=2.82% | go: median=3.6364, mean=3.6364, range=0.0000, CV=0.00% |
| `fasta_generation` | `compiled` | 5 | median=0.0400, mean=0.0400, range=0.0000, CV=0.00% | go: median=0.0169, mean=0.0173, range=0.0052, CV=10.91% | go: median=2.3121, mean=2.3121, range=0.0000, CV=0.00% |
| `fib` | `compiled` | 5 | median=3.3200, mean=3.4160, range=0.5800, CV=7.14% | go: median=3.4677, mean=3.5367, range=0.7392, CV=8.42% | go: median=0.9659, mean=0.9659, range=0.0000, CV=0.00% |
| `fixed_width_128` | `compiled` | 5 | median=0.0800, mean=0.0840, range=0.0200, CV=10.65% | go: median=0.0063, mean=0.0064, range=0.0009, CV=5.46% | go: median=13.1250, mean=13.1250, range=0.0000, CV=0.00% |
| `future_await_race` | `compiled` | 5 | median=0.0300, mean=0.0300, range=0.0000, CV=0.00% | go: median=0.0045, mean=0.0045, range=0.0004, CV=3.48% | go: median=6.6667, mean=6.6667, range=0.0000, CV=0.00% |
| `future_pipeline` | `compiled` | 5 | median=0.0300, mean=0.0300, range=0.0000, CV=0.00% | go: median=0.0070, mean=0.0070, range=0.0017, CV=10.00% | go: median=4.2857, mean=4.2857, range=0.0000, CV=0.00% |
| `generic_slot_buffer` | `compiled` | 5 | median=0.0300, mean=0.0300, range=0.0000, CV=0.00% | go: median=0.0057, mean=0.0058, range=0.0008, CV=5.45% | go: median=5.1724, mean=5.1724, range=0.0000, CV=0.00% |
| `i_before_e` | `compiled` | 5 | median=0.0400, mean=0.0400, range=0.0000, CV=0.00% | go: median=0.0722, mean=0.0718, range=0.0062, CV=3.33% | go: median=0.5571, mean=0.5571, range=0.0000, CV=0.00% |
| `inventory_reconciliation` | `compiled` | 5 | median=0.1000, mean=0.1000, range=0.0000, CV=0.00% | go: median=0.0091, mean=0.0092, range=0.0006, CV=2.39% | go: median=10.8696, mean=10.8696, range=0.0000, CV=0.00% |
| `json` | `compiled` | 5 | median=0.5600, mean=0.5600, range=0.0000, CV=0.00% | go: median=1.5698, mean=1.5717, range=0.1138, CV=2.78% | go: median=0.3563, mean=0.3563, range=0.0000, CV=0.00% |
| `k_nucleotide` | `compiled` | 5 | median=1.2800, mean=1.3040, range=0.1000, CV=3.19% | go: median=0.0579, mean=0.0620, range=0.0168, CV=11.66% | go: median=21.0323, mean=21.0323, range=0.0000, CV=0.00% |
| `lexical_rollup` | `compiled` | 5 | median=0.0300, mean=0.0400, range=0.0500, CV=55.90% | go: median=0.0041, mean=0.0041, range=0.0005, CV=4.80% | go: median=9.7561, mean=9.7561, range=0.0000, CV=0.00% |
| `log_routing_redaction` | `compiled` | 5 | median=0.0400, mean=0.0400, range=0.0000, CV=0.00% | go: median=0.0050, mean=0.0049, range=0.0017, CV=12.47% | go: median=8.1633, mean=8.1633, range=0.0000, CV=0.00% |
| `mandelbrot` | `compiled` | 5 | median=0.0700, mean=0.0720, range=0.0100, CV=6.21% | go: median=0.0626, mean=0.0620, range=0.0033, CV=2.08% | go: median=1.1613, mean=1.1613, range=0.0000, CV=0.00% |
| `manifest_normalization` | `compiled` | 5 | median=0.0300, mean=0.0300, range=0.0000, CV=0.00% | go: median=0.0043, mean=0.0043, range=0.0003, CV=2.98% | go: median=6.9767, mean=6.9767, range=0.0000, CV=0.00% |
| `matrixmultiply` | `compiled` | 5 | median=0.9900, mean=1.0400, range=0.2300, CV=8.97% | go: median=0.9723, mean=1.0344, range=0.2130, CV=9.12% | go: median=1.0054, mean=1.0054, range=0.0000, CV=0.00% |
| `monte_carlo_pi` | `compiled` | 5 | median=0.1400, mean=0.1400, range=0.0000, CV=0.00% | go: median=0.2588, mean=0.2537, range=0.0578, CV=9.39% | go: median=0.5518, mean=0.5518, range=0.0000, CV=0.00% |
| `mutex_await_journal` | `compiled` | 5 | median=0.0200, mean=0.0220, range=0.0100, CV=20.33% | go: median=0.0046, mean=0.0046, range=0.0013, CV=11.24% | go: median=4.7826, mean=4.7826, range=0.0000, CV=0.00% |
| `mutex_ledger` | `compiled` | 5 | median=0.0300, mean=0.0300, range=0.0000, CV=0.00% | go: median=0.0053, mean=0.0053, range=0.0004, CV=3.21% | go: median=5.6604, mean=5.6604, range=0.0000, CV=0.00% |
| `mutex_work_queue` | `compiled` | 5 | median=0.0300, mean=0.0300, range=0.0000, CV=0.00% | go: median=0.0049, mean=0.0050, range=0.0012, CV=10.45% | go: median=6.0000, mean=6.0000, range=0.0000, CV=0.00% |
| `nbody` | `compiled` | 5 | median=0.0800, mean=0.0800, range=0.0000, CV=0.00% | go: median=0.0360, mean=0.0387, range=0.0145, CV=16.42% | go: median=2.0672, mean=2.0672, range=0.0000, CV=0.00% |
| `option_result_config` | `compiled` | 5 | median=0.0400, mean=0.0420, range=0.0200, CV=19.92% | go: median=0.0038, mean=0.0038, range=0.0003, CV=3.70% | go: median=11.0526, mean=11.0526, range=0.0000, CV=0.00% |
| `pidigits` | `compiled` | 5 | median=1.1700, mean=1.2000, range=0.1000, CV=4.21% | go: median=1.2264, mean=1.2436, range=0.0551, CV=2.05% | go: median=0.9649, mean=0.9649, range=0.0000, CV=0.00% |
| `policy_record_dispatch` | `compiled` | 5 | median=0.0700, mean=0.0680, range=0.0100, CV=6.58% | go: median=0.0051, mean=0.0053, range=0.0011, CV=8.09% | go: median=12.8302, mean=12.8302, range=0.0000, CV=0.00% |
| `quicksort` | `compiled` | 5 | median=1.7900, mean=1.7820, range=0.1500, CV=3.01% | go: median=2.6765, mean=2.7145, range=0.2673, CV=4.29% | go: median=0.6565, mean=0.6565, range=0.0000, CV=0.00% |
| `rational_series` | `compiled` | 5 | median=0.0500, mean=0.0500, range=0.0000, CV=0.00% | go: median=0.0135, mean=0.0135, range=0.0009, CV=2.61% | go: median=3.7037, mean=3.7037, range=0.0000, CV=0.00% |
| `regex_set_audit` | `compiled` | 5 | median=0.0500, mean=0.0460, range=0.0100, CV=11.91% | go: median=0.0045, mean=0.0046, range=0.0003, CV=2.49% | go: median=10.0000, mean=10.0000, range=0.0000, CV=0.00% |
| `regex_stream_audit` | `compiled` | 5 | median=0.0500, mean=0.0480, range=0.0200, CV=17.43% | go: median=0.0045, mean=0.0046, range=0.0006, CV=5.27% | go: median=10.4348, mean=10.4348, range=0.0000, CV=0.00% |
| `regex_suffix_audit` | `compiled` | 5 | median=0.0400, mean=0.0440, range=0.0100, CV=12.45% | go: median=0.0046, mean=0.0047, range=0.0007, CV=6.36% | go: median=9.3617, mean=9.3617, range=0.0000, CV=0.00% |
| `reverse_complement` | `compiled` | 5 | median=0.0400, mean=0.0400, range=0.0000, CV=0.00% | go: median=0.0177, mean=0.0172, range=0.0046, CV=11.07% | go: median=2.3256, mean=2.3256, range=0.0000, CV=0.00% |
| `rms_norm` | `compiled` | 5 | median=0.0400, mean=0.0400, range=0.0000, CV=0.00% | go: median=0.0125, mean=0.0126, range=0.0018, CV=5.62% | go: median=3.1746, mean=3.1746, range=0.0000, CV=0.00% |
| `sensor_calibration` | `compiled` | 5 | median=0.0300, mean=0.0340, range=0.0100, CV=16.11% | go: median=0.0055, mean=0.0055, range=0.0006, CV=4.06% | go: median=6.1818, mean=6.1818, range=0.0000, CV=0.00% |
| `sudoku_masks` | `compiled` | 5 | median=1.5700, mean=1.6020, range=0.5400, CV=13.33% | go: median=0.7610, mean=0.7574, range=0.0204, CV=1.14% | go: median=2.1151, mean=2.1151, range=0.0000, CV=0.00% |
| `tapelang_alphabet` | `compiled` | 5 | median=3.6400, mean=3.6680, range=0.6300, CV=7.06% | go: median=3.1680, mean=3.2099, range=0.3900, CV=4.57% | go: median=1.1427, mean=1.1427, range=0.0000, CV=0.00% |
| `transaction_ledger_audit` | `compiled` | 5 | median=0.0300, mean=0.0300, range=0.0000, CV=0.00% | go: median=0.0069, mean=0.0071, range=0.0024, CV=14.11% | go: median=4.2254, mean=4.2254, range=0.0000, CV=0.00% |
| `unicode_scalar_pipeline` | `compiled` | 5 | median=0.1000, mean=0.1020, range=0.0100, CV=4.38% | go: median=0.0098, mean=0.0098, range=0.0008, CV=3.31% | go: median=10.4082, mean=10.4082, range=0.0000, CV=0.00% |
| `validated_job_pipeline` | `compiled` | 5 | median=0.0600, mean=0.0600, range=0.0000, CV=0.00% | go: median=0.0042, mean=0.0043, range=0.0004, CV=4.37% | go: median=13.9535, mean=13.9535, range=0.0000, CV=0.00% |
| `versioned_telemetry_pipeline` | `compiled` | 5 | median=1.9600, mean=1.9880, range=0.2400, CV=4.70% | go: median=0.2055, mean=0.2057, range=0.0244, CV=4.74% | go: median=9.6646, mean=9.6646, range=0.0000, CV=0.00% |
| `wide_integer_records` | `compiled` | 5 | median=0.0600, mean=0.0600, range=0.0000, CV=0.00% | go: median=0.0281, mean=0.0278, range=0.0014, CV=2.09% | go: median=2.1583, mean=2.1583, range=0.0000, CV=0.00% |
| `word_frequency` | `compiled` | 5 | median=0.0300, mean=0.0320, range=0.0100, CV=13.98% | go: median=0.0058, mean=0.0059, range=0.0009, CV=5.41% | go: median=5.4237, mean=5.4237, range=0.0000, CV=0.00% |

## Inputs

- `v12/docs/perf-baselines/2026-07-31-compiled-closure-refresh-scorecard.json` — generated `2026-07-31T22:45:38.198870Z`, CPU `12-15`
