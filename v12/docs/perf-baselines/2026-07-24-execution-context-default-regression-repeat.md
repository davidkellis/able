# External Benchmark Comparison

- Generated: `2026-07-24T18:03:10.470399Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-24-execution-context-go-reference.json`
- Suite: `custom`
- Able benchmarks: `matrixmultiply, base64, distance_field, fixed_width_128, rational_series, wide_integer_records, unicode_scalar_pipeline, mutex_work_queue, concurrent_event_routing, manifest_normalization, future_pipeline, concurrent_stencil_reduction`
- Able modes: `compiled`
- Reference languages: `go`
- CPU pool: `0-3` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `matrixmultiply` | `compiled` | ok (5) | verified (5) | bce2053ea5fdc1351bf8d28398b346b894f8cf35c0e047effdca9dbb47e81d0c | 1.2040 | 1.0259 | 1.17x |
| `base64` | `compiled` | ok (5) | verified (5) | 5f4c00cd811078942fc98cd5dbca3b47fac2f8d8210f07ea116c1a7c0d6ac316 | 2.5360 | 2.6618 | 0.95x |
| `distance_field` | `compiled` | ok (5) | verified (5) | cdaaf4451b236346af59b6a407f3136da96004e0c7c39c165546b7b9b21eda94 | 0.1100 | 0.0149 | 7.38x |
| `fixed_width_128` | `compiled` | ok (5) | verified (5) | eceabf5869b1abca8d6dd228b64a09f89e4e98ba8cabc4833ffee1218dafa56a | 0.2260 | 0.0061 | 37.05x |
| `rational_series` | `compiled` | ok (5) | verified (5) | 127f0f44ee4870b57a188a7948f80a0a5d14584a326c345a48d4285594069f0c | 0.1360 | 0.0154 | 8.83x |
| `wide_integer_records` | `compiled` | ok (5) | verified (5) | f373537521cc6bfb0fb9e1a1eb36eb93a057654b526a4521878bc269261713e5 | 0.2220 | 0.0266 | 8.35x |
| `unicode_scalar_pipeline` | `compiled` | ok (5) | verified (5) | c9efadb7f22969600334daa4a4eed2edde38c8e86d2c81d354d6f3979c854eb9 | 0.2720 | 0.0103 | 26.41x |
| `mutex_work_queue` | `compiled` | ok (5) | verified (5) | 57d3c0d15899da95d375a749cfe34dcc6942eb82a45427f197b9244b85ff8e58 | 1.7200 | 0.0048 | 358.33x |
| `concurrent_event_routing` | `compiled` | ok (5) | verified (5) | 672e4542ce0d231428dd910348225e7c8b058824d7b8ceffbfb616d77a939d39 | 2.7240 | 0.0053 | 513.96x |
| `manifest_normalization` | `compiled` | ok (5) | verified (5) | 2d6d55d5a76f3e45c6eb4fc3c0b892c2c5d8e02f3e38fa916d4f1c9a1579e9cb | 0.2140 | 0.0052 | 41.15x |
| `future_pipeline` | `compiled` | ok (5) | verified (5) | 3db937fdd21b5719ab57c25679a1944e8f30128695c8edbfae4940aba7aa6d98 | 0.3200 | 0.0066 | 48.48x |
| `concurrent_stencil_reduction` | `compiled` | ok (5) | verified (5) | 42870ec44f0b8a860e066ec155ce13e2916bbff632d74a5c87704f7f81fa4a3b | 0.2200 | 0.0060 | 36.67x |
