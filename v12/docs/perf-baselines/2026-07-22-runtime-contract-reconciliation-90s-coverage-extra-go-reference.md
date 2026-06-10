# Pinned Go Reference Refresh

- Generated: `2026-07-23T03:04:56Z`
- Go: `go version go1.26.4 linux/amd64`
- Runs: `5`
- Timeout: `90s`
- CPU pool: `0-15`
- GOMAXPROCS: `catalog CPU budget per row`

| Benchmark | Go status | Validation | Go real (s) | Source SHA-256 | Stdout SHA-256 |
| --- | --- | --- | ---: | --- | --- |
| `fixed_width_128` | 5/5 (timeouts 0, failures 0) | verified | 0.0093 | 4125d090eb3cb657b79c2fc02d3170d7aaa62588f92a1e504f0140d1dbd65094 | eceabf5869b1abca8d6dd228b64a09f89e4e98ba8cabc4833ffee1218dafa56a |
| `rational_series` | 5/5 (timeouts 0, failures 0) | verified | 0.0161 | 118406800ba92ccce86446d5a61b837f9f75e7edcc771e7b04ddd33f260bd652 | 127f0f44ee4870b57a188a7948f80a0a5d14584a326c345a48d4285594069f0c |
| `wide_integer_records` | 5/5 (timeouts 0, failures 0) | verified | 0.0309 | e473d9e7fa9ed63d16d1229245f95f422412380ca16650fed918776d9c2509cd | f373537521cc6bfb0fb9e1a1eb36eb93a057654b526a4521878bc269261713e5 |
| `binary_event_log` | 5/5 (timeouts 0, failures 0) | verified | 0.0108 | f9d5afe5d21a6bd2f73cb9c8ca9dd003fcd0b1993f695846ea82343f5c2bd2bf | fb075dc8606582c1e6a1d5e520fa8dda237fc7304044b84b3f8f3a2c6b1c36e9 |
| `word_frequency` | 5/5 (timeouts 0, failures 0) | verified | 0.0073 | ab257b9efad063b14c0eb0ad130b845cc2721a91ba21e14bf781ed7fef2c2c53 | 7dc1dae393e2c070eb0b9c9e611e154b2e6cce1b4a4268aa1bc73f8ff0e2fd07 |
| `document_audit` | 5/5 (timeouts 0, failures 0) | verified | 0.0054 | efca73563a21cd01ff3d844acc1f666176d810b92de4ad61db362a93781cf39c | 0dad030a80c8a883cbb56fbcfebfd530d521075e15d5d91ba538bc93e66c0aab |
| `lexical_rollup` | 5/5 (timeouts 0, failures 0) | verified | 0.0052 | e1ce0907e22b60082acbc836da6b66b9629895ce44fefd441e451d4744936c99 | a6a1f91069e8c95a38fba1a3cb7fb3f582434245605091f200ee90cdb190e604 |
| `regex_suffix_audit` | 5/5 (timeouts 0, failures 0) | verified | 0.0056 | 4424b20eadfabba04dc90132f1606bdcd33be424fbe37b99041a2d10b356ca98 | b5d5ccfabbfd4dc5952406cb1c42d62b807f75828661c4c3774b251abe38380f |
| `regex_set_audit` | 5/5 (timeouts 0, failures 0) | verified | 0.0064 | ce6d01f4d9331f3ee52ff73bcca39637e226c14f3c045942f5a4765e0c3995ed | 3d8f861a312416f95b95d59be62614f0ffc7918e86fb984fe6035ca7b7b28da2 |
| `regex_stream_audit` | 5/5 (timeouts 0, failures 0) | verified | 0.0057 | 1f45fe2a76d146be18659ac9dd6eeab88548c5afe22e2a3ccc5ac3c8c0d82be9 | dd7801f0b104d8bd47aaef64d685bc65a06263851f12c8df8cf75260d09e717b |
| `log_routing_redaction` | 5/5 (timeouts 0, failures 0) | verified | 0.0072 | 4b9ee58bfba7e36b5c2bd530cc0b3977eb5a8ce22e14c1d5efaa656e224ce855 | 0d9585b01f83904fdf11d47b2902678c1718c8442ed1d84410d61d5d90f60bf4 |
| `config_validation_extraction` | 5/5 (timeouts 0, failures 0) | verified | 0.0063 | dd0a7582e9ac0b83d571d57000fa30cf2c09fbed5f9bd0988283755aeef74f3e | c1aa99b9a13bb6e0c7731cb2aea77e300cd3cecc695df7fd4af90036939341d1 |
| `array_slice_window` | 5/5 (timeouts 0, failures 0) | verified | 0.0063 | b05ef51494077012bf4dd3822d8b36edaf10331d7de66ea6532a02f86e5d0402 | 155f89122475c7b282637dbf2ecba6d19771d396e801b581cb1d1b0cef64103e |
| `dependency_plan` | 5/5 (timeouts 0, failures 0) | verified | 0.0048 | c0f1e22de8cdbb92823d70897d9fd87f52df8a2b814789d8831ffd655618c8d3 | 96dc74508d9b7a476bafdef453b11e11f2f70279c58ccaa5dcb6d85c529c4a38 |
| `inventory_reconciliation` | 5/5 (timeouts 0, failures 0) | verified | 0.0119 | 958b1a85bff5e471739a502ff0af5eaf886fade65c6601aeee74b69d8ae76b44 | 37dc8fb8d62ba404bd385852de356d7ccbd468da2b3d20b4f0a535263d6ffcf3 |
| `option_result_config` | 5/5 (timeouts 0, failures 0) | verified | 0.0049 | 3ced806f4f42b53e457f102494807d04b0bab5a93f76c9a4a1338efedcd2d68a | 28e46b27a6dceeaa15968e9db7a6728f4a2b35f87a89ff7bf561db18cad53112 |
| `unicode_scalar_pipeline` | 5/5 (timeouts 0, failures 0) | verified | 0.0151 | f90b820814de944d7e8ed3e139dc16f8637e2fc12e31e347db8a933765bc64c5 | c9efadb7f22969600334daa4a4eed2edde38c8e86d2c81d354d6f3979c854eb9 |
| `concurrent_text_index` | 5/5 (timeouts 0, failures 0) | verified | 0.0084 | 8600a1c50b2dd113912d501c320b86dbffcb83cbeda9b73d96b089831d07db59 | 3bdc437d713ad4ddcd2c37e7367ca95879766c927643a80803b3d0eba3c249d8 |
| `validated_job_pipeline` | 5/5 (timeouts 0, failures 0) | verified | 0.0056 | 503d3dbbbf6767c1e0717ab659109a0adc4ef24557276f97e90ce770780157a3 | ee5a3553094de6253bd71daaafd30e1db3eb9a17d11d3411b18ca652feafc40a |
| `dependency_wave_validation` | 5/5 (timeouts 0, failures 0) | verified | 0.0054 | a1f24cf7900327dfb259cda9086ae131d6528e6eda53bdae1c925009362ea2ef | de00786abc00ecdc5be17fb12025be79aa7405a31b73c6c08ab7dbad9b555dc6 |
| `concurrent_event_routing` | 5/5 (timeouts 0, failures 0) | verified | 0.0068 | b6707e2193637e4a257a84f8ef9cf9208671e0d1bf9ed71b978ce5ef71659e1c | 672e4542ce0d231428dd910348225e7c8b058824d7b8ceffbfb616d77a939d39 |
| `concurrent_document_pipeline` | 5/5 (timeouts 0, failures 0) | verified | 0.0051 | bfdfcfe86bf39725d48357a348aded4a167aa0853a3a229998a6aebba18b6976 | 60b369f137cf022522072c4abfd911091aa3c77597528906f58b62610f438120 |
| `manifest_normalization` | 5/5 (timeouts 0, failures 0) | verified | 0.0060 | 6bbbec6c228c0d5c525bac71bee099af423e6af69f02e5e4d0e9d394352919b6 | 2d6d55d5a76f3e45c6eb4fc3c0b892c2c5d8e02f3e38fa916d4f1c9a1579e9cb |
| `policy_record_dispatch` | 5/5 (timeouts 0, failures 0) | verified | 0.0059 | a31c758f4f52941e200b527621253bf7559e55b3da8561671aab5dab0c813ed5 | f15c66e53e4c89650ae12aff6c2f466f2cf0c7adb6ca3fe5be0127bfba217c05 |
| `sensor_calibration` | 5/5 (timeouts 0, failures 0) | verified | 0.0086 | 89c3de53add27fe7b865a16fe379bb264c4681ae4d2e5d8802c5b83489f5e0d6 | e96cf1e366228f34478289660b4478b345bc069ac6e6633900d9805f0340edbb |
