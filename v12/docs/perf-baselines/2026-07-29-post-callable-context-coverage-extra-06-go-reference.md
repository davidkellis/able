# Pinned Go Reference Refresh

- Generated: `2026-07-29T12:24:47Z`
- Go: `go version go1.26.5 linux/amd64`
- Go toolchain selector: `go1.26.5`
- Runs: `5`
- Timeout: `90s`
- CPU pool: `7-10`
- GOMAXPROCS: `catalog CPU budget per row`

| Benchmark | Go status | Validation | Go real (s) | Source SHA-256 | Stdout SHA-256 |
| --- | --- | --- | ---: | --- | --- |
| `concurrent_text_index` | 5/5 (timeouts 0, failures 0) | verified | 0.0084 | 8600a1c50b2dd113912d501c320b86dbffcb83cbeda9b73d96b089831d07db59 | 3bdc437d713ad4ddcd2c37e7367ca95879766c927643a80803b3d0eba3c249d8 |
| `validated_job_pipeline` | 5/5 (timeouts 0, failures 0) | verified | 0.0053 | 503d3dbbbf6767c1e0717ab659109a0adc4ef24557276f97e90ce770780157a3 | ee5a3553094de6253bd71daaafd30e1db3eb9a17d11d3411b18ca652feafc40a |
| `dependency_wave_validation` | 5/5 (timeouts 0, failures 0) | verified | 0.0060 | a1f24cf7900327dfb259cda9086ae131d6528e6eda53bdae1c925009362ea2ef | de00786abc00ecdc5be17fb12025be79aa7405a31b73c6c08ab7dbad9b555dc6 |
| `concurrent_event_routing` | 5/5 (timeouts 0, failures 0) | verified | 0.0060 | b6707e2193637e4a257a84f8ef9cf9208671e0d1bf9ed71b978ce5ef71659e1c | 672e4542ce0d231428dd910348225e7c8b058824d7b8ceffbfb616d77a939d39 |
| `concurrent_document_pipeline` | 5/5 (timeouts 0, failures 0) | verified | 0.0046 | bfdfcfe86bf39725d48357a348aded4a167aa0853a3a229998a6aebba18b6976 | 60b369f137cf022522072c4abfd911091aa3c77597528906f58b62610f438120 |
| `manifest_normalization` | 5/5 (timeouts 0, failures 0) | verified | 0.0055 | 6bbbec6c228c0d5c525bac71bee099af423e6af69f02e5e4d0e9d394352919b6 | 2d6d55d5a76f3e45c6eb4fc3c0b892c2c5d8e02f3e38fa916d4f1c9a1579e9cb |
| `policy_record_dispatch` | 5/5 (timeouts 0, failures 0) | verified | 0.0059 | a31c758f4f52941e200b527621253bf7559e55b3da8561671aab5dab0c813ed5 | f15c66e53e4c89650ae12aff6c2f466f2cf0c7adb6ca3fe5be0127bfba217c05 |
| `sensor_calibration` | 5/5 (timeouts 0, failures 0) | verified | 0.0066 | 89c3de53add27fe7b865a16fe379bb264c4681ae4d2e5d8802c5b83489f5e0d6 | e96cf1e366228f34478289660b4478b345bc069ac6e6633900d9805f0340edbb |
| `concurrent_stencil_reduction` | 5/5 (timeouts 0, failures 0) | verified | 0.0059 | ac0b3b8210a6c956b2fd4b97bdd0e477a6063dfc451e62d05efc72b82107cc20 | 42870ec44f0b8a860e066ec155ce13e2916bbff632d74a5c87704f7f81fa4a3b |
| `concurrent_signal_dispatch` | 5/5 (timeouts 0, failures 0) | verified | 0.0067 | fa2d920afaf8cc1fc02a8dd9182eace3a9d84797de5caec72dbd9c816d05bf50 | cb24d4b4ebe05455c69d9232b5bee08e27e1f864776c80996a0253ece9d69d94 |
| `concurrent_transform_chain` | 5/5 (timeouts 0, failures 0) | verified | 0.0074 | 5c5a2a2dba692a072f34bdc07a1ffdb69dbaa549b10b784dafc624dfca206995 | 4695dbb758688b107a6c7917406311bad54f32f2f42310b2e6c2f58702468f57 |
| `concurrent_policy_callbacks` | 5/5 (timeouts 0, failures 0) | verified | 0.0062 | 6a5074e704e2c8b0c33f7d85505ba27942e4e8c5754892f51fd7f21367fe457b | 7f1eeebf4548e851a416d06dddf41316bb9ebd4baa5f0e9e6b5265405883f210 |
| `concurrent_graph_visitors` | 5/5 (timeouts 0, failures 0) | verified | 0.0047 | 43bd8daf0556f88913cb7e883d7afbaf51247523ffe25152bf239c80761630d4 | 399e7fd8eae623db1fb6fe83ba5d08b46a747fca2d7e51ec6536aab65b50c9ee |
| `concurrent_audio_voices` | 5/5 (timeouts 0, failures 0) | verified | 0.0061 | e96bcffed4da40c14d70eff8a672acecaa03c232feb3f4a1735788a4613abd5c | 6ec28390bc9c749cf18e4a6fbd4bc03154345e1acf0a2f6601baab16884a6e28 |
| `concurrent_packet_codecs` | 5/5 (timeouts 0, failures 0) | verified | 0.0052 | 2f39269e10153deb8b9777298344d4f5f6750d682b14fb01583921e4da045323 | cf10b00cfd2619f5162ee99687f8e059e9333c4597169846846468fa20c230a5 |
| `concurrent_scene_tiles` | 5/5 (timeouts 0, failures 0) | verified | 0.0050 | ed49208100202d9390b5f7dad3d895cf475dcb444c8e60c137411f0c12828eff | 2c60d0d812049ff61485905f50f92e678610e9708c24849927292dd7566a15fe |
| `concurrent_tree_folds` | 5/5 (timeouts 0, failures 0) | verified | 0.0055 | d3529a8d32b897661b0246b2d3a4c200efea2a4b9938e958e6cf3761d1b07cbe | ea4c3694c769a63e64e13afa98ee33a5a2edcd8d1ec0cd821c01bb257b2166d7 |
| `concurrent_state_machines` | 5/5 (timeouts 0, failures 0) | verified | 0.0056 | 1d4548675ab3465bdc7acc64fccd236dd7b130aa0d7561180ddc027cbd02ce98 | 96296c1ea028df4cae0d4dde3e2f8a91533b7bb4daf1f19a611ea9b0ec2b0103 |
| `concurrent_stateful_pipeline` | 5/5 (timeouts 0, failures 0) | verified | 0.0052 | ffed025d01fa740bbe3c656a1a85514bea3316698138b8919be3dd53c2390047 | b76b6e24e86beed9a7fc734ccfdf62266d67dea722d758656456824ecee96b67 |
