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
| `versioned_telemetry_pipeline` | 10 | 26 | 120 | 10 | 0.2078 | 2.0680 | 1.8493 | yes | yes |
| `concurrent_event_routing` | 11 | 29 | 165 | 0 | 0.0055 | 0.0540 | 0.0482 | yes | no |
| `concurrent_state_machines` | 11 | 29 | 165 | 5 | 0.0048 | 0.0260 | 0.0209 | yes | no |
| `concurrent_tree_folds` | 11 | 29 | 165 | 5 | 0.0046 | 0.0320 | 0.0272 | yes | no |
| `concurrent_policy_callbacks` | 11 | 29 | 165 | 2 | 0.0045 | 0.0340 | 0.0293 | yes | no |
| `concurrent_scene_tiles` | 11 | 29 | 165 | 5 | 0.0045 | 0.0300 | 0.0253 | yes | no |
| `concurrent_graph_visitors` | 11 | 29 | 165 | 6 | 0.0044 | 0.0300 | 0.0254 | yes | no |
| `concurrent_audio_voices` | 11 | 29 | 165 | 5 | 0.0043 | 0.0400 | 0.0355 | yes | no |
| `concurrent_packet_codecs` | 11 | 29 | 165 | 6 | 0.0040 | 0.0300 | 0.0258 | yes | no |
| `concurrent_stateful_pipeline` | 10 | 27 | 120 | 3 | 0.0056 | 0.0660 | 0.0601 | yes | no |
| `concurrent_text_index` | 10 | 26 | 120 | 0 | 0.0075 | 0.0400 | 0.0321 | yes | no |
| `policy_record_dispatch` | 10 | 26 | 120 | 0 | 0.0053 | 0.0960 | 0.0904 | yes | no |
| `concurrent_transform_chain` | 10 | 26 | 120 | 2 | 0.0048 | 0.0420 | 0.0369 | yes | no |
| `concurrent_signal_dispatch` | 10 | 26 | 120 | 10 | 0.0046 | 0.0380 | 0.0332 | yes | no |
| `concurrent_document_pipeline` | 10 | 26 | 120 | 0 | 0.0043 | 0.0440 | 0.0395 | yes | no |
| `generic_slot_buffer` | 9 | 25 | 84 | 0 | 0.0052 | 0.0760 | 0.0705 | yes | no |
| `dependency_wave_validation` | 9 | 25 | 84 | 0 | 0.0048 | 0.0460 | 0.0409 | yes | no |
| `binary_event_log` | 9 | 24 | 84 | 9 | 0.0097 | 0.0900 | 0.0798 | yes | no |
| `manifest_normalization` | 9 | 24 | 84 | 0 | 0.0048 | 0.0440 | 0.0389 | yes | no |
| `transaction_ledger_audit` | 9 | 23 | 84 | 0 | 0.0075 | 0.0680 | 0.0601 | yes | no |
| `sensor_calibration` | 9 | 23 | 84 | 7 | 0.0056 | 0.0580 | 0.0521 | yes | no |
| `concurrent_stencil_reduction` | 9 | 23 | 84 | 9 | 0.0050 | 0.0420 | 0.0367 | yes | no |
| `validated_job_pipeline` | 9 | 23 | 84 | 0 | 0.0044 | 0.0700 | 0.0654 | yes | no |
| `backup_dedup` | 8 | 20 | 56 | 8 | 0.0110 | 0.0720 | 0.0604 | yes | no |
| `discrete_event_simulation` | 5 | 14 | 10 | 4 | 0.0138 | 0.0420 | 0.0275 | no | no |
| `word_frequency` | 4 | 11 | 4 | 3 | 0.0052 | 0.0480 | 0.0425 | no | no |
| `lexical_rollup` | 4 | 11 | 4 | 0 | 0.0041 | 0.0660 | 0.0617 | no | no |
| `document_audit` | 4 | 11 | 4 | 4 | 0.0041 | 0.0380 | 0.0337 | no | no |
| `option_result_config` | 4 | 11 | 4 | 4 | 0.0038 | 0.0460 | 0.0420 | no | no |
| `dependency_plan` | 4 | 10 | 4 | 2 | 0.0038 | 0.0200 | 0.0160 | no | no |
| `log_routing_redaction` | 4 | 9 | 4 | 1 | 0.0041 | 0.0420 | 0.0377 | no | no |
| `config_validation_extraction` | 4 | 9 | 4 | 1 | 0.0038 | 0.0480 | 0.0440 | no | no |
| `binarytrees` | 3 | 8 | 1 | 4 | 10.6498 | 10.7140 | 0.0000 | no | yes |
| `k_nucleotide` | 3 | 8 | 1 | 3 | 0.0669 | 1.5920 | 1.5216 | no | no |
| `fasta_generation` | 3 | 8 | 1 | 0 | 0.0163 | 0.0480 | 0.0308 | no | no |
| `inventory_reconciliation` | 3 | 8 | 1 | 1 | 0.0095 | 0.1400 | 0.1300 | no | no |
| `quicksort` | 3 | 7 | 1 | 2 | 2.6816 | 1.9140 | 0.0000 | no | yes |
| `base64` | 3 | 7 | 1 | 3 | 2.6198 | 2.4480 | 0.0000 | no | yes |
| `sudoku_masks` | 3 | 7 | 1 | 2 | 0.7184 | 1.9660 | 1.2098 | no | yes |
| `fib` | 2 | 6 | 0 | 1 | 3.1597 | 4.0080 | 0.6820 | no | yes |
| `json` | 2 | 6 | 0 | 1 | 1.4767 | 1.2160 | 0.0000 | no | yes |
| `distance_field` | 2 | 6 | 0 | 1 | 0.0147 | 0.0380 | 0.0225 | no | no |
| `rms_norm` | 2 | 6 | 0 | 0 | 0.0140 | 0.0300 | 0.0153 | no | no |
| `wide_integer_records` | 2 | 5 | 0 | 1 | 0.0247 | 0.0740 | 0.0480 | no | no |
| `rational_series` | 2 | 5 | 0 | 2 | 0.0135 | 0.0700 | 0.0558 | no | no |
| `unicode_scalar_pipeline` | 2 | 5 | 0 | 1 | 0.0107 | 0.1200 | 0.1087 | no | no |
| `fixed_width_128` | 2 | 5 | 0 | 2 | 0.0059 | 0.1060 | 0.0998 | no | no |
| `regex_suffix_audit` | 2 | 5 | 0 | 2 | 0.0054 | 0.0620 | 0.0563 | no | no |
| `regex_stream_audit` | 2 | 5 | 0 | 1 | 0.0050 | 0.0700 | 0.0647 | no | no |
| `array_slice_window` | 2 | 5 | 0 | 0 | 0.0044 | 0.0300 | 0.0254 | no | no |
| `i_before_e` | 2 | 4 | 0 | 0 | 0.0634 | 0.0680 | 0.0013 | no | no |
| `reverse_complement` | 2 | 4 | 0 | 1 | 0.0160 | 0.0600 | 0.0432 | no | no |
| `tapelang_alphabet` | 1 | 3 | 0 | 0 | 3.0243 | 3.7680 | 0.5845 | no | yes |
| `pidigits` | 1 | 3 | 0 | 0 | 1.2218 | 1.2340 | 0.0000 | no | yes |
| `matrixmultiply` | 1 | 3 | 0 | 1 | 1.0241 | 1.1860 | 0.1080 | no | yes |
| `monte_carlo_pi` | 1 | 3 | 0 | 0 | 0.2024 | 0.2040 | 0.0000 | no | yes |
| `mandelbrot` | 1 | 3 | 0 | 1 | 0.0533 | 0.1100 | 0.0539 | no | no |
| `nbody` | 1 | 3 | 0 | 1 | 0.0354 | 0.0980 | 0.0607 | no | no |
| `future_pipeline` | 1 | 3 | 0 | 3 | 0.0067 | 0.0360 | 0.0289 | no | no |
| `channel_rollup` | 1 | 3 | 0 | 1 | 0.0065 | 0.0420 | 0.0352 | no | no |
| `await_channel_mux` | 1 | 3 | 0 | 2 | 0.0057 | 0.1120 | 0.1060 | no | no |
| `mutex_ledger` | 1 | 3 | 0 | 3 | 0.0054 | 0.0360 | 0.0303 | no | no |
| `mutex_work_queue` | 1 | 3 | 0 | 1 | 0.0047 | 0.0460 | 0.0411 | no | no |
| `future_await_race` | 1 | 3 | 0 | 0 | 0.0044 | 0.0380 | 0.0334 | no | no |
| `mutex_await_journal` | 1 | 3 | 0 | 1 | 0.0043 | 0.0300 | 0.0255 | no | no |
| `regex_set_audit` | 1 | 2 | 0 | 1 | 0.0054 | 0.0680 | 0.0623 | no | no |
