# External Benchmark Comparison

- Generated: `2026-07-10T16:42:27.147653Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Suite: `generality`
- Able benchmarks: `fib, binarytrees, matrixmultiply, quicksort, sudoku, i_before_e, base64, json, monte_carlo_pi, pidigits, mandelbrot, reverse_complement, k_nucleotide, nbody, tapelang_alphabet`
- Able modes: `compiled, bytecode`
- Reference languages: `go, ruby, python`
- CPU affinity: `0`

| Benchmark | Mode | Able Status | Able Real (s) | go Real (s) | Able/go | ruby Real (s) | Able/ruby | python Real (s) | Able/python |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `fib` | `compiled` | ok (1) | 3.3800 | 2.8400 | 1.19x | 46.6400 | 0.07x | 60.6700 | 0.06x |
| `fib` | `bytecode` | ok (1) | 0.1600 | 2.8400 | 0.06x | 46.6400 | 0.00x | 60.6700 | 0.00x |
| `binarytrees` | `compiled` | timeout (1) | n/a | 3.8300 | n/a | 20.3900 | n/a | 12.2500 | n/a |
| `binarytrees` | `bytecode` | timeout (1) | n/a | 3.8300 | n/a | 20.3900 | n/a | 12.2500 | n/a |
| `matrixmultiply` | `compiled` | ok (1) | 1.1900 | 0.8800 | 1.35x | 42.9300 | 0.03x | 56.2900 | 0.02x |
| `matrixmultiply` | `bytecode` | ok (1) | 3.7500 | 0.8800 | 4.26x | 42.9300 | 0.09x | 56.2900 | 0.07x |
| `quicksort` | `compiled` | ok (1) | 2.0200 | 2.0100 | 1.00x | 14.5800 | 0.14x | 20.3200 | 0.10x |
| `quicksort` | `bytecode` | timeout (1) | n/a | 2.0100 | n/a | 14.5800 | n/a | 20.3200 | n/a |
| `sudoku` | `compiled` | ok (1) | 0.2000 | 0.1300 | 1.54x | 5.6700 | 0.04x | 3.0200 | 0.07x |
| `sudoku` | `bytecode` | ok (1) | 0.4800 | 0.1300 | 3.69x | 5.6700 | 0.08x | 3.0200 | 0.16x |
| `i_before_e` | `compiled` | ok (1) | 0.2000 | 0.0500 | 4.00x | 0.1000 | 2.00x | 0.1300 | 1.54x |
| `i_before_e` | `bytecode` | ok (1) | 0.5600 | 0.0500 | 11.20x | 0.1000 | 5.60x | 0.1300 | 4.31x |
| `base64` | `compiled` | ok (1) | 2.5100 | 2.2000 | 1.14x | 2.2100 | 1.14x | 3.3100 | 0.76x |
| `base64` | `bytecode` | ok (1) | 3.5900 | 2.2000 | 1.63x | 2.2100 | 1.62x | 3.3100 | 1.08x |
| `json` | `compiled` | ok (1) | 0.7300 | 1.3600 | 0.54x | 1.5600 | 0.47x | 2.8700 | 0.25x |
| `json` | `bytecode` | ok (1) | 1.2000 | 1.3600 | 0.88x | 1.5600 | 0.77x | 2.8700 | 0.42x |
| `monte_carlo_pi` | `compiled` | ok (1) | 0.2000 | 0.1800 | 1.11x | 1.4200 | 0.14x | 1.6800 | 0.12x |
| `monte_carlo_pi` | `bytecode` | ok (1) | 2.6100 | 0.1800 | 14.50x | 1.4200 | 1.84x | 1.6800 | 1.55x |
| `pidigits` | `compiled` | ok (1) | 1.3300 | 0.7400 | 1.80x | 9.1800 | 0.14x | n/a | n/a |
| `pidigits` | `bytecode` | ok (1) | 2.1900 | 0.7400 | 2.96x | 9.1800 | 0.24x | n/a | n/a |
| `mandelbrot` | `compiled` | ok (1) | 0.1300 | 0.0400 | 3.25x | n/a | n/a | n/a | n/a |
| `mandelbrot` | `bytecode` | ok (1) | 7.0700 | 0.0400 | 176.75x | n/a | n/a | n/a | n/a |
| `reverse_complement` | `compiled` | ok (1) | 0.1200 | 0.0100 | 12.00x | n/a | n/a | n/a | n/a |
| `reverse_complement` | `bytecode` | ok (1) | 5.8200 | 0.0100 | 582.00x | n/a | n/a | n/a | n/a |
| `k_nucleotide` | `compiled` | ok (1) | 4.6400 | n/a | n/a | n/a | n/a | n/a | n/a |
| `k_nucleotide` | `bytecode` | timeout (1) | n/a | n/a | n/a | n/a | n/a | n/a | n/a |
| `nbody` | `compiled` | error (1) | n/a | n/a | n/a | n/a | n/a | n/a | n/a |
| `nbody` | `bytecode` | error (1) | n/a | n/a | n/a | n/a | n/a | n/a | n/a |
| `tapelang_alphabet` | `compiled` | ok (1) | 3.7100 | 1.7500 | 2.12x | 67.8300 | 0.05x | 58.9900 | 0.06x |
| `tapelang_alphabet` | `bytecode` | timeout (1) | n/a | 1.7500 | n/a | 67.8300 | n/a | 58.9900 | n/a |
