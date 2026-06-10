# External Benchmark Comparison

- Generated: `2026-07-11T10:10:26.116858Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Suite: `generality`
- Able benchmarks: `fib, binarytrees, matrixmultiply, quicksort, sudoku, i_before_e, base64, json, monte_carlo_pi, pidigits, mandelbrot, reverse_complement, k_nucleotide, nbody, tapelang_alphabet`
- Able modes: `compiled, bytecode`
- Reference languages: `go, ruby, python`
- CPU affinity: `2-3`

| Benchmark | Mode | Able Status | Able Real (s) | go Real (s) | Able/go | ruby Real (s) | Able/ruby | python Real (s) | Able/python |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `fib` | `compiled` | ok (1) | 3.3300 | 2.8400 | 1.17x | 46.6400 | 0.07x | 60.6700 | 0.05x |
| `fib` | `bytecode` | ok (1) | 0.1400 | 2.8400 | 0.05x | 46.6400 | 0.00x | 60.6700 | 0.00x |
| `binarytrees` | `compiled` | ok (1) | 17.1600 | 3.8300 | 4.48x | 20.3900 | 0.84x | 12.2500 | 1.40x |
| `binarytrees` | `bytecode` | timeout (1) | n/a | 3.8300 | n/a | 20.3900 | n/a | 12.2500 | n/a |
| `matrixmultiply` | `compiled` | ok (1) | 1.3000 | 0.8800 | 1.48x | 42.9300 | 0.03x | 56.2900 | 0.02x |
| `matrixmultiply` | `bytecode` | ok (1) | 4.0000 | 0.8800 | 4.55x | 42.9300 | 0.09x | 56.2900 | 0.07x |
| `quicksort` | `compiled` | ok (1) | 1.8900 | 2.0100 | 0.94x | 14.5800 | 0.13x | 20.3200 | 0.09x |
| `quicksort` | `bytecode` | timeout (1) | n/a | 2.0100 | n/a | 14.5800 | n/a | 20.3200 | n/a |
| `sudoku` | `compiled` | ok (1) | 0.1900 | 0.1300 | 1.46x | 5.6700 | 0.03x | 3.0200 | 0.06x |
| `sudoku` | `bytecode` | ok (1) | 0.4300 | 0.1300 | 3.31x | 5.6700 | 0.08x | 3.0200 | 0.14x |
| `i_before_e` | `compiled` | ok (1) | 0.1000 | 0.0500 | 2.00x | 0.1000 | 1.00x | 0.1300 | 0.77x |
| `i_before_e` | `bytecode` | ok (1) | 0.5100 | 0.0500 | 10.20x | 0.1000 | 5.10x | 0.1300 | 3.92x |
| `base64` | `compiled` | ok (1) | 2.2700 | 2.2000 | 1.03x | 2.2100 | 1.03x | 3.3100 | 0.69x |
| `base64` | `bytecode` | ok (1) | 3.0100 | 2.2000 | 1.37x | 2.2100 | 1.36x | 3.3100 | 0.91x |
| `json` | `compiled` | ok (1) | 0.6800 | 1.3600 | 0.50x | 1.5600 | 0.44x | 2.8700 | 0.24x |
| `json` | `bytecode` | ok (1) | 0.7900 | 1.3600 | 0.58x | 1.5600 | 0.51x | 2.8700 | 0.28x |
| `monte_carlo_pi` | `compiled` | ok (1) | 0.2000 | 0.1800 | 1.11x | 1.4200 | 0.14x | 1.6800 | 0.12x |
| `monte_carlo_pi` | `bytecode` | ok (1) | 2.4900 | 0.1800 | 13.83x | 1.4200 | 1.75x | 1.6800 | 1.48x |
| `pidigits` | `compiled` | ok (1) | 1.1600 | 0.7400 | 1.57x | 9.1800 | 0.13x | n/a | n/a |
| `pidigits` | `bytecode` | ok (1) | 2.1200 | 0.7400 | 2.86x | 9.1800 | 0.23x | n/a | n/a |
| `mandelbrot` | `compiled` | ok (1) | 0.1200 | 0.0400 | 3.00x | n/a | n/a | n/a | n/a |
| `mandelbrot` | `bytecode` | ok (1) | 6.3700 | 0.0400 | 159.25x | n/a | n/a | n/a | n/a |
| `reverse_complement` | `compiled` | ok (1) | 0.0900 | 0.0100 | 9.00x | n/a | n/a | n/a | n/a |
| `reverse_complement` | `bytecode` | ok (1) | 5.7000 | 0.0100 | 570.00x | n/a | n/a | n/a | n/a |
| `k_nucleotide` | `compiled` | ok (1) | 3.5700 | n/a | n/a | n/a | n/a | n/a | n/a |
| `k_nucleotide` | `bytecode` | ok (1) | 38.5800 | n/a | n/a | n/a | n/a | n/a | n/a |
| `nbody` | `compiled` | ok (1) | 0.4800 | n/a | n/a | n/a | n/a | n/a | n/a |
| `nbody` | `bytecode` | timeout (1) | n/a | n/a | n/a | n/a | n/a | n/a | n/a |
| `tapelang_alphabet` | `compiled` | ok (1) | 3.8500 | 1.7500 | 2.20x | 67.8300 | 0.06x | 58.9900 | 0.07x |
| `tapelang_alphabet` | `bytecode` | timeout (1) | n/a | 1.7500 | n/a | 67.8300 | n/a | 58.9900 | n/a |
