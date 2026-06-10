# External Benchmark Comparison

- Generated: `2026-07-15T06:29:56.641505Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-15-post-extern-source-selection-generality-go-reference.json`
- Suite: `custom`
- Able benchmarks: `nbody, tapelang_alphabet`
- Able modes: `compiled`
- Reference languages: `go`
- CPU affinity: `14`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `nbody` | `compiled` | ok (1) | verified (1) | 2b1471417bee5179b1f9278ec51262ac9827e98d2df01cb93f954a22c0cd3e5d | 0.3700 | 0.0303 | 12.21x |
| `tapelang_alphabet` | `compiled` | ok (1) | verified (1) | a8ac3a1054c1aa7ac25f9b1e652a96a7ac86a1c1130687fc53b90e20c766d149 | 3.4300 | 1.7337 | 1.98x |
