# Pinned Go Reference Refresh

- Generated: `2026-07-28T05:03:46Z`
- Go: `go version go1.26.4 linux/amd64`
- Runs: `5`
- Timeout: `60s`
- CPU pool: `0-3`
- GOMAXPROCS: `catalog CPU budget per row`

| Benchmark | Go status | Validation | Go real (s) | Source SHA-256 | Stdout SHA-256 |
| --- | --- | --- | ---: | --- | --- |
| `future_await_race` | 5/5 (timeouts 0, failures 0) | verified | 0.0050 | 04d296c4183b7e00c93369218788cac51325c8ba3a3a2cb97b64c81a5c2ac94b | 33297798eeb96d6d471a6afe97ec8a8bf09eda0ce30e13917aed1db40fb930e4 |
| `await_channel_mux` | 5/5 (timeouts 0, failures 0) | verified | 0.0064 | cbbe5f94b6c169e778c687ac010d9683ec1d49d17597a4a876b9d5ec4ab59745 | 0a32fa2e9481ecb635562079d993288d911326d9f36aeea0d9d3058d1a494693 |
| `mutex_await_journal` | 5/5 (timeouts 0, failures 0) | verified | 0.0039 | e81b9ab6eae0badebaa7bce0797cd551702e55f064baa12313992933a525449b | e6f87830a6a1bb1bf3889a6b77eae053bfd3630055a06dd8fd500201586bf61e |
| `mutex_work_queue` | 5/5 (timeouts 0, failures 0) | verified | 0.0046 | c346c087fb68c8db955a88902c4c14c28531228465764bad82014008d5f1f736 | 57d3c0d15899da95d375a749cfe34dcc6942eb82a45427f197b9244b85ff8e58 |
