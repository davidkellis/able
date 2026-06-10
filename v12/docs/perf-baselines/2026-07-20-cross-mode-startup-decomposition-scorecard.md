# External Benchmark Comparison

- Generated: `2026-07-20T22:15:18.740532Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Suite: `custom`
- Able benchmarks: `document_audit, dependency_plan, future_await_race`
- Able modes: `compiled, bytecode, bytecode-runtime`
- Warmed bytecode main() calls per run: `1`
- Reference languages: `go, ruby, python`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go | ruby Real (s) | Able/ruby | python Real (s) | Able/python |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `document_audit` | `compiled` | ok (10) | verified (10) | 0dad030a80c8a883cbb56fbcfebfd530d521075e15d5d91ba538bc93e66c0aab | 0.0900 | n/a | n/a | n/a | n/a | n/a | n/a |
| `document_audit` | `bytecode` | ok (10) | verified (10) | 0dad030a80c8a883cbb56fbcfebfd530d521075e15d5d91ba538bc93e66c0aab | 0.3530 | n/a | n/a | n/a | n/a | n/a | n/a |
| `document_audit` | `bytecode-runtime` | ok (10) | unavailable | n/a | 0.4120 | n/a | n/a | n/a | n/a | n/a | n/a |
| `dependency_plan` | `compiled` | ok (10) | verified (10) | 96dc74508d9b7a476bafdef453b11e11f2f70279c58ccaa5dcb6d85c529c4a38 | 0.0700 | n/a | n/a | n/a | n/a | n/a | n/a |
| `dependency_plan` | `bytecode` | ok (10) | verified (10) | 96dc74508d9b7a476bafdef453b11e11f2f70279c58ccaa5dcb6d85c529c4a38 | 0.5060 | n/a | n/a | n/a | n/a | n/a | n/a |
| `dependency_plan` | `bytecode-runtime` | ok (10) | unavailable | n/a | 0.8260 | n/a | n/a | n/a | n/a | n/a | n/a |
| `future_await_race` | `compiled` | ok (10) | verified (10) | 33297798eeb96d6d471a6afe97ec8a8bf09eda0ce30e13917aed1db40fb930e4 | 0.0830 | n/a | n/a | n/a | n/a | n/a | n/a |
| `future_await_race` | `bytecode` | ok (10) | verified (10) | 33297798eeb96d6d471a6afe97ec8a8bf09eda0ce30e13917aed1db40fb930e4 | 0.1440 | n/a | n/a | n/a | n/a | n/a | n/a |
| `future_await_race` | `bytecode-runtime` | ok (10) | unavailable | n/a | 0.2170 | n/a | n/a | n/a | n/a | n/a | n/a |
