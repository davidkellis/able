# External Benchmark Scoreboard

- Generated: `2026-05-02T23:26:24Z`
- External results: `../benchmarks/results.json`
- External results generated: `2026-02-27T00:26:22Z`
- Able measurement source: kept measurements recorded in `LOG.md`, `PLAN.md`,
  and `v12/docs/performance-benchmarks.md` through 2026-05-02.

This is the checked-in current scoreboard for implemented external benchmark
families. Ratios below `1.00x` mean Able is faster than that reference row.

| Benchmark | Mode | Able status | Able real (s) | Go real (s) | Able/Go | Ruby real (s) | Able/Ruby | Python real (s) | Able/Python |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `fib` | `compiled` | ok (5) | 2.9940 | 2.8400 | 1.05x | 46.6400 | 0.06x | 60.6700 | 0.05x |
| `fib` | `bytecode` | ok | 67.8200 | 2.8400 | 23.88x | 46.6400 | 1.45x | 60.6700 | 1.12x |
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
| `sudoku` | `bytecode` | ok (5) | 0.5640 | 0.1300 | 4.34x | 5.6700 | 0.10x | 3.0200 | 0.19x |
| `sudoku` | `treewalker` | ok | 6.7100 | 0.1300 | 51.62x | 5.6700 | 1.18x | 3.0200 | 2.22x |
| `i_before_e` | `compiled` | ok (5) | 0.0620 | 0.0500 | 1.24x | 0.1000 | 0.62x | 0.1300 | 0.48x |
| `i_before_e` | `bytecode` | ok (5) | 0.4480 | 0.0500 | 8.96x | 0.1000 | 4.48x | 0.1300 | 3.45x |
| `i_before_e` | `treewalker` | ok | 3.5400 | 0.0500 | 70.80x | 0.1000 | 35.40x | 0.1300 | 27.23x |
