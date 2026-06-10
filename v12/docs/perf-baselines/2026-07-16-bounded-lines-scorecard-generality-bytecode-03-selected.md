# External Benchmark Comparison

- Generated: `2026-07-16T09:13:11.369579Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-16-bounded-lines-scorecard-generality-interpreter-reference-selected.json`
- Suite: `custom`
- Able benchmarks: `i_before_e, base64, json`
- Able modes: `bytecode`
- Reference languages: `python, ruby`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `i_before_e` | `bytecode` | ok (5) | verified (5) | 981f0d37a277be25f359c097e10df2ef68009fea2d3e322aeed3a65d0fbaca39 | 0.5400 | 0.0760 | 7.11x | 0.1070 | 5.05x |
| `base64` | `bytecode` | ok (5) | verified (5) | 5f4c00cd811078942fc98cd5dbca3b47fac2f8d8210f07ea116c1a7c0d6ac316 | 2.8260 | 3.5852 | 0.79x | 2.3263 | 1.21x |
| `json` | `bytecode` | ok (5) | verified (5) | 7cbf9736119f7683d59646bea566591a2ef278db54286b8e2525aa39e907839e | 0.8700 | 2.4928 | 0.35x | 1.5667 | 0.56x |
