# Pinned Python/Ruby Reference Refresh

- Generated: `2026-07-22T00:37:53Z`
- Suite: `custom`
- Python: `Python 3.14.5` from `python-3.14`
- Ruby: `ruby 4.0.5 (2026-05-20 revision 64336ffd0e) +PRISM [x86_64-linux]` from `ruby-4.0`
- Runs: `5`
- Timeout: `55s`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Language | Status | Validation | Real (s) | Source SHA-256 | Stdout SHA-256 |
| --- | --- | --- | --- | ---: | --- | --- |
| `config_validation_extraction` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0234 | bce3a0f940a4ec17897935df39c0fb9aee8d1c779dc31575211fcb63b0bb0755 | c1aa99b9a13bb6e0c7731cb2aea77e300cd3cecc695df7fd4af90036939341d1 |
| `config_validation_extraction` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0479 | 19c25888ed8c17bd510cb68ef64786f0fa438a65ed7b792f910b58bc929fb571 | c1aa99b9a13bb6e0c7731cb2aea77e300cd3cecc695df7fd4af90036939341d1 |
| `log_routing_redaction` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0176 | e780f4640bf1b52f7a793b1bf308578cb462ac8e13a56609894102b17e0ff088 | 0d9585b01f83904fdf11d47b2902678c1718c8442ed1d84410d61d5d90f60bf4 |
| `log_routing_redaction` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0449 | 19ce20224801dfde924d6a90c5f6cb7f8d48dea71e6fa5c1fb78aead5bf94d62 | 0d9585b01f83904fdf11d47b2902678c1718c8442ed1d84410d61d5d90f60bf4 |
| `policy_record_dispatch` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0250 | 9b1deb13f7883db77316cff9fa4fdd3bfba53e730adab5a3098966f68cca98e1 | f15c66e53e4c89650ae12aff6c2f466f2cf0c7adb6ca3fe5be0127bfba217c05 |
| `policy_record_dispatch` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0459 | af815023090cd27bf2338575d47bd39a6309d0aa5df14b6ef2b2c8cb8d5a3f0e | f15c66e53e4c89650ae12aff6c2f466f2cf0c7adb6ca3fe5be0127bfba217c05 |
| `regex_set_audit` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0199 | 3dfa49b7d1b958dc2ac7e278cbdf207e70ee9cb524c2d19f483d31baa1331b73 | 3d8f861a312416f95b95d59be62614f0ffc7918e86fb984fe6035ca7b7b28da2 |
| `regex_set_audit` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0460 | 914b1064f4f14c3c55c6bf50c3ca9a73b4f627fd3854c84c4feb68db7c4be1b1 | 3d8f861a312416f95b95d59be62614f0ffc7918e86fb984fe6035ca7b7b28da2 |
| `regex_stream_audit` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0188 | 711b2bfd7229437c0e6da454f303a890cd9a2bfdba0ed67cd0f83c84c0c94414 | dd7801f0b104d8bd47aaef64d685bc65a06263851f12c8df8cf75260d09e717b |
| `regex_stream_audit` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0452 | 132ebe36d07b64b7f8542a0c1150c70735b47be31c44a924a66c967c42cec8e5 | dd7801f0b104d8bd47aaef64d685bc65a06263851f12c8df8cf75260d09e717b |
| `regex_suffix_audit` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0181 | 2f2e6c0e6a25f93c70f68b687e0367b408c83ab06ec82c50f3f4c37c7d7b603e | b5d5ccfabbfd4dc5952406cb1c42d62b807f75828661c4c3774b251abe38380f |
| `regex_suffix_audit` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0389 | 86cb62160bc16cb5f08d80ac37dce5493ae7f15bbe4977f32ecb77269a728595 | b5d5ccfabbfd4dc5952406cb1c42d62b807f75828661c4c3774b251abe38380f |
