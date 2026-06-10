# External Benchmark Comparison

- Generated: `2026-07-24T17:45:21.640063Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-24-execution-context-go-reference.json`
- Suite: `coverage`
- Able benchmarks: `fib, binarytrees, matrixmultiply, quicksort, sudoku_masks, i_before_e, base64, binary_event_log, json, monte_carlo_pi, pidigits, mandelbrot, reverse_complement, k_nucleotide, nbody, tapelang_alphabet, distance_field, rms_norm, fasta_generation, fixed_width_128, rational_series, wide_integer_records, word_frequency, document_audit, lexical_rollup, channel_rollup, future_pipeline, future_await_race, await_channel_mux, mutex_ledger, mutex_await_journal, mutex_work_queue, regex_suffix_audit, regex_set_audit, regex_stream_audit, log_routing_redaction, config_validation_extraction, unicode_scalar_pipeline, array_slice_window, dependency_plan, inventory_reconciliation, option_result_config, concurrent_text_index, validated_job_pipeline, dependency_wave_validation, concurrent_event_routing, concurrent_document_pipeline, concurrent_stencil_reduction, concurrent_signal_dispatch, concurrent_transform_chain, concurrent_policy_callbacks, concurrent_graph_visitors, concurrent_audio_voices, concurrent_packet_codecs, concurrent_scene_tiles, concurrent_tree_folds, concurrent_state_machines, concurrent_stateful_pipeline, manifest_normalization, policy_record_dispatch, sensor_calibration`
- Able modes: `compiled`
- Reference languages: `go`
- CPU pool: `0-3` (each row records its resolved catalog budget)
- Experimental execution context: `enabled for compiled mode`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `fib` | `compiled` | ok (5) | verified (5) | 231e154363ef69683f6dca862f5a448e786ccc90485636bc20a86e2e470a2188 | 3.4160 | 3.4907 | 0.98x |
| `binarytrees` | `compiled` | ok (5) | verified (5) | 341de11a51feab3d8122b4b5d6a68b038a2d14434aa9bc2372f39300bf5f48e1 | 9.7320 | 11.4050 | 0.85x |
| `matrixmultiply` | `compiled` | ok (5) | verified (5) | bce2053ea5fdc1351bf8d28398b346b894f8cf35c0e047effdca9dbb47e81d0c | 1.2700 | 1.0259 | 1.24x |
| `quicksort` | `compiled` | ok (5) | verified (5) | d0d07db0afd4266c1b6de5e76438bfa6aa974727e06c74e280aa7b497ca0e8b3 | 1.8880 | 2.7249 | 0.69x |
| `sudoku_masks` | `compiled` | ok (5) | verified (5) | 35a81e448daf9986f2a9b7c3a873dc6216bd55c969efec50c9b1de6d866659ec | 1.8380 | 0.6385 | 2.88x |
| `i_before_e` | `compiled` | ok (5) | verified (5) | 981f0d37a277be25f359c097e10df2ef68009fea2d3e322aeed3a65d0fbaca39 | 0.1420 | 0.0658 | 2.16x |
| `base64` | `compiled` | ok (5) | verified (5) | 5f4c00cd811078942fc98cd5dbca3b47fac2f8d8210f07ea116c1a7c0d6ac316 | 2.6900 | 2.6618 | 1.01x |
| `binary_event_log` | `compiled` | ok (5) | verified (5) | fb075dc8606582c1e6a1d5e520fa8dda237fc7304044b84b3f8f3a2c6b1c36e9 | 0.5560 | 0.0091 | 61.10x |
| `json` | `compiled` | ok (5) | verified (5) | 7cbf9736119f7683d59646bea566591a2ef278db54286b8e2525aa39e907839e | 0.7400 | 1.5555 | 0.48x |
| `monte_carlo_pi` | `compiled` | ok (5) | verified (5) | dab6ddc0a42f6644dd620c574b8b552697477be92433e886b7116ec8723123d2 | 0.2320 | 0.2058 | 1.13x |
| `pidigits` | `compiled` | ok (5) | verified (5) | bdfa7b6c756d96492f472f97aee9cc139bee954d271eacedfd7ace5d2875f06c | 1.4400 | 1.2871 | 1.12x |
| `mandelbrot` | `compiled` | ok (5) | verified (5) | d1560cadc49ab9dca30a4cfbe4123c1fcc0240ed80de27ad37d7fcc97159173e | 0.1420 | 0.0540 | 2.63x |
| `reverse_complement` | `compiled` | ok (5) | verified (5) | db06a593d68950aa91293fb57ffed87ad57040c4ac7ab81ae9dc43d50b400bb7 | 0.1100 | 0.0255 | 4.31x |
| `k_nucleotide` | `compiled` | ok (5) | verified (5) | d628623daa677c00673df8d4961d14eb271b4112cdcb65b8233f07b69d7b49b8 | 2.8920 | 0.0908 | 31.85x |
| `nbody` | `compiled` | ok (5) | verified (5) | 40799ff8af9b84a416e8bf940921658787c57be38f638fb4d98c735c8d39e820 | 0.1540 | 0.0375 | 4.11x |
| `tapelang_alphabet` | `compiled` | ok (5) | verified (5) | a8ac3a1054c1aa7ac25f9b1e652a96a7ac86a1c1130687fc53b90e20c766d149 | 3.7240 | 2.1771 | 1.71x |
| `distance_field` | `compiled` | ok (5) | verified (5) | cdaaf4451b236346af59b6a407f3136da96004e0c7c39c165546b7b9b21eda94 | 0.1340 | 0.0149 | 8.99x |
| `rms_norm` | `compiled` | ok (5) | verified (5) | 255c3e1c7ae7f523918e96244a6ac395b58699c4d2220549b097702faaa1037b | 0.0860 | 0.0116 | 7.41x |
| `fasta_generation` | `compiled` | ok (5) | verified (5) | 2907f3fb66fea247549c0f26b5b5d5cd1940a055574b72dad344283e1eb0fd10 | 0.1120 | 0.0142 | 7.89x |
| `fixed_width_128` | `compiled` | ok (5) | verified (5) | eceabf5869b1abca8d6dd228b64a09f89e4e98ba8cabc4833ffee1218dafa56a | 0.3380 | 0.0061 | 55.41x |
| `rational_series` | `compiled` | ok (5) | verified (5) | 127f0f44ee4870b57a188a7948f80a0a5d14584a326c345a48d4285594069f0c | 0.1780 | 0.0154 | 11.56x |
| `wide_integer_records` | `compiled` | ok (5) | verified (5) | f373537521cc6bfb0fb9e1a1eb36eb93a057654b526a4521878bc269261713e5 | 0.3400 | 0.0266 | 12.78x |
| `word_frequency` | `compiled` | ok (5) | verified (5) | 7dc1dae393e2c070eb0b9c9e611e154b2e6cce1b4a4268aa1bc73f8ff0e2fd07 | 0.1800 | 0.0057 | 31.58x |
| `document_audit` | `compiled` | ok (5) | verified (5) | 0dad030a80c8a883cbb56fbcfebfd530d521075e15d5d91ba538bc93e66c0aab | 0.1040 | 0.0043 | 24.19x |
| `lexical_rollup` | `compiled` | ok (5) | verified (5) | a6a1f91069e8c95a38fba1a3cb7fb3f582434245605091f200ee90cdb190e604 | 0.1020 | 0.0045 | 22.67x |
| `channel_rollup` | `compiled` | ok (5) | verified (5) | a6a1f91069e8c95a38fba1a3cb7fb3f582434245605091f200ee90cdb190e604 | 0.4060 | 0.0068 | 59.71x |
| `future_pipeline` | `compiled` | ok (5) | verified (5) | 3db937fdd21b5719ab57c25679a1944e8f30128695c8edbfae4940aba7aa6d98 | 0.1520 | 0.0066 | 23.03x |
| `future_await_race` | `compiled` | ok (5) | verified (5) | 33297798eeb96d6d471a6afe97ec8a8bf09eda0ce30e13917aed1db40fb930e4 | 0.0880 | 0.0051 | 17.25x |
| `await_channel_mux` | `compiled` | ok (5) | verified (5) | 0a32fa2e9481ecb635562079d993288d911326d9f36aeea0d9d3058d1a494693 | 0.3380 | 0.0059 | 57.29x |
| `mutex_ledger` | `compiled` | ok (5) | verified (5) | 58f2e2fda882a7c1e5032a2f34c4d9d05f40ced51199c0be0672ecc4d4cf14e4 | 0.7620 | 0.0055 | 138.55x |
| `mutex_await_journal` | `compiled` | ok (5) | verified (5) | e6f87830a6a1bb1bf3889a6b77eae053bfd3630055a06dd8fd500201586bf61e | 0.6360 | 0.0046 | 138.26x |
| `mutex_work_queue` | `compiled` | ok (5) | verified (5) | 57d3c0d15899da95d375a749cfe34dcc6942eb82a45427f197b9244b85ff8e58 | 1.9560 | 0.0048 | 407.50x |
| `regex_suffix_audit` | `compiled` | ok (5) | verified (5) | b5d5ccfabbfd4dc5952406cb1c42d62b807f75828661c4c3774b251abe38380f | 0.1220 | 0.0064 | 19.06x |
| `regex_set_audit` | `compiled` | ok (5) | verified (5) | 3d8f861a312416f95b95d59be62614f0ffc7918e86fb984fe6035ca7b7b28da2 | 0.1080 | 0.0066 | 16.36x |
| `regex_stream_audit` | `compiled` | ok (5) | verified (5) | dd7801f0b104d8bd47aaef64d685bc65a06263851f12c8df8cf75260d09e717b | 0.1020 | 0.0066 | 15.45x |
| `log_routing_redaction` | `compiled` | ok (5) | verified (5) | 0d9585b01f83904fdf11d47b2902678c1718c8442ed1d84410d61d5d90f60bf4 | 0.1100 | 0.0052 | 21.15x |
| `config_validation_extraction` | `compiled` | ok (5) | verified (5) | c1aa99b9a13bb6e0c7731cb2aea77e300cd3cecc695df7fd4af90036939341d1 | 0.0900 | 0.0047 | 19.15x |
| `unicode_scalar_pipeline` | `compiled` | ok (5) | verified (5) | c9efadb7f22969600334daa4a4eed2edde38c8e86d2c81d354d6f3979c854eb9 | 0.3520 | 0.0103 | 34.17x |
| `array_slice_window` | `compiled` | ok (5) | verified (5) | 155f89122475c7b282637dbf2ecba6d19771d396e801b581cb1d1b0cef64103e | 0.0860 | 0.0045 | 19.11x |
| `dependency_plan` | `compiled` | ok (5) | verified (5) | 96dc74508d9b7a476bafdef453b11e11f2f70279c58ccaa5dcb6d85c529c4a38 | 0.0900 | 0.0046 | 19.57x |
| `inventory_reconciliation` | `compiled` | ok (5) | verified (5) | 37dc8fb8d62ba404bd385852de356d7ccbd468da2b3d20b4f0a535263d6ffcf3 | 0.1940 | 0.0128 | 15.16x |
| `option_result_config` | `compiled` | ok (5) | verified (5) | 28e46b27a6dceeaa15968e9db7a6728f4a2b35f87a89ff7bf561db18cad53112 | 0.1680 | 0.0048 | 35.00x |
| `concurrent_text_index` | `compiled` | ok (5) | verified (5) | 3bdc437d713ad4ddcd2c37e7367ca95879766c927643a80803b3d0eba3c249d8 | 0.6460 | 0.0066 | 97.88x |
| `validated_job_pipeline` | `compiled` | ok (5) | verified (5) | ee5a3553094de6253bd71daaafd30e1db3eb9a17d11d3411b18ca652feafc40a | 1.1160 | 0.0047 | 237.45x |
| `dependency_wave_validation` | `compiled` | ok (5) | verified (5) | de00786abc00ecdc5be17fb12025be79aa7405a31b73c6c08ab7dbad9b555dc6 | 1.3040 | 0.0054 | 241.48x |
| `concurrent_event_routing` | `compiled` | ok (5) | verified (5) | 672e4542ce0d231428dd910348225e7c8b058824d7b8ceffbfb616d77a939d39 | 3.1040 | 0.0053 | 585.66x |
| `concurrent_document_pipeline` | `compiled` | ok (5) | verified (5) | 60b369f137cf022522072c4abfd911091aa3c77597528906f58b62610f438120 | 0.2800 | 0.0045 | 62.22x |
| `concurrent_stencil_reduction` | `compiled` | ok (5) | verified (5) | 42870ec44f0b8a860e066ec155ce13e2916bbff632d74a5c87704f7f81fa4a3b | 0.1740 | 0.0060 | 29.00x |
| `concurrent_signal_dispatch` | `compiled` | ok (5) | verified (5) | cb24d4b4ebe05455c69d9232b5bee08e27e1f864776c80996a0253ece9d69d94 | 0.1920 | 0.0050 | 38.40x |
| `concurrent_transform_chain` | `compiled` | ok (5) | verified (5) | 4695dbb758688b107a6c7917406311bad54f32f2f42310b2e6c2f58702468f57 | 8.4400 | 0.0061 | 1383.61x |
| `concurrent_policy_callbacks` | `compiled` | ok (5) | verified (5) | 7f1eeebf4548e851a416d06dddf41316bb9ebd4baa5f0e9e6b5265405883f210 | 0.5660 | 0.0048 | 117.92x |
| `concurrent_graph_visitors` | `compiled` | ok (5) | verified (5) | 399e7fd8eae623db1fb6fe83ba5d08b46a747fca2d7e51ec6536aab65b50c9ee | 0.2560 | 0.0051 | 50.20x |
| `concurrent_audio_voices` | `compiled` | ok (5) | verified (5) | 6ec28390bc9c749cf18e4a6fbd4bc03154345e1acf0a2f6601baab16884a6e28 | 0.8100 | 0.0046 | 176.09x |
| `concurrent_packet_codecs` | `compiled` | ok (5) | verified (5) | cf10b00cfd2619f5162ee99687f8e059e9333c4597169846846468fa20c230a5 | 0.4620 | 0.0043 | 107.44x |
| `concurrent_scene_tiles` | `compiled` | ok (5) | verified (5) | 2c60d0d812049ff61485905f50f92e678610e9708c24849927292dd7566a15fe | 0.3400 | 0.0045 | 75.56x |
| `concurrent_tree_folds` | `compiled` | ok (5) | verified (5) | ea4c3694c769a63e64e13afa98ee33a5a2edcd8d1ec0cd821c01bb257b2166d7 | 0.4140 | 0.0045 | 92.00x |
| `concurrent_state_machines` | `compiled` | ok (5) | verified (5) | 96296c1ea028df4cae0d4dde3e2f8a91533b7bb4daf1f19a611ea9b0ec2b0103 | 0.2300 | 0.0044 | 52.27x |
| `concurrent_stateful_pipeline` | `compiled` | ok (5) | verified (5) | b76b6e24e86beed9a7fc734ccfdf62266d67dea722d758656456824ecee96b67 | 0.7700 | 0.0057 | 135.09x |
| `manifest_normalization` | `compiled` | ok (5) | verified (5) | 2d6d55d5a76f3e45c6eb4fc3c0b892c2c5d8e02f3e38fa916d4f1c9a1579e9cb | 0.2160 | 0.0052 | 41.54x |
| `policy_record_dispatch` | `compiled` | ok (5) | verified (5) | f15c66e53e4c89650ae12aff6c2f466f2cf0c7adb6ca3fe5be0127bfba217c05 | 0.1840 | 0.0057 | 32.28x |
| `sensor_calibration` | `compiled` | ok (5) | verified (5) | e96cf1e366228f34478289660b4478b345bc069ac6e6633900d9805f0340edbb | 0.2660 | 0.0055 | 48.36x |
