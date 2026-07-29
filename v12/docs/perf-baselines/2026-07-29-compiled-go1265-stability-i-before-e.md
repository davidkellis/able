# External Benchmark Comparison

- Generated: `2026-07-29T06:16:39.870494Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-29-compiled-go1265-stability-i-before-e-go-reference.json`
- Suite: `custom`
- Able benchmarks: `i_before_e`
- Able modes: `compiled`
- Reference languages: `go`
- CPU pool: `6` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `i_before_e` | `compiled` | ok (5) | verified (5) | 981f0d37a277be25f359c097e10df2ef68009fea2d3e322aeed3a65d0fbaca39 | 0.0520 | 0.0567 | 0.92x |
