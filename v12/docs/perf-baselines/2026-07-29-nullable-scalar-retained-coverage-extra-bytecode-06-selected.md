# External Benchmark Comparison

- Generated: `2026-07-29T23:59:12.829005Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-29-nullable-scalar-retained-coverage-extra-interpreter-reference-selected.json`
- Suite: `custom`
- Able benchmarks: `concurrent_text_index, validated_job_pipeline, dependency_wave_validation, concurrent_event_routing, concurrent_document_pipeline, manifest_normalization, policy_record_dispatch, sensor_calibration, transaction_ledger_audit, generic_slot_buffer, concurrent_stencil_reduction, concurrent_signal_dispatch, concurrent_transform_chain, concurrent_policy_callbacks, concurrent_graph_visitors, concurrent_audio_voices, concurrent_packet_codecs, concurrent_scene_tiles, concurrent_tree_folds, concurrent_state_machines, concurrent_stateful_pipeline`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU pool: `12-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `concurrent_text_index` | `bytecode` | ok (5) | verified (5) | 3bdc437d713ad4ddcd2c37e7367ca95879766c927643a80803b3d0eba3c249d8 | 0.7540 | 0.1017 | 7.41x | 0.0815 | 9.25x |
| `validated_job_pipeline` | `bytecode` | ok (5) | verified (5) | ee5a3553094de6253bd71daaafd30e1db3eb9a17d11d3411b18ca652feafc40a | 0.6260 | 0.0255 | 24.55x | 0.0595 | 10.52x |
| `dependency_wave_validation` | `bytecode` | ok (5) | verified (5) | de00786abc00ecdc5be17fb12025be79aa7405a31b73c6c08ab7dbad9b555dc6 | 0.8020 | 0.0323 | 24.83x | 0.0745 | 10.77x |
| `concurrent_event_routing` | `bytecode` | ok (5) | verified (5) | 672e4542ce0d231428dd910348225e7c8b058824d7b8ceffbfb616d77a939d39 | 4.6940 | 0.0516 | 90.97x | 0.0691 | 67.93x |
| `concurrent_document_pipeline` | `bytecode` | ok (5) | verified (5) | 60b369f137cf022522072c4abfd911091aa3c77597528906f58b62610f438120 | 0.2800 | 0.0235 | 11.91x | 0.0464 | 6.03x |
| `manifest_normalization` | `bytecode` | ok (5) | verified (5) | 2d6d55d5a76f3e45c6eb4fc3c0b892c2c5d8e02f3e38fa916d4f1c9a1579e9cb | 1.4920 | 0.0209 | 71.39x | 0.0683 | 21.84x |
| `policy_record_dispatch` | `bytecode` | ok (5) | verified (5) | f15c66e53e4c89650ae12aff6c2f466f2cf0c7adb6ca3fe5be0127bfba217c05 | 7.8620 | 0.0276 | 284.86x | 0.0610 | 128.89x |
| `sensor_calibration` | `bytecode` | ok (5) | verified (5) | e96cf1e366228f34478289660b4478b345bc069ac6e6633900d9805f0340edbb | 2.8680 | 0.0322 | 89.07x | 0.0961 | 29.84x |
| `transaction_ledger_audit` | `bytecode` | ok (5) | verified (5) | aa5a4fe7f85ce13998797ef506647a93f16e0ee747613683268fd801d609c812 | 4.5880 | 0.0382 | 120.10x | 0.1235 | 37.15x |
| `generic_slot_buffer` | `bytecode` | ok (5) | verified (5) | 149cd95dcb57f9309c82ccd148336280f98baa95ea3d91ba34be7989fdab06fe | 2.2400 | 0.1923 | 11.65x | 0.1149 | 19.50x |
| `concurrent_stencil_reduction` | `bytecode` | ok (5) | verified (5) | 42870ec44f0b8a860e066ec155ce13e2916bbff632d74a5c87704f7f81fa4a3b | 1.8020 | 0.0846 | 21.30x | 0.1105 | 16.31x |
| `concurrent_signal_dispatch` | `bytecode` | ok (5) | verified (5) | cb24d4b4ebe05455c69d9232b5bee08e27e1f864776c80996a0253ece9d69d94 | 1.6100 | 0.0716 | 22.49x | 0.0887 | 18.15x |
| `concurrent_transform_chain` | `bytecode` | ok (5) | verified (5) | 4695dbb758688b107a6c7917406311bad54f32f2f42310b2e6c2f58702468f57 | 2.6920 | 0.1297 | 20.76x | 0.1331 | 20.23x |
| `concurrent_policy_callbacks` | `bytecode` | ok (5) | verified (5) | 7f1eeebf4548e851a416d06dddf41316bb9ebd4baa5f0e9e6b5265405883f210 | 0.4280 | 0.0566 | 7.56x | 0.0544 | 7.87x |
| `concurrent_graph_visitors` | `bytecode` | ok (5) | verified (5) | 399e7fd8eae623db1fb6fe83ba5d08b46a747fca2d7e51ec6536aab65b50c9ee | 1.2980 | 0.0715 | 18.15x | 0.0698 | 18.60x |
| `concurrent_audio_voices` | `bytecode` | ok (5) | verified (5) | 6ec28390bc9c749cf18e4a6fbd4bc03154345e1acf0a2f6601baab16884a6e28 | 1.4960 | 0.1374 | 10.89x | 0.1177 | 12.71x |
| `concurrent_packet_codecs` | `bytecode` | ok (5) | verified (5) | cf10b00cfd2619f5162ee99687f8e059e9333c4597169846846468fa20c230a5 | 0.8200 | 0.0923 | 8.88x | 0.1003 | 8.18x |
| `concurrent_scene_tiles` | `bytecode` | ok (5) | verified (5) | 2c60d0d812049ff61485905f50f92e678610e9708c24849927292dd7566a15fe | 0.6480 | 0.0773 | 8.38x | 0.0778 | 8.33x |
| `concurrent_tree_folds` | `bytecode` | ok (5) | verified (5) | ea4c3694c769a63e64e13afa98ee33a5a2edcd8d1ec0cd821c01bb257b2166d7 | 0.4540 | 0.0689 | 6.59x | 0.0691 | 6.57x |
| `concurrent_state_machines` | `bytecode` | ok (5) | verified (5) | 96296c1ea028df4cae0d4dde3e2f8a91533b7bb4daf1f19a611ea9b0ec2b0103 | 0.3660 | 0.0647 | 5.66x | 0.0631 | 5.80x |
| `concurrent_stateful_pipeline` | `bytecode` | ok (5) | verified (5) | b76b6e24e86beed9a7fc734ccfdf62266d67dea722d758656456824ecee96b67 | 0.5000 | 0.0687 | 7.28x | 0.0573 | 8.73x |
