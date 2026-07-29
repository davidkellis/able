# Pinned Go Reference Refresh

- Generated: `2026-07-29T15:41:47Z`
- Go: `go version go1.26.5 linux/amd64`
- Go toolchain selector: `go1.26.5`
- Runs: `5`
- Timeout: `90s`
- CPU pool: `5,10,3,6`
- GOMAXPROCS: `catalog CPU budget per row`

| Benchmark | Go status | Validation | Go real (s) | Source SHA-256 | Stdout SHA-256 |
| --- | --- | --- | ---: | --- | --- |
| `fib` | 5/5 (timeouts 0, failures 0) | verified | 3.1297 | d35840666fee5c1c2c45f44086e1c379331fb844bbba93e8542d8a848898d705 | 231e154363ef69683f6dca862f5a448e786ccc90485636bc20a86e2e470a2188 |
| `binarytrees` | 5/5 (timeouts 0, failures 0) | verified | 11.0316 | 27d067b4bfe5f501f1fe2d6a3e0254d699759c750296d9414161c2b5d3623b9f | 341de11a51feab3d8122b4b5d6a68b038a2d14434aa9bc2372f39300bf5f48e1 |
| `matrixmultiply` | 5/5 (timeouts 0, failures 0) | verified | 1.0960 | 4b77c1e4f0add1763c830c9da22a9d92c43d14e48432bab021e2146e3ccb1e42 | 0dfcf69f5c73589f22465d7054ec20cd1aa43a7a1829c57673b147a49290fc13 |
| `channel_rollup` | 5/5 (timeouts 0, failures 0) | verified | 0.0057 | ce97e496db8d6e776af7e6714d2ff51c0c55c3e7fcf098fa0db48c616f060ef7 | a6a1f91069e8c95a38fba1a3cb7fb3f582434245605091f200ee90cdb190e604 |
| `future_pipeline` | 5/5 (timeouts 0, failures 0) | verified | 0.0061 | 6b8d6b83a6086d48e1c62ec05f192bb2bc7a488f8e3ac6143cbb4a6b5c3df73c | 3db937fdd21b5719ab57c25679a1944e8f30128695c8edbfae4940aba7aa6d98 |
| `future_await_race` | 5/5 (timeouts 0, failures 0) | verified | 0.0049 | 04d296c4183b7e00c93369218788cac51325c8ba3a3a2cb97b64c81a5c2ac94b | 33297798eeb96d6d471a6afe97ec8a8bf09eda0ce30e13917aed1db40fb930e4 |
| `await_channel_mux` | 5/5 (timeouts 0, failures 0) | verified | 0.0054 | cbbe5f94b6c169e778c687ac010d9683ec1d49d17597a4a876b9d5ec4ab59745 | 0a32fa2e9481ecb635562079d993288d911326d9f36aeea0d9d3058d1a494693 |
| `mutex_ledger` | 5/5 (timeouts 0, failures 0) | verified | 0.0054 | 59d9c8fc20ccbf5807fa06c407ffcad1824b497d9d56fbf99d4106520aef5bef | 58f2e2fda882a7c1e5032a2f34c4d9d05f40ced51199c0be0672ecc4d4cf14e4 |
| `mutex_await_journal` | 5/5 (timeouts 0, failures 0) | verified | 0.0046 | e81b9ab6eae0badebaa7bce0797cd551702e55f064baa12313992933a525449b | e6f87830a6a1bb1bf3889a6b77eae053bfd3630055a06dd8fd500201586bf61e |
| `mutex_work_queue` | 5/5 (timeouts 0, failures 0) | verified | 0.0047 | c346c087fb68c8db955a88902c4c14c28531228465764bad82014008d5f1f736 | 57d3c0d15899da95d375a749cfe34dcc6942eb82a45427f197b9244b85ff8e58 |
| `concurrent_text_index` | 5/5 (timeouts 0, failures 0) | verified | 0.0072 | 8600a1c50b2dd113912d501c320b86dbffcb83cbeda9b73d96b089831d07db59 | 3bdc437d713ad4ddcd2c37e7367ca95879766c927643a80803b3d0eba3c249d8 |
| `validated_job_pipeline` | 5/5 (timeouts 0, failures 0) | verified | 0.0049 | 503d3dbbbf6767c1e0717ab659109a0adc4ef24557276f97e90ce770780157a3 | ee5a3553094de6253bd71daaafd30e1db3eb9a17d11d3411b18ca652feafc40a |
| `dependency_wave_validation` | 5/5 (timeouts 0, failures 0) | verified | 0.0049 | a1f24cf7900327dfb259cda9086ae131d6528e6eda53bdae1c925009362ea2ef | de00786abc00ecdc5be17fb12025be79aa7405a31b73c6c08ab7dbad9b555dc6 |
| `concurrent_event_routing` | 5/5 (timeouts 0, failures 0) | verified | 0.0048 | b6707e2193637e4a257a84f8ef9cf9208671e0d1bf9ed71b978ce5ef71659e1c | 672e4542ce0d231428dd910348225e7c8b058824d7b8ceffbfb616d77a939d39 |
| `concurrent_document_pipeline` | 5/5 (timeouts 0, failures 0) | verified | 0.0046 | bfdfcfe86bf39725d48357a348aded4a167aa0853a3a229998a6aebba18b6976 | 60b369f137cf022522072c4abfd911091aa3c77597528906f58b62610f438120 |
| `manifest_normalization` | 5/5 (timeouts 0, failures 0) | verified | 0.0056 | 6bbbec6c228c0d5c525bac71bee099af423e6af69f02e5e4d0e9d394352919b6 | 2d6d55d5a76f3e45c6eb4fc3c0b892c2c5d8e02f3e38fa916d4f1c9a1579e9cb |
| `policy_record_dispatch` | 5/5 (timeouts 0, failures 0) | verified | 0.0060 | a31c758f4f52941e200b527621253bf7559e55b3da8561671aab5dab0c813ed5 | f15c66e53e4c89650ae12aff6c2f466f2cf0c7adb6ca3fe5be0127bfba217c05 |
| `sensor_calibration` | 5/5 (timeouts 0, failures 0) | verified | 0.0049 | 89c3de53add27fe7b865a16fe379bb264c4681ae4d2e5d8802c5b83489f5e0d6 | e96cf1e366228f34478289660b4478b345bc069ac6e6633900d9805f0340edbb |
| `concurrent_stencil_reduction` | 5/5 (timeouts 0, failures 0) | verified | 0.0055 | ac0b3b8210a6c956b2fd4b97bdd0e477a6063dfc451e62d05efc72b82107cc20 | 42870ec44f0b8a860e066ec155ce13e2916bbff632d74a5c87704f7f81fa4a3b |
| `concurrent_signal_dispatch` | 5/5 (timeouts 0, failures 0) | verified | 0.0055 | fa2d920afaf8cc1fc02a8dd9182eace3a9d84797de5caec72dbd9c816d05bf50 | cb24d4b4ebe05455c69d9232b5bee08e27e1f864776c80996a0253ece9d69d94 |
| `concurrent_transform_chain` | 5/5 (timeouts 0, failures 0) | verified | 0.0057 | 5c5a2a2dba692a072f34bdc07a1ffdb69dbaa549b10b784dafc624dfca206995 | 4695dbb758688b107a6c7917406311bad54f32f2f42310b2e6c2f58702468f57 |
| `concurrent_policy_callbacks` | 5/5 (timeouts 0, failures 0) | verified | 0.0047 | 6a5074e704e2c8b0c33f7d85505ba27942e4e8c5754892f51fd7f21367fe457b | 7f1eeebf4548e851a416d06dddf41316bb9ebd4baa5f0e9e6b5265405883f210 |
| `concurrent_graph_visitors` | 5/5 (timeouts 0, failures 0) | verified | 0.0045 | 43bd8daf0556f88913cb7e883d7afbaf51247523ffe25152bf239c80761630d4 | 399e7fd8eae623db1fb6fe83ba5d08b46a747fca2d7e51ec6536aab65b50c9ee |
| `concurrent_audio_voices` | 5/5 (timeouts 0, failures 0) | verified | 0.0047 | e96bcffed4da40c14d70eff8a672acecaa03c232feb3f4a1735788a4613abd5c | 6ec28390bc9c749cf18e4a6fbd4bc03154345e1acf0a2f6601baab16884a6e28 |
| `concurrent_packet_codecs` | 5/5 (timeouts 0, failures 0) | verified | 0.0047 | 2f39269e10153deb8b9777298344d4f5f6750d682b14fb01583921e4da045323 | cf10b00cfd2619f5162ee99687f8e059e9333c4597169846846468fa20c230a5 |
| `concurrent_scene_tiles` | 5/5 (timeouts 0, failures 0) | verified | 0.0043 | ed49208100202d9390b5f7dad3d895cf475dcb444c8e60c137411f0c12828eff | 2c60d0d812049ff61485905f50f92e678610e9708c24849927292dd7566a15fe |
| `concurrent_tree_folds` | 5/5 (timeouts 0, failures 0) | verified | 0.0040 | d3529a8d32b897661b0246b2d3a4c200efea2a4b9938e958e6cf3761d1b07cbe | ea4c3694c769a63e64e13afa98ee33a5a2edcd8d1ec0cd821c01bb257b2166d7 |
| `concurrent_state_machines` | 5/5 (timeouts 0, failures 0) | verified | 0.0041 | 1d4548675ab3465bdc7acc64fccd236dd7b130aa0d7561180ddc027cbd02ce98 | 96296c1ea028df4cae0d4dde3e2f8a91533b7bb4daf1f19a611ea9b0ec2b0103 |
| `concurrent_stateful_pipeline` | 5/5 (timeouts 0, failures 0) | verified | 0.0052 | ffed025d01fa740bbe3c656a1a85514bea3316698138b8919be3dd53c2390047 | b76b6e24e86beed9a7fc734ccfdf62266d67dea722d758656456824ecee96b67 |
