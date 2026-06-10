# Pinned Python/Ruby Reference Refresh

- Generated: `2026-07-23T03:26:12Z`
- Suite: `custom`
- Python: `Python 3.14.5` from `python-3.14`
- Ruby: `ruby 4.0.5 (2026-05-20 revision 64336ffd0e) +PRISM [x86_64-linux]` from `ruby-4.0`
- Runs: `5`
- Timeout: `90s`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Language | Status | Validation | Real (s) | Source SHA-256 | Stdout SHA-256 |
| --- | --- | --- | --- | ---: | --- | --- |
| `fixed_width_128` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.9960 | 9b1e7b1cb5f737c0726e0ef0490152098786925ccb090a40ca1caea7543279f7 | eceabf5869b1abca8d6dd228b64a09f89e4e98ba8cabc4833ffee1218dafa56a |
| `fixed_width_128` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 1.7428 | 7655ea9f640810e2cdb796e9323389ec22367e63bb50c2715577aa25a0f51e1f | eceabf5869b1abca8d6dd228b64a09f89e4e98ba8cabc4833ffee1218dafa56a |
| `rational_series` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.2394 | 88feefdf0cb2d5d729a992e91e2b31da5c5528ed4e27c51ffb498446437f4df8 | 127f0f44ee4870b57a188a7948f80a0a5d14584a326c345a48d4285594069f0c |
| `rational_series` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.3173 | 7cddf887e05d9c2795420f2bd2ba3150fe5ca544b3a1d208882192d53aa0f135 | 127f0f44ee4870b57a188a7948f80a0a5d14584a326c345a48d4285594069f0c |
| `wide_integer_records` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.1591 | 7a26e96bc7aab8344ec9da95037d7feb31935ff297b4a02b702599a337d15b47 | f373537521cc6bfb0fb9e1a1eb36eb93a057654b526a4521878bc269261713e5 |
| `wide_integer_records` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.3386 | 5d9c0a97b32ffab5dc4106f0d9baf50ba9f38213a8e2a535ef0553e880854059 | f373537521cc6bfb0fb9e1a1eb36eb93a057654b526a4521878bc269261713e5 |
| `binary_event_log` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.3633 | d2ee02735398e5717b2112985f5ec1cd7a203330b1cf1a09a0cf6c15e304e9e2 | fb075dc8606582c1e6a1d5e520fa8dda237fc7304044b84b3f8f3a2c6b1c36e9 |
| `binary_event_log` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.4953 | 423a4b7dc89e049315e8f7dd7efd47d6554356166048e23de36026fb2dba0a59 | fb075dc8606582c1e6a1d5e520fa8dda237fc7304044b84b3f8f3a2c6b1c36e9 |
| `word_frequency` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0361 | 07f6d7bf1629dfa4d3c06c74c5e5b7593bc01cded12d30659da36f4fe641a18c | 7dc1dae393e2c070eb0b9c9e611e154b2e6cce1b4a4268aa1bc73f8ff0e2fd07 |
| `word_frequency` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.1064 | 6c31037565a1d96128a4874a8dd99ef1c91ab3905682ebb99a37caaf585d6758 | 7dc1dae393e2c070eb0b9c9e611e154b2e6cce1b4a4268aa1bc73f8ff0e2fd07 |
| `document_audit` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0140 | e05af7da8840ba98d35af010b9c0e4448fea394baef41ef9ec4614536a626019 | 0dad030a80c8a883cbb56fbcfebfd530d521075e15d5d91ba538bc93e66c0aab |
| `document_audit` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0466 | 4c3156ec19dc17d4b1054a321032cb52aa78a904bd9cf63762b0ba8f3ebda056 | 0dad030a80c8a883cbb56fbcfebfd530d521075e15d5d91ba538bc93e66c0aab |
| `lexical_rollup` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0227 | 0565427240bf21dcb3be8d61d8a6175358a4cc4353dfbda9b23c1faf537fe849 | a6a1f91069e8c95a38fba1a3cb7fb3f582434245605091f200ee90cdb190e604 |
| `lexical_rollup` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0602 | d890311899c5ed62c0aa148024c658fd2a3495053f3ac48caf31fa8405ceab87 | a6a1f91069e8c95a38fba1a3cb7fb3f582434245605091f200ee90cdb190e604 |
| `regex_suffix_audit` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0286 | 2f2e6c0e6a25f93c70f68b687e0367b408c83ab06ec82c50f3f4c37c7d7b603e | b5d5ccfabbfd4dc5952406cb1c42d62b807f75828661c4c3774b251abe38380f |
| `regex_suffix_audit` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0949 | 86cb62160bc16cb5f08d80ac37dce5493ae7f15bbe4977f32ecb77269a728595 | b5d5ccfabbfd4dc5952406cb1c42d62b807f75828661c4c3774b251abe38380f |
| `regex_set_audit` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0507 | 3dfa49b7d1b958dc2ac7e278cbdf207e70ee9cb524c2d19f483d31baa1331b73 | 3d8f861a312416f95b95d59be62614f0ffc7918e86fb984fe6035ca7b7b28da2 |
| `regex_set_audit` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.1141 | 914b1064f4f14c3c55c6bf50c3ca9a73b4f627fd3854c84c4feb68db7c4be1b1 | 3d8f861a312416f95b95d59be62614f0ffc7918e86fb984fe6035ca7b7b28da2 |
| `regex_stream_audit` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0404 | 711b2bfd7229437c0e6da454f303a890cd9a2bfdba0ed67cd0f83c84c0c94414 | dd7801f0b104d8bd47aaef64d685bc65a06263851f12c8df8cf75260d09e717b |
| `regex_stream_audit` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.1209 | 132ebe36d07b64b7f8542a0c1150c70735b47be31c44a924a66c967c42cec8e5 | dd7801f0b104d8bd47aaef64d685bc65a06263851f12c8df8cf75260d09e717b |
| `log_routing_redaction` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0437 | e780f4640bf1b52f7a793b1bf308578cb462ac8e13a56609894102b17e0ff088 | 0d9585b01f83904fdf11d47b2902678c1718c8442ed1d84410d61d5d90f60bf4 |
| `log_routing_redaction` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0942 | 19ce20224801dfde924d6a90c5f6cb7f8d48dea71e6fa5c1fb78aead5bf94d62 | 0d9585b01f83904fdf11d47b2902678c1718c8442ed1d84410d61d5d90f60bf4 |
| `config_validation_extraction` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0409 | bce3a0f940a4ec17897935df39c0fb9aee8d1c779dc31575211fcb63b0bb0755 | c1aa99b9a13bb6e0c7731cb2aea77e300cd3cecc695df7fd4af90036939341d1 |
| `config_validation_extraction` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0989 | 19c25888ed8c17bd510cb68ef64786f0fa438a65ed7b792f910b58bc929fb571 | c1aa99b9a13bb6e0c7731cb2aea77e300cd3cecc695df7fd4af90036939341d1 |
| `array_slice_window` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0642 | 81a554ebbdae62eef642e766d86f3d49d33666842435eb4e9a6765d979af55e9 | 155f89122475c7b282637dbf2ecba6d19771d396e801b581cb1d1b0cef64103e |
| `array_slice_window` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.1602 | 5fa9c8c230542bb1bc30cbfc44302181df6880c6e3bb0646cb90178d4b06d017 | 155f89122475c7b282637dbf2ecba6d19771d396e801b581cb1d1b0cef64103e |
| `dependency_plan` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0360 | 6ffa5c69fac638090ad0af67f3a06f5b70e1afe74a8c5eff8e86a0c82504b3f0 | 96dc74508d9b7a476bafdef453b11e11f2f70279c58ccaa5dcb6d85c529c4a38 |
| `dependency_plan` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.1168 | 0a0d8b9b09b0dd9a603dcac911ee588c380f3e3acd0aefd1ce9f389f6d07a3d1 | 96dc74508d9b7a476bafdef453b11e11f2f70279c58ccaa5dcb6d85c529c4a38 |
| `inventory_reconciliation` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.1492 | 8b716b3f89aab77b211784aed263716c7fd1925d24479b1664957806a17285b6 | 37dc8fb8d62ba404bd385852de356d7ccbd468da2b3d20b4f0a535263d6ffcf3 |
| `inventory_reconciliation` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.1860 | f995acffe2d3bde3f769ea4a07073c131c839e38ca4d49f713403d303d5bca9c | 37dc8fb8d62ba404bd385852de356d7ccbd468da2b3d20b4f0a535263d6ffcf3 |
| `option_result_config` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0378 | 50349cea5bc033ab63f3f7846d680068b5ad7e1158d0e985756b4f2110e02bc2 | 28e46b27a6dceeaa15968e9db7a6728f4a2b35f87a89ff7bf561db18cad53112 |
| `option_result_config` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.1144 | bfe5349856b45010b6e4c0126bac468adcb242952fefe4ff5cde2ff4216bfd0f | 28e46b27a6dceeaa15968e9db7a6728f4a2b35f87a89ff7bf561db18cad53112 |
| `unicode_scalar_pipeline` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.4833 | b35e6b7a83baaf1d46449c214cc487bb31fc9de614ad74b6a0526024c8c167ea | c9efadb7f22969600334daa4a4eed2edde38c8e86d2c81d354d6f3979c854eb9 |
| `unicode_scalar_pipeline` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.8154 | 729dc9a83929d4b3a07aa6800d59cae078a4063407111fb69c993a32a6b38c6d | c9efadb7f22969600334daa4a4eed2edde38c8e86d2c81d354d6f3979c854eb9 |
| `concurrent_text_index` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.1984 | 03f2dde874de320eb3ab9cf6059326686865079fce9118c732bc1b24bff1b5d2 | 3bdc437d713ad4ddcd2c37e7367ca95879766c927643a80803b3d0eba3c249d8 |
| `concurrent_text_index` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.2859 | 51a39d2834654a1885d12fd1e82adfb71c26c8ba5740f17f1c0639a1d021d3ae | 3bdc437d713ad4ddcd2c37e7367ca95879766c927643a80803b3d0eba3c249d8 |
| `validated_job_pipeline` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0791 | ec4d1dd04fa66c0a9177c8881f6c3576843c6c019c0e3ff081c5fc956b661140 | ee5a3553094de6253bd71daaafd30e1db3eb9a17d11d3411b18ca652feafc40a |
| `validated_job_pipeline` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.1494 | 37aa6d08f16b3091090b76595c409d15df9d638fa60608c83dbd18e2651372f2 | ee5a3553094de6253bd71daaafd30e1db3eb9a17d11d3411b18ca652feafc40a |
| `dependency_wave_validation` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0674 | 2bd1b5c121a0e5b8eae191a57ad49112aa5c6bb161373c06f5c0d7401de592da | de00786abc00ecdc5be17fb12025be79aa7405a31b73c6c08ab7dbad9b555dc6 |
| `dependency_wave_validation` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.1307 | ada281583851ca5590752081976ede4bd9a336d0d21f9699db3742b198313d83 | de00786abc00ecdc5be17fb12025be79aa7405a31b73c6c08ab7dbad9b555dc6 |
| `concurrent_event_routing` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0731 | e6d26914edd3bd99c1999b95ed463675b795970683339356e7a179d451a8d34b | 672e4542ce0d231428dd910348225e7c8b058824d7b8ceffbfb616d77a939d39 |
| `concurrent_event_routing` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.1617 | 9e1de4cd0d2f8fb60807a5d5bae77c250f190c7eb4ae9546530c838e1afde681 | 672e4542ce0d231428dd910348225e7c8b058824d7b8ceffbfb616d77a939d39 |
| `concurrent_document_pipeline` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0486 | 41e176fb2a83211ed5fd75e6ed4315712db6154270f574e23f885dc66d317c99 | 60b369f137cf022522072c4abfd911091aa3c77597528906f58b62610f438120 |
| `concurrent_document_pipeline` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.1165 | 7f31d21b363c9273656588d3eb321339a5fdfd70d786209217958a5887ab03e8 | 60b369f137cf022522072c4abfd911091aa3c77597528906f58b62610f438120 |
| `manifest_normalization` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0434 | 434dcae93591c107a540b74e7b9773d121b11e35d6574e1009498c08aa45e267 | 2d6d55d5a76f3e45c6eb4fc3c0b892c2c5d8e02f3e38fa916d4f1c9a1579e9cb |
| `manifest_normalization` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.1322 | db78e02ae03d358429da3c7d8e2376d9be83526c58666ce7b88fe89af2dae409 | 2d6d55d5a76f3e45c6eb4fc3c0b892c2c5d8e02f3e38fa916d4f1c9a1579e9cb |
| `policy_record_dispatch` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0515 | 9b1deb13f7883db77316cff9fa4fdd3bfba53e730adab5a3098966f68cca98e1 | f15c66e53e4c89650ae12aff6c2f466f2cf0c7adb6ca3fe5be0127bfba217c05 |
| `policy_record_dispatch` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.1346 | af815023090cd27bf2338575d47bd39a6309d0aa5df14b6ef2b2c8cb8d5a3f0e | f15c66e53e4c89650ae12aff6c2f466f2cf0c7adb6ca3fe5be0127bfba217c05 |
| `sensor_calibration` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0787 | 03ddbcc562997e1901ff546c9a8e5212b59de2b51a8bd69d7295ffbc618d29bf | e96cf1e366228f34478289660b4478b345bc069ac6e6633900d9805f0340edbb |
| `sensor_calibration` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.1760 | 18fc9a0df5f593ed69981b767c2d9061df49bb803aa61fc39c21a0007ecc6db8 | e96cf1e366228f34478289660b4478b345bc069ac6e6633900d9805f0340edbb |
