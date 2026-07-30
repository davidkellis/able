# External Benchmark Comparison

- Generated: `2026-07-30T13:39:56.942227Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Go reference rows: `/var/tmp/able-versioned-telemetry-20260730/measurements/go-reference-final.json`
- Suite: `custom`
- Able benchmarks: `versioned_telemetry_pipeline`
- Able modes: `compiled`
- Reference languages: `go`
- CPU pool: `12` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `versioned_telemetry_pipeline` | `compiled` | ok (5) | verified (5) | 824f93580f56e01b938c047701218b04041ebaaab783db5d29c0f2eafae11a86 | 51.4180 | 0.2201 | 233.61x |
