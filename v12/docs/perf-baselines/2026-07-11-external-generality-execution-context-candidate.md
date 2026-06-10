# External Benchmark Comparison

- Generated: `2026-07-11T10:15:53.831756Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Suite: `generality`
- Able benchmarks: `fib, binarytrees, matrixmultiply, quicksort, sudoku, i_before_e, base64, json, monte_carlo_pi, pidigits, mandelbrot, reverse_complement, k_nucleotide, nbody, tapelang_alphabet`
- Able modes: `compiled`
- Reference languages: `go, ruby, python`
- CPU affinity: `2-3`
- Experimental execution context: `enabled for compiled mode`

| Benchmark | Mode | Able Status | Able Real (s) | go Real (s) | Able/go | ruby Real (s) | Able/ruby | python Real (s) | Able/python |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `fib` | `compiled` | ok (1) | 3.4800 | 2.8400 | 1.23x | 46.6400 | 0.07x | 60.6700 | 0.06x |
| `binarytrees` | `compiled` | ok (1) | 18.5900 | 3.8300 | 4.85x | 20.3900 | 0.91x | 12.2500 | 1.52x |
| `matrixmultiply` | `compiled` | ok (1) | 1.1800 | 0.8800 | 1.34x | 42.9300 | 0.03x | 56.2900 | 0.02x |
| `quicksort` | `compiled` | ok (1) | 1.8900 | 2.0100 | 0.94x | 14.5800 | 0.13x | 20.3200 | 0.09x |
| `sudoku` | `compiled` | ok (1) | 0.1200 | 0.1300 | 0.92x | 5.6700 | 0.02x | 3.0200 | 0.04x |
| `i_before_e` | `compiled` | ok (1) | 0.1100 | 0.0500 | 2.20x | 0.1000 | 1.10x | 0.1300 | 0.85x |
| `base64` | `compiled` | ok (1) | 2.3600 | 2.2000 | 1.07x | 2.2100 | 1.07x | 3.3100 | 0.71x |
| `json` | `compiled` | ok (1) | 0.7000 | 1.3600 | 0.51x | 1.5600 | 0.45x | 2.8700 | 0.24x |
| `monte_carlo_pi` | `compiled` | ok (1) | 0.2000 | 0.1800 | 1.11x | 1.4200 | 0.14x | 1.6800 | 0.12x |
| `pidigits` | `compiled` | ok (1) | 1.1900 | 0.7400 | 1.61x | 9.1800 | 0.13x | n/a | n/a |
| `mandelbrot` | `compiled` | ok (1) | 0.1200 | 0.0400 | 3.00x | n/a | n/a | n/a | n/a |
| `reverse_complement` | `compiled` | ok (1) | 0.1000 | 0.0100 | 10.00x | n/a | n/a | n/a | n/a |
| `k_nucleotide` | `compiled` | ok (1) | 3.8600 | n/a | n/a | n/a | n/a | n/a | n/a |
| `nbody` | `compiled` | ok (1) | 0.6100 | n/a | n/a | n/a | n/a | n/a | n/a |
| `tapelang_alphabet` | `compiled` | ok (1) | 3.9700 | 1.7500 | 2.27x | 67.8300 | 0.06x | 58.9900 | 0.07x |
