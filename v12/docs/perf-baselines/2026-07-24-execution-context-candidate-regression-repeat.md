# External Benchmark Comparison

- Generated: `2026-07-24T17:54:05.419274Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-24-execution-context-go-reference.json`
- Suite: `custom`
- Able benchmarks: `manifest_normalization, concurrent_event_routing, mutex_work_queue, unicode_scalar_pipeline, wide_integer_records, rational_series, fixed_width_128, distance_field, base64, matrixmultiply`
- Able modes: `compiled`
- Reference languages: `go`
- CPU pool: `0-3` (each row records its resolved catalog budget)
- Experimental execution context: `enabled for compiled mode`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `manifest_normalization` | `compiled` | ok (5) | verified (5) | 2d6d55d5a76f3e45c6eb4fc3c0b892c2c5d8e02f3e38fa916d4f1c9a1579e9cb | 0.1900 | 0.0052 | 36.54x |
| `concurrent_event_routing` | `compiled` | ok (5) | verified (5) | 672e4542ce0d231428dd910348225e7c8b058824d7b8ceffbfb616d77a939d39 | 3.0100 | 0.0053 | 567.92x |
| `mutex_work_queue` | `compiled` | ok (5) | verified (5) | 57d3c0d15899da95d375a749cfe34dcc6942eb82a45427f197b9244b85ff8e58 | 1.7960 | 0.0048 | 374.17x |
| `unicode_scalar_pipeline` | `compiled` | ok (5) | verified (5) | c9efadb7f22969600334daa4a4eed2edde38c8e86d2c81d354d6f3979c854eb9 | 0.3800 | 0.0103 | 36.89x |
| `wide_integer_records` | `compiled` | ok (5) | verified (5) | f373537521cc6bfb0fb9e1a1eb36eb93a057654b526a4521878bc269261713e5 | 0.3040 | 0.0266 | 11.43x |
| `rational_series` | `compiled` | ok (5) | verified (5) | 127f0f44ee4870b57a188a7948f80a0a5d14584a326c345a48d4285594069f0c | 0.1700 | 0.0154 | 11.04x |
| `fixed_width_128` | `compiled` | ok (5) | verified (5) | eceabf5869b1abca8d6dd228b64a09f89e4e98ba8cabc4833ffee1218dafa56a | 0.3460 | 0.0061 | 56.72x |
| `distance_field` | `compiled` | ok (5) | verified (5) | cdaaf4451b236346af59b6a407f3136da96004e0c7c39c165546b7b9b21eda94 | 0.1000 | 0.0149 | 6.71x |
| `base64` | `compiled` | ok (5) | verified (5) | 5f4c00cd811078942fc98cd5dbca3b47fac2f8d8210f07ea116c1a7c0d6ac316 | 2.3980 | 2.6618 | 0.90x |
| `matrixmultiply` | `compiled` | ok (5) | verified (5) | bce2053ea5fdc1351bf8d28398b346b894f8cf35c0e047effdca9dbb47e81d0c | 1.1560 | 1.0259 | 1.13x |
