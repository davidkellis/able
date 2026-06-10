# External Numeric-Kernel Scorecard

This is a bounded scorecard tranche, not a runtime, compiler, or standard
library change. It compares independently shaped numeric kernels before
selecting a performance candidate.

## Method

`v12/bench_compare_external` ran `matrixmultiply`, `monte_carlo_pi`, and
`mandelbrot` in compiled and bytecode modes, three times each, against the
checked-in Go, Ruby, and Python rows in `../benchmarks/results.json`. The Able
runs used the canonical stdlib at `/home/david/sync/projects/able-stdlib/src`,
CPU affinity `2`, `GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, and a
45-second per-run cap. Every requested mode completed all three runs without a
timeout or launch failure.

Artifacts are retained at
`v12/tmp/scorecard-numeric-kernels/numeric-kernels.{json,md}`. As with the
text/byte shard, the wrapper retains its `core` preset label after explicit
benchmark selection; the artifact's benchmark list defines the actual scope.

## Results

| Benchmark | Mode | Able (s) | Relevant reference (s) | Able/reference |
| --- | --- | ---: | ---: | ---: |
| MatrixMultiply | compiled | 1.1433 | Go 0.8800 | 1.30x |
| MatrixMultiply | bytecode | 4.1200 | Ruby 42.9300; Python 56.2900 | 0.10x; 0.07x |
| Monte Carlo Pi | compiled | 0.2333 | Go 0.1800 | 1.30x |
| Monte Carlo Pi | bytecode | 2.6800 | Ruby 1.4200; Python 1.6800 | 1.89x; 1.60x |
| Mandelbrot | compiled | 0.1533 | Go 0.0400 | 3.83x |
| Mandelbrot | bytecode | 6.8667 | Go 0.0400 | 171.67x |

A ratio above approximately `1.053x` misses the stated 95%-of-reference-speed
floor. MatrixMultiply and Monte Carlo Pi repeat a moderate compiled-versus-Go
miss despite substantially different source shapes (array/matrix operations
versus scalar pseudo-random sampling). MatrixMultiply clears the bytecode
Ruby/Python comparison, while Monte Carlo Pi does not. Mandelbrot has no
checked-in Ruby or Python reference row and is therefore only a compiled-Go
and profile guard in this tranche.

## Decision

Keep no code. Similar ratios alone are not a concrete shared implementation
wall: MatrixMultiply may exercise Array/struct transport while Monte Carlo Pi
may exercise scalar numeric lowering, and Mandelbrot's much larger float-loop
gap could be distinct. The bytecode comparison is also not repeated across two
reference-backed kernels. No array-, float-, random-, matrix-, or
benchmark-specific rule is authorized, and no `able-stdlib` source changed.

## Next recommendation

Collect bounded compiled CPU and allocation profiles for MatrixMultiply and
Monte Carlo Pi, with Mandelbrot as an unrelated float-control guard. The pair
shares a repeatable Go-relative miss; profiles are needed to determine whether
one compiler-generated helper or lowering boundary actually repeats. Use the
same pinned runtime settings and output checks, then accept a candidate only if
the same concrete generic descendant is material in both pair members and does
not regress the Mandelbrot guard.
