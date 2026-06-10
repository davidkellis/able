# Compiled generality status ledger (2026-07-12)

## Method

This ledger combines the fresh rows collected in this tranche with the
immediately preceding current-checkout scorecards. This tranche makes no
compiler, VM, stdlib, or benchmark source change. Every ratio is only
interpreted within its own CPU-2-pinned Go/Able measurement lane
(`GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`); raw seconds are not compared
across applications.

New three-run verified rows in this tranche are Fib (`3.2800s` Able versus
`3.1338s` Go) and N-body (`0.4600s` versus `0.0327s`). An attempted Go
refresh also established that MatrixMultiply and TapeLang currently have Go
sources but no suite `verify.rb`; they are recorded as coverage gaps rather
than assigned unverified timing ratios.

## Current status

| Feature family / benchmark | Compiled Able / fresh Go | Status | Selection role |
| --- | ---: | --- | --- |
| recursion: Fib | 3.2800s / 3.1338s (1.05x) | verified 3/3 | narrow miss |
| parallel allocation: BinaryTrees | 31.2600s / 34.3861s (0.91x) | prior verified 3/3, one-core goroutine lane | healthy control |
| dense f64 arrays: MatrixMultiply | n/a | no semantic verifier | coverage gap |
| mutation/recursion: QuickSort | 2.0420s / 2.8749s (0.71x) | verified 10/10 | healthy control |
| scan-based backtracking: Sudoku | >45s / 0.1600s | Able timeout; Go verified 10/10 | bounded status |
| mask search: Sudoku Masks | 9.5400s / 0.6652s (14.34x) | prior verifier-backed bounded row | isolated recursive-array miss |
| file/string search: I-Before-E | 0.1267s / 0.0664s (1.91x) | verified 3/3 | text-process miss |
| host byte/MD5: Base64 | 2.8590s / 2.9549s (0.97x) | verified 10/10 | healthy control |
| host parsing: JSON | 0.7600s / 1.4582s (0.52x) | verified 3/3 | healthy control |
| scalar PRNG/float: Monte Carlo Pi | 0.2600s / 0.2389s (1.09x) | verified 10/10 | narrow miss |
| arbitrary precision: PiDigits | 1.5230s / 1.3861s (1.10x) | verified 10/10 | narrow miss |
| escape-time float: Mandelbrot | 0.1900s / 0.0494s (3.85x) | verified 3/3 | numeric miss |
| primitive byte transform: Reverse Complement | 0.1367s / 0.0164s (8.34x) | verified 3/3 | file/byte miss |
| text/n-gram counting: K-Nucleotide | 5.8033s / 0.0586s (99.03x) | verified 3/3 | conversion/map miss |
| five-body dynamics: N-body | 0.4600s / 0.0327s (14.07x) | verified 3/3 | numeric miss |
| program-defined tape: TapeLang Alphabet | n/a | no semantic verifier | coverage gap |

The supporting fresh scorecards are cited in the plan entry for this tranche.
The cached one-core BinaryTrees row remains a valid status/control lane, not a
variance comparison with a parallel default reference. Sudoku's timeout is
also status, not a ratio.

## Verifier-closure update

The subsequent verifier-closure tranche resolves the two coverage gaps without
changing any benchmark algorithm: MatrixMultiply is now a verified `1.23x`
compiled-Able/Go row and TapeLang Alphabet is a verified `2.02x` row. The
Matrix catalog also now passes the Docker-standard `1000` input to the local
reference lane. See `2026-07-12-generality-verifier-closure.md` for the
scripts, preflights, and exact measurements.

## Selection decision

Keep no compiler, VM, runtime, `able-stdlib`, or benchmark-algorithm change.
The broad ledger adds N-body as a material verified miss, but it does not
create a generic pair: current evidence attributes N-body to direct f64
dynamics, square-root/absolute-value work, and bridge environment swaps, while
Mandelbrot is its own pixel/escape kernel. The existing numeric audit already
separates MatrixMultiply's typed-array triple loop, Monte Carlo's checked
integer recurrence, and Mandelbrot's float iteration. The repeated
Word-Frequency/K-Nucleotide conversion helper was separately rejected because
it occurs under named map/string boundaries and lacks a safe language-wide
value representation.

Accordingly, do not add a N-body, float, matrix, tape, HashMap, string-key,
FASTA, Sudoku, or source-shape lowering. The two missing verifiers are not
permission to use stored or unvalidated results for target decisions.

## Next recommendation

Add semantic verifiers for MatrixMultiply and TapeLang Alphabet in the sibling
benchmark repository, then run their fresh Go and compiled Able rows through
the normal reference harness. Why: these are the only generality coverage gaps
in the compiler ledger, and trusted output validation is a prerequisite for
using their numeric-array and user-defined-control results to select a broad
optimization. The work entails verifier scripts that accept the established
cross-language output contract, preflight checks against Go/Able reference
programs, catalog/harness confirmation, and a bounded fresh rerun; it must not
change either benchmark algorithm or introduce compiler special cases.
