# External Benchmark Scoreboard

- Generated: `2026-05-28T05:28:20Z`
- External results: `../benchmarks/results.json`
- External results generated: `2026-02-27T00:26:22Z`
- Able measurement source: kept measurements recorded in `LOG.md`, `PLAN.md`,
  and `v12/docs/performance-benchmarks.md` through 2026-05-28.

This is the checked-in current scoreboard for implemented external benchmark
families. Ratios below `1.00x` mean Able is faster than that reference row.
Tree-walker rows are retained only for semantic and historical context; new
performance work targets compiled and bytecode.

| Benchmark | Mode | Able status | Able real (s) | Go real (s) | Able/Go | Ruby real (s) | Able/Ruby | Python real (s) | Able/Python |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `fib` | `compiled` | ok (5) | 2.9940 | 2.8400 | 1.05x | 46.6400 | 0.06x | 60.6700 | 0.05x |
| `fib` | `bytecode` | ok (3) | 3.7633 | 2.8400 | 1.33x | 46.6400 | 0.08x | 60.6700 | 0.06x |
| `fib` | `treewalker` | timeout | timeout | 2.8400 | n/a | 46.6400 | n/a | 60.6700 | n/a |
| `binarytrees` | `compiled` | ok (3) | 3.6400 | 3.8300 | 0.95x | 20.3900 | 0.18x | 12.2500 | 0.30x |
| `binarytrees` | `bytecode` | timeout | timeout | 3.8300 | n/a | 20.3900 | n/a | 12.2500 | n/a |
| `binarytrees` | `treewalker` | timeout | timeout | 3.8300 | n/a | 20.3900 | n/a | 12.2500 | n/a |
| `matrixmultiply` | `compiled` | ok (5) | 0.9660 | 0.8800 | 1.10x | 42.9300 | 0.02x | 56.2900 | 0.02x |
| `matrixmultiply` | `bytecode` | timeout | timeout | 0.8800 | n/a | 42.9300 | n/a | 56.2900 | n/a |
| `matrixmultiply` | `treewalker` | timeout | timeout | 0.8800 | n/a | 42.9300 | n/a | 56.2900 | n/a |
| `quicksort` | `compiled` | ok (3) | 1.7500 | 2.0100 | 0.87x | 14.5800 | 0.12x | 20.3200 | 0.09x |
| `quicksort` | `bytecode` | timeout | timeout | 2.0100 | n/a | 14.5800 | n/a | 20.3200 | n/a |
| `quicksort` | `treewalker` | timeout | timeout | 2.0100 | n/a | 14.5800 | n/a | 20.3200 | n/a |
| `sudoku` | `compiled` | ok (5) | 0.0600 | 0.1300 | 0.46x | 5.6700 | 0.01x | 3.0200 | 0.02x |
| `sudoku` | `bytecode` | ok (5) | 0.3360 | 0.1300 | 2.58x | 5.6700 | 0.06x | 3.0200 | 0.11x |
| `sudoku` | `treewalker` | ok | 6.7100 | 0.1300 | 51.62x | 5.6700 | 1.18x | 3.0200 | 2.22x |
| `i_before_e` | `compiled` | ok (5) | 0.0620 | 0.0500 | 1.24x | 0.1000 | 0.62x | 0.1300 | 0.48x |
| `i_before_e` | `bytecode` | ok (5) | 0.4480 | 0.0500 | 8.96x | 0.1000 | 4.48x | 0.1300 | 3.45x |
| `i_before_e` | `treewalker` | ok | 3.5400 | 0.0500 | 70.80x | 0.1000 | 35.40x | 0.1300 | 27.23x |
| `base64` | `compiled` | ok (1) | 2.8600 | 2.2000 | 1.30x | 2.2100 | 1.29x | 3.3100 | 0.86x |
| `base64` | `bytecode` | ok (1) | 8.4400 | 2.2000 | 3.84x | 2.2100 | 3.82x | 3.3100 | 2.55x |
| `base64` | `treewalker` | ok (1) | 26.8400 | 2.2000 | 12.20x | 2.2100 | 12.14x | 3.3100 | 8.11x |
| `json` | `compiled` | ok (3) | 0.6700 | 1.3600 | 0.49x | 1.5600 | 0.43x | 2.8700 | 0.23x |
| `json` | `bytecode` | ok (3) | 0.7233 | 1.3600 | 0.53x | 1.5600 | 0.46x | 2.8700 | 0.25x |
| `monte_carlo_pi` | `compiled` | ok (3) | 0.3233 | 0.1800 | 1.80x | 1.4200 | 0.23x | 1.6800 | 0.19x |
| `monte_carlo_pi` | `bytecode` | ok (3) | 18.7967 | 0.1800 | 104.43x | 1.4200 | 13.24x | 1.6800 | 11.19x |
| `pidigits` | `compiled` | ok (3) | 1.3367 | 0.7400 | 1.81x | 9.1800 | 0.15x | n/a | n/a |
| `pidigits` | `bytecode` | ok (3) | 2.0300 | 0.7400 | 2.74x | 9.1800 | 0.22x | n/a | n/a |
