# Pinned Python/Ruby Reference Refresh

- Generated: `2026-07-16T05:23:16Z`
- Suite: `custom`
- Python: `Python 3.14.5` from `python-3.14`
- Ruby: `ruby 4.0.5 (2026-05-20 revision 64336ffd0e) +PRISM [x86_64-linux]` from `ruby-4.0`
- Runs: `5`
- Timeout: `90s`
- CPU affinity: `none`

| Benchmark | Language | Status | Validation | Real (s) | Source SHA-256 | Stdout SHA-256 |
| --- | --- | --- | --- | ---: | --- | --- |
| `fixed_width_128` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.3428 | 9b1e7b1cb5f737c0726e0ef0490152098786925ccb090a40ca1caea7543279f7 | eceabf5869b1abca8d6dd228b64a09f89e4e98ba8cabc4833ffee1218dafa56a |
| `fixed_width_128` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.6456 | 7655ea9f640810e2cdb796e9323389ec22367e63bb50c2715577aa25a0f51e1f | eceabf5869b1abca8d6dd228b64a09f89e4e98ba8cabc4833ffee1218dafa56a |
| `rational_series` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0987 | 88feefdf0cb2d5d729a992e91e2b31da5c5528ed4e27c51ffb498446437f4df8 | 127f0f44ee4870b57a188a7948f80a0a5d14584a326c345a48d4285594069f0c |
| `rational_series` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.1317 | 7cddf887e05d9c2795420f2bd2ba3150fe5ca544b3a1d208882192d53aa0f135 | 127f0f44ee4870b57a188a7948f80a0a5d14584a326c345a48d4285594069f0c |
| `word_frequency` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0187 | 07f6d7bf1629dfa4d3c06c74c5e5b7593bc01cded12d30659da36f4fe641a18c | 7dc1dae393e2c070eb0b9c9e611e154b2e6cce1b4a4268aa1bc73f8ff0e2fd07 |
| `word_frequency` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0512 | 6c31037565a1d96128a4874a8dd99ef1c91ab3905682ebb99a37caaf585d6758 | 7dc1dae393e2c070eb0b9c9e611e154b2e6cce1b4a4268aa1bc73f8ff0e2fd07 |
| `document_audit` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0131 | e05af7da8840ba98d35af010b9c0e4448fea394baef41ef9ec4614536a626019 | 0dad030a80c8a883cbb56fbcfebfd530d521075e15d5d91ba538bc93e66c0aab |
| `document_audit` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0436 | 4c3156ec19dc17d4b1054a321032cb52aa78a904bd9cf63762b0ba8f3ebda056 | 0dad030a80c8a883cbb56fbcfebfd530d521075e15d5d91ba538bc93e66c0aab |
| `lexical_rollup` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0163 | 0565427240bf21dcb3be8d61d8a6175358a4cc4353dfbda9b23c1faf537fe849 | a6a1f91069e8c95a38fba1a3cb7fb3f582434245605091f200ee90cdb190e604 |
| `lexical_rollup` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0505 | d890311899c5ed62c0aa148024c658fd2a3495053f3ac48caf31fa8405ceab87 | a6a1f91069e8c95a38fba1a3cb7fb3f582434245605091f200ee90cdb190e604 |
| `regex_suffix_audit` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0396 | c4e8c524a52175b22d63a016e76826a346308cb3cf42bece55a399389e42b6a5 | 48835ea1a1741c659d1b6b215a56e6611e525366596e08e9a10ec985106f598a |
| `regex_suffix_audit` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0748 | 06b7ff3282e0f1da6e5f1a5deca02addec25d060d30af1c85a3163b4e0579217 | 48835ea1a1741c659d1b6b215a56e6611e525366596e08e9a10ec985106f598a |
| `regex_set_audit` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0181 | 3dfa49b7d1b958dc2ac7e278cbdf207e70ee9cb524c2d19f483d31baa1331b73 | 3d8f861a312416f95b95d59be62614f0ffc7918e86fb984fe6035ca7b7b28da2 |
| `regex_set_audit` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0416 | 914b1064f4f14c3c55c6bf50c3ca9a73b4f627fd3854c84c4feb68db7c4be1b1 | 3d8f861a312416f95b95d59be62614f0ffc7918e86fb984fe6035ca7b7b28da2 |
| `regex_stream_audit` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0181 | 711b2bfd7229437c0e6da454f303a890cd9a2bfdba0ed67cd0f83c84c0c94414 | dd7801f0b104d8bd47aaef64d685bc65a06263851f12c8df8cf75260d09e717b |
| `regex_stream_audit` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0427 | 132ebe36d07b64b7f8542a0c1150c70735b47be31c44a924a66c967c42cec8e5 | dd7801f0b104d8bd47aaef64d685bc65a06263851f12c8df8cf75260d09e717b |
| `array_slice_window` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0270 | 81a554ebbdae62eef642e766d86f3d49d33666842435eb4e9a6765d979af55e9 | 155f89122475c7b282637dbf2ecba6d19771d396e801b581cb1d1b0cef64103e |
| `array_slice_window` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0606 | 5fa9c8c230542bb1bc30cbfc44302181df6880c6e3bb0646cb90178d4b06d017 | 155f89122475c7b282637dbf2ecba6d19771d396e801b581cb1d1b0cef64103e |
| `dependency_plan` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0156 | 6ffa5c69fac638090ad0af67f3a06f5b70e1afe74a8c5eff8e86a0c82504b3f0 | 96dc74508d9b7a476bafdef453b11e11f2f70279c58ccaa5dcb6d85c529c4a38 |
| `dependency_plan` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0473 | 0a0d8b9b09b0dd9a603dcac911ee588c380f3e3acd0aefd1ce9f389f6d07a3d1 | 96dc74508d9b7a476bafdef453b11e11f2f70279c58ccaa5dcb6d85c529c4a38 |
| `option_result_config` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0164 | 50349cea5bc033ab63f3f7846d680068b5ad7e1158d0e985756b4f2110e02bc2 | 28e46b27a6dceeaa15968e9db7a6728f4a2b35f87a89ff7bf561db18cad53112 |
| `option_result_config` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0423 | bfe5349856b45010b6e4c0126bac468adcb242952fefe4ff5cde2ff4216bfd0f | 28e46b27a6dceeaa15968e9db7a6728f4a2b35f87a89ff7bf561db18cad53112 |
