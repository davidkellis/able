# Compiled Numeric-Kernel Phase Profiles

This evidence tranche profiles the repeated compiled-Go scorecard miss without
changing compiler, runtime, or standard-library behavior.

## Method

Current-source compiled binaries were built once for MatrixMultiply, Monte
Carlo Pi, and Mandelbrot with the canonical stdlib at
`/home/david/sync/projects/able-stdlib/src`. Each clean process ran from its
external benchmark directory on CPU `2` with `GOMEMLIMIT=1GiB`, `GOGC=50`, and
`GOMAXPROCS=1`. `ABLE_GO_PHASE_PROFILE_DIR` separated bootstrap from the
registered `main()` phase; the analysis uses only merged `main.cpu.pprof`
captures.

MatrixMultiply contributed three main-phase launches (3.22 seconds of CPU
samples). Monte Carlo Pi and Mandelbrot are short processes, so each
contributed twelve launches (1.53 and 0.59 seconds of CPU samples). Every
process completed within 45 seconds. The stdout hash was identical across all
runs of each program (3/3, 12/12, and 12/12 respectively).

Artifacts, generated source, per-process phase profiles, and merged profiles
are retained under `v12/tmp/compiled-numeric-profiles/`.

## CPU evidence

| Workload | Main-phase CPU leaf evidence |
| --- | --- |
| MatrixMultiply | `__able_compiled_fn_matmul` is 91.9% flat (95.3% cumulative); annotated source places the work in the generated triple loop's typed Array reads, `f64` multiply/add, and loop increment. |
| Monte Carlo Pi | `__able_compiled_fn_approx_pi` is 66.7% flat; generic checked signed multiply is 17.0% and signed div/mod is 15.7%. The source is the scalar Park--Miller state update plus float point-in-circle test. |
| Mandelbrot guard | `__able_compiled_fn_pixel_byte` is 88.1% flat (93.2% cumulative); checked shift is 5.1%. The source is its own complex-number iteration and escape test. |

The phase allocation summaries are not used to select a leaf. They include
the intentional CPU-profiler initialization cost, and allocation-profile leaf
reports are correspondingly dominated by profile serialization. Their totals
also differ substantially: MatrixMultiply's main phase reports 35.4 MB and
two GCs per run, versus 2.45 MB/zero GCs for Monte Carlo Pi and 2.91 MB/zero
GCs for Mandelbrot.

## Decision

Keep no code. The scorecard's matching 1.30x compiled-Go ratios do not trace
to a common helper or lowering boundary. MatrixMultiply's typed Array/floating
triple loop, Monte Carlo Pi's overflow-safe scalar arithmetic, and
Mandelbrot's float pixel kernel are separate execution walls. In particular,
do not specialize Matrix, random sampling, Mandelbrot, f64 arrays, or one
numeric expression shape. No `able-stdlib` source changed.

## Next recommendation

Refresh the next bounded external structural/control shard for Fib, Sudoku,
and QuickSort in compiled and bytecode modes, retaining timeout rows as status
evidence. Those programs add recursion, backtracking, mutation, and partition
control flow to the scorecard. Use the same pinned multi-run method, then
profile only a concrete leaf that repeats across at least two completed,
independently shaped workloads with an unrelated guard.
