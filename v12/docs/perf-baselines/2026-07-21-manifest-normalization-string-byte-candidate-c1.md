# External Benchmark Comparison

- Generated: `2026-07-22T03:57:31.507483Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-21-manifest-normalization-go-reference.json`
- Suite: `custom`
- Able benchmarks: `manifest_normalization`
- Able modes: `compiled`
- Reference languages: `go`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `manifest_normalization` | `compiled` | ok (5) | verified (5) | 2d6d55d5a76f3e45c6eb4fc3c0b892c2c5d8e02f3e38fa916d4f1c9a1579e9cb | 0.2120 | 0.0054 | 39.26x |
