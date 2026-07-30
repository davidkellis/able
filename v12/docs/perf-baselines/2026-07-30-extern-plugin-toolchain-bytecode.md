# External Benchmark Comparison

- Generated: `2026-07-30T18:18:40.955882Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-30-extern-plugin-toolchain-interpreter-reference.json`
- Suite: `custom`
- Able benchmarks: `base64, json`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `base64` | `bytecode` | ok (5) | verified (5) | 5f4c00cd811078942fc98cd5dbca3b47fac2f8d8210f07ea116c1a7c0d6ac316 | 2.9760 | 4.0818 | 0.73x | 2.7211 | 1.09x |
| `json` | `bytecode` | ok (5) | verified (5) | 7cbf9736119f7683d59646bea566591a2ef278db54286b8e2525aa39e907839e | 0.8800 | 2.6764 | 0.33x | 1.7105 | 0.51x |
