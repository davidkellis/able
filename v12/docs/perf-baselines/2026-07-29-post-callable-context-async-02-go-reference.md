# Pinned Go Reference Refresh

- Generated: `2026-07-29T12:07:30Z`
- Go: `go version go1.26.5 linux/amd64`
- Go toolchain selector: `go1.26.5`
- Runs: `5`
- Timeout: `90s`
- CPU pool: `7-10`
- GOMAXPROCS: `catalog CPU budget per row`

| Benchmark | Go status | Validation | Go real (s) | Source SHA-256 | Stdout SHA-256 |
| --- | --- | --- | ---: | --- | --- |
| `await_channel_mux` | 5/5 (timeouts 0, failures 0) | verified | 0.0055 | cbbe5f94b6c169e778c687ac010d9683ec1d49d17597a4a876b9d5ec4ab59745 | 0a32fa2e9481ecb635562079d993288d911326d9f36aeea0d9d3058d1a494693 |
| `mutex_ledger` | 5/5 (timeouts 0, failures 0) | verified | 0.0051 | 59d9c8fc20ccbf5807fa06c407ffcad1824b497d9d56fbf99d4106520aef5bef | 58f2e2fda882a7c1e5032a2f34c4d9d05f40ced51199c0be0672ecc4d4cf14e4 |
| `mutex_await_journal` | 5/5 (timeouts 0, failures 0) | verified | 0.0052 | e81b9ab6eae0badebaa7bce0797cd551702e55f064baa12313992933a525449b | e6f87830a6a1bb1bf3889a6b77eae053bfd3630055a06dd8fd500201586bf61e |
| `mutex_work_queue` | 5/5 (timeouts 0, failures 0) | verified | 0.0055 | c346c087fb68c8db955a88902c4c14c28531228465764bad82014008d5f1f736 | 57d3c0d15899da95d375a749cfe34dcc6942eb82a45427f197b9244b85ff8e58 |
