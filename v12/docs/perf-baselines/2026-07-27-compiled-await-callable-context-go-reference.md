# Pinned Go Reference Refresh

- Generated: `2026-07-28T03:57:01Z`
- Go: `go version go1.26.4 linux/amd64`
- Runs: `5`
- Timeout: `60s`
- CPU pool: `0-3`
- GOMAXPROCS: `catalog CPU budget per row`

| Benchmark | Go status | Validation | Go real (s) | Source SHA-256 | Stdout SHA-256 |
| --- | --- | --- | ---: | --- | --- |
| `await_channel_mux` | 5/5 (timeouts 0, failures 0) | verified | 0.0049 | cbbe5f94b6c169e778c687ac010d9683ec1d49d17597a4a876b9d5ec4ab59745 | 0a32fa2e9481ecb635562079d993288d911326d9f36aeea0d9d3058d1a494693 |
| `mutex_await_journal` | 5/5 (timeouts 0, failures 0) | verified | 0.0040 | e81b9ab6eae0badebaa7bce0797cd551702e55f064baa12313992933a525449b | e6f87830a6a1bb1bf3889a6b77eae053bfd3630055a06dd8fd500201586bf61e |
| `mutex_work_queue` | 5/5 (timeouts 0, failures 0) | verified | 0.0046 | c346c087fb68c8db955a88902c4c14c28531228465764bad82014008d5f1f736 | 57d3c0d15899da95d375a749cfe34dcc6942eb82a45427f197b9244b85ff8e58 |
| `concurrent_packet_codecs` | 5/5 (timeouts 0, failures 0) | verified | 0.0039 | 2f39269e10153deb8b9777298344d4f5f6750d682b14fb01583921e4da045323 | cf10b00cfd2619f5162ee99687f8e059e9333c4597169846846468fa20c230a5 |
| `concurrent_audio_voices` | 5/5 (timeouts 0, failures 0) | verified | 0.0050 | e96bcffed4da40c14d70eff8a672acecaa03c232feb3f4a1735788a4613abd5c | 6ec28390bc9c749cf18e4a6fbd4bc03154345e1acf0a2f6601baab16884a6e28 |
| `concurrent_scene_tiles` | 5/5 (timeouts 0, failures 0) | verified | 0.0047 | ed49208100202d9390b5f7dad3d895cf475dcb444c8e60c137411f0c12828eff | 2c60d0d812049ff61485905f50f92e678610e9708c24849927292dd7566a15fe |
| `concurrent_graph_visitors` | 5/5 (timeouts 0, failures 0) | verified | 0.0040 | 43bd8daf0556f88913cb7e883d7afbaf51247523ffe25152bf239c80761630d4 | 399e7fd8eae623db1fb6fe83ba5d08b46a747fca2d7e51ec6536aab65b50c9ee |
| `nbody` | 5/5 (timeouts 0, failures 0) | verified | 0.0389 | ecbfb6dbba972da47cd6b2a4e377d053eb47b1fe1427236c625574aec82d294d | 7fee18aa4de449f07aa173bae7a37df103ff0317ed8055566e6f3c9358c09b2c |
