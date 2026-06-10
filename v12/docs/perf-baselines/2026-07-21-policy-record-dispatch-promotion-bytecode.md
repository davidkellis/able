# External Benchmark Comparison

- Generated: `2026-07-21T16:50:18.320836Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-21-policy-record-dispatch-interpreter-reference.json`
- Suite: `custom`
- Able benchmarks: `policy_record_dispatch`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `policy_record_dispatch` | `bytecode` | ok (5) | verified (5) | f15c66e53e4c89650ae12aff6c2f466f2cf0c7adb6ca3fe5be0127bfba217c05 | 8.7820 | 0.0245 | 358.45x | 0.0519 | 169.21x |
