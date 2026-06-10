# External Benchmark Comparison

- Generated: `2026-07-22T03:23:41.112575Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-21-concurrent-document-pipeline-c2-interpreter-reference.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-21-concurrent-document-pipeline-c2-go-reference.json`
- Suite: `custom`
- Able benchmarks: `concurrent_document_pipeline`
- Able modes: `compiled, bytecode`
- Reference languages: `go, ruby, python`
- CPU pool: `0-3` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go | ruby Real (s) | Able/ruby | python Real (s) | Able/python |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `concurrent_document_pipeline` | `compiled` | ok (5) | verified (5) | 60b369f137cf022522072c4abfd911091aa3c77597528906f58b62610f438120 | 0.2520 | 0.0041 | 61.46x | n/a | n/a | n/a | n/a |
| `concurrent_document_pipeline` | `bytecode` | ok (5) | verified (5) | 60b369f137cf022522072c4abfd911091aa3c77597528906f58b62610f438120 | 0.2620 | n/a | n/a | 0.0399 | 6.57x | 0.0194 | 13.51x |
