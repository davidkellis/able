# Current application coverage ledger (2026-07-12)

## Purpose and target

This is the durable, provenance-preserving join of the versioned 16-program
`generality` sweep and the six newly measured `coverage` applications. It is a
status ledger, not a new timing run: no ratio is inferred from a timeout,
parallel-only row, or a different source revision.

The compiled target is Able/Go no greater than `1.0526x` (Able is at least 95%
as fast as Go). The bytecode targets use the same bound against Python and
Ruby. A ratio below `1.00x` means Able finished faster. Each source report
states its own process counts, CPU pinning, versions, verifier, and timeout
policy; the 16-program rows come from the versioned compiled and bytecode
reports, while the six added application rows come from the coverage-lane
report. JSON is deliberately listed once from `generality`; its independent
coverage-control recheck had the same healthy direction but is not a second
application.

## Cross-runtime status

| Application | Compiled Able/Go | Bytecode Able/Python | Bytecode Able/Ruby | Status and provenance |
| --- | ---: | ---: | ---: | --- |
| Fib | 1.05x | n/a | n/a | compiled verified 3/3; bytecode verified, Python/Ruby cap-bound |
| BinaryTrees | n/a | n/a | n/a | compiled parallel status only; all bytecode/reference lanes cap-bound |
| MatrixMultiply | 1.17x | n/a | n/a | compiled verified 3/3; bytecode verified, Python/Ruby cap-bound |
| QuickSort | 0.74x | n/a | n/a | compiled verified 3/3; bytecode cap-bound |
| Sudoku | n/a | n/a | n/a | compiled and bytecode cap-bound |
| Sudoku Masks | 15.94x | n/a | n/a | compiled verified 3/3; bytecode cap-bound |
| I-Before-E | 2.33x | 6.90x | 5.32x | verified |
| Base64 | 1.00x | 0.75x | 1.19x | verified |
| JSON | 0.51x | 0.33x | 0.50x | verified; independently rechecked as coverage control |
| Monte Carlo Pi | 2.00x | 1.70x | 1.61x | verified; valid nondeterministic output |
| PiDigits | 1.40x | 0.53x | 0.21x | verified |
| Mandelbrot | 3.23x | 5.28x | 3.29x | verified |
| Reverse Complement | 6.99x | 243.36x | 84.65x | verified |
| K-Nucleotide | 80.52x | 30.45x | 31.15x | verified |
| N-body | 13.70x | n/a | n/a | compiled verified 3/3; bytecode cap-bound |
| TapeLang Alphabet | 2.24x | n/a | n/a | compiled verified 3/3; bytecode and references cap-bound |
| Fixed Width 128 | 1268.40x | 22.17x | 12.03x | coverage lane verified 3/3 |
| Rational Series | 156.64x | 38.45x | 29.62x | coverage lane verified 3/3 |
| Word Frequency | 46.79x | 74.13x | 27.96x | coverage lane verified 3/3 |
| Document Audit | 25.20x | 22.72x | 7.32x | coverage lane verified 3/3 |
| Lexical Rollup | 30.17x | 23.14x | 7.91x | coverage lane verified 3/3 |
| Channel Rollup | 302.56x | 11.28x | 8.52x | coverage lane verified 3/3 |

The 16-program compiled numbers are from
`2026-07-12-compiled-generalities-versioned.md`; its successful serial rows
have three fresh verifier-backed Able and Go processes, except the explicitly
recorded Sudoku timeout and BinaryTrees parallel status. The corresponding
bytecode ratios and cap statuses are from
`2026-07-12-bytecode-generalities-versioned.md`; each successful Able bytecode
row is one current, verifier-backed status run and each Python/Ruby reference
has three processes. The six coverage rows are three-process verified runs for
every Able and reference mode in
`2026-07-12-coverage-application-lanes.md`.

## What the ledger says

Of 20 rankable compiled/Go applications, four meet the 95% target in this
snapshot: Fib, QuickSort, Base64, and JSON. Of 14 rankable bytecode/Python
applications, three meet it (Base64, JSON, and PiDigits); two of 14 rankable
bytecode/Ruby applications meet it (JSON and PiDigits). These counts include
only completed comparison rows and make no claim about cap-bound rows.

This is substantial remaining work, not evidence for a local fast path. The
large misses do not share a concrete leaf after the source-aligned profile and
guard work already recorded in the companion reports:

- Fixed Width 128, Rational Series, and the remaining numerical applications
  split between checked multiword values, nominal ratio arithmetic, raw float
  work, and unrelated numeric kernels.
- Word Frequency and K-Nucleotide have map/name/conversion costs; Document
  Audit and Lexical Rollup have iterator, member, generator, and filesystem
  costs; their common dispatcher parents are not one reusable implementation
  boundary.
- Channel Rollup is scheduler/task work, while the recursive, array-heavy,
  byte-heavy, and call-heavy misses use different leaf paths. The healthy JSON
  and Base64/PiDigits controls rule out a blanket process, startup, or
  reference-runner explanation.

Accordingly, this ledger authorizes neither a benchmark-shaped optimization nor
a rule for any named nominal type or standard-library container. It also does
not justify repeating profiles whose current-source attribution is already
known. No compiler, VM, runtime, benchmark, or `able-stdlib` source changes
follow from this consolidation.

## Decision and next recommendation

Begin a constrained canonical runtime-value architecture design, not a code
prototype. Why: the ledger confirms performance gaps across numeric, text,
collection, call, iterator, and concurrent application families, while all
currently observed local helper candidates are either disjoint or have failed
broad guards. The primitive-value feasibility audit leaves a universal tagged
runtime cell as the only potentially broad representation direction; it also
correctly rejects the current `RawValue` as a partial substitute.

The next tranche should specify the proposed stable carrier's complete data
layout, primitive and reference ownership rules, dynamic dispatch and
reflection behavior, hashing/equality/type-match semantics, host/extern and
future hand-off, and exact tree-walker/bytecode/compiler migration boundaries.
It must turn every row of the existing proof matrix into a concrete
cross-runtime fixture plan, with a no-regression gate on this 22-application
ledger. Do not write a runtime prototype unless that design makes every
materialization boundary and semantic obligation explicit; an incomplete
design is a rejection rather than authorization for a partial optimization.
