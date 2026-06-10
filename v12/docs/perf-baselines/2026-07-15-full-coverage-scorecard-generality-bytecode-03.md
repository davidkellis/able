# External Benchmark Comparison

- Generated: `2026-07-15T08:44:15.399511Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-15-full-coverage-scorecard-generality-interpreter-reference.json`
- Suite: `custom`
- Able benchmarks: `i_before_e, base64, json`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU affinity: `14`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `i_before_e` | `bytecode` | ok (3) | verified (3) | 981f0d37a277be25f359c097e10df2ef68009fea2d3e322aeed3a65d0fbaca39 | 0.5100 | 0.0733 | 6.96x | 0.1017 | 5.01x |
| `base64` | `bytecode` | ok (3) | verified (3) | 5f4c00cd811078942fc98cd5dbca3b47fac2f8d8210f07ea116c1a7c0d6ac316 | 2.6633 | 3.4421 | 0.77x | 2.1771 | 1.22x |
| `json` | `bytecode` | ok (3) | verified (3) | 7cbf9736119f7683d59646bea566591a2ef278db54286b8e2525aa39e907839e | 0.7267 | 2.3201 | 0.31x | 1.4611 | 0.50x |
