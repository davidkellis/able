# External Benchmark Comparison

- Generated: `2026-07-23T22:33:22.600246Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-23-concurrent-audio-voices-interpreter-b.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-23-concurrent-audio-voices-go-b.json`
- Suite: `concurrent-audio-voices`
- Able benchmarks: `concurrent_audio_voices`
- Able modes: `compiled, bytecode`
- Reference languages: `go, python, ruby`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `concurrent_audio_voices` | `compiled` | ok (5) | verified (5) | 6ec28390bc9c749cf18e4a6fbd4bc03154345e1acf0a2f6601baab16884a6e28 | 0.9420 | 0.0045 | 209.33x | n/a | n/a | n/a | n/a |
| `concurrent_audio_voices` | `bytecode` | ok (5) | verified (5) | 6ec28390bc9c749cf18e4a6fbd4bc03154345e1acf0a2f6601baab16884a6e28 | 1.2600 | n/a | n/a | 0.1285 | 9.81x | 0.1239 | 10.17x |
