# External Benchmark Comparison

- Generated: `2026-07-16T01:43:06.120776Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-15-error-reconcile-a-generality-go-reference.json`
- Suite: `custom`
- Able benchmarks: `nbody, tapelang_alphabet`
- Able modes: `compiled`
- Reference languages: `go`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `nbody` | `compiled` | ok (5) | verified (5) | 2b1471417bee5179b1f9278ec51262ac9827e98d2df01cb93f954a22c0cd3e5d | 0.3860 | 0.0302 | 12.78x |
| `tapelang_alphabet` | `compiled` | ok (5) | verified (5) | a8ac3a1054c1aa7ac25f9b1e652a96a7ac86a1c1130687fc53b90e20c766d149 | 3.2860 | 1.7589 | 1.87x |
