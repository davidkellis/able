# External Benchmark Comparison

- Generated: `2026-07-28T20:58:59.396723Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/var/tmp/able-mode-aware-stability-20260728-JbVhIT/interpreter-cpu6.json`
- Suite: `custom`
- Able benchmarks: `fib, matrixmultiply, json, pidigits`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU pool: `6` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `fib` | `bytecode` | ok (5) | verified (5) | a9c936c441fe6280cabd79ab1abac782a93ac7ad3495f87b8040d653a046ff36 | 0.2040 | 5.7455 | 0.04x | 4.1900 | 0.05x |
| `matrixmultiply` | `bytecode` | ok (5) | verified (5) | 4841d03fe93e0d4db2f42144f0a035ad7b5443bfdfca828012d5dbeed584a144 | 0.9580 | 3.2223 | 0.30x | 3.0445 | 0.31x |
| `json` | `bytecode` | ok (5) | verified (5) | 7cbf9736119f7683d59646bea566591a2ef278db54286b8e2525aa39e907839e | 0.9060 | 3.0176 | 0.30x | 1.7206 | 0.53x |
| `pidigits` | `bytecode` | ok (5) | verified (5) | bdfa7b6c756d96492f472f97aee9cc139bee954d271eacedfd7ace5d2875f06c | 2.4920 | 4.1902 | 0.59x | 9.8635 | 0.25x |
