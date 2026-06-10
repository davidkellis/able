# External Benchmark Comparison

- Generated: `2026-07-13T16:11:42.786575Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-13-bytecode-generality-interpreter-refresh.json`
- Suite: `generality`
- Able benchmarks: `fib, binarytrees, matrixmultiply, quicksort, sudoku, sudoku_masks, i_before_e, base64, json, monte_carlo_pi, pidigits, mandelbrot, reverse_complement, k_nucleotide, nbody, tapelang_alphabet`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU affinity: `15`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `fib` | `bytecode` | ok (1) | verified (1) | 231e154363ef69683f6dca862f5a448e786ccc90485636bc20a86e2e470a2188 | 0.1600 | n/a | n/a | n/a | n/a |
| `binarytrees` | `bytecode` | timeout (1) | not run | n/a | n/a | n/a | n/a | n/a | n/a |
| `matrixmultiply` | `bytecode` | ok (1) | verified (1) | bce2053ea5fdc1351bf8d28398b346b894f8cf35c0e047effdca9dbb47e81d0c | 4.4800 | n/a | n/a | n/a | n/a |
| `quicksort` | `bytecode` | timeout (1) | not run | n/a | n/a | 24.7090 | n/a | 15.0942 | n/a |
| `sudoku` | `bytecode` | timeout (1) | not run | n/a | n/a | 3.0964 | n/a | 6.4054 | n/a |
| `sudoku_masks` | `bytecode` | timeout (1) | not run | n/a | n/a | 18.4054 | n/a | 21.6825 | n/a |
| `i_before_e` | `bytecode` | ok (1) | verified (1) | 981f0d37a277be25f359c097e10df2ef68009fea2d3e322aeed3a65d0fbaca39 | 0.5500 | 0.1018 | 5.40x | 0.1311 | 4.20x |
| `base64` | `bytecode` | ok (1) | verified (1) | 5f4c00cd811078942fc98cd5dbca3b47fac2f8d8210f07ea116c1a7c0d6ac316 | 4.3500 | 3.9394 | 1.10x | 2.5055 | 1.74x |
| `json` | `bytecode` | ok (1) | verified (1) | 7cbf9736119f7683d59646bea566591a2ef278db54286b8e2525aa39e907839e | 1.2000 | 2.6063 | 0.46x | 1.7365 | 0.69x |
| `monte_carlo_pi` | `bytecode` | ok (1) | verified (1) | dab6ddc0a42f6644dd620c574b8b552697477be92433e886b7116ec8723123d2 | 2.6800 | 1.5413 | 1.74x | 1.6674 | 1.61x |
| `pidigits` | `bytecode` | ok (1) | verified (1) | bdfa7b6c756d96492f472f97aee9cc139bee954d271eacedfd7ace5d2875f06c | 2.7100 | 4.0582 | 0.67x | 10.2528 | 0.26x |
| `mandelbrot` | `bytecode` | ok (1) | verified (1) | d1560cadc49ab9dca30a4cfbe4123c1fcc0240ed80de27ad37d7fcc97159173e | 6.7900 | 1.2999 | 5.22x | 1.9488 | 3.48x |
| `reverse_complement` | `bytecode` | ok (1) | verified (1) | db06a593d68950aa91293fb57ffed87ad57040c4ac7ab81ae9dc43d50b400bb7 | 6.7500 | 0.0420 | 160.71x | 0.0841 | 80.26x |
| `k_nucleotide` | `bytecode` | timeout (1) | not run | n/a | n/a | 1.3914 | n/a | 1.2755 | n/a |
| `nbody` | `bytecode` | timeout (1) | not run | n/a | n/a | 2.0879 | n/a | 3.2619 | n/a |
| `tapelang_alphabet` | `bytecode` | timeout (1) | not run | n/a | n/a | n/a | n/a | n/a | n/a |
