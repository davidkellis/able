# External Benchmark Comparison

- Generated: `2026-07-20T16:25:50.895685Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-20-mutex-work-queue-cohort-a-generality-interpreter-reference-selected.json`
- Suite: `custom`
- Able benchmarks: `i_before_e, base64, json`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU pool: `0-3` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `i_before_e` | `bytecode` | ok (5) | verified (5) | 981f0d37a277be25f359c097e10df2ef68009fea2d3e322aeed3a65d0fbaca39 | 0.5360 | 0.1091 | 4.91x | 0.1219 | 4.40x |
| `base64` | `bytecode` | ok (5) | verified (5) | 5f4c00cd811078942fc98cd5dbca3b47fac2f8d8210f07ea116c1a7c0d6ac316 | 2.7700 | 3.9303 | 0.70x | 2.4816 | 1.12x |
| `json` | `bytecode` | ok (5) | verified (5) | 7cbf9736119f7683d59646bea566591a2ef278db54286b8e2525aa39e907839e | 0.8820 | 2.6015 | 0.34x | 1.6757 | 0.53x |
