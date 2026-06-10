# Versioned compiled generalities baseline (2026-07-12)

## Method

This is the first single-provenance compiled-Able-versus-Go 1.26.4 sweep after
the MatrixMultiply and TapeLang verifier closure. Fresh Go references and
compiled Able used CPU 2, `GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, and a
45-second per-process limit. The reference runner builds Go binaries outside
timing; every successful result runs its suite `verify.rb`.

Fifteen serial lanes received three fresh Go runs. Fourteen of those completed
three compiled Able runs, for 42 verifier-backed compiled executions. Sudoku
has one separately captured formal 45-second timeout row. BinaryTrees is
explicitly excluded from serial ratio selection: two fresh one-core Go
processes verified, then its third was interrupted rather than spend further
cap-bound work on the benchmark's explicitly parallel goroutine lane.

| Benchmark | Fresh Go (s) | Compiled Able (s) | Able/Go | Status |
| --- | ---: | ---: | ---: | --- |
| Fib | 3.1322 | 3.2800 | 1.05x | verified 3/3 |
| BinaryTrees | n/a | n/a | n/a | parallel one-core status; 2 Go outputs verified, not ranked |
| MatrixMultiply | 0.9741 | 1.1400 | 1.17x | verified 3/3 |
| QuickSort | 2.5395 | 1.8733 | 0.74x | verified 3/3 |
| Sudoku | 0.1438 | n/a | n/a | compiled timeout 1/1 at 45s |
| Sudoku Masks | 0.5761 | 9.1833 | 15.94x | verified 3/3 |
| I-Before-E | 0.0644 | 0.1500 | 2.33x | verified 3/3 |
| Base64 | 2.4634 | 2.4733 | 1.00x | verified 3/3 |
| JSON | 1.4644 | 0.7500 | 0.51x | verified 3/3 |
| Monte Carlo Pi | 0.2170 | 0.4333 | 2.00x | verified 3/3; Go outputs valid but nondeterministic |
| PiDigits | 1.1940 | 1.6700 | 1.40x | verified 3/3 |
| Mandelbrot | 0.0496 | 0.1600 | 3.23x | verified 3/3 |
| Reverse Complement | 0.0162 | 0.1133 | 6.99x | verified 3/3 |
| K-Nucleotide | 0.0556 | 4.4767 | 80.52x | verified 3/3 |
| N-body | 0.0326 | 0.4467 | 13.70x | verified 3/3 |
| TapeLang Alphabet | 1.8321 | 4.1033 | 2.24x | verified 3/3 |

The report deliberately retains timeout and parallel-lane status instead of
inventing ratios. Its temporary raw JSON, generated sources, binaries, and
captured outputs are removed after this record; the table is the durable
versioned baseline.

## Selection decision

Keep no compiler, VM, runtime, `able-stdlib`, or benchmark-algorithm change.
The complete source-aligned report strengthens the measurement surface but does
not create a repeated concrete helper:

- MatrixMultiply's typed f64 Array triple loop, Monte Carlo's checked PRNG
  recurrence, Mandelbrot's pixel iteration, and N-body's direct dynamics are
  distinct numeric paths in the existing compiled profiles.
- K-Nucleotide's generic conversion/allocation wall was independently paired
  with Word Frequency and rejected as a named map/string context without a
  semantics-complete value representation.
- Reverse Complement's primitive-byte path, I-Before-E's line/string search,
  Sudoku Masks' recursive search, and TapeLang's program-defined interpreter
  have no matching non-nominal descendant.

The 95%-of-Go target is met by QuickSort, Base64, and JSON in this snapshot;
Fib is just outside it. The remaining misses are real priorities, but a ratio
or broad category is not authorization for an application or named-type
special case.

## Next recommendation

Refresh the bytecode interpreter's Python/Ruby generality status using the now
verifier-complete benchmark suite, with the same timeout/status discipline.
Why: the compiler baseline has no eligible shared helper, while the bytecode
target is independent and MatrixMultiply/TapeLang can now contribute trusted
validation. The work entails fresh pinned Python/Ruby references and bounded
bytecode runs, followed by profiles only if two material interpreter misses
share one concrete VM/runtime helper that remains neutral on broad controls.
