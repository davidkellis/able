# External Benchmark Comparison

- Generated: `2026-07-22T03:26:45.168787Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-21-concurrent-document-pipeline-interpreter-reference.json`
- Suite: `custom`
- Able benchmarks: `concurrent_document_pipeline`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU pool: `0` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `concurrent_document_pipeline` | `bytecode` | ok (5) | verified (5) | 60b369f137cf022522072c4abfd911091aa3c77597528906f58b62610f438120 | 0.3480 | 0.0223 | 15.61x | 0.0419 | 8.31x |
