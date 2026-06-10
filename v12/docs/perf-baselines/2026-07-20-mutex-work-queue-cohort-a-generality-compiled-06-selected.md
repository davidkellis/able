# External Benchmark Comparison

- Generated: `2026-07-20T16:21:03.873861Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-20-mutex-work-queue-cohort-a-generality-go-reference.json`
- Suite: `custom`
- Able benchmarks: `nbody, tapelang_alphabet`
- Able modes: `compiled`
- Reference languages: `go`
- CPU pool: `0-3` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `nbody` | `compiled` | ok (5) | verified (5) | 40799ff8af9b84a416e8bf940921658787c57be38f638fb4d98c735c8d39e820 | 0.1720 | 0.0457 | 3.76x |
| `tapelang_alphabet` | `compiled` | ok (5) | verified (5) | a8ac3a1054c1aa7ac25f9b1e652a96a7ac86a1c1130687fc53b90e20c766d149 | 3.4780 | 2.1681 | 1.60x |
