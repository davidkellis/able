# External Benchmark Comparison

- Generated: `2026-07-22T03:26:40.965451Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-21-concurrent-document-pipeline-go-reference.json`
- Suite: `custom`
- Able benchmarks: `concurrent_document_pipeline`
- Able modes: `compiled`
- Reference languages: `go`
- CPU pool: `0-3` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `concurrent_document_pipeline` | `compiled` | ok (5) | verified (5) | 60b369f137cf022522072c4abfd911091aa3c77597528906f58b62610f438120 | 0.2520 | 0.0037 | 68.11x |
