# External Benchmark Comparison

- Generated: `2026-07-15T07:33:18.215516Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Suite: `custom`
- Able benchmarks: `quicksort, pidigits`
- Able modes: `compiled`
- Reference languages: `go, ruby, python`
- CPU affinity: `14`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go | ruby Real (s) | Able/ruby | python Real (s) | Able/python |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `quicksort` | `compiled` | ok (3) | verified (3) | d0d07db0afd4266c1b6de5e76438bfa6aa974727e06c74e280aa7b497ca0e8b3 | 1.6600 | 2.0100 | 0.83x | 14.5800 | 0.11x | 20.3200 | 0.08x |
| `pidigits` | `compiled` | ok (3) | verified (3) | bdfa7b6c756d96492f472f97aee9cc139bee954d271eacedfd7ace5d2875f06c | 1.0767 | 0.7400 | 1.46x | 9.1800 | 0.12x | n/a | n/a |
