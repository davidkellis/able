# Pinned Python/Ruby Reference Refresh

- Generated: `2026-07-15T08:17:59Z`
- Suite: `custom`
- Python: `Python 3.14.5` from `python-3.14`
- Ruby: `ruby 4.0.5 (2026-05-20 revision 64336ffd0e) +PRISM [x86_64-linux]` from `ruby-4.0`
- Runs: `3`
- Timeout: `45s`
- CPU affinity: `14`

| Benchmark | Language | Status | Validation | Real (s) | Stdout SHA-256 |
| --- | --- | --- | --- | ---: | --- |
| `channel_rollup` | python | 3/3 (attempted 3, timeouts 0, failures 0) | verified | 0.0365 | a6a1f91069e8c95a38fba1a3cb7fb3f582434245605091f200ee90cdb190e604 |
| `channel_rollup` | ruby | 3/3 (attempted 3, timeouts 0, failures 0) | verified | 0.0471 | a6a1f91069e8c95a38fba1a3cb7fb3f582434245605091f200ee90cdb190e604 |
| `future_pipeline` | python | 3/3 (attempted 3, timeouts 0, failures 0) | verified | 0.0542 | 3db937fdd21b5719ab57c25679a1944e8f30128695c8edbfae4940aba7aa6d98 |
| `future_pipeline` | ruby | 3/3 (attempted 3, timeouts 0, failures 0) | verified | 0.0650 | 3db937fdd21b5719ab57c25679a1944e8f30128695c8edbfae4940aba7aa6d98 |
| `future_await_race` | python | 3/3 (attempted 3, timeouts 0, failures 0) | verified | 0.0280 | 33297798eeb96d6d471a6afe97ec8a8bf09eda0ce30e13917aed1db40fb930e4 |
| `future_await_race` | ruby | 3/3 (attempted 3, timeouts 0, failures 0) | verified | 0.0489 | 33297798eeb96d6d471a6afe97ec8a8bf09eda0ce30e13917aed1db40fb930e4 |
| `await_channel_mux` | python | 3/3 (attempted 3, timeouts 0, failures 0) | verified | 0.1084 | 0a32fa2e9481ecb635562079d993288d911326d9f36aeea0d9d3058d1a494693 |
| `await_channel_mux` | ruby | 3/3 (attempted 3, timeouts 0, failures 0) | verified | 0.0847 | 0a32fa2e9481ecb635562079d993288d911326d9f36aeea0d9d3058d1a494693 |
| `mutex_ledger` | python | 3/3 (attempted 3, timeouts 0, failures 0) | verified | 0.0361 | 58f2e2fda882a7c1e5032a2f34c4d9d05f40ced51199c0be0672ecc4d4cf14e4 |
| `mutex_ledger` | ruby | 3/3 (attempted 3, timeouts 0, failures 0) | verified | 0.0481 | 58f2e2fda882a7c1e5032a2f34c4d9d05f40ced51199c0be0672ecc4d4cf14e4 |
| `mutex_await_journal` | python | 3/3 (attempted 3, timeouts 0, failures 0) | verified | 0.0186 | e6f87830a6a1bb1bf3889a6b77eae053bfd3630055a06dd8fd500201586bf61e |
| `mutex_await_journal` | ruby | 3/3 (attempted 3, timeouts 0, failures 0) | verified | 0.0412 | e6f87830a6a1bb1bf3889a6b77eae053bfd3630055a06dd8fd500201586bf61e |
