# External Benchmark Comparison

- Generated: `2026-07-28T20:26:09.989405Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-28-mode-aware-benchmark-contract-coverage-extra-interpreter-reference-selected.json`
- Suite: `custom`
- Able benchmarks: `concurrent_text_index, validated_job_pipeline, dependency_wave_validation, concurrent_event_routing, concurrent_document_pipeline, manifest_normalization, policy_record_dispatch, sensor_calibration, concurrent_stencil_reduction, concurrent_signal_dispatch, concurrent_transform_chain, concurrent_policy_callbacks, concurrent_graph_visitors, concurrent_audio_voices, concurrent_packet_codecs, concurrent_scene_tiles, concurrent_tree_folds, concurrent_state_machines, concurrent_stateful_pipeline`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU pool: `7-10` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `concurrent_text_index` | `bytecode` | ok (5) | verified (5) | 3bdc437d713ad4ddcd2c37e7367ca95879766c927643a80803b3d0eba3c249d8 | 0.6020 | 0.0584 | 10.31x | 0.0783 | 7.69x |
| `validated_job_pipeline` | `bytecode` | ok (5) | verified (5) | ee5a3553094de6253bd71daaafd30e1db3eb9a17d11d3411b18ca652feafc40a | 0.3700 | 0.0232 | 15.95x | 0.0485 | 7.63x |
| `dependency_wave_validation` | `bytecode` | ok (5) | verified (5) | de00786abc00ecdc5be17fb12025be79aa7405a31b73c6c08ab7dbad9b555dc6 | 0.5080 | 0.0312 | 16.28x | 0.0494 | 10.28x |
| `concurrent_event_routing` | `bytecode` | ok (5) | verified (5) | 672e4542ce0d231428dd910348225e7c8b058824d7b8ceffbfb616d77a939d39 | 2.9540 | 0.0312 | 94.68x | 0.0578 | 51.11x |
| `concurrent_document_pipeline` | `bytecode` | ok (5) | verified (5) | 60b369f137cf022522072c4abfd911091aa3c77597528906f58b62610f438120 | 0.2860 | 0.0246 | 11.63x | 0.0509 | 5.62x |
| `manifest_normalization` | `bytecode` | ok (5) | verified (5) | 2d6d55d5a76f3e45c6eb4fc3c0b892c2c5d8e02f3e38fa916d4f1c9a1579e9cb | 1.5360 | 0.0172 | 89.30x | 0.0522 | 29.43x |
| `policy_record_dispatch` | `bytecode` | ok (5) | verified (5) | f15c66e53e4c89650ae12aff6c2f466f2cf0c7adb6ca3fe5be0127bfba217c05 | 7.5820 | 0.0202 | 375.35x | 0.0450 | 168.49x |
| `sensor_calibration` | `bytecode` | ok (5) | verified (5) | e96cf1e366228f34478289660b4478b345bc069ac6e6633900d9805f0340edbb | 2.6240 | 0.0277 | 94.73x | 0.0715 | 36.70x |
| `concurrent_stencil_reduction` | `bytecode` | ok (5) | verified (5) | 42870ec44f0b8a860e066ec155ce13e2916bbff632d74a5c87704f7f81fa4a3b | 1.8100 | 0.0751 | 24.10x | 0.0956 | 18.93x |
| `concurrent_signal_dispatch` | `bytecode` | ok (5) | verified (5) | cb24d4b4ebe05455c69d9232b5bee08e27e1f864776c80996a0253ece9d69d94 | 1.5600 | 0.0631 | 24.72x | 0.1054 | 14.80x |
| `concurrent_transform_chain` | `bytecode` | ok (5) | verified (5) | 4695dbb758688b107a6c7917406311bad54f32f2f42310b2e6c2f58702468f57 | 2.7360 | 0.1521 | 17.99x | 0.1392 | 19.66x |
| `concurrent_policy_callbacks` | `bytecode` | ok (5) | verified (5) | 7f1eeebf4548e851a416d06dddf41316bb9ebd4baa5f0e9e6b5265405883f210 | 0.3860 | 0.0493 | 7.83x | 0.0657 | 5.88x |
| `concurrent_graph_visitors` | `bytecode` | ok (5) | verified (5) | 399e7fd8eae623db1fb6fe83ba5d08b46a747fca2d7e51ec6536aab65b50c9ee | 1.2900 | 0.0762 | 16.93x | 0.0640 | 20.16x |
| `concurrent_audio_voices` | `bytecode` | ok (5) | verified (5) | 6ec28390bc9c749cf18e4a6fbd4bc03154345e1acf0a2f6601baab16884a6e28 | 1.3540 | 0.1256 | 10.78x | 0.1269 | 10.67x |
| `concurrent_packet_codecs` | `bytecode` | ok (5) | verified (5) | cf10b00cfd2619f5162ee99687f8e059e9333c4597169846846468fa20c230a5 | 0.8380 | 0.0779 | 10.76x | 0.0822 | 10.19x |
| `concurrent_scene_tiles` | `bytecode` | ok (5) | verified (5) | 2c60d0d812049ff61485905f50f92e678610e9708c24849927292dd7566a15fe | 0.6400 | 0.0717 | 8.93x | 0.0711 | 9.00x |
| `concurrent_tree_folds` | `bytecode` | ok (5) | verified (5) | ea4c3694c769a63e64e13afa98ee33a5a2edcd8d1ec0cd821c01bb257b2166d7 | 0.4260 | 0.0658 | 6.47x | 0.0607 | 7.02x |
| `concurrent_state_machines` | `bytecode` | ok (5) | verified (5) | 96296c1ea028df4cae0d4dde3e2f8a91533b7bb4daf1f19a611ea9b0ec2b0103 | 0.3900 | 0.0609 | 6.40x | 0.0613 | 6.36x |
| `concurrent_stateful_pipeline` | `bytecode` | ok (5) | verified (5) | b76b6e24e86beed9a7fc734ccfdf62266d67dea722d758656456824ecee96b67 | 0.4560 | 0.0679 | 6.72x | 0.0561 | 8.13x |
