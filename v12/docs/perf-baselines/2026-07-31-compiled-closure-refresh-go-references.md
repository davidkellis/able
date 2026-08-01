# Pinned Go Reference Refresh

- Generated: `2026-07-31T22:19:46Z`
- Go: `go version go1.26.5 linux/amd64`
- Go toolchain selector: `go1.26.5`
- Runs: `5`
- Timeout: `55s`
- CPU pool: `12-15`
- GOMAXPROCS: `catalog CPU budget per row`

| Benchmark | Go status | Validation | Go real (s) | Source SHA-256 | Stdout SHA-256 |
| --- | --- | --- | ---: | --- | --- |
| `array_slice_window` | 5/5 (timeouts 0, failures 0) | verified | 0.0050 | b05ef51494077012bf4dd3822d8b36edaf10331d7de66ea6532a02f86e5d0402 | 155f89122475c7b282637dbf2ecba6d19771d396e801b581cb1d1b0cef64103e |
| `await_channel_mux` | 5/5 (timeouts 0, failures 0) | verified | 0.0047 | cbbe5f94b6c169e778c687ac010d9683ec1d49d17597a4a876b9d5ec4ab59745 | 0a32fa2e9481ecb635562079d993288d911326d9f36aeea0d9d3058d1a494693 |
| `backup_dedup` | 5/5 (timeouts 0, failures 0) | verified | 0.0138 | 6df8f4b980b668c9e5a01926ff3ce1b50eaad803296f55728328a5e69a803c5e | bf4d5c89239bd78c6dcb9d755b8df4e90bc092c2a64bf15e45786e815918504e |
| `base64` | 5/5 (timeouts 0, failures 0) | verified | 2.4163 | 61a41a0ec45d3b3a8c890853d0a7839ce371b758ded8d8b8d0129a5b28390af6 | 5f4c00cd811078942fc98cd5dbca3b47fac2f8d8210f07ea116c1a7c0d6ac316 |
| `binary_event_log` | 5/5 (timeouts 0, failures 0) | verified | 0.0080 | f9d5afe5d21a6bd2f73cb9c8ca9dd003fcd0b1993f695846ea82343f5c2bd2bf | fb075dc8606582c1e6a1d5e520fa8dda237fc7304044b84b3f8f3a2c6b1c36e9 |
| `binarytrees` | 5/5 (timeouts 0, failures 0) | verified | 11.9591 | 27d067b4bfe5f501f1fe2d6a3e0254d699759c750296d9414161c2b5d3623b9f | 341de11a51feab3d8122b4b5d6a68b038a2d14434aa9bc2372f39300bf5f48e1 |
| `channel_rollup` | 5/5 (timeouts 0, failures 0) | verified | 0.0061 | ce97e496db8d6e776af7e6714d2ff51c0c55c3e7fcf098fa0db48c616f060ef7 | a6a1f91069e8c95a38fba1a3cb7fb3f582434245605091f200ee90cdb190e604 |
| `concurrent_audio_voices` | 5/5 (timeouts 0, failures 0) | verified | 0.0053 | e96bcffed4da40c14d70eff8a672acecaa03c232feb3f4a1735788a4613abd5c | 6ec28390bc9c749cf18e4a6fbd4bc03154345e1acf0a2f6601baab16884a6e28 |
| `concurrent_document_pipeline` | 5/5 (timeouts 0, failures 0) | verified | 0.0047 | bfdfcfe86bf39725d48357a348aded4a167aa0853a3a229998a6aebba18b6976 | 60b369f137cf022522072c4abfd911091aa3c77597528906f58b62610f438120 |
| `concurrent_event_routing` | 5/5 (timeouts 0, failures 0) | verified | 0.0064 | b6707e2193637e4a257a84f8ef9cf9208671e0d1bf9ed71b978ce5ef71659e1c | 672e4542ce0d231428dd910348225e7c8b058824d7b8ceffbfb616d77a939d39 |
| `concurrent_graph_visitors` | 5/5 (timeouts 0, failures 0) | verified | 0.0048 | 43bd8daf0556f88913cb7e883d7afbaf51247523ffe25152bf239c80761630d4 | 399e7fd8eae623db1fb6fe83ba5d08b46a747fca2d7e51ec6536aab65b50c9ee |
| `concurrent_packet_codecs` | 5/5 (timeouts 0, failures 0) | verified | 0.0048 | 2f39269e10153deb8b9777298344d4f5f6750d682b14fb01583921e4da045323 | cf10b00cfd2619f5162ee99687f8e059e9333c4597169846846468fa20c230a5 |
| `concurrent_policy_callbacks` | 5/5 (timeouts 0, failures 0) | verified | 0.0070 | 6a5074e704e2c8b0c33f7d85505ba27942e4e8c5754892f51fd7f21367fe457b | 7f1eeebf4548e851a416d06dddf41316bb9ebd4baa5f0e9e6b5265405883f210 |
| `concurrent_scene_tiles` | 5/5 (timeouts 0, failures 0) | verified | 0.0052 | ed49208100202d9390b5f7dad3d895cf475dcb444c8e60c137411f0c12828eff | 2c60d0d812049ff61485905f50f92e678610e9708c24849927292dd7566a15fe |
| `concurrent_signal_dispatch` | 5/5 (timeouts 0, failures 0) | verified | 0.0058 | fa2d920afaf8cc1fc02a8dd9182eace3a9d84797de5caec72dbd9c816d05bf50 | cb24d4b4ebe05455c69d9232b5bee08e27e1f864776c80996a0253ece9d69d94 |
| `concurrent_state_machines` | 5/5 (timeouts 0, failures 0) | verified | 0.0065 | 1d4548675ab3465bdc7acc64fccd236dd7b130aa0d7561180ddc027cbd02ce98 | 96296c1ea028df4cae0d4dde3e2f8a91533b7bb4daf1f19a611ea9b0ec2b0103 |
| `concurrent_stateful_pipeline` | 5/5 (timeouts 0, failures 0) | verified | 0.0074 | ffed025d01fa740bbe3c656a1a85514bea3316698138b8919be3dd53c2390047 | b76b6e24e86beed9a7fc734ccfdf62266d67dea722d758656456824ecee96b67 |
| `concurrent_stencil_reduction` | 5/5 (timeouts 0, failures 0) | verified | 0.0061 | ac0b3b8210a6c956b2fd4b97bdd0e477a6063dfc451e62d05efc72b82107cc20 | 42870ec44f0b8a860e066ec155ce13e2916bbff632d74a5c87704f7f81fa4a3b |
| `concurrent_text_index` | 5/5 (timeouts 0, failures 0) | verified | 0.0123 | 8600a1c50b2dd113912d501c320b86dbffcb83cbeda9b73d96b089831d07db59 | 3bdc437d713ad4ddcd2c37e7367ca95879766c927643a80803b3d0eba3c249d8 |
| `concurrent_transform_chain` | 5/5 (timeouts 0, failures 0) | verified | 0.0068 | 5c5a2a2dba692a072f34bdc07a1ffdb69dbaa549b10b784dafc624dfca206995 | 4695dbb758688b107a6c7917406311bad54f32f2f42310b2e6c2f58702468f57 |
| `concurrent_tree_folds` | 5/5 (timeouts 0, failures 0) | verified | 0.0049 | d3529a8d32b897661b0246b2d3a4c200efea2a4b9938e958e6cf3761d1b07cbe | ea4c3694c769a63e64e13afa98ee33a5a2edcd8d1ec0cd821c01bb257b2166d7 |
| `config_validation_extraction` | 5/5 (timeouts 0, failures 0) | verified | 0.0057 | dd0a7582e9ac0b83d571d57000fa30cf2c09fbed5f9bd0988283755aeef74f3e | c1aa99b9a13bb6e0c7731cb2aea77e300cd3cecc695df7fd4af90036939341d1 |
| `dependency_plan` | 5/5 (timeouts 0, failures 0) | verified | 0.0054 | c0f1e22de8cdbb92823d70897d9fd87f52df8a2b814789d8831ffd655618c8d3 | 96dc74508d9b7a476bafdef453b11e11f2f70279c58ccaa5dcb6d85c529c4a38 |
| `dependency_wave_validation` | 5/5 (timeouts 0, failures 0) | verified | 0.0055 | a1f24cf7900327dfb259cda9086ae131d6528e6eda53bdae1c925009362ea2ef | de00786abc00ecdc5be17fb12025be79aa7405a31b73c6c08ab7dbad9b555dc6 |
| `discrete_event_simulation` | 5/5 (timeouts 0, failures 0) | verified | 0.0148 | 48504c9fbf21421f3738007ccfa3cac00c0e1abba0156c3133578fcaa121250f | 6aebca9b31a78441438d2321290a7b66dc831ddbc7671d783e4a725aed6e7405 |
| `distance_field` | 5/5 (timeouts 0, failures 0) | verified | 0.0156 | 418832deca45aa420b724aeca415542cc5d3eac0ce5b9ad1c6305654e52a03b8 | 114cd92849943d55ca4824ca4f820d00a4f7c732223a5da0fc6fa937c1a3a113 |
| `document_audit` | 5/5 (timeouts 0, failures 0) | verified | 0.0055 | efca73563a21cd01ff3d844acc1f666176d810b92de4ad61db362a93781cf39c | 0dad030a80c8a883cbb56fbcfebfd530d521075e15d5d91ba538bc93e66c0aab |
| `fasta_generation` | 5/5 (timeouts 0, failures 0) | verified | 0.0173 | 116e80452758cc68483d2a8da625204794ff1260878a84b1907c2703c82ae029 | 2907f3fb66fea247549c0f26b5b5d5cd1940a055574b72dad344283e1eb0fd10 |
| `fib` | 5/5 (timeouts 0, failures 0) | verified | 3.5367 | d35840666fee5c1c2c45f44086e1c379331fb844bbba93e8542d8a848898d705 | 231e154363ef69683f6dca862f5a448e786ccc90485636bc20a86e2e470a2188 |
| `fixed_width_128` | 5/5 (timeouts 0, failures 0) | verified | 0.0064 | 4125d090eb3cb657b79c2fc02d3170d7aaa62588f92a1e504f0140d1dbd65094 | eceabf5869b1abca8d6dd228b64a09f89e4e98ba8cabc4833ffee1218dafa56a |
| `future_await_race` | 5/5 (timeouts 0, failures 0) | verified | 0.0045 | 04d296c4183b7e00c93369218788cac51325c8ba3a3a2cb97b64c81a5c2ac94b | 33297798eeb96d6d471a6afe97ec8a8bf09eda0ce30e13917aed1db40fb930e4 |
| `future_pipeline` | 5/5 (timeouts 0, failures 0) | verified | 0.0070 | 6b8d6b83a6086d48e1c62ec05f192bb2bc7a488f8e3ac6143cbb4a6b5c3df73c | 3db937fdd21b5719ab57c25679a1944e8f30128695c8edbfae4940aba7aa6d98 |
| `generic_slot_buffer` | 5/5 (timeouts 0, failures 0) | verified | 0.0058 | c3b6e6b4003fb32e6a5a68d2a9fd37e95a3b532a4e5bdd979273a109a2dcd5da | 149cd95dcb57f9309c82ccd148336280f98baa95ea3d91ba34be7989fdab06fe |
| `i_before_e` | 5/5 (timeouts 0, failures 0) | verified | 0.0718 | 63386a111f2fd35ff949092a419c99ce7dcf21e81b24575ece330f7729df65c0 | 981f0d37a277be25f359c097e10df2ef68009fea2d3e322aeed3a65d0fbaca39 |
| `inventory_reconciliation` | 5/5 (timeouts 0, failures 0) | verified | 0.0092 | 958b1a85bff5e471739a502ff0af5eaf886fade65c6601aeee74b69d8ae76b44 | 37dc8fb8d62ba404bd385852de356d7ccbd468da2b3d20b4f0a535263d6ffcf3 |
| `json` | 5/5 (timeouts 0, failures 0) | verified | 1.5717 | 6ea05b659c17322d9fd009c1b6f7ca3862e4621d2d854bbae25cd2eaccbc91c1 | 16d35d81fed277412d781cad8942b00f198974d4de5981bfb61142fe73b7e8e1 |
| `k_nucleotide` | 5/5 (timeouts 0, failures 0) | verified | 0.0620 | 8f1ec0923f819b16a7a63adbb1f3d7165c526e7df73462e8edf865d6a39c9a29 | d37cb398c9d9b1f1f02b33a8861aac8490334241ec92b18f325a7789e619d515 |
| `lexical_rollup` | 5/5 (timeouts 0, failures 0) | verified | 0.0041 | e1ce0907e22b60082acbc836da6b66b9629895ce44fefd441e451d4744936c99 | a6a1f91069e8c95a38fba1a3cb7fb3f582434245605091f200ee90cdb190e604 |
| `log_routing_redaction` | 5/5 (timeouts 0, failures 0) | verified | 0.0049 | 4b9ee58bfba7e36b5c2bd530cc0b3977eb5a8ce22e14c1d5efaa656e224ce855 | 0d9585b01f83904fdf11d47b2902678c1718c8442ed1d84410d61d5d90f60bf4 |
| `mandelbrot` | 5/5 (timeouts 0, failures 0) | verified | 0.0620 | c0a81d428de3e5b86ec9980b514441c8388dadfc550ed345a1f6271dcd8f0b4c | d1560cadc49ab9dca30a4cfbe4123c1fcc0240ed80de27ad37d7fcc97159173e |
| `manifest_normalization` | 5/5 (timeouts 0, failures 0) | verified | 0.0043 | 6bbbec6c228c0d5c525bac71bee099af423e6af69f02e5e4d0e9d394352919b6 | 2d6d55d5a76f3e45c6eb4fc3c0b892c2c5d8e02f3e38fa916d4f1c9a1579e9cb |
| `matrixmultiply` | 5/5 (timeouts 0, failures 0) | verified | 1.0344 | 4b77c1e4f0add1763c830c9da22a9d92c43d14e48432bab021e2146e3ccb1e42 | 0dfcf69f5c73589f22465d7054ec20cd1aa43a7a1829c57673b147a49290fc13 |
| `monte_carlo_pi` | 5/5 (timeouts 0, failures 0) | verified-nondeterministic | 0.2537 | 34b6ad655126f97e453b17350d49552144c2bcd332fb3b3ca7192382554e3877 | 1a14bf7bbe1e28c16059b3fd513061819ec93f84f14a8e574b1b518680e3844b,3785482cfcc84716df6155daf7042da06cb6bd39780aa9cee8657c3ebc1c63de,854f55a874c144d03bbc204fe68858635531c209a04f340e7d1338aad22002e2,929db04e16061d756e4da3e451342bf3f9acbb4004a9540879ddbe26c88f52d2,eeb4cf61e6f7aab030d9c2f6decebf4bc7833c8d0c7a4ea0f42a0cab3871710a |
| `mutex_await_journal` | 5/5 (timeouts 0, failures 0) | verified | 0.0046 | e81b9ab6eae0badebaa7bce0797cd551702e55f064baa12313992933a525449b | e6f87830a6a1bb1bf3889a6b77eae053bfd3630055a06dd8fd500201586bf61e |
| `mutex_ledger` | 5/5 (timeouts 0, failures 0) | verified | 0.0053 | 59d9c8fc20ccbf5807fa06c407ffcad1824b497d9d56fbf99d4106520aef5bef | 58f2e2fda882a7c1e5032a2f34c4d9d05f40ced51199c0be0672ecc4d4cf14e4 |
| `mutex_work_queue` | 5/5 (timeouts 0, failures 0) | verified | 0.0050 | c346c087fb68c8db955a88902c4c14c28531228465764bad82014008d5f1f736 | 57d3c0d15899da95d375a749cfe34dcc6942eb82a45427f197b9244b85ff8e58 |
| `nbody` | 5/5 (timeouts 0, failures 0) | verified | 0.0387 | 8c7ecf3a42b99d2b35b69b861f844edbe7bb160f76c6530384ddf894cff4d628 | 7fee18aa4de449f07aa173bae7a37df103ff0317ed8055566e6f3c9358c09b2c |
| `option_result_config` | 5/5 (timeouts 0, failures 0) | verified | 0.0038 | 3ced806f4f42b53e457f102494807d04b0bab5a93f76c9a4a1338efedcd2d68a | 28e46b27a6dceeaa15968e9db7a6728f4a2b35f87a89ff7bf561db18cad53112 |
| `pidigits` | 5/5 (timeouts 0, failures 0) | verified | 1.2436 | c8669a71e52ce32ed4e6852547efec42f67e8b8f5656e88653720d74b37da58c | bdfa7b6c756d96492f472f97aee9cc139bee954d271eacedfd7ace5d2875f06c |
| `policy_record_dispatch` | 5/5 (timeouts 0, failures 0) | verified | 0.0053 | a31c758f4f52941e200b527621253bf7559e55b3da8561671aab5dab0c813ed5 | f15c66e53e4c89650ae12aff6c2f466f2cf0c7adb6ca3fe5be0127bfba217c05 |
| `quicksort` | 5/5 (timeouts 0, failures 0) | verified | 2.7145 | 24e8c148e54ff9ecad02af8477de362ff6da6ba5707709a44c4ff1360e393103 | d0d07db0afd4266c1b6de5e76438bfa6aa974727e06c74e280aa7b497ca0e8b3 |
| `rational_series` | 5/5 (timeouts 0, failures 0) | verified | 0.0135 | 118406800ba92ccce86446d5a61b837f9f75e7edcc771e7b04ddd33f260bd652 | 127f0f44ee4870b57a188a7948f80a0a5d14584a326c345a48d4285594069f0c |
| `regex_set_audit` | 5/5 (timeouts 0, failures 0) | verified | 0.0046 | ce6d01f4d9331f3ee52ff73bcca39637e226c14f3c045942f5a4765e0c3995ed | 3d8f861a312416f95b95d59be62614f0ffc7918e86fb984fe6035ca7b7b28da2 |
| `regex_stream_audit` | 5/5 (timeouts 0, failures 0) | verified | 0.0046 | 1f45fe2a76d146be18659ac9dd6eeab88548c5afe22e2a3ccc5ac3c8c0d82be9 | dd7801f0b104d8bd47aaef64d685bc65a06263851f12c8df8cf75260d09e717b |
| `regex_suffix_audit` | 5/5 (timeouts 0, failures 0) | verified | 0.0047 | 4424b20eadfabba04dc90132f1606bdcd33be424fbe37b99041a2d10b356ca98 | b5d5ccfabbfd4dc5952406cb1c42d62b807f75828661c4c3774b251abe38380f |
| `reverse_complement` | 5/5 (timeouts 0, failures 0) | verified | 0.0172 | d6ab5b73111cf5dbc06e1f3879ccc8548d082d88b778cfcba731a2fa3aaacd74 | db06a593d68950aa91293fb57ffed87ad57040c4ac7ab81ae9dc43d50b400bb7 |
| `rms_norm` | 5/5 (timeouts 0, failures 0) | verified | 0.0126 | 70f4c52128ffdffef690939025f0bf4bbf56ef11a72ca69f669a1aa31fd0dd83 | 8130eb54a255c77ccb95bf467a0eb70755e1ede11902672e6b0dbf951d1a627d |
| `sensor_calibration` | 5/5 (timeouts 0, failures 0) | verified | 0.0055 | 89c3de53add27fe7b865a16fe379bb264c4681ae4d2e5d8802c5b83489f5e0d6 | e96cf1e366228f34478289660b4478b345bc069ac6e6633900d9805f0340edbb |
| `sudoku_masks` | 5/5 (timeouts 0, failures 0) | verified | 0.7574 | 9c6ac0e722416ff7dc703ad01c3fe521d8277f2c00263749d077aebe86705821 | 35a81e448daf9986f2a9b7c3a873dc6216bd55c969efec50c9b1de6d866659ec |
| `tapelang_alphabet` | 5/5 (timeouts 0, failures 0) | verified | 3.2099 | e53f36775b30177debe262c6693c40ff0c3db9a552fcf2dd8cb6c5a37174f775 | a8ac3a1054c1aa7ac25f9b1e652a96a7ac86a1c1130687fc53b90e20c766d149 |
| `transaction_ledger_audit` | 5/5 (timeouts 0, failures 0) | verified | 0.0071 | e362f9aeca17298a8263c8cb72ef8fde1ebbed611e4f3d758aa73a1a372011f8 | aa5a4fe7f85ce13998797ef506647a93f16e0ee747613683268fd801d609c812 |
| `unicode_scalar_pipeline` | 5/5 (timeouts 0, failures 0) | verified | 0.0098 | f90b820814de944d7e8ed3e139dc16f8637e2fc12e31e347db8a933765bc64c5 | c9efadb7f22969600334daa4a4eed2edde38c8e86d2c81d354d6f3979c854eb9 |
| `validated_job_pipeline` | 5/5 (timeouts 0, failures 0) | verified | 0.0043 | 503d3dbbbf6767c1e0717ab659109a0adc4ef24557276f97e90ce770780157a3 | ee5a3553094de6253bd71daaafd30e1db3eb9a17d11d3411b18ca652feafc40a |
| `versioned_telemetry_pipeline` | 5/5 (timeouts 0, failures 0) | verified | 0.2057 | df321a44087aff77a18fc4bb8eaed98dec670a11d86e99e66b23485788197c67 | 824f93580f56e01b938c047701218b04041ebaaab783db5d29c0f2eafae11a86 |
| `wide_integer_records` | 5/5 (timeouts 0, failures 0) | verified | 0.0278 | e473d9e7fa9ed63d16d1229245f95f422412380ca16650fed918776d9c2509cd | f373537521cc6bfb0fb9e1a1eb36eb93a057654b526a4521878bc269261713e5 |
| `word_frequency` | 5/5 (timeouts 0, failures 0) | verified | 0.0059 | ab257b9efad063b14c0eb0ad130b845cc2721a91ba21e14bf781ed7fef2c2c53 | 7dc1dae393e2c070eb0b9c9e611e154b2e6cce1b4a4268aa1bc73f8ff0e2fd07 |
