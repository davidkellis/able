# External Benchmark Comparison

- Generated: `2026-07-11T18:33:45.590860Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Suite: `generality`
- Able benchmarks: `fib, binarytrees, matrixmultiply, quicksort, sudoku, i_before_e, base64, json, monte_carlo_pi, pidigits, mandelbrot, reverse_complement, k_nucleotide, nbody, tapelang_alphabet`
- Able modes: `compiled, bytecode`
- Reference languages: `go, ruby, python`
- CPU affinity: `2`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | go Real (s) | Able/go | ruby Real (s) | Able/ruby | python Real (s) | Able/python |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `fib` | `compiled` | ok (1) | verified (1) | 231e154363ef69683f6dca862f5a448e786ccc90485636bc20a86e2e470a2188 | 3.3300 | 2.8400 | 1.17x | 46.6400 | 0.07x | 60.6700 | 0.05x |
| `fib` | `bytecode` | ok (1) | verified (1) | 231e154363ef69683f6dca862f5a448e786ccc90485636bc20a86e2e470a2188 | 0.1500 | 2.8400 | 0.05x | 46.6400 | 0.00x | 60.6700 | 0.00x |
| `binarytrees` | `compiled` | ok (1) | verified (1) | 341de11a51feab3d8122b4b5d6a68b038a2d14434aa9bc2372f39300bf5f48e1 | 30.7800 | 3.8300 | 8.04x | 20.3900 | 1.51x | 12.2500 | 2.51x |
| `binarytrees` | `bytecode` | timeout (1) | not run | n/a | n/a | 3.8300 | n/a | 20.3900 | n/a | 12.2500 | n/a |
| `matrixmultiply` | `compiled` | ok (1) | unavailable | bce2053ea5fdc1351bf8d28398b346b894f8cf35c0e047effdca9dbb47e81d0c | 1.2300 | 0.8800 | 1.40x | 42.9300 | 0.03x | 56.2900 | 0.02x |
| `matrixmultiply` | `bytecode` | ok (1) | unavailable | bce2053ea5fdc1351bf8d28398b346b894f8cf35c0e047effdca9dbb47e81d0c | 4.4900 | 0.8800 | 5.10x | 42.9300 | 0.10x | 56.2900 | 0.08x |
| `quicksort` | `compiled` | ok (1) | verified (1) | d0d07db0afd4266c1b6de5e76438bfa6aa974727e06c74e280aa7b497ca0e8b3 | 1.9600 | 2.0100 | 0.98x | 14.5800 | 0.13x | 20.3200 | 0.10x |
| `quicksort` | `bytecode` | timeout (1) | not run | n/a | n/a | 2.0100 | n/a | 14.5800 | n/a | 20.3200 | n/a |
| `sudoku` | `compiled` | timeout (1) | not run | n/a | n/a | 0.1300 | n/a | 5.6700 | n/a | 3.0200 | n/a |
| `sudoku` | `bytecode` | timeout (1) | not run | n/a | n/a | 0.1300 | n/a | 5.6700 | n/a | 3.0200 | n/a |
| `i_before_e` | `compiled` | ok (1) | verified (1) | 981f0d37a277be25f359c097e10df2ef68009fea2d3e322aeed3a65d0fbaca39 | 0.2000 | 0.0500 | 4.00x | 0.1000 | 2.00x | 0.1300 | 1.54x |
| `i_before_e` | `bytecode` | ok (1) | verified (1) | 981f0d37a277be25f359c097e10df2ef68009fea2d3e322aeed3a65d0fbaca39 | 0.5600 | 0.0500 | 11.20x | 0.1000 | 5.60x | 0.1300 | 4.31x |
| `base64` | `compiled` | ok (1) | verified (1) | 5f4c00cd811078942fc98cd5dbca3b47fac2f8d8210f07ea116c1a7c0d6ac316 | 2.4500 | 2.2000 | 1.11x | 2.2100 | 1.11x | 3.3100 | 0.74x |
| `base64` | `bytecode` | ok (1) | verified (1) | 5f4c00cd811078942fc98cd5dbca3b47fac2f8d8210f07ea116c1a7c0d6ac316 | 3.0900 | 2.2000 | 1.40x | 2.2100 | 1.40x | 3.3100 | 0.93x |
| `json` | `compiled` | ok (1) | unavailable | 7cbf9736119f7683d59646bea566591a2ef278db54286b8e2525aa39e907839e | 0.8500 | 1.3600 | 0.62x | 1.5600 | 0.54x | 2.8700 | 0.30x |
| `json` | `bytecode` | ok (1) | unavailable | 7cbf9736119f7683d59646bea566591a2ef278db54286b8e2525aa39e907839e | 3.7300 | 1.3600 | 2.74x | 1.5600 | 2.39x | 2.8700 | 1.30x |
| `monte_carlo_pi` | `compiled` | ok (1) | verified (1) | dab6ddc0a42f6644dd620c574b8b552697477be92433e886b7116ec8723123d2 | 0.2100 | 0.1800 | 1.17x | 1.4200 | 0.15x | 1.6800 | 0.12x |
| `monte_carlo_pi` | `bytecode` | ok (1) | verified (1) | dab6ddc0a42f6644dd620c574b8b552697477be92433e886b7116ec8723123d2 | 2.7300 | 0.1800 | 15.17x | 1.4200 | 1.92x | 1.6800 | 1.62x |
| `pidigits` | `compiled` | ok (1) | verified (1) | bdfa7b6c756d96492f472f97aee9cc139bee954d271eacedfd7ace5d2875f06c | 1.4200 | 0.7400 | 1.92x | 9.1800 | 0.15x | n/a | n/a |
| `pidigits` | `bytecode` | ok (1) | verified (1) | bdfa7b6c756d96492f472f97aee9cc139bee954d271eacedfd7ace5d2875f06c | 3.3900 | 0.7400 | 4.58x | 9.1800 | 0.37x | n/a | n/a |
| `mandelbrot` | `compiled` | ok (1) | verified (1) | d1560cadc49ab9dca30a4cfbe4123c1fcc0240ed80de27ad37d7fcc97159173e | 0.2000 | 0.0400 | 5.00x | n/a | n/a | n/a | n/a |
| `mandelbrot` | `bytecode` | ok (1) | verified (1) | d1560cadc49ab9dca30a4cfbe4123c1fcc0240ed80de27ad37d7fcc97159173e | 7.2800 | 0.0400 | 182.00x | n/a | n/a | n/a | n/a |
| `reverse_complement` | `compiled` | ok (1) | verified (1) | db06a593d68950aa91293fb57ffed87ad57040c4ac7ab81ae9dc43d50b400bb7 | 0.1900 | 0.0100 | 19.00x | n/a | n/a | n/a | n/a |
| `reverse_complement` | `bytecode` | ok (1) | verified (1) | db06a593d68950aa91293fb57ffed87ad57040c4ac7ab81ae9dc43d50b400bb7 | 6.9000 | 0.0100 | 690.00x | n/a | n/a | n/a | n/a |
| `k_nucleotide` | `compiled` | ok (1) | verified (1) | d628623daa677c00673df8d4961d14eb271b4112cdcb65b8233f07b69d7b49b8 | 4.1900 | n/a | n/a | n/a | n/a | n/a | n/a |
| `k_nucleotide` | `bytecode` | ok (1) | verified (1) | d628623daa677c00673df8d4961d14eb271b4112cdcb65b8233f07b69d7b49b8 | 40.9300 | n/a | n/a | n/a | n/a | n/a | n/a |
| `nbody` | `compiled` | ok (1) | verified (1) | 2b1471417bee5179b1f9278ec51262ac9827e98d2df01cb93f954a22c0cd3e5d | 0.4900 | n/a | n/a | n/a | n/a | n/a | n/a |
| `nbody` | `bytecode` | timeout (1) | not run | n/a | n/a | n/a | n/a | n/a | n/a | n/a | n/a |
| `tapelang_alphabet` | `compiled` | ok (1) | unavailable | a8ac3a1054c1aa7ac25f9b1e652a96a7ac86a1c1130687fc53b90e20c766d149 | 3.8100 | 1.7500 | 2.18x | 67.8300 | 0.06x | 58.9900 | 0.06x |
| `tapelang_alphabet` | `bytecode` | timeout (1) | unavailable | n/a | n/a | 1.7500 | n/a | 67.8300 | n/a | 58.9900 | n/a |
