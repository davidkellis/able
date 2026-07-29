# Bytecode current profile coverage

All 59 current bytecode target misses have source-identity-checked CPU and allocation evidence. The four bytecode target guards are intentionally excluded.

Total current target excess is 221.503684 seconds. The retained closure ledger has 23 entries; no production optimization was admitted.

## Evidence groups

| Group | Rows | Target excess | CPU | Allocation | Disposition |
| --- | ---: | ---: | --- | --- | --- |
| `bytecode-portable-workload-admission` | 5 | 74.548947 s | current | current | `closed-no-shared-leaf` |
| `bytecode-text-map` | 9 | 60.738421 s | current | current | `closed-rejected-candidate` |
| `bytecode-regex` | 6 | 23.137895 s | current | current | `closed-rejected-candidate` |
| `bytecode-concurrency` | 23 | 17.716947 s | current | current | `closed-no-shared-leaf` |
| `bytecode-wide-numeric` | 3 | 17.174000 s | current | current | `closed-rejected-candidate` |
| `bytecode-float-numeric` | 4 | 15.811789 s | current | current | `closed-rejected-candidate` |
| `bytecode-iterator-control` | 6 | 7.028737 s | current | current | `closed-no-shared-leaf` |
| `bytecode-byte-output` | 3 | 5.346947 s | current | current | `closed-no-shared-leaf` |

## Source identity map

| Application | Group | Source SHA-256 | Target excess |
| --- | --- | --- | ---: |
| `k_nucleotide` | `bytecode-text-map` | `933749cb33f84a88274e010f7459d027be839a42162f0d559eca8a1920aa8a2a` | 41.316000 s |
| `sudoku_masks` | `bytecode-portable-workload-admission` | `222b321f579d7b2a84f4bc0fd379064a7ebe554bd83169782b28d04eaaab90e0` | 22.503263 s |
| `tapelang_alphabet` | `bytecode-portable-workload-admission` | `426a40e33840f3a0e9e62d5f9b9519a6840edd2733b8031df6a280e4b782fdb8` | 19.909263 s |
| `binarytrees` | `bytecode-portable-workload-admission` | `b2e5cd3b3f439960e39e04b2da675321e9bc45c1b5d811bdb9d61069bb168eeb` | 11.707474 s |
| `quicksort` | `bytecode-portable-workload-admission` | `13dc68cc43b87d80d21943a45ca24f0722614f3b151f2ed26ef4f1025b103338` | 11.481684 s |
| `nbody` | `bytecode-portable-workload-admission` | `7d0cb3f9291be2577e7726e6868deaaf2399223acfe48a666226d7ac697f8e49` | 8.947263 s |
| `fixed_width_128` | `bytecode-wide-numeric` | `330ce59771b6dc41a1c147c53d82bb65b43d9a09c018d11ee527c1431a829f0e` | 7.640316 s |
| `policy_record_dispatch` | `bytecode-regex` | `cd6089f205ff542c9c9c07c5526d47b9fc4287f9fd9c7081947d363610ec5263` | 7.560737 s |
| `binary_event_log` | `bytecode-text-map` | `658bf062ac063af7302544b1e9d2bc30d21c5c3b535e48634f40c20ba2c8d2ea` | 5.644632 s |
| `wide_integer_records` | `bytecode-wide-numeric` | `8ba1e245f5d169409d2d35ba53c3ba94a554fd6767ec8d9d4602fef064b8405d` | 5.501579 s |
| `distance_field` | `bytecode-float-numeric` | `80f96ee942dafbf81b7e09edfc3410c3af37ed3190357fcf77c5d9641a30cf88` | 5.402526 s |
| `mandelbrot` | `bytecode-float-numeric` | `663385828c4e7318bac285c4fcce27d3b435f9a3536a1a20b3cfa09af2a32315` | 5.089263 s |
| `discrete_event_simulation` | `bytecode-iterator-control` | `f4229da6b02e7e79c49b81f266a08efb87476edade1255277a55fda77f10f22f` | 4.521158 s |
| `rms_norm` | `bytecode-float-numeric` | `b289c78432f7c4077e259810c9e788341824899188821df669af345eed04dc59` | 4.128000 s |
| `regex_set_audit` | `bytecode-regex` | `124fcfb1435ab1160adff38f7d50ba442c22b4d63cce25fa06e7799a26458b9b` | 4.110421 s |
| `rational_series` | `bytecode-wide-numeric` | `20a58261d852834d425de755ed35e7c34b0bda80945caf97a94d4ba9b2e0bf46` | 4.032105 s |
| `regex_stream_audit` | `bytecode-regex` | `ed58e3975b12776a03ace5bd7e066b570e7538f303f631db9c1704d3a1eaa999` | 3.597158 s |
| `unicode_scalar_pipeline` | `bytecode-text-map` | `9ff36de9cbdd1e2b1138611fc20453e0803e3f8734769a30929eb52528ab7bdb` | 3.563684 s |
| `regex_suffix_audit` | `bytecode-regex` | `7f8a1b031ae20c9af490cd21ac9846ea472b43c62cd082b22cb6e600fea2c5fe` | 3.429895 s |
| `reverse_complement` | `bytecode-byte-output` | `4bd16d4b6c65362efc5b2515548f59686144553a91eadcc65a0cde6df537e5f9` | 3.264316 s |
| `log_routing_redaction` | `bytecode-regex` | `958ad9cdee6f91694757dbca9748e6f23bff84340e52e0e1375589726be0629b` | 3.095368 s |
| `concurrent_event_routing` | `bytecode-concurrency` | `23d12e189b92ee44f83917df8a8fa8a84f2ab77242f9785474df345fbe5e274b` | 2.921158 s |
| `sensor_calibration` | `bytecode-text-map` | `28a29926e6511185eda8621d69f09109a3f16bcff69de6f41e0d4cdc4c0d480d` | 2.594842 s |
| `concurrent_transform_chain` | `bytecode-concurrency` | `378d0a572c0a23d65a80d7681c6309ee0eb1e147cdd2e3442ce628ca04c40027` | 2.589474 s |
| `inventory_reconciliation` | `bytecode-text-map` | `16c408486e1f3fb1003c795b5ed0ee63b310cd2fecba4a85d9b4081c381e8d9f` | 2.416526 s |
| `backup_dedup` | `bytecode-text-map` | `e29c9ccd3c94346397695303b1fb83d4e705aa9ce2d5b7194b8964268822f944` | 1.775053 s |
| `concurrent_stencil_reduction` | `bytecode-concurrency` | `db33fbbe0624cda3958b64f1fb51a07e950a023ca480b3469f1b09215f22501a` | 1.730947 s |
| `fasta_generation` | `bytecode-byte-output` | `7b30ce2139b20f4b30495a44e5afc99bb4ab664c2ec9e0817512784adcb06c0e` | 1.668842 s |
| `manifest_normalization` | `bytecode-text-map` | `4738af1fb5aabcdafd7581c2b84289824d6bf0a5dae779e70f4db6746f63aec9` | 1.517895 s |
| `concurrent_signal_dispatch` | `bytecode-concurrency` | `79d9f7945ab5278b5725a0c28ee1e1bdf66a81530f95777c6478e9efa01d63d0` | 1.493579 s |
| `word_frequency` | `bytecode-text-map` | `ad6e5d39c1685931dfb7b86fb7ce2afe783a1088ffc646a5482540606cd79a3f` | 1.480526 s |
| `config_validation_extraction` | `bytecode-regex` | `ffdba7256c645683bd9e53ef5a3422ca332ad51dae98816ef902906691ac8b73` | 1.344316 s |
| `concurrent_graph_visitors` | `bytecode-concurrency` | `b445e78cd4c706fc09a3ab008df58fb4a5c90237b2caee4cd52420d04298db56` | 1.222632 s |
| `concurrent_audio_voices` | `bytecode-concurrency` | `b6a5d1e0d2901bd3341b02d1f3445601ef05670cb2442c9f706305f4ec8b7cad` | 1.221789 s |
| `monte_carlo_pi` | `bytecode-float-numeric` | `9afb1620c5f41eed0b519e225cf7a7cc32e837ccfccd2e047d56175c1805aa0c` | 1.192000 s |
| `concurrent_packet_codecs` | `bytecode-concurrency` | `2184c24ec515d57551c57440c9351b58cb4a43c1824e5f1df43e839b3de57481` | 0.756000 s |
| `option_result_config` | `bytecode-iterator-control` | `9a0cd261ba89aad1fe1cea7da9ce78bdf356a350151ba7f17277ed5a125d7c48` | 0.721158 s |
| `array_slice_window` | `bytecode-iterator-control` | `64c65a525023e5fc7000f13b9d36875716b7d8b936060fede904381b58156b4e` | 0.657368 s |
| `future_pipeline` | `bytecode-concurrency` | `823bcdb46878c9d57ab6438bed3767861974227b96318572dc823f3da69067b7` | 0.604000 s |
| `concurrent_scene_tiles` | `bytecode-concurrency` | `dd1b5c6eb972af1480eb87bca3096672e44280e3c8c473df95942c3f06ada71f` | 0.565158 s |
| `concurrent_text_index` | `bytecode-concurrency` | `5a2429f00e30e21af242850073e6e8c24d4022b8cad9bc7aee5e6e7860999eef` | 0.540526 s |
| `dependency_wave_validation` | `bytecode-concurrency` | `7d200dbe3ee867e95668ae8c74a2e45893f4d369db6bf00962363348dc632a50` | 0.475158 s |
| `dependency_plan` | `bytecode-iterator-control` | `f2579982dbe25debb2e40f2ec1f2d0187016a3801c5373c82f45258e16210d8e` | 0.475053 s |
| `i_before_e` | `bytecode-text-map` | `4cabfaeebf777072eb41a2411a7b9467386be7fcb55303ede708d8da50f057d9` | 0.429263 s |
| `base64` | `bytecode-byte-output` | `b4676ab1b4392ed4433d7a2ce57c7388907e4719494e6edce32728b071750108` | 0.413789 s |
| `channel_rollup` | `bytecode-concurrency` | `cc7ddd6dca16348087e17b89f07a30afa536750d088606744f9c22ce0704808a` | 0.403158 s |
| `concurrent_stateful_pipeline` | `bytecode-concurrency` | `99c6ecfb4eccd2c93925701f8a9845a53de88b9ac042ba6d34517ca073922bf5` | 0.396947 s |
| `mutex_ledger` | `bytecode-concurrency` | `2975ac5ae318fd528fd353ee8846f4cb86104e8525a43c2aa1c31e05dfdb1a61` | 0.373684 s |
| `lexical_rollup` | `bytecode-iterator-control` | `567a239892f4f6b00e5d61e4d262857ff14687c31e16cd1d390d9e9dfb6b3159` | 0.368526 s |
| `concurrent_tree_folds` | `bytecode-concurrency` | `521a80923670cdbf78d12110adb8e7b8fc4aa540ac1aea39133f029aa6b2e321` | 0.362105 s |
| `validated_job_pipeline` | `bytecode-concurrency` | `792dd2fbf2d09643082e523999f768745eb5c08af64f2b747e82901e3fe4bd8e` | 0.345579 s |
| `mutex_work_queue` | `bytecode-concurrency` | `8341715e210cd8342a49c32dd2d9ba01331f893b680e8a418f5ee2542a456113` | 0.344737 s |
| `concurrent_policy_callbacks` | `bytecode-concurrency` | `ae4b730a18e1f64e85b17f3eb2a53c5d7b54bf9d0c99b748e67bb18265451266` | 0.334105 s |
| `concurrent_state_machines` | `bytecode-concurrency` | `a8270e45edd0da619ec0c49ff1c09bf80a920e37ceff046e1887963059ea221e` | 0.325895 s |
| `document_audit` | `bytecode-iterator-control` | `4f3fe4b3ce2a782c67e79d1178ba0ef3165ff8917b1970a6c96ccd7e550f27f2` | 0.285474 s |
| `concurrent_document_pipeline` | `bytecode-concurrency` | `5c5d8a577a404fdc73e355a9b8997ff36acd04e8ca974e07bde0729e78a7dc16` | 0.260105 s |
| `mutex_await_journal` | `bytecode-concurrency` | `3d17db99f2a06f2d87868dcb00e129d6f0cb18d4209b329e09bac281adaf8696` | 0.208526 s |
| `await_channel_mux` | `bytecode-concurrency` | `31bd5d760fff06cdd813c5520f961fbaadb5222908608f2630c183c8598f5b0c` | 0.132316 s |
| `future_await_race` | `bytecode-concurrency` | `7950a425ac6225a577d7c234395ee4293d5c89c2ddfcb19fd87732e8d4d0335b` | 0.109368 s |

Every row is `current` for both CPU and allocation coverage. Source files are rehashed when this ledger is generated; stale scorecard identities, missing evidence, incomplete miss coverage, and non-closed frontier groups are errors.
