# External Benchmark Comparison

- Generated: `2026-07-10T19:09:43.468253Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Suite: `generality`
- Able benchmarks: `fib, binarytrees, matrixmultiply, quicksort, sudoku, i_before_e, base64, json, monte_carlo_pi, pidigits, mandelbrot, reverse_complement, k_nucleotide, nbody, tapelang_alphabet`
- Able modes: `compiled, bytecode`
- Reference languages: `go, ruby, python`
- CPU affinity: `0`

| Benchmark | Mode | Able Status | Able Real (s) | go Real (s) | Able/go | ruby Real (s) | Able/ruby | python Real (s) | Able/python |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `fib` | `compiled` | ok (1) | 3.6200 | 2.8400 | 1.27x | 46.6400 | 0.08x | 60.6700 | 0.06x |
| `fib` | `bytecode` | ok (1) | 0.1600 | 2.8400 | 0.06x | 46.6400 | 0.00x | 60.6700 | 0.00x |
| `binarytrees` | `compiled` | timeout (1) | n/a | 3.8300 | n/a | 20.3900 | n/a | 12.2500 | n/a |
| `binarytrees` | `bytecode` | timeout (1) | n/a | 3.8300 | n/a | 20.3900 | n/a | 12.2500 | n/a |
| `matrixmultiply` | `compiled` | ok (1) | 1.3100 | 0.8800 | 1.49x | 42.9300 | 0.03x | 56.2900 | 0.02x |
| `matrixmultiply` | `bytecode` | ok (1) | 5.4800 | 0.8800 | 6.23x | 42.9300 | 0.13x | 56.2900 | 0.10x |
| `quicksort` | `compiled` | ok (1) | 2.5000 | 2.0100 | 1.24x | 14.5800 | 0.17x | 20.3200 | 0.12x |
| `quicksort` | `bytecode` | timeout (1) | n/a | 2.0100 | n/a | 14.5800 | n/a | 20.3200 | n/a |
| `sudoku` | `compiled` | ok (1) | 0.1300 | 0.1300 | 1.00x | 5.6700 | 0.02x | 3.0200 | 0.04x |
| `sudoku` | `bytecode` | ok (1) | 0.5100 | 0.1300 | 3.92x | 5.6700 | 0.09x | 3.0200 | 0.17x |
| `i_before_e` | `compiled` | ok (1) | 0.1800 | 0.0500 | 3.60x | 0.1000 | 1.80x | 0.1300 | 1.38x |
| `i_before_e` | `bytecode` | ok (1) | 0.5800 | 0.0500 | 11.60x | 0.1000 | 5.80x | 0.1300 | 4.46x |
| `base64` | `compiled` | ok (1) | 2.4600 | 2.2000 | 1.12x | 2.2100 | 1.11x | 3.3100 | 0.74x |
| `base64` | `bytecode` | ok (1) | 3.0700 | 2.2000 | 1.40x | 2.2100 | 1.39x | 3.3100 | 0.93x |
| `json` | `compiled` | ok (1) | 0.8400 | 1.3600 | 0.62x | 1.5600 | 0.54x | 2.8700 | 0.29x |
| `json` | `bytecode` | ok (1) | 0.8400 | 1.3600 | 0.62x | 1.5600 | 0.54x | 2.8700 | 0.29x |
| `monte_carlo_pi` | `compiled` | ok (1) | 0.2100 | 0.1800 | 1.17x | 1.4200 | 0.15x | 1.6800 | 0.12x |
| `monte_carlo_pi` | `bytecode` | ok (1) | 2.7700 | 0.1800 | 15.39x | 1.4200 | 1.95x | 1.6800 | 1.65x |
| `pidigits` | `compiled` | ok (1) | 1.4000 | 0.7400 | 1.89x | 9.1800 | 0.15x | n/a | n/a |
| `pidigits` | `bytecode` | ok (1) | 2.2100 | 0.7400 | 2.99x | 9.1800 | 0.24x | n/a | n/a |
| `mandelbrot` | `compiled` | ok (1) | 0.2000 | 0.0400 | 5.00x | n/a | n/a | n/a | n/a |
| `mandelbrot` | `bytecode` | ok (1) | 6.7700 | 0.0400 | 169.25x | n/a | n/a | n/a | n/a |
| `reverse_complement` | `compiled` | ok (1) | 0.1900 | 0.0100 | 19.00x | n/a | n/a | n/a | n/a |
| `reverse_complement` | `bytecode` | ok (1) | 5.6100 | 0.0100 | 561.00x | n/a | n/a | n/a | n/a |
| `k_nucleotide` | `compiled` | ok (1) | 5.1000 | n/a | n/a | n/a | n/a | n/a | n/a |
| `k_nucleotide` | `bytecode` | timeout (1) | n/a | n/a | n/a | n/a | n/a | n/a | n/a |
| `nbody` | `compiled` | error (1) | n/a | n/a | n/a | n/a | n/a | n/a | n/a |
| `nbody` | `bytecode` | error (1) | n/a | n/a | n/a | n/a | n/a | n/a | n/a |
| `tapelang_alphabet` | `compiled` | ok (1) | 3.8000 | 1.7500 | 2.17x | 67.8300 | 0.06x | 58.9900 | 0.06x |
| `tapelang_alphabet` | `bytecode` | timeout (1) | n/a | 1.7500 | n/a | 67.8300 | n/a | 58.9900 | n/a |
