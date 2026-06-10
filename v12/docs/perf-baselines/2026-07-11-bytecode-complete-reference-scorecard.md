# Complete external bytecode reference scorecard (2026-07-11)

## Method

This is the first pinned Python 3.14/Ruby 4.0 reference sweep across every
completed entry in the external `generality` catalog. It closes the runnable
source gaps with standard-library Python/Ruby N-body and standard-library
Python PiDigits. Python PiDigits intentionally uses built-in arbitrary-
precision integers instead of the separately named GMP variant, which is not
available in the shared local Python lane.

References used CPU `2`, a 45-second cap, and three verifier-backed processes.
Timeouts stop that row; missing verifiers stay explicitly unavailable. The
matching current Able bytecode sweep used the same pinning, cap, canonical
stdlib, and verifier, with one process per row. It is a status snapshot, not a
one-run optimization decision.

## Complete reference coverage

| Benchmark | Python 3.14 | Ruby 4.0 | Status |
| --- | ---: | ---: | --- |
| Fib | timeout | 44.8451 s (1/3) | cap-bound |
| BinaryTrees | timeout | timeout | cap-bound |
| MatrixMultiply | n/a | n/a | no verifier |
| QuickSort | 34.0081 s | 16.0195 s | verified 3/3 |
| Sudoku | 3.0792 s | 6.4205 s | verified 3/3 |
| I-Before-E | 0.0859 s | 0.1217 s | verified 3/3 |
| Base64 | 3.8752 s | 2.4864 s | verified 3/3 |
| JSON | 2.5752 s | 1.7783 s | verified 3/3 |
| Monte Carlo Pi | 1.6637 s | 1.7778 s | verifier-accepted nondeterministic 3/3 |
| PiDigits | 4.3131 s | 12.6597 s | verified 3/3 |
| Mandelbrot | 1.5763 s | 2.2010 s | verified 3/3 |
| Reverse Complement | 0.0274 s | 0.0854 s | exact output, verified 3/3 |
| K-Nucleotide | 1.4828 s | 1.6414 s | verified 3/3 |
| N-body | 2.6253 s | 3.5391 s | verified 3/3 |
| Tapelang Alphabet | n/a | n/a | no verifier |

## Able bytecode status

| Benchmark | Bytecode | Python ratio | Ruby ratio | Status |
| --- | ---: | ---: | ---: | --- |
| Fib | 0.2400 s | n/a | 0.01x | Ruby reference cap-bound; status only |
| BinaryTrees | timeout | n/a | n/a | status only |
| MatrixMultiply | 5.0100 s | n/a | n/a | no reference verifier |
| QuickSort | timeout | n/a | n/a | status only |
| Sudoku | timeout | n/a | n/a | status only |
| I-Before-E | 0.7300 s | 8.50x | 6.00x | verified miss |
| Base64 | 3.4200 s | 0.88x | 1.38x | Python-near control; Ruby miss |
| JSON | 1.0200 s | 0.40x | 0.57x | faster than both controls |
| Monte Carlo Pi | 2.8300 s | 1.70x | 1.59x | verified float-lane miss |
| PiDigits | 2.6200 s | 0.61x | 0.21x | faster than both controls |
| Mandelbrot | 7.4500 s | 4.73x | 3.38x | verified float-lane miss |
| Reverse Complement | 7.7200 s | 281.75x | 90.40x | verified byte-array miss |
| K-Nucleotide | timeout | n/a | n/a | status only |
| N-body | timeout | n/a | n/a | status only |
| Tapelang Alphabet | timeout | n/a | n/a | no reference verifier |

## Decision

Keep no VM, compiler, tree-walker, or `able-stdlib` code. The fresh complete
coverage does not reveal a new shared bytecode leaf. I-Before-E is member/slot
dispatch; Mandelbrot and Monte Carlo repeat the already-rejected float lane;
Reverse Complement is direct `Array u8` read/push, stack materialization, and
boxing; Base64 is host codec/MD5 work. JSON and PiDigits are controls. Do not
treat a timeout, missing verifier, or one-process status row as a ratio.

## Next recommendation

Return to the compiler target with an initialization-attribution pair:
repeated verified process launches for I-Before-E and Reverse Complement, with
JSON as a control, separating generated bootstrap/registration work from
`main()`. The complete bytecode scorecard has no new shared VM leaf, while the
compiler still has material Go gaps on short binaries whose normal CPU profiles
cannot sample the body reliably. Retain a change only when the same concrete
bootstrap helper is material in both misses and neutral on JSON.
