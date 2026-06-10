# Versioned bytecode generalities baseline (2026-07-12)

## Method

This refresh measures the verifier-complete `generality` catalog under one
interpreter-reference protocol. Local Python 3.14.5 and Ruby 4.0.5 sources ran
three verifier-backed, CPU-2-pinned processes per row, stopping a row after its
first 45-second timeout. One current Able bytecode process ran under the same
CPU pin, timeout, canonical stdlib, and output verifier. A one-process Able
row is intentionally status evidence, not a variance claim.

MatrixMultiply and Tapelang Alphabet are included now that their sibling
benchmark verifiers exist. The external interpreters time out on both at their
canonical workloads, so those rows remain status-only instead of manufactured
ratios.

## Fresh status

| Benchmark | Bytecode | Python | Bytecode/Python | Ruby | Bytecode/Ruby | Status |
| --- | ---: | ---: | ---: | ---: | ---: | --- |
| Fib | 0.2200 s | timeout | n/a | timeout | n/a | bytecode verified; references cap-bound |
| BinaryTrees | timeout | timeout | n/a | timeout | n/a | cap-bound |
| MatrixMultiply | 4.3100 s | timeout | n/a | timeout | n/a | bytecode verified; references cap-bound |
| QuickSort | timeout | 23.9992 s | n/a | 14.9136 s | n/a | bytecode cap-bound |
| Sudoku | timeout | 3.0289 s | n/a | 6.3266 s | n/a | bytecode cap-bound |
| Sudoku Masks | timeout | 17.9513 s | n/a | 21.2810 s | n/a | bytecode cap-bound |
| I-Before-E | 0.6000 s | 0.0870 s | 6.90x | 0.1127 s | 5.32x | verified |
| Base64 | 2.9200 s | 3.8697 s | 0.75x | 2.4641 s | 1.19x | verified control |
| JSON | 0.8400 s | 2.5612 s | 0.33x | 1.6863 s | 0.50x | verified control |
| Monte Carlo Pi | 2.5300 s | 1.4872 s | 1.70x | 1.5700 s | 1.61x | verified nondeterministic output |
| PiDigits | 2.1600 s | 4.0535 s | 0.53x | 10.0814 s | 0.21x | verified control |
| Mandelbrot | 6.2900 s | 1.1902 s | 5.28x | 1.9137 s | 3.29x | verified |
| Reverse Complement | 6.2300 s | 0.0256 s | 243.36x | 0.0736 s | 84.65x | verified |
| K-Nucleotide | 39.3600 s | 1.2924 s | 30.45x | 1.2637 s | 31.15x | verified |
| N-body | timeout | 2.0105 s | n/a | 3.1960 s | n/a | bytecode cap-bound |
| Tapelang Alphabet | timeout | timeout | n/a | timeout | n/a | cap-bound |

The complete reference artifact records Python/Ruby version, verifier status,
attempt count, and stdout hashes; the paired Able artifact records the same
Able validation evidence. Both are intentionally temporary run artifacts,
removed after this report was written.

## Decision

Keep no VM, compiler, tree-walker, benchmark, or `able-stdlib` code. The
fresh complete status makes the performance deficit auditable, but it does not
create a new general bytecode candidate:

- The two verified numeric misses repeat the already-exhausted raw-float
  compare/store lane; its representation and quickening variants regressed
  broad guards.
- Reverse Complement is primitive-byte array reads/pushes, stack
  materialization, and boxed integers. K-Nucleotide is call-name/inline-return,
  raw-u64 binary, and map work. Their common dispatcher parents are not a
  concrete leaf, and the relevant generic array, raw-carrier, stack, map,
  call-name, and return variants have already failed broad benchmarks.
- The fresh timeout rows have already received bounded current-source profile
  attribution: BinaryTrees is recursive named-struct construction, QuickSort
  i32-array snapshots/boxing, N-body raw-float arithmetic materialization, and
  Tapelang's former stack-growth issue is repaired but its remaining Array/
  member compute path is distinct. Sudoku and Sudoku Masks are recursive
  mutable-array searches, not evidence for a solver-specific rule.

Do not turn a timeout into a ratio, repeat unchanged profiles, or introduce a
DNA, `Array u8`, HashMap, recursive-search, float-kernel, or benchmark-shaped
shortcut.

## Next recommendation

Take a language-wide primitive-value representation feasibility tranche before
another VM micro-optimization. The suite now repeatedly shows that local
boxing/raw-cell/call-return changes are either disjoint or regress broad
guards, while a semantically complete representation boundary could benefit
ordinary numeric, byte, map, call, and collection programs without naming an
application or nominal container. The work entails extending the existing raw
value boundary audit into a concrete cross-runtime proposal and fixture matrix
for all primitive widths, public `runtime.Value` behavior, calls/returns,
closures, arrays/maps, structs/interfaces/unions, native/extern calls, and
tree-walker/bytecode/compiler parity. Prototype work is authorized only after
that matrix demonstrates an unambiguous language-level carrier; otherwise the
audit must reject the design before runtime code changes.
