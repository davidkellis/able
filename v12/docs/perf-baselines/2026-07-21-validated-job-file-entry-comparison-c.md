# External Benchmark Comparison

- Generated: `2026-07-22T04:48:34.358369Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-21-validated-job-file-entry-interpreter-reference-c.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-21-validated-job-file-entry-go-reference-c.json`
- Suite: `custom`
- Able benchmarks: `validated_job_pipeline`
- Able modes: `compiled, bytecode`
- Reference languages: `go, ruby, python`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go | ruby Real (s) | Able/ruby | python Real (s) | Able/python |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `validated_job_pipeline` | `compiled` | ok (5) | verified (5) | ee5a3553094de6253bd71daaafd30e1db3eb9a17d11d3411b18ca652feafc40a | 1.2380 | 0.0049 | 252.65x | n/a | n/a | n/a | n/a |
| `validated_job_pipeline` | `bytecode` | ok (5) | verified (5) | ee5a3553094de6253bd71daaafd30e1db3eb9a17d11d3411b18ca652feafc40a | 0.5120 | n/a | n/a | 0.0541 | 9.46x | 0.0266 | 19.25x |
