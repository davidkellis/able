# Bytecode Opcode Audit

- Generated: `2026-07-20T20:43:00Z`
- Suite: `corpus-full`
- Benchmarks: `120`
- Functions lowered: `460`
- Instructions lowered: `21955`

| Benchmark | Functions | Instructions | `LoadName` | `LoadSlot` | `LoadImplicitSlot` | `LoadSlotI32` | `StoreImplicitSlot` | `Match` | `JumpIfNotTypedPattern` | `LoadSlotStructField` | `TryFloatUpdatePair` | `JumpIfFloatMulAddMulCompareConstFalse` | `JumpIfFloatAddCompareConstFalse` | `StoreSlotFloatAddMulSlot` | `StoreSlotFloatAddSub` |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| `fib` | 2 | 9 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `binarytrees` | 3 | 228 | 37 | 7 | 0 | 0 | 0 | 0 | 2 | 2 | 0 | 0 | 0 | 0 | 0 |
| `matrixmultiply` | 3 | 231 | 0 | 42 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `quicksort` | 4 | 225 | 0 | 32 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `sudoku_masks` | 9 | 589 | 0 | 126 | 0 | 0 | 0 | 0 | 2 | 0 | 0 | 0 | 0 | 0 | 0 |
| `i_before_e` | 2 | 87 | 0 | 12 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `base64` | 3 | 98 | 0 | 13 | 0 | 0 | 0 | 0 | 2 | 0 | 0 | 0 | 0 | 0 | 0 |
| `json` | 3 | 65 | 0 | 12 | 0 | 0 | 0 | 0 | 2 | 0 | 0 | 0 | 0 | 0 | 0 |
| `monte_carlo_pi` | 3 | 65 | 0 | 4 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 1 | 0 | 0 | 0 |
| `pidigits` | 6 | 338 | 0 | 86 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `mandelbrot` | 3 | 140 | 5 | 17 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 1 | 0 | 1 | 0 |
| `reverse_complement` | 6 | 359 | 0 | 60 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `k_nucleotide` | 17 | 652 | 19 | 79 | 0 | 0 | 0 | 0 | 1 | 6 | 0 | 0 | 0 | 0 | 0 |
| `nbody` | 4 | 710 | 0 | 225 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `tapelang_alphabet` | 13 | 420 | 13 | 95 | 0 | 0 | 0 | 0 | 2 | 6 | 0 | 0 | 0 | 0 | 0 |
| `distance_field` | 1 | 58 | 0 | 7 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `rms_norm` | 1 | 79 | 0 | 8 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `fasta_generation` | 8 | 220 | 9 | 25 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `fixed_width_128` | 5 | 137 | 0 | 19 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `rational_series` | 3 | 155 | 0 | 18 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `wide_integer_records` | 5 | 305 | 6 | 46 | 0 | 0 | 0 | 0 | 4 | 0 | 0 | 0 | 0 | 0 | 0 |
| `word_frequency` | 2 | 121 | 0 | 19 | 4 | 0 | 3 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| `document_audit` | 3 | 116 | 6 | 8 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `lexical_rollup` | 3 | 103 | 5 | 8 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `channel_rollup` | 3 | 219 | 14 | 8 | 0 | 0 | 0 | 0 | 2 | 0 | 0 | 0 | 0 | 0 | 0 |
| `future_pipeline` | 4 | 309 | 23 | 15 | 0 | 0 | 0 | 0 | 3 | 0 | 0 | 0 | 0 | 0 | 0 |
| `future_await_race` | 2 | 175 | 13 | 1 | 0 | 0 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| `await_channel_mux` | 2 | 165 | 13 | 3 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `mutex_ledger` | 3 | 169 | 27 | 3 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `mutex_await_journal` | 3 | 170 | 26 | 3 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `mutex_work_queue` | 5 | 160 | 20 | 8 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `regex_suffix_audit` | 4 | 262 | 9 | 19 | 4 | 0 | 4 | 0 | 6 | 0 | 0 | 0 | 0 | 0 | 0 |
| `regex_set_audit` | 3 | 186 | 0 | 23 | 4 | 0 | 4 | 0 | 2 | 0 | 0 | 0 | 0 | 0 | 0 |
| `regex_stream_audit` | 3 | 256 | 0 | 30 | 6 | 0 | 6 | 0 | 6 | 0 | 0 | 0 | 0 | 0 | 0 |
| `log_routing_redaction` | 6 | 323 | 4 | 49 | 4 | 0 | 4 | 0 | 6 | 0 | 0 | 0 | 0 | 0 | 0 |
| `config_validation_extraction` | 5 | 347 | 0 | 56 | 6 | 0 | 6 | 0 | 4 | 0 | 0 | 0 | 0 | 0 | 0 |
| `unicode_scalar_pipeline` | 4 | 140 | 0 | 14 | 5 | 0 | 4 | 0 | 4 | 0 | 0 | 0 | 0 | 0 | 0 |
| `array_slice_window` | 4 | 179 | 5 | 23 | 4 | 0 | 4 | 0 | 4 | 0 | 0 | 0 | 0 | 0 | 0 |
| `dependency_plan` | 4 | 299 | 2 | 27 | 21 | 0 | 8 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| `inventory_reconciliation` | 6 | 193 | 8 | 22 | 6 | 0 | 5 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| `option_result_config` | 10 | 174 | 15 | 2 | 3 | 1 | 3 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `array_filter_i32_small` | 3 | 90 | 0 | 4 | 6 | 0 | 5 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `array_fold_i32_small` | 2 | 77 | 0 | 3 | 4 | 0 | 3 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `array_map_i32_small` | 3 | 90 | 0 | 4 | 6 | 0 | 5 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `ascii_lower_small` | 4 | 121 | 0 | 13 | 4 | 0 | 4 | 0 | 4 | 0 | 0 | 0 | 0 | 0 | 0 |
| `automata_dfa_small` | 6 | 223 | 0 | 30 | 6 | 0 | 5 | 0 | 2 | 0 | 0 | 0 | 0 | 0 | 0 |
| `await_batch_i64_small` | 3 | 98 | 6 | 2 | 4 | 0 | 4 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `base64_roundtrip_small` | 3 | 98 | 0 | 13 | 0 | 0 | 0 | 0 | 2 | 0 | 0 | 0 | 0 | 0 | 0 |
| `bigint_add_mul_small` | 1 | 58 | 0 | 8 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `bigint_ref_newton_small` | 3 | 188 | 0 | 42 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `biguint_add_mul_small` | 1 | 58 | 0 | 8 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `binarytrees_small` | 3 | 228 | 37 | 7 | 0 | 0 | 0 | 0 | 2 | 2 | 0 | 0 | 0 | 0 | 0 |
| `bit_set_small` | 3 | 134 | 13 | 3 | 2 | 0 | 2 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `boolean_reconciliation_small` | 3 | 121 | 0 | 9 | 5 | 0 | 4 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `byte_histogram_small` | 4 | 183 | 0 | 19 | 8 | 0 | 5 | 0 | 4 | 0 | 0 | 0 | 0 | 0 | 0 |
| `channel_pipeline_i32_small` | 2 | 156 | 10 | 0 | 2 | 0 | 2 | 0 | 2 | 0 | 0 | 0 | 0 | 0 | 0 |
| `channel_roundtrip_i32_small` | 2 | 108 | 8 | 0 | 2 | 0 | 2 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| `concurrent_queue_i32_small` | 3 | 128 | 10 | 3 | 4 | 0 | 3 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `crc32_small` | 3 | 125 | 0 | 9 | 4 | 0 | 3 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `deque_i32_small` | 3 | 143 | 12 | 4 | 5 | 0 | 3 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `dijkstra_heap_small` | 7 | 305 | 1 | 37 | 21 | 0 | 9 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| `fib_i32_small` | 2 | 9 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `future_fanout_i32_small` | 1 | 142 | 20 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `future_yield_i32_small` | 1 | 118 | 18 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `graph_bfs_small` | 4 | 245 | 0 | 27 | 24 | 0 | 10 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| `hash_set_i32_small` | 3 | 150 | 15 | 2 | 4 | 0 | 3 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `hashmap_i32_small` | 3 | 137 | 17 | 2 | 2 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `heap_i32_small` | 3 | 91 | 0 | 6 | 6 | 0 | 5 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| `i_before_e_small` | 2 | 87 | 0 | 12 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `int128_accumulate_small` | 2 | 80 | 0 | 9 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `iterator_match_identifier_small` | 3 | 85 | 0 | 7 | 4 | 0 | 3 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| `json_means_small` | 5 | 139 | 0 | 18 | 0 | 0 | 0 | 0 | 2 | 0 | 0 | 0 | 0 | 0 | 0 |
| `k_nucleotide_small` | 17 | 652 | 19 | 79 | 0 | 0 | 0 | 0 | 1 | 6 | 0 | 0 | 0 | 0 | 0 |
| `knapsack_i32_small` | 4 | 205 | 0 | 19 | 13 | 0 | 6 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `lazy_seq_cache_i32_small` | 5 | 157 | 0 | 12 | 8 | 0 | 6 | 0 | 2 | 0 | 0 | 0 | 0 | 0 | 0 |
| `lazy_seq_take_i32_small` | 3 | 90 | 0 | 4 | 6 | 0 | 5 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `levenshtein_small` | 5 | 278 | 0 | 30 | 14 | 3 | 5 | 0 | 2 | 0 | 0 | 0 | 0 | 0 | 0 |
| `linked_list_enumerable_i32_small` | 3 | 66 | 0 | 4 | 4 | 0 | 3 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `linked_list_for_i32_small` | 4 | 108 | 0 | 5 | 9 | 0 | 8 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `linked_list_iterator_collect_i64_small` | 3 | 62 | 0 | 2 | 4 | 0 | 3 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `linked_list_iterator_filter_map_i64_small` | 3 | 60 | 0 | 2 | 4 | 0 | 3 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `linked_list_iterator_pipeline_i64_small` | 3 | 89 | 0 | 6 | 6 | 0 | 5 | 0 | 2 | 0 | 0 | 0 | 0 | 0 | 0 |
| `list_i32_small` | 3 | 90 | 0 | 4 | 6 | 0 | 6 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `mandelbrot_small` | 3 | 145 | 7 | 17 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 1 | 0 | 1 | 0 |
| `matrixmultiply_f64_small` | 3 | 225 | 0 | 36 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `md5_hex_small` | 3 | 95 | 0 | 7 | 6 | 0 | 5 | 0 | 2 | 0 | 0 | 0 | 0 | 0 | 0 |
| `monte_carlo_pi_small` | 3 | 65 | 0 | 4 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 1 | 0 | 0 | 0 |
| `mutex_counter_i32_small` | 2 | 72 | 6 | 1 | 2 | 0 | 2 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `nbody_small` | 4 | 713 | 0 | 225 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `nominal_recurrence_small` | 2 | 50 | 0 | 2 | 0 | 0 | 0 | 0 | 0 | 5 | 0 | 0 | 0 | 0 | 0 |
| `noop` | 1 | 3 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `persistent_map_i32_small` | 3 | 192 | 24 | 4 | 0 | 0 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| `persistent_map_string_small` | 4 | 189 | 24 | 3 | 0 | 0 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| `persistent_queue_i32_small` | 3 | 126 | 10 | 1 | 4 | 0 | 4 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `persistent_set_i32_small` | 3 | 150 | 18 | 3 | 2 | 0 | 2 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `persistent_sorted_set_i32_small` | 3 | 199 | 28 | 1 | 2 | 0 | 2 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `pidigits_small` | 6 | 338 | 0 | 86 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `queue_i32_small` | 3 | 116 | 8 | 2 | 4 | 0 | 3 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `quicksort_file_small` | 4 | 242 | 0 | 35 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `random_lcg_i64_small` | 1 | 43 | 0 | 3 | 2 | 0 | 2 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `rational_series_small` | 1 | 68 | 0 | 5 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `regex_is_match_small` | 4 | 202 | 0 | 22 | 6 | 0 | 5 | 0 | 4 | 0 | 0 | 0 | 0 | 0 | 0 |
| `reverse_complement_small` | 6 | 359 | 0 | 60 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `run_length_encode_small` | 4 | 184 | 0 | 28 | 2 | 0 | 2 | 0 | 4 | 0 | 0 | 0 | 0 | 0 | 0 |
| `sieve_count` | 1 | 73 | 0 | 6 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `sieve_full` | 1 | 140 | 0 | 16 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `string_builder_small` | 4 | 170 | 0 | 21 | 4 | 0 | 4 | 0 | 4 | 0 | 0 | 0 | 0 | 0 | 0 |
| `string_contains_small` | 3 | 124 | 0 | 12 | 7 | 0 | 7 | 0 | 2 | 0 | 0 | 0 | 0 | 0 | 0 |
| `string_split_join_small` | 3 | 130 | 0 | 15 | 4 | 0 | 4 | 0 | 2 | 0 | 0 | 0 | 0 | 0 | 0 |
| `sudoku_file_small` | 6 | 413 | 0 | 68 | 0 | 0 | 0 | 0 | 2 | 0 | 0 | 0 | 0 | 0 | 0 |
| `sum_u32_small` | 3 | 90 | 0 | 2 | 6 | 0 | 5 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `tapelang_small` | 14 | 448 | 13 | 97 | 2 | 0 | 2 | 0 | 2 | 6 | 0 | 0 | 0 | 0 | 0 |
| `toposort_small` | 4 | 292 | 0 | 25 | 21 | 0 | 8 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| `tree_map_i32_small` | 3 | 177 | 18 | 2 | 4 | 0 | 3 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| `tree_set_i32_small` | 3 | 140 | 14 | 1 | 4 | 0 | 3 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `uint128_accumulate_small` | 2 | 78 | 0 | 9 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `union_find_small` | 6 | 192 | 0 | 18 | 6 | 0 | 4 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `vector_i32_small` | 3 | 138 | 17 | 1 | 2 | 0 | 2 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| `word_count_small` | 7 | 227 | 0 | 31 | 7 | 0 | 5 | 0 | 3 | 0 | 0 | 0 | 0 | 0 | 0 |
| `zigzag_char_small` | 4 | 239 | 0 | 16 | 9 | 0 | 4 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |

## Totals

| Opcode | Total Hits | Benchmarks With Hits |
| --- | ---: | ---: |
| `LoadName` | 652 | 46 |
| `LoadSlot` | 2627 | 113 |
| `LoadImplicitSlot` | 370 | 62 |
| `LoadSlotI32` | 4 | 2 |
| `StoreImplicitSlot` | 260 | 62 |
| `Match` | 0 | 0 |
| `JumpIfNotTypedPattern` | 115 | 49 |
| `LoadSlotStructField` | 33 | 7 |
| `TryFloatUpdatePair` | 0 | 0 |
| `JumpIfFloatMulAddMulCompareConstFalse` | 4 | 4 |
| `JumpIfFloatAddCompareConstFalse` | 0 | 0 |
| `StoreSlotFloatAddMulSlot` | 2 | 2 |
| `StoreSlotFloatAddSub` | 0 | 0 |
