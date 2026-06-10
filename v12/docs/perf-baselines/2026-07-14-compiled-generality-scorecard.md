# External Benchmark Comparison

- Generated: `2026-07-14T03:55:31.023990Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-14-compiled-generality-go-refresh.json`
- Suite: `generality`
- Able benchmarks: `fib, binarytrees, matrixmultiply, quicksort, sudoku, sudoku_masks, i_before_e, base64, json, monte_carlo_pi, pidigits, mandelbrot, reverse_complement, k_nucleotide, nbody, tapelang_alphabet`
- Able modes: `compiled`
- Reference languages: `go`
- CPU affinity: `15`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `fib` | `compiled` | ok (1) | verified (1) | 231e154363ef69683f6dca862f5a448e786ccc90485636bc20a86e2e470a2188 | 4.2900 | 4.2224 | 1.02x |
| `binarytrees` | `compiled` | ok (1) | verified (1) | 341de11a51feab3d8122b4b5d6a68b038a2d14434aa9bc2372f39300bf5f48e1 | 32.7800 | 38.4634 | 0.85x |
| `matrixmultiply` | `compiled` | ok (1) | verified (1) | bce2053ea5fdc1351bf8d28398b346b894f8cf35c0e047effdca9dbb47e81d0c | 1.1500 | 1.0185 | 1.13x |
| `quicksort` | `compiled` | ok (1) | verified (1) | d0d07db0afd4266c1b6de5e76438bfa6aa974727e06c74e280aa7b497ca0e8b3 | 1.9000 | 2.6425 | 0.72x |
| `sudoku` | `compiled` | timeout (1) | not run | n/a | n/a | 0.1464 | n/a |
| `sudoku_masks` | `compiled` | ok (1) | verified (1) | 35a81e448daf9986f2a9b7c3a873dc6216bd55c969efec50c9b1de6d866659ec | 9.1200 | 0.5895 | 15.47x |
| `i_before_e` | `compiled` | ok (1) | verified (1) | 981f0d37a277be25f359c097e10df2ef68009fea2d3e322aeed3a65d0fbaca39 | 0.1000 | 0.0610 | 1.64x |
| `base64` | `compiled` | ok (1) | verified (1) | 5f4c00cd811078942fc98cd5dbca3b47fac2f8d8210f07ea116c1a7c0d6ac316 | 2.5500 | 2.6446 | 0.96x |
| `json` | `compiled` | ok (1) | verified (1) | 7cbf9736119f7683d59646bea566591a2ef278db54286b8e2525aa39e907839e | 0.9700 | 1.5141 | 0.64x |
| `monte_carlo_pi` | `compiled` | ok (1) | verified (1) | dab6ddc0a42f6644dd620c574b8b552697477be92433e886b7116ec8723123d2 | 0.3800 | 0.2195 | 1.73x |
| `pidigits` | `compiled` | ok (1) | verified (1) | bdfa7b6c756d96492f472f97aee9cc139bee954d271eacedfd7ace5d2875f06c | 1.3600 | 1.2374 | 1.10x |
| `mandelbrot` | `compiled` | ok (1) | verified (1) | d1560cadc49ab9dca30a4cfbe4123c1fcc0240ed80de27ad37d7fcc97159173e | 0.1300 | 0.0503 | 2.58x |
| `reverse_complement` | `compiled` | ok (1) | verified (1) | db06a593d68950aa91293fb57ffed87ad57040c4ac7ab81ae9dc43d50b400bb7 | 0.1400 | 0.0160 | 8.75x |
| `k_nucleotide` | `compiled` | ok (1) | verified (1) | d628623daa677c00673df8d4961d14eb271b4112cdcb65b8233f07b69d7b49b8 | 4.4900 | 0.0559 | 80.32x |
| `nbody` | `compiled` | ok (1) | verified (1) | 2b1471417bee5179b1f9278ec51262ac9827e98d2df01cb93f954a22c0cd3e5d | 0.5300 | 0.0366 | 14.48x |
| `tapelang_alphabet` | `compiled` | ok (1) | verified (1) | a8ac3a1054c1aa7ac25f9b1e652a96a7ac86a1c1130687fc53b90e20c766d149 | 3.8400 | 1.9705 | 1.95x |
