# Sustained Multi-Feature Workload-Depth Audit

- Applications: `66`
- Discriminating feature families: `11`
- Broad threshold: `6` families
- Sustained native threshold: `0.200s`
- Broad applications: `24`
- Sustained applications: `11`
- Applications satisfying both: `1`
- Material depth gap: `no`

The duration predicate uses the five-process Go mean, not Able time, so a
large runtime/compiler penalty cannot make a short reference workload look
sustained. The breadth predicate uses at least half of the portable feature
families selected by the existing interaction-priority contract.

| Application | Families | Weight | Triples | Operations | Go s | Able s | Excess s | Broad | Sustained |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- | --- |
| `versioned_telemetry_pipeline` | 10 | 26 | 120 | 10 | 0.2057 | 1.9880 | 1.7715 | yes | yes |
| `concurrent_policy_callbacks` | 11 | 29 | 165 | 2 | 0.0070 | 0.0320 | 0.0246 | yes | no |
| `concurrent_state_machines` | 11 | 29 | 165 | 5 | 0.0065 | 0.0320 | 0.0252 | yes | no |
| `concurrent_event_routing` | 11 | 29 | 165 | 0 | 0.0064 | 0.0340 | 0.0273 | yes | no |
| `concurrent_audio_voices` | 11 | 29 | 165 | 5 | 0.0053 | 0.0300 | 0.0244 | yes | no |
| `concurrent_scene_tiles` | 11 | 29 | 165 | 5 | 0.0052 | 0.0300 | 0.0245 | yes | no |
| `concurrent_tree_folds` | 11 | 29 | 165 | 5 | 0.0049 | 0.0280 | 0.0228 | yes | no |
| `concurrent_packet_codecs` | 11 | 29 | 165 | 6 | 0.0048 | 0.0300 | 0.0249 | yes | no |
| `concurrent_graph_visitors` | 11 | 29 | 165 | 6 | 0.0048 | 0.0200 | 0.0149 | yes | no |
| `concurrent_stateful_pipeline` | 10 | 27 | 120 | 3 | 0.0074 | 0.0640 | 0.0562 | yes | no |
| `concurrent_text_index` | 10 | 26 | 120 | 0 | 0.0123 | 0.0400 | 0.0271 | yes | no |
| `concurrent_transform_chain` | 10 | 26 | 120 | 2 | 0.0068 | 0.0300 | 0.0228 | yes | no |
| `concurrent_signal_dispatch` | 10 | 26 | 120 | 10 | 0.0058 | 0.0360 | 0.0299 | yes | no |
| `policy_record_dispatch` | 10 | 26 | 120 | 0 | 0.0053 | 0.0680 | 0.0624 | yes | no |
| `concurrent_document_pipeline` | 10 | 26 | 120 | 0 | 0.0047 | 0.0260 | 0.0211 | yes | no |
| `generic_slot_buffer` | 9 | 25 | 84 | 0 | 0.0058 | 0.0300 | 0.0239 | yes | no |
| `dependency_wave_validation` | 9 | 25 | 84 | 0 | 0.0055 | 0.0420 | 0.0362 | yes | no |
| `binary_event_log` | 9 | 24 | 84 | 9 | 0.0080 | 0.0720 | 0.0636 | yes | no |
| `manifest_normalization` | 9 | 24 | 84 | 0 | 0.0043 | 0.0300 | 0.0255 | yes | no |
| `transaction_ledger_audit` | 9 | 23 | 84 | 0 | 0.0071 | 0.0300 | 0.0225 | yes | no |
| `concurrent_stencil_reduction` | 9 | 23 | 84 | 9 | 0.0061 | 0.0340 | 0.0276 | yes | no |
| `sensor_calibration` | 9 | 23 | 84 | 7 | 0.0055 | 0.0340 | 0.0282 | yes | no |
| `validated_job_pipeline` | 9 | 23 | 84 | 0 | 0.0043 | 0.0600 | 0.0555 | yes | no |
| `backup_dedup` | 8 | 20 | 56 | 8 | 0.0138 | 0.0420 | 0.0275 | yes | no |
| `discrete_event_simulation` | 5 | 14 | 10 | 4 | 0.0148 | 0.0400 | 0.0244 | no | no |
| `word_frequency` | 4 | 11 | 4 | 3 | 0.0059 | 0.0320 | 0.0258 | no | no |
| `document_audit` | 4 | 11 | 4 | 4 | 0.0055 | 0.0200 | 0.0142 | no | no |
| `lexical_rollup` | 4 | 11 | 4 | 0 | 0.0041 | 0.0400 | 0.0357 | no | no |
| `option_result_config` | 4 | 11 | 4 | 4 | 0.0038 | 0.0420 | 0.0380 | no | no |
| `dependency_plan` | 4 | 10 | 4 | 2 | 0.0054 | 0.0360 | 0.0303 | no | no |
| `config_validation_extraction` | 4 | 9 | 4 | 1 | 0.0057 | 0.0360 | 0.0300 | no | no |
| `log_routing_redaction` | 4 | 9 | 4 | 1 | 0.0049 | 0.0400 | 0.0348 | no | no |
| `binarytrees` | 3 | 8 | 1 | 4 | 11.9591 | 7.3740 | 0.0000 | no | yes |
| `k_nucleotide` | 3 | 8 | 1 | 3 | 0.0620 | 1.3040 | 1.2387 | no | no |
| `fasta_generation` | 3 | 8 | 1 | 0 | 0.0173 | 0.0400 | 0.0218 | no | no |
| `inventory_reconciliation` | 3 | 8 | 1 | 1 | 0.0092 | 0.1000 | 0.0903 | no | no |
| `quicksort` | 3 | 7 | 1 | 2 | 2.7145 | 1.7820 | 0.0000 | no | yes |
| `base64` | 3 | 7 | 1 | 3 | 2.4163 | 2.3560 | 0.0000 | no | yes |
| `sudoku_masks` | 3 | 7 | 1 | 2 | 0.7574 | 1.6020 | 0.8047 | no | yes |
| `fib` | 2 | 6 | 0 | 1 | 3.5367 | 3.4160 | 0.0000 | no | yes |
| `json` | 2 | 6 | 0 | 1 | 1.5717 | 0.5600 | 0.0000 | no | yes |
| `distance_field` | 2 | 6 | 0 | 1 | 0.0156 | 0.0600 | 0.0436 | no | no |
| `rms_norm` | 2 | 6 | 0 | 0 | 0.0126 | 0.0400 | 0.0267 | no | no |
| `wide_integer_records` | 2 | 5 | 0 | 1 | 0.0278 | 0.0600 | 0.0307 | no | no |
| `rational_series` | 2 | 5 | 0 | 2 | 0.0135 | 0.0500 | 0.0358 | no | no |
| `unicode_scalar_pipeline` | 2 | 5 | 0 | 1 | 0.0098 | 0.1020 | 0.0917 | no | no |
| `fixed_width_128` | 2 | 5 | 0 | 2 | 0.0064 | 0.0840 | 0.0773 | no | no |
| `array_slice_window` | 2 | 5 | 0 | 0 | 0.0050 | 0.0360 | 0.0307 | no | no |
| `regex_suffix_audit` | 2 | 5 | 0 | 2 | 0.0047 | 0.0440 | 0.0391 | no | no |
| `regex_stream_audit` | 2 | 5 | 0 | 1 | 0.0046 | 0.0480 | 0.0432 | no | no |
| `i_before_e` | 2 | 4 | 0 | 0 | 0.0718 | 0.0400 | 0.0000 | no | no |
| `reverse_complement` | 2 | 4 | 0 | 1 | 0.0172 | 0.0400 | 0.0219 | no | no |
| `tapelang_alphabet` | 1 | 3 | 0 | 0 | 3.2099 | 3.6680 | 0.2892 | no | yes |
| `pidigits` | 1 | 3 | 0 | 0 | 1.2436 | 1.2000 | 0.0000 | no | yes |
| `matrixmultiply` | 1 | 3 | 0 | 1 | 1.0344 | 1.0400 | 0.0000 | no | yes |
| `monte_carlo_pi` | 1 | 3 | 0 | 0 | 0.2537 | 0.1400 | 0.0000 | no | yes |
| `mandelbrot` | 1 | 3 | 0 | 1 | 0.0620 | 0.0720 | 0.0067 | no | no |
| `nbody` | 1 | 3 | 0 | 1 | 0.0387 | 0.0800 | 0.0393 | no | no |
| `future_pipeline` | 1 | 3 | 0 | 3 | 0.0070 | 0.0300 | 0.0226 | no | no |
| `channel_rollup` | 1 | 3 | 0 | 1 | 0.0061 | 0.0400 | 0.0336 | no | no |
| `mutex_ledger` | 1 | 3 | 0 | 3 | 0.0053 | 0.0300 | 0.0244 | no | no |
| `mutex_work_queue` | 1 | 3 | 0 | 1 | 0.0050 | 0.0300 | 0.0247 | no | no |
| `await_channel_mux` | 1 | 3 | 0 | 2 | 0.0047 | 0.1120 | 0.1071 | no | no |
| `mutex_await_journal` | 1 | 3 | 0 | 1 | 0.0046 | 0.0220 | 0.0172 | no | no |
| `future_await_race` | 1 | 3 | 0 | 0 | 0.0045 | 0.0300 | 0.0253 | no | no |
| `regex_set_audit` | 1 | 2 | 0 | 1 | 0.0046 | 0.0460 | 0.0412 | no | no |
