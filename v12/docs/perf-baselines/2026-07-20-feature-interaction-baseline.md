# External Benchmark Comparison

- Generated: `2026-07-21T04:00:09.344838Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-20-feature-interaction-interpreter-reference.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-20-feature-interaction-go-reference.json`
- Suite: `custom`
- Able benchmarks: `concurrent_text_index, validated_job_pipeline`
- Able modes: `compiled, bytecode`
- Reference languages: `go, ruby, python`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go | ruby Real (s) | Able/ruby | python Real (s) | Able/python |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `concurrent_text_index` | `compiled` | ok (5) | verified (5) | 3bdc437d713ad4ddcd2c37e7367ca95879766c927643a80803b3d0eba3c249d8 | 1.0180 | 0.0057 | 178.60x | n/a | n/a | n/a | n/a |
| `concurrent_text_index` | `bytecode` | ok (5) | verified (5) | 3bdc437d713ad4ddcd2c37e7367ca95879766c927643a80803b3d0eba3c249d8 | 0.5760 | n/a | n/a | 0.0997 | 5.78x | 0.0863 | 6.67x |
| `validated_job_pipeline` | `compiled` | ok (5) | verified (5) | 96cca38f1e5b45bea159f191a7a49507fc3cc26613c759617a30a27af27db9e2 | 3.0840 | 0.0059 | 522.71x | n/a | n/a | n/a | n/a |
| `validated_job_pipeline` | `bytecode` | ok (5) | verified (5) | 96cca38f1e5b45bea159f191a7a49507fc3cc26613c759617a30a27af27db9e2 | 0.7240 | n/a | n/a | 0.0806 | 8.98x | 0.0939 | 7.71x |
