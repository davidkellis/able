# Pinned Go Reference Refresh

- Generated: `2026-07-20T19:07:38Z`
- Go: `go version go1.26.4 linux/amd64`
- Runs: `5`
- Timeout: `55s`
- CPU pool: `0-15`
- GOMAXPROCS: `catalog CPU budget per row`

| Benchmark | Go status | Validation | Go real (s) | Source SHA-256 | Stdout SHA-256 |
| --- | --- | --- | ---: | --- | --- |
| `channel_rollup` | 5/5 (timeouts 0, failures 0) | verified | 0.0062 | ce97e496db8d6e776af7e6714d2ff51c0c55c3e7fcf098fa0db48c616f060ef7 | a6a1f91069e8c95a38fba1a3cb7fb3f582434245605091f200ee90cdb190e604 |
| `future_pipeline` | 5/5 (timeouts 0, failures 0) | verified | 0.0059 | 6b8d6b83a6086d48e1c62ec05f192bb2bc7a488f8e3ac6143cbb4a6b5c3df73c | 3db937fdd21b5719ab57c25679a1944e8f30128695c8edbfae4940aba7aa6d98 |
| `future_await_race` | 5/5 (timeouts 0, failures 0) | verified | 0.0038 | 04d296c4183b7e00c93369218788cac51325c8ba3a3a2cb97b64c81a5c2ac94b | 33297798eeb96d6d471a6afe97ec8a8bf09eda0ce30e13917aed1db40fb930e4 |
| `await_channel_mux` | 5/5 (timeouts 0, failures 0) | verified | 0.0048 | cbbe5f94b6c169e778c687ac010d9683ec1d49d17597a4a876b9d5ec4ab59745 | 0a32fa2e9481ecb635562079d993288d911326d9f36aeea0d9d3058d1a494693 |
| `mutex_ledger` | 5/5 (timeouts 0, failures 0) | verified | 0.0046 | 59d9c8fc20ccbf5807fa06c407ffcad1824b497d9d56fbf99d4106520aef5bef | 58f2e2fda882a7c1e5032a2f34c4d9d05f40ced51199c0be0672ecc4d4cf14e4 |
| `mutex_await_journal` | 5/5 (timeouts 0, failures 0) | verified | 0.0044 | e81b9ab6eae0badebaa7bce0797cd551702e55f064baa12313992933a525449b | e6f87830a6a1bb1bf3889a6b77eae053bfd3630055a06dd8fd500201586bf61e |
| `mutex_work_queue` | 5/5 (timeouts 0, failures 0) | verified | 0.0048 | c346c087fb68c8db955a88902c4c14c28531228465764bad82014008d5f1f736 | 57d3c0d15899da95d375a749cfe34dcc6942eb82a45427f197b9244b85ff8e58 |
