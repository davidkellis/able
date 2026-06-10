# External Benchmark Comparison

- Generated: `2026-07-26T00:55:13.377710Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-25-current-compiled-go-reference.json`
- Suite: `custom`
- Able benchmarks: `fib, binarytrees, matrixmultiply, quicksort, sudoku_masks, i_before_e, base64, binary_event_log, json, monte_carlo_pi, pidigits, mandelbrot, reverse_complement, k_nucleotide, nbody, tapelang_alphabet, distance_field, rms_norm, fasta_generation, fixed_width_128, rational_series, wide_integer_records, word_frequency, document_audit, lexical_rollup, channel_rollup, future_pipeline, future_await_race, await_channel_mux, mutex_ledger, mutex_await_journal, mutex_work_queue, regex_suffix_audit, regex_set_audit, regex_stream_audit, log_routing_redaction, config_validation_extraction, unicode_scalar_pipeline, array_slice_window, dependency_plan, inventory_reconciliation, option_result_config, concurrent_text_index, validated_job_pipeline, dependency_wave_validation, concurrent_event_routing, concurrent_document_pipeline, concurrent_stencil_reduction, concurrent_signal_dispatch, concurrent_transform_chain, concurrent_policy_callbacks, concurrent_graph_visitors, concurrent_audio_voices, concurrent_packet_codecs, concurrent_scene_tiles, concurrent_tree_folds, concurrent_state_machines, concurrent_stateful_pipeline, manifest_normalization, policy_record_dispatch, sensor_calibration`
- Able modes: `compiled`
- Reference languages: `go`
- CPU pool: `5,10,15,11` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `fib` | `compiled` | ok (5) | verified (5) | 231e154363ef69683f6dca862f5a448e786ccc90485636bc20a86e2e470a2188 | 3.6340 | 3.3510 | 1.08x |
| `binarytrees` | `compiled` | ok (5) | verified (5) | 341de11a51feab3d8122b4b5d6a68b038a2d14434aa9bc2372f39300bf5f48e1 | 10.7780 | 10.6699 | 1.01x |
| `matrixmultiply` | `compiled` | ok (5) | verified (5) | bce2053ea5fdc1351bf8d28398b346b894f8cf35c0e047effdca9dbb47e81d0c | 1.1820 | 0.9801 | 1.21x |
| `quicksort` | `compiled` | ok (5) | verified (5) | d0d07db0afd4266c1b6de5e76438bfa6aa974727e06c74e280aa7b497ca0e8b3 | 1.8600 | 2.7835 | 0.67x |
| `sudoku_masks` | `compiled` | ok (5) | verified (5) | 35a81e448daf9986f2a9b7c3a873dc6216bd55c969efec50c9b1de6d866659ec | 1.8120 | 0.6968 | 2.60x |
| `i_before_e` | `compiled` | ok (5) | verified (5) | 981f0d37a277be25f359c097e10df2ef68009fea2d3e322aeed3a65d0fbaca39 | 0.0720 | 0.0831 | 0.87x |
| `base64` | `compiled` | ok (5) | verified (5) | 5f4c00cd811078942fc98cd5dbca3b47fac2f8d8210f07ea116c1a7c0d6ac316 | 2.2460 | 2.6730 | 0.84x |
| `binary_event_log` | `compiled` | error (1) | unavailable | n/a | n/a | 0.0085 | n/a |
| `json` | `compiled` | ok (5) | verified (5) | 7cbf9736119f7683d59646bea566591a2ef278db54286b8e2525aa39e907839e | 0.6760 | 1.5470 | 0.44x |
| `monte_carlo_pi` | `compiled` | ok (5) | verified (5) | dab6ddc0a42f6644dd620c574b8b552697477be92433e886b7116ec8723123d2 | 0.1560 | 0.2018 | 0.77x |
| `pidigits` | `compiled` | ok (5) | verified (5) | bdfa7b6c756d96492f472f97aee9cc139bee954d271eacedfd7ace5d2875f06c | 1.2640 | 1.2117 | 1.04x |
| `mandelbrot` | `compiled` | ok (5) | verified (5) | d1560cadc49ab9dca30a4cfbe4123c1fcc0240ed80de27ad37d7fcc97159173e | 0.0860 | 0.0504 | 1.71x |
| `reverse_complement` | `compiled` | ok (5) | verified (5) | db06a593d68950aa91293fb57ffed87ad57040c4ac7ab81ae9dc43d50b400bb7 | 0.0460 | 0.0143 | 3.22x |
| `k_nucleotide` | `compiled` | ok (5) | verified (5) | d628623daa677c00673df8d4961d14eb271b4112cdcb65b8233f07b69d7b49b8 | 1.6460 | 0.0556 | 29.60x |
| `nbody` | `compiled` | ok (5) | verified (5) | 40799ff8af9b84a416e8bf940921658787c57be38f638fb4d98c735c8d39e820 | 0.0900 | 0.0323 | 2.79x |
| `tapelang_alphabet` | `compiled` | ok (5) | verified (5) | a8ac3a1054c1aa7ac25f9b1e652a96a7ac86a1c1130687fc53b90e20c766d149 | 3.8180 | 2.0027 | 1.91x |
| `distance_field` | `compiled` | ok (5) | verified (5) | cdaaf4451b236346af59b6a407f3136da96004e0c7c39c165546b7b9b21eda94 | 0.0300 | 0.0139 | 2.16x |
| `rms_norm` | `compiled` | ok (5) | verified (5) | 255c3e1c7ae7f523918e96244a6ac395b58699c4d2220549b097702faaa1037b | 0.0340 | 0.0116 | 2.93x |
| `fasta_generation` | `compiled` | ok (5) | verified (5) | 2907f3fb66fea247549c0f26b5b5d5cd1940a055574b72dad344283e1eb0fd10 | 0.0500 | 0.0145 | 3.45x |
| `fixed_width_128` | `compiled` | ok (5) | verified (5) | eceabf5869b1abca8d6dd228b64a09f89e4e98ba8cabc4833ffee1218dafa56a | 0.1000 | 0.0079 | 12.66x |
| `rational_series` | `compiled` | ok (5) | verified (5) | 127f0f44ee4870b57a188a7948f80a0a5d14584a326c345a48d4285594069f0c | 0.0600 | 0.0139 | 4.32x |
| `wide_integer_records` | `compiled` | ok (5) | verified (5) | f373537521cc6bfb0fb9e1a1eb36eb93a057654b526a4521878bc269261713e5 | 0.1180 | 0.0258 | 4.57x |
| `word_frequency` | `compiled` | ok (5) | verified (5) | 7dc1dae393e2c070eb0b9c9e611e154b2e6cce1b4a4268aa1bc73f8ff0e2fd07 | 0.0520 | 0.0051 | 10.20x |
| `document_audit` | `compiled` | ok (5) | verified (5) | 0dad030a80c8a883cbb56fbcfebfd530d521075e15d5d91ba538bc93e66c0aab | 0.0340 | 0.0043 | 7.91x |
| `lexical_rollup` | `compiled` | ok (5) | verified (5) | a6a1f91069e8c95a38fba1a3cb7fb3f582434245605091f200ee90cdb190e604 | 0.0420 | 0.0048 | 8.75x |
| `channel_rollup` | `compiled` | ok (5) | verified (5) | a6a1f91069e8c95a38fba1a3cb7fb3f582434245605091f200ee90cdb190e604 | 0.0420 | 0.0058 | 7.24x |
| `future_pipeline` | `compiled` | ok (5) | verified (5) | 3db937fdd21b5719ab57c25679a1944e8f30128695c8edbfae4940aba7aa6d98 | 0.0400 | 0.0052 | 7.69x |
| `future_await_race` | `compiled` | ok (5) | verified (5) | 33297798eeb96d6d471a6afe97ec8a8bf09eda0ce30e13917aed1db40fb930e4 | 0.0520 | 0.0044 | 11.82x |
| `await_channel_mux` | `compiled` | ok (5) | verified (5) | 0a32fa2e9481ecb635562079d993288d911326d9f36aeea0d9d3058d1a494693 | 0.3160 | 0.0053 | 59.62x |
| `mutex_ledger` | `compiled` | error (1) | verified (5) | 58f2e2fda882a7c1e5032a2f34c4d9d05f40ced51199c0be0672ecc4d4cf14e4 | 0.1050 | 0.0048 | n/a |
| `mutex_await_journal` | `compiled` | ok (5) | verified (5) | e6f87830a6a1bb1bf3889a6b77eae053bfd3630055a06dd8fd500201586bf61e | 0.4280 | 0.0041 | 104.39x |
| `mutex_work_queue` | `compiled` | ok (5) | verified (5) | 57d3c0d15899da95d375a749cfe34dcc6942eb82a45427f197b9244b85ff8e58 | 0.9880 | 0.0047 | 210.21x |
| `regex_suffix_audit` | `compiled` | error (1) | unavailable | n/a | n/a | 0.0046 | n/a |
| `regex_set_audit` | `compiled` | error (1) | unavailable | n/a | n/a | 0.0051 | n/a |
| `regex_stream_audit` | `compiled` | error (1) | unavailable | n/a | n/a | 0.0051 | n/a |
| `log_routing_redaction` | `compiled` | error (1) | unavailable | n/a | n/a | 0.0045 | n/a |
| `config_validation_extraction` | `compiled` | error (1) | unavailable | n/a | n/a | 0.0041 | n/a |
| `unicode_scalar_pipeline` | `compiled` | ok (5) | verified (5) | c9efadb7f22969600334daa4a4eed2edde38c8e86d2c81d354d6f3979c854eb9 | 0.1460 | 0.0093 | 15.70x |
| `array_slice_window` | `compiled` | ok (5) | verified (5) | 155f89122475c7b282637dbf2ecba6d19771d396e801b581cb1d1b0cef64103e | 0.0500 | 0.0041 | 12.20x |
| `dependency_plan` | `compiled` | ok (5) | verified (5) | 96dc74508d9b7a476bafdef453b11e11f2f70279c58ccaa5dcb6d85c529c4a38 | 0.0240 | 0.0053 | 4.53x |
| `inventory_reconciliation` | `compiled` | ok (5) | verified (5) | 37dc8fb8d62ba404bd385852de356d7ccbd468da2b3d20b4f0a535263d6ffcf3 | 0.1520 | 0.0138 | 11.01x |
| `option_result_config` | `compiled` | ok (5) | verified (5) | 28e46b27a6dceeaa15968e9db7a6728f4a2b35f87a89ff7bf561db18cad53112 | 0.0460 | 0.0046 | 10.00x |
| `concurrent_text_index` | `compiled` | ok (5) | verified (5) | 3bdc437d713ad4ddcd2c37e7367ca95879766c927643a80803b3d0eba3c249d8 | 0.0520 | 0.0071 | 7.32x |
| `validated_job_pipeline` | `compiled` | ok (5) | verified (5) | ee5a3553094de6253bd71daaafd30e1db3eb9a17d11d3411b18ca652feafc40a | 0.0560 | 0.0044 | 12.73x |
| `dependency_wave_validation` | `compiled` | error (1) | verified (5) | de00786abc00ecdc5be17fb12025be79aa7405a31b73c6c08ab7dbad9b555dc6 | 0.0400 | 0.0045 | n/a |
| `concurrent_event_routing` | `compiled` | ok (5) | verified (5) | 672e4542ce0d231428dd910348225e7c8b058824d7b8ceffbfb616d77a939d39 | 0.0320 | 0.0053 | 6.04x |
| `concurrent_document_pipeline` | `compiled` | ok (5) | verified (5) | 60b369f137cf022522072c4abfd911091aa3c77597528906f58b62610f438120 | 0.0320 | 0.0042 | 7.62x |
| `concurrent_stencil_reduction` | `compiled` | ok (5) | verified (5) | 42870ec44f0b8a860e066ec155ce13e2916bbff632d74a5c87704f7f81fa4a3b | 0.0320 | 0.0060 | 5.33x |
| `concurrent_signal_dispatch` | `compiled` | ok (5) | verified (5) | cb24d4b4ebe05455c69d9232b5bee08e27e1f864776c80996a0253ece9d69d94 | 0.0400 | 0.0051 | 7.84x |
| `concurrent_transform_chain` | `compiled` | ok (5) | verified (5) | 4695dbb758688b107a6c7917406311bad54f32f2f42310b2e6c2f58702468f57 | 0.0380 | 0.0052 | 7.31x |
| `concurrent_policy_callbacks` | `compiled` | ok (5) | verified (5) | 7f1eeebf4548e851a416d06dddf41316bb9ebd4baa5f0e9e6b5265405883f210 | 0.0380 | 0.0045 | 8.44x |
| `concurrent_graph_visitors` | `compiled` | ok (5) | verified (5) | 399e7fd8eae623db1fb6fe83ba5d08b46a747fca2d7e51ec6536aab65b50c9ee | 0.0460 | 0.0039 | 11.79x |
| `concurrent_audio_voices` | `compiled` | ok (5) | verified (5) | 6ec28390bc9c749cf18e4a6fbd4bc03154345e1acf0a2f6601baab16884a6e28 | 0.0360 | 0.0038 | 9.47x |
| `concurrent_packet_codecs` | `compiled` | ok (5) | verified (5) | cf10b00cfd2619f5162ee99687f8e059e9333c4597169846846468fa20c230a5 | 0.0340 | 0.0038 | 8.95x |
| `concurrent_scene_tiles` | `compiled` | ok (5) | verified (5) | 2c60d0d812049ff61485905f50f92e678610e9708c24849927292dd7566a15fe | 0.0300 | 0.0038 | 7.89x |
| `concurrent_tree_folds` | `compiled` | ok (5) | verified (5) | ea4c3694c769a63e64e13afa98ee33a5a2edcd8d1ec0cd821c01bb257b2166d7 | 0.0320 | 0.0040 | 8.00x |
| `concurrent_state_machines` | `compiled` | error (4) | verified (5) | 96296c1ea028df4cae0d4dde3e2f8a91533b7bb4daf1f19a611ea9b0ec2b0103 | 0.0300 | 0.0043 | n/a |
| `concurrent_stateful_pipeline` | `compiled` | ok (5) | verified (5) | b76b6e24e86beed9a7fc734ccfdf62266d67dea722d758656456824ecee96b67 | 0.0480 | 0.0045 | 10.67x |
| `manifest_normalization` | `compiled` | ok (5) | verified (5) | 2d6d55d5a76f3e45c6eb4fc3c0b892c2c5d8e02f3e38fa916d4f1c9a1579e9cb | 0.0520 | 0.0063 | 8.25x |
| `policy_record_dispatch` | `compiled` | error (1) | unavailable | n/a | n/a | 0.0067 | n/a |
| `sensor_calibration` | `compiled` | ok (5) | verified (5) | e96cf1e366228f34478289660b4478b345bc069ac6e6633900d9805f0340edbb | 0.0520 | 0.0047 | 11.06x |
