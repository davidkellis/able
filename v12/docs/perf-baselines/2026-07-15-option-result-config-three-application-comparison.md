# External Benchmark Comparison

- Generated: `2026-07-15T17:04:54.471620Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-15-option-result-config-interpreter-reference.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-15-option-result-config-go-reference.json`
- Suite: `custom`
- Able benchmarks: `option_result_config, dependency_plan, document_audit`
- Able modes: `compiled, bytecode`
- Reference languages: `go, ruby, python`
- CPU affinity: `11`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go | ruby Real (s) | Able/ruby | python Real (s) | Able/python |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `option_result_config` | `compiled` | ok (3) | verified (3) | 28e46b27a6dceeaa15968e9db7a6728f4a2b35f87a89ff7bf561db18cad53112 | 0.2200 | 0.0032 | 68.75x | 0.0384 | 5.73x | 0.0154 | 14.29x |
| `option_result_config` | `bytecode` | ok (3) | verified (3) | 28e46b27a6dceeaa15968e9db7a6728f4a2b35f87a89ff7bf561db18cad53112 | 3.2633 | 0.0032 | 1019.78x | 0.0384 | 84.98x | 0.0154 | 211.90x |
| `dependency_plan` | `compiled` | ok (3) | verified (3) | 96dc74508d9b7a476bafdef453b11e11f2f70279c58ccaa5dcb6d85c529c4a38 | 0.1033 | n/a | n/a | n/a | n/a | n/a | n/a |
| `dependency_plan` | `bytecode` | ok (3) | verified (3) | 96dc74508d9b7a476bafdef453b11e11f2f70279c58ccaa5dcb6d85c529c4a38 | 0.4700 | n/a | n/a | n/a | n/a | n/a | n/a |
| `document_audit` | `compiled` | ok (3) | verified (3) | 0dad030a80c8a883cbb56fbcfebfd530d521075e15d5d91ba538bc93e66c0aab | 0.1033 | n/a | n/a | n/a | n/a | n/a | n/a |
| `document_audit` | `bytecode` | ok (3) | verified (3) | 0dad030a80c8a883cbb56fbcfebfd530d521075e15d5d91ba538bc93e66c0aab | 0.3300 | n/a | n/a | n/a | n/a | n/a | n/a |
