# External Benchmark Comparison

- Generated: `2026-07-29T23:20:12.679332Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-29-nullable-scalar-retained-generality-interpreter-reference-selected.json`
- Suite: `custom`
- Able benchmarks: `nbody, tapelang_alphabet`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU pool: `12-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `nbody` | `bytecode` | ok (5) | verified (5) | 2fbaf2f7d37b8b9be70d2ee8f8fb8c5ac6d91c030077463353667e2e3fad5208 | 9.9220 | 0.2174 | 45.64x | 0.3560 | 27.87x |
| `tapelang_alphabet` | `bytecode` | ok (5) | verified (5) | a8ac3a1054c1aa7ac25f9b1e652a96a7ac86a1c1130687fc53b90e20c766d149 | 25.2740 | 0.6164 | 41.00x | 0.7878 | 32.08x |
