# External Benchmark Comparison

- Generated: `2026-07-21T23:08:25.932725Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-21-truthiness-cast-numeric-next-closures-compiled-c2-go-reference.json`
- Suite: `custom`
- Able benchmarks: `distance_field, nbody`
- Able modes: `compiled`
- Reference languages: `go`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `distance_field` | `compiled` | ok (5) | verified (5) | cdaaf4451b236346af59b6a407f3136da96004e0c7c39c165546b7b9b21eda94 | 0.1140 | 0.0139 | 8.20x |
| `nbody` | `compiled` | ok (5) | verified (5) | 40799ff8af9b84a416e8bf940921658787c57be38f638fb4d98c735c8d39e820 | 0.1780 | 0.0418 | 4.26x |
