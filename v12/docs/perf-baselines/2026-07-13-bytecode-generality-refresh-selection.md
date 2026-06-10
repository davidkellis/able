# Bytecode Generality Refresh Selection — 2026-07-13

## Decision

Keep no VM, compiler, runtime, canonical-stdlib, or benchmark-source change.
Fresh Python/Ruby references identify current bytecode product gaps, but do not
re-open the raw-float, frame, call-name, inline-return, or member-cache paths
that already failed broad guards. The prior Base64, Monte Carlo Pi, and
I-Before-E profiles are host-codec, rejected float-slot, and text-dispatch
costs respectively; none is a shared explanation for the new ranking.

## Method and Selection

All processes used CPU 15, `GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, a
45-second cap, and each benchmark's Ruby verifier. Fresh Python 3.14 and Ruby
4.0 references were rebuilt outside timing, followed by one current Able
bytecode process per application. Twelve applications have both verified
references; unavailable reference rows and Able timeouts remain unranked.

| Rankable benchmark | Able / Python | Able / Ruby | Selection reading |
| --- | ---: | ---: | --- |
| K-Nucleotide | 33.69x | 33.72x | largest completed text/counting miss |
| Reverse Complement | 239.64x | 86.14x | largest byte/file transformation miss |
| I-Before-E | 7.89x | 5.47x | existing text-dispatch profile control |
| Mandelbrot | 5.45x | 3.60x | float escape-loop miss |
| Monte Carlo Pi | 1.81x | 1.65x | existing rejected float-slot control |
| Base64 | 0.80x | 1.25x | host-codec control |
| JSON / PiDigits | below both references | below both references | clear controls |

Fib, BinaryTrees, MatrixMultiply, and TapeLang lack fresh reference rows;
QuickSort, Sudoku, Sudoku Masks, N-body, and TapeLang bytecode runs timed out.
The raw reports are cleanup-eligible under
`v12/tmp/perf/2026-07-13-bytecode-generality-refresh/`.

## Next Recommendation

Collect bounded one-process bytecode CPU profiles for K-Nucleotide and Reverse
Complement, then compare their concrete descendants with the retained
I-Before-E text profile. These are independently authored, rankable severe
misses with text/byte and counting/map shapes; a source candidate is permitted
only if the same concrete VM/runtime leaf repeats across them. If their leaves
diverge, keep no change rather than optimizing a benchmark family.
