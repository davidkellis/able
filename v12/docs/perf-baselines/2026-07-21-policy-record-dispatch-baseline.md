# External Benchmark Comparison

- Generated: `2026-07-21T16:22:51.712455Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-21-policy-record-dispatch-interpreter-reference.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-21-policy-record-dispatch-go-reference.json`
- Suite: `custom`
- Able benchmarks: `policy_record_dispatch`
- Able modes: `compiled, bytecode`
- Reference languages: `go, python, ruby`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `policy_record_dispatch` | `compiled` | ok (5) | verified (5) | f15c66e53e4c89650ae12aff6c2f466f2cf0c7adb6ca3fe5be0127bfba217c05 | 0.2340 | 0.0092 | 25.43x | n/a | n/a | n/a | n/a |
| `policy_record_dispatch` | `bytecode` | ok (5) | verified (5) | f15c66e53e4c89650ae12aff6c2f466f2cf0c7adb6ca3fe5be0127bfba217c05 | 7.7500 | n/a | n/a | 0.0237 | 327.00x | 0.0451 | 171.84x |
