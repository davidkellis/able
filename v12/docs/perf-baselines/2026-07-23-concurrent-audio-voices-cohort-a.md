# External Benchmark Comparison

- Generated: `2026-07-23T22:32:10.017930Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-23-concurrent-audio-voices-interpreter-a.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-23-concurrent-audio-voices-go-a.json`
- Suite: `concurrent-audio-voices`
- Able benchmarks: `concurrent_audio_voices`
- Able modes: `compiled, bytecode`
- Reference languages: `go, python, ruby`
- CPU pool: `0-15` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `concurrent_audio_voices` | `compiled` | ok (5) | verified (5) | 6ec28390bc9c749cf18e4a6fbd4bc03154345e1acf0a2f6601baab16884a6e28 | 0.9260 | 0.0049 | 188.98x | n/a | n/a | n/a | n/a |
| `concurrent_audio_voices` | `bytecode` | ok (5) | verified (5) | 6ec28390bc9c749cf18e4a6fbd4bc03154345e1acf0a2f6601baab16884a6e28 | 1.2900 | n/a | n/a | 0.1335 | 9.66x | 0.1240 | 10.40x |
