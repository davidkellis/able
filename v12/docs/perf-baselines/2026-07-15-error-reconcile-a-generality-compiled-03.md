# External Benchmark Comparison

- Generated: `2026-07-16T01:37:08.136196Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-15-error-reconcile-a-generality-go-reference.json`
- Suite: `custom`
- Able benchmarks: `i_before_e, base64, json`
- Able modes: `compiled`
- Reference languages: `go`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `i_before_e` | `compiled` | ok (5) | verified (5) | 981f0d37a277be25f359c097e10df2ef68009fea2d3e322aeed3a65d0fbaca39 | 0.1000 | 0.0561 | 1.78x |
| `base64` | `compiled` | ok (5) | verified (5) | 5f4c00cd811078942fc98cd5dbca3b47fac2f8d8210f07ea116c1a7c0d6ac316 | 2.2160 | 2.2209 | 1.00x |
| `json` | `compiled` | ok (5) | verified (5) | 7cbf9736119f7683d59646bea566591a2ef278db54286b8e2525aa39e907839e | 0.6700 | 1.2796 | 0.52x |
