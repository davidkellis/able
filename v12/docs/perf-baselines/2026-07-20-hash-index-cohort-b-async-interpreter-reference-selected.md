# Pinned Python/Ruby Reference Refresh

- Generated: `2026-07-20T14:05:01Z`
- Suite: `custom`
- Python: `Python 3.14.5` from `python-3.14`
- Ruby: `ruby 4.0.5 (2026-05-20 revision 64336ffd0e) +PRISM [x86_64-linux]` from `ruby-4.0`
- Runs: `5`
- Timeout: `55s`
- CPU pool: `0-3` (each row records its resolved catalog budget)

| Benchmark | Language | Status | Validation | Real (s) | Source SHA-256 | Stdout SHA-256 |
| --- | --- | --- | --- | ---: | --- | --- |
| `channel_rollup` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0457 | b05cb8c5e36f4bd30dbbcf5c51b338c94fecb4cebc521e229e4f420e241e3eae | a6a1f91069e8c95a38fba1a3cb7fb3f582434245605091f200ee90cdb190e604 |
| `channel_rollup` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0720 | 3da7a0b3933ce3b12b0311fdfeb86e45b88d68919d1b24426387d41286136e6d | a6a1f91069e8c95a38fba1a3cb7fb3f582434245605091f200ee90cdb190e604 |
| `future_pipeline` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0619 | 4332cc5163ce314ec7b513c755ffae3619e9b0a9d7f55de37e8bf41862e4726f | 3db937fdd21b5719ab57c25679a1944e8f30128695c8edbfae4940aba7aa6d98 |
| `future_pipeline` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0797 | 1c5a410ce7ed870395bc9d616a949e40d36902f12efac99e156e16315a8dc36c | 3db937fdd21b5719ab57c25679a1944e8f30128695c8edbfae4940aba7aa6d98 |
| `future_await_race` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0305 | be85b10446a504991b9c911c6302a2043fc5a7d14a8d5310ec6d720a5823a3d4 | 33297798eeb96d6d471a6afe97ec8a8bf09eda0ce30e13917aed1db40fb930e4 |
| `future_await_race` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0606 | 5884a3d44aaa2776c494285eb26d9b28c0eb0973673ce1a3c30e3031090f13ec | 33297798eeb96d6d471a6afe97ec8a8bf09eda0ce30e13917aed1db40fb930e4 |
| `await_channel_mux` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.1285 | 7d8f70c24001b98261028d86c6438c0eb54dfc2dfaa6fe83f7e587ba9a4017ad | 0a32fa2e9481ecb635562079d993288d911326d9f36aeea0d9d3058d1a494693 |
| `await_channel_mux` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.1003 | 3ffb141e0e2ea79cad28c12eb8de625fc852f5fc6e4359b1f79084bd34e7e2ea | 0a32fa2e9481ecb635562079d993288d911326d9f36aeea0d9d3058d1a494693 |
| `mutex_ledger` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0304 | 72ba4e3ad7ce73b0972491d9a5097f632805e0fa360c055abb139f8fe74fb67c | 58f2e2fda882a7c1e5032a2f34c4d9d05f40ced51199c0be0672ecc4d4cf14e4 |
| `mutex_ledger` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0581 | 51f827fa60164f0292c8d30fc9356fcacb2220283902f721b7b3d011a5066391 | 58f2e2fda882a7c1e5032a2f34c4d9d05f40ced51199c0be0672ecc4d4cf14e4 |
| `mutex_await_journal` | python | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0241 | 9c164216c74ab106e96bff4d8416232690e3e9b2ee2bed34deeac2691b800fcc | e6f87830a6a1bb1bf3889a6b77eae053bfd3630055a06dd8fd500201586bf61e |
| `mutex_await_journal` | ruby | 5/5 (attempted 5, timeouts 0, failures 0) | verified | 0.0448 | 5bd8014c605c78ae706b6d936d84321bd62176a07372f027e736e6a7512051ef | e6f87830a6a1bb1bf3889a6b77eae053bfd3630055a06dd8fd500201586bf61e |
