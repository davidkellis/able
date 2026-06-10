# External Benchmark Comparison

- Generated: `2026-07-14T04:01:56.267498Z`
- External results: `/home/david/sync/projects/benchmarks/results.json`
- Fresh Python/Ruby reference rows: `/home/david/sync/projects/able/v12/docs/perf-baselines/2026-07-14-bytecode-generality-interpreter-refresh.json`
- Suite: `generality`
- Able benchmarks: `fib, binarytrees, matrixmultiply, quicksort, sudoku, sudoku_masks, i_before_e, base64, json, monte_carlo_pi, pidigits, mandelbrot, reverse_complement, k_nucleotide, nbody, tapelang_alphabet`
- Able modes: `bytecode`
- Reference languages: `python, ruby`
- CPU affinity: `15`

| Benchmark | Mode | Able Status | Validation | Stdout SHA-256 | Able Real (s) | python Real (s) | Able/python | ruby Real (s) | Able/ruby |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `fib` | `bytecode` | ok (1) | verified (1) | 231e154363ef69683f6dca862f5a448e786ccc90485636bc20a86e2e470a2188 | 0.1800 | n/a | n/a | n/a | n/a |
| `binarytrees` | `bytecode` | timeout (1) | not run | n/a | n/a | n/a | n/a | n/a | n/a |
| `matrixmultiply` | `bytecode` | ok (1) | verified (1) | bce2053ea5fdc1351bf8d28398b346b894f8cf35c0e047effdca9dbb47e81d0c | 5.9800 | n/a | n/a | n/a | n/a |
| `quicksort` | `bytecode` | timeout (1) | not run | n/a | n/a | 25.8520 | n/a | 15.4697 | n/a |
| `sudoku` | `bytecode` | timeout (1) | not run | n/a | n/a | 3.2666 | n/a | 6.7622 | n/a |
| `sudoku_masks` | `bytecode` | timeout (1) | not run | n/a | n/a | 19.3157 | n/a | 23.0681 | n/a |
| `i_before_e` | `bytecode` | ok (1) | verified (1) | 981f0d37a277be25f359c097e10df2ef68009fea2d3e322aeed3a65d0fbaca39 | 0.6200 | 0.0923 | 6.72x | 0.1216 | 5.10x |
| `base64` | `bytecode` | ok (1) | verified (1) | 5f4c00cd811078942fc98cd5dbca3b47fac2f8d8210f07ea116c1a7c0d6ac316 | 3.1400 | 4.1055 | 0.76x | 2.5417 | 1.24x |
| `json` | `bytecode` | ok (1) | verified (1) | 7cbf9736119f7683d59646bea566591a2ef278db54286b8e2525aa39e907839e | 0.9200 | 2.7039 | 0.34x | 1.7460 | 0.53x |
| `monte_carlo_pi` | `bytecode` | ok (1) | verified (1) | dab6ddc0a42f6644dd620c574b8b552697477be92433e886b7116ec8723123d2 | 2.8800 | 1.6012 | 1.80x | 1.6940 | 1.70x |
| `pidigits` | `bytecode` | ok (1) | verified (1) | bdfa7b6c756d96492f472f97aee9cc139bee954d271eacedfd7ace5d2875f06c | 2.4200 | 4.2290 | 0.57x | 10.6187 | 0.23x |
| `mandelbrot` | `bytecode` | ok (1) | verified (1) | d1560cadc49ab9dca30a4cfbe4123c1fcc0240ed80de27ad37d7fcc97159173e | 7.1400 | 1.3266 | 5.38x | 2.0528 | 3.48x |
| `reverse_complement` | `bytecode` | ok (1) | verified (1) | db06a593d68950aa91293fb57ffed87ad57040c4ac7ab81ae9dc43d50b400bb7 | 6.9700 | 0.0288 | 242.01x | 0.0760 | 91.71x |
| `k_nucleotide` | `bytecode` | ok (1) | verified (1) | d628623daa677c00673df8d4961d14eb271b4112cdcb65b8233f07b69d7b49b8 | 44.5000 | 1.4370 | 30.97x | 1.3673 | 32.55x |
| `nbody` | `bytecode` | timeout (1) | not run | n/a | n/a | 2.2141 | n/a | 3.3427 | n/a |
| `tapelang_alphabet` | `bytecode` | timeout (1) | not run | n/a | n/a | n/a | n/a | n/a | n/a |
