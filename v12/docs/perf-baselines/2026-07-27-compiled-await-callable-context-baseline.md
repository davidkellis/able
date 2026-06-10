# External Benchmark Comparison

- Generated: `2026-07-28T03:59:05.379171Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-27-compiled-await-callable-context-go-reference.json`
- Suite: `custom`
- Able benchmarks: `await_channel_mux, mutex_await_journal, mutex_work_queue, concurrent_packet_codecs, concurrent_audio_voices, concurrent_scene_tiles, concurrent_graph_visitors, nbody`
- Able modes: `compiled`
- Reference languages: `go`
- CPU pool: `0-3` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `await_channel_mux` | `compiled` | ok (5) | verified (5) | 0a32fa2e9481ecb635562079d993288d911326d9f36aeea0d9d3058d1a494693 | 0.2880 | 0.0049 | 58.78x |
| `mutex_await_journal` | `compiled` | ok (5) | verified (5) | e6f87830a6a1bb1bf3889a6b77eae053bfd3630055a06dd8fd500201586bf61e | 0.3480 | 0.0040 | 87.00x |
| `mutex_work_queue` | `compiled` | ok (5) | verified (5) | 57d3c0d15899da95d375a749cfe34dcc6942eb82a45427f197b9244b85ff8e58 | 0.8560 | 0.0046 | 186.09x |
| `concurrent_packet_codecs` | `compiled` | ok (5) | verified (5) | cf10b00cfd2619f5162ee99687f8e059e9333c4597169846846468fa20c230a5 | 0.0220 | 0.0039 | 5.64x |
| `concurrent_audio_voices` | `compiled` | ok (5) | verified (5) | 6ec28390bc9c749cf18e4a6fbd4bc03154345e1acf0a2f6601baab16884a6e28 | 0.0300 | 0.0050 | 6.00x |
| `concurrent_scene_tiles` | `compiled` | ok (5) | verified (5) | 2c60d0d812049ff61485905f50f92e678610e9708c24849927292dd7566a15fe | 0.0260 | 0.0047 | 5.53x |
| `concurrent_graph_visitors` | `compiled` | ok (5) | verified (5) | 399e7fd8eae623db1fb6fe83ba5d08b46a747fca2d7e51ec6536aab65b50c9ee | 0.0300 | 0.0040 | 7.50x |
| `nbody` | `compiled` | ok (5) | verified (5) | 40799ff8af9b84a416e8bf940921658787c57be38f638fb4d98c735c8d39e820 | 0.0920 | 0.0389 | 2.37x |
