# External Benchmark Comparison

- Generated: `2026-07-19T01:34:42.638750Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-18-fasta-generation-interpreter-reference.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-18-fasta-generation-go-reference.json`
- Suite: `bulk-output`
- Able benchmarks: `mandelbrot, reverse_complement, fasta_generation`
- Able modes: `compiled, bytecode`
- Reference languages: `go, python, ruby`
- CPU pool: `0` (each row records its resolved catalog budget)

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `mandelbrot` | `compiled` | ok (5) | verified (5) | d1560cadc49ab9dca30a4cfbe4123c1fcc0240ed80de27ad37d7fcc97159173e | 0.1260 | n/a | n/a | n/a | n/a | n/a | n/a |
| `mandelbrot` | `bytecode` | ok (5) | verified (5) | d1560cadc49ab9dca30a4cfbe4123c1fcc0240ed80de27ad37d7fcc97159173e | 6.7620 | n/a | n/a | n/a | n/a | n/a | n/a |
| `reverse_complement` | `compiled` | ok (5) | verified (5) | db06a593d68950aa91293fb57ffed87ad57040c4ac7ab81ae9dc43d50b400bb7 | 0.1040 | n/a | n/a | n/a | n/a | n/a | n/a |
| `reverse_complement` | `bytecode` | ok (5) | verified (5) | db06a593d68950aa91293fb57ffed87ad57040c4ac7ab81ae9dc43d50b400bb7 | 4.2420 | n/a | n/a | n/a | n/a | n/a | n/a |
| `fasta_generation` | `compiled` | ok (5) | verified (5) | 2907f3fb66fea247549c0f26b5b5d5cd1940a055574b72dad344283e1eb0fd10 | 0.1120 | 0.0258 | 4.34x | n/a | n/a | n/a | n/a |
| `fasta_generation` | `bytecode` | ok (5) | verified (5) | 2907f3fb66fea247549c0f26b5b5d5cd1940a055574b72dad344283e1eb0fd10 | 2.1660 | n/a | n/a | 0.2262 | 9.58x | 0.3106 | 6.97x |
