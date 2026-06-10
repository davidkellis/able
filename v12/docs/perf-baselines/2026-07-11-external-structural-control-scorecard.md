# External Structural and Control-Flow Scorecard

This is a bounded scorecard tranche, not a runtime, compiler, or standard
library change. It adds recursive, backtracking, and partition-control
workloads to the cross-language performance evidence.

## Method

`v12/bench_compare_external` ran `fib`, `sudoku`, and `quicksort` in compiled
and bytecode modes, three times each, against the checked-in Go, Ruby, and
Python rows in `../benchmarks/results.json`. Able used the canonical stdlib at
`/home/david/sync/projects/able-stdlib/src`, CPU affinity `2`,
`GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, and a 45-second cap per process.

The retained comparison artifacts are
`v12/tmp/scorecard-structural-control/structural-control.{json,md}`. Explicit
benchmark selection retains the wrapper's `core` preset label; the artifact's
benchmark list defines the actual shard.

## Results

| Benchmark | Mode | Status | Able (s) | Relevant reference (s) | Able/reference |
| --- | --- | --- | ---: | ---: | ---: |
| Fib | compiled | 3/3 ok | 3.2900 | Go 2.8400 | 1.16x |
| Fib | bytecode | 3/3 ok | 0.1700 | Ruby 46.6400; Python 60.6700 | 0.00x; 0.00x |
| Sudoku | compiled | 3/3 ok | 0.2367 | Go 0.1300 | 1.82x |
| Sudoku | bytecode | 3/3 ok | 0.4933 | Ruby 5.6700; Python 3.0200 | 0.09x; 0.16x |
| QuickSort | compiled | 3/3 ok | 1.8933 | Go 2.0100 | 0.94x |
| QuickSort | bytecode | 3/3 timeout | n/a | Go/Ruby/Python available | n/a |

A ratio above approximately `1.053x` misses the stated 95%-of-reference-speed
floor. Fib and Sudoku miss the compiled-Go target; QuickSort clears it.
Bytecode Fib and Sudoku clear both Ruby/Python comparisons, while QuickSort's
three bounded timeouts are status evidence rather than a synthetic ratio.

## Decision

Keep no code. The shard exposes no common completed runtime target: the
bytecode rows split between two strong comparisons and one timeout, and the
compiled rows split between recursive Fib, backtracking Sudoku, and a
Go-competitive mutable partition program. No recursion-, Sudoku-, QuickSort-,
or partition-shape special case is authorized. No `able-stdlib` source
changed.

## Next recommendation

Collect bounded phase-separated compiled CPU/allocation profiles for Fib and
Sudoku, using compiled QuickSort as a control. The two completed compiled-Go
misses may share a general call/control or allocation boundary, but only a
repeated concrete generated helper can justify a compiler change. Preserve the
same pinned settings and deterministic output checks, and reject any candidate
that does not also keep the QuickSort guard healthy.
