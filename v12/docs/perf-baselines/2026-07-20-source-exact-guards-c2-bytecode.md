# External Benchmark Comparison

- Generated: `2026-07-21T03:28:52.043981Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-20-source-exact-guards-c2-interpreter-reference.json`
- Suite: `custom`
- Able benchmarks: `pidigits, json`
- Able modes: `bytecode`
- Reference languages: `ruby, python`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | ruby Real (s) | Able/ruby | python Real (s) | Able/python |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `pidigits` | `bytecode` | ok (5) | verified (5) | bdfa7b6c756d96492f472f97aee9cc139bee954d271eacedfd7ace5d2875f06c | 2.5180 | 10.6965 | 0.24x | 4.0615 | 0.62x |
| `json` | `bytecode` | ok (5) | verified (5) | 7cbf9736119f7683d59646bea566591a2ef278db54286b8e2525aa39e907839e | 0.9980 | 1.6817 | 0.59x | 2.8027 | 0.36x |
