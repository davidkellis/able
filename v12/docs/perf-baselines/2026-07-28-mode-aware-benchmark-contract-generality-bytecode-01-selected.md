# External Benchmark Comparison

- Generated: `2026-07-28T19:41:52.326251Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-28-mode-aware-benchmark-contract-generality-interpreter-reference-selected.json`
- Suite: `custom`
- Able benchmarks: `fib, binarytrees, matrixmultiply`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU pool: `7-10` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `fib` | `bytecode` | ok (5) | verified (5) | a9c936c441fe6280cabd79ab1abac782a93ac7ad3495f87b8040d653a046ff36 | 0.2480 | 5.2074 | 0.05x | 4.3390 | 0.06x |
| `binarytrees` | `bytecode` | ok (5) | verified (5) | 92b6df65f712164fc10a53dbc1085312406b233110001316a85b78ed0a16cfab | 12.2980 | 0.5610 | 21.92x | 0.5744 | 21.41x |
| `matrixmultiply` | `bytecode` | ok (5) | verified (5) | 4841d03fe93e0d4db2f42144f0a035ad7b5443bfdfca828012d5dbeed584a144 | 1.0160 | 3.3338 | 0.30x | 3.1064 | 0.33x |
