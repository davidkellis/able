# External Benchmark Comparison

- Generated: `2026-07-20T23:49:53.124840Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Suite: `custom`
- Able benchmarks: `document_audit, dependency_plan, future_await_race, config_validation_extraction`
- Able modes: `bytecode, bytecode-runtime`
- Warmed bytecode main() calls per run: `1`
- Reference languages: `go, ruby, python`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go | ruby Real (s) | Able/ruby | python Real (s) | Able/python |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `document_audit` | `bytecode` | ok (10) | verified (10) | 0dad030a80c8a883cbb56fbcfebfd530d521075e15d5d91ba538bc93e66c0aab | 0.2940 | n/a | n/a | n/a | n/a | n/a | n/a |
| `document_audit` | `bytecode-runtime` | ok (10) | unavailable | n/a | 0.3600 | n/a | n/a | n/a | n/a | n/a | n/a |
| `dependency_plan` | `bytecode` | ok (10) | verified (10) | 96dc74508d9b7a476bafdef453b11e11f2f70279c58ccaa5dcb6d85c529c4a38 | 0.4460 | n/a | n/a | n/a | n/a | n/a | n/a |
| `dependency_plan` | `bytecode-runtime` | ok (10) | unavailable | n/a | 0.8480 | n/a | n/a | n/a | n/a | n/a | n/a |
| `future_await_race` | `bytecode` | ok (10) | verified (10) | 33297798eeb96d6d471a6afe97ec8a8bf09eda0ce30e13917aed1db40fb930e4 | 0.1260 | n/a | n/a | n/a | n/a | n/a | n/a |
| `future_await_race` | `bytecode-runtime` | ok (10) | unavailable | n/a | 0.2080 | n/a | n/a | n/a | n/a | n/a | n/a |
| `config_validation_extraction` | `bytecode` | ok (10) | verified (10) | c1aa99b9a13bb6e0c7731cb2aea77e300cd3cecc695df7fd4af90036939341d1 | 1.2680 | n/a | n/a | n/a | n/a | n/a | n/a |
| `config_validation_extraction` | `bytecode-runtime` | ok (10) | unavailable | n/a | 2.0970 | n/a | n/a | n/a | n/a | n/a | n/a |
