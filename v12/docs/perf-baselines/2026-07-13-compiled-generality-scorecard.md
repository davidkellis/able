# External Benchmark Comparison

- Generated: `2026-07-13T15:42:32.833893Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Go reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-13-compiled-generality-go-refresh.json`
- Suite: `generality`
- Able benchmarks: `fib, binarytrees, matrixmultiply, quicksort, sudoku, sudoku_masks, i_before_e, base64, json, monte_carlo_pi, pidigits, mandelbrot, reverse_complement, k_nucleotide, nbody, tapelang_alphabet`
- Able modes: `compiled`
- Reference languages: `go`
- CPU affinity: `15`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `fib` | `compiled` | ok (1) | verified (1) | 231e154363ef69683f6dca862f5a448e786ccc90485636bc20a86e2e470a2188 | 3.8100 | 3.2278 | 1.18x |
| `binarytrees` | `compiled` | ok (1) | verified (1) | 341de11a51feab3d8122b4b5d6a68b038a2d14434aa9bc2372f39300bf5f48e1 | 33.3800 | 34.0543 | 0.98x |
| `matrixmultiply` | `compiled` | ok (1) | verified (1) | bce2053ea5fdc1351bf8d28398b346b894f8cf35c0e047effdca9dbb47e81d0c | 1.1800 | 0.9830 | 1.20x |
| `quicksort` | `compiled` | ok (1) | verified (1) | d0d07db0afd4266c1b6de5e76438bfa6aa974727e06c74e280aa7b497ca0e8b3 | 1.8900 | 2.7048 | 0.70x |
| `sudoku` | `compiled` | timeout (1) | not run | n/a | n/a | 0.1479 | n/a |
| `sudoku_masks` | `compiled` | ok (1) | verified (1) | 35a81e448daf9986f2a9b7c3a873dc6216bd55c969efec50c9b1de6d866659ec | 10.8100 | n/a | n/a |
| `i_before_e` | `compiled` | ok (1) | verified (1) | 981f0d37a277be25f359c097e10df2ef68009fea2d3e322aeed3a65d0fbaca39 | 0.2700 | 0.0607 | 4.45x |
| `base64` | `compiled` | ok (1) | verified (1) | 5f4c00cd811078942fc98cd5dbca3b47fac2f8d8210f07ea116c1a7c0d6ac316 | 2.6900 | 2.4982 | 1.08x |
| `json` | `compiled` | ok (1) | verified (1) | 7cbf9736119f7683d59646bea566591a2ef278db54286b8e2525aa39e907839e | 0.8500 | 1.5526 | 0.55x |
| `monte_carlo_pi` | `compiled` | ok (1) | verified (1) | dab6ddc0a42f6644dd620c574b8b552697477be92433e886b7116ec8723123d2 | 0.2300 | 0.2022 | 1.14x |
| `pidigits` | `compiled` | ok (1) | verified (1) | bdfa7b6c756d96492f472f97aee9cc139bee954d271eacedfd7ace5d2875f06c | 1.4200 | 1.2443 | 1.14x |
| `mandelbrot` | `compiled` | ok (1) | verified (1) | d1560cadc49ab9dca30a4cfbe4123c1fcc0240ed80de27ad37d7fcc97159173e | 0.2000 | 0.0505 | 3.96x |
| `reverse_complement` | `compiled` | ok (1) | verified (1) | db06a593d68950aa91293fb57ffed87ad57040c4ac7ab81ae9dc43d50b400bb7 | 0.1200 | 0.0179 | 6.70x |
| `k_nucleotide` | `compiled` | ok (1) | verified (1) | d628623daa677c00673df8d4961d14eb271b4112cdcb65b8233f07b69d7b49b8 | 3.7200 | 0.0710 | 52.39x |
| `nbody` | `compiled` | ok (1) | verified (1) | 2b1471417bee5179b1f9278ec51262ac9827e98d2df01cb93f954a22c0cd3e5d | 0.4300 | 0.0400 | 10.75x |
| `tapelang_alphabet` | `compiled` | ok (1) | verified (1) | a8ac3a1054c1aa7ac25f9b1e652a96a7ac86a1c1130687fc53b90e20c766d149 | 3.7300 | 1.9688 | 1.89x |
