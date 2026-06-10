# Bytecode primitive-integer region census (2026-07-18)

## Decision

Keep no compiler, VM, runtime, stdlib, benchmark, or fixture change. Statically
proven multi-operation primitive-integer expression regions are highly material
in Monte Carlo Pi and in a reduced, validated Sudoku Masks run, but not in a
third unlike application. The predeclared three-application breadth gate fails,
so an integer-region opcode was not built.

All temporary planner metadata, instruction annotations, counters, environment
switches, snapshot fields, and tests were removed before final verification.

## Admission rule

The temporary planner admitted only trees composed of exact local primitive
integer slots and integer literals. It covered checked `+`, `-`, `*`, and `^`,
Euclidean `//` and `%`, and dotted bitwise/shift operations. It rejected calls,
casts, members, indexing, dynamic values, `i128`/`u128`, trees with fewer than
two operations, and trees deeper than the retained float-region scratch bound.

Before collecting results, a workload was defined as dynamically material only
when the admitted trees:

- executed at least 10,000 times in main;
- represented at least 1% of main's bytecode instructions when weighted by
  their integer-operation count; and
- repeated in at least three unlike applications, with at least two primitive
  integer result kinds across those applications.

The observer counted the instruction already emitted for an expression; it did
not alter execution. Production recurrence kernels remained enabled, and the
observer did not claim regions bypassed by those kernels.

## Coverage run

One census-only CLI was reused for the exact 35-application external coverage
selection with `GOMAXPROCS=1`, `GOMEMLIMIT=1GiB`, `GOGC=50`, canonical external
stdlib sources, main-only counters, and a 55-second process bound. Twenty-nine
applications completed and passed their public Ruby verifiers. Binary Trees,
QuickSort, Sudoku Masks, N-Body, TapeLang Alphabet, and Regex Suffix exceeded
the bound and remain explicit full-size exclusions.

The first pass used the existing comprehensive bytecode stats mode. A second
temporary census-only switch removed unrelated atomic stats work and restored
production recurrence-kernel selection. This allowed Fib, K-Nucleotide, and
Regex Set to complete; the six remaining exclusions still exceeded 55 seconds.

Only five completed applications executed any admitted region:

| Application | Main instructions | Region executions | Integer operations in regions | Share | Kind / operations |
| --- | ---: | ---: | ---: | ---: | --- |
| Monte Carlo Pi | 106,341,149 | 11,111,000 | 22,222,000 | 20.8969% | `i64`, 2 |
| K-Nucleotide | 750,775,159 | 46 | 92 | <0.0001% | `u64`, 2 |
| Await Channel Mux | 365,370 | 512 | 1,536 | 0.4204% | `i64`, 3 |
| Mutex Ledger | 3,080,369 | 8,192 | 24,576 | 0.7978% | `i64`, 3 |
| Mutex Await Journal | 968,879 | 2,048 | 6,144 | 0.6341% | `i64`, 3 |

Distance Field and RMS Norm each lowered one eligible `u32` tree but executed
it zero times on the production path. The other 22 completed applications had
zero eligible executions. Thus only Monte Carlo Pi passed both dynamic
thresholds in the full external run.

## Exclusion closure

The existing `binarytrees_small` benchmark uses the same candidate expression
as full Binary Trees and completed 16,549,714 instructions with zero eligible
executions. QuickSort's nested integer expressions contain casts or ordinary
division; N-Body's multi-operation trees are floating point; and TapeLang and
Regex Suffix have no local-only multi-operation integer tree admitted by the
planner.

For Sudoku Masks, the exact external source was run against one canonical
puzzle while retaining its ten repetitions. All ten emitted boards were
independently checked for givens, rows, columns, and 3x3 squares. The verified
run executed 1,795,457 instructions and 27,430 two-operation `i32` regions,
representing 54,860 integer operations or 3.0555% of instructions. Sudoku is
therefore a second material application and supplies a second integer kind,
but there is still no third application.

## Interpretation

The representation is technically feasible, and it could remove substantial
dispatch/materialization work from Monte Carlo Pi and Sudoku. Current suite
evidence says that benefit is concentrated in two algorithm shapes, while the
same general mechanism is absent or immaterial in text, map, iterator,
concurrency, wide-integer, float, and allocation-heavy applications. Building
the region now would violate the project's broad-applicability rule.

## Next recommendation

Measure repeated straight-line loop bodies as whole typed basic blocks rather
than searching for another expression or adjacent-opcode micro-pattern. Admit a
prototype only when one side-effect-safe block class dominates at least three
unlike applications and spans multiple workload families. This is the next
larger structural batching boundary: it could amortize dispatch across loads,
primitive operations, branches, and stores without inventing per-benchmark
opcodes. The tranche should use a lightweight main-only observer, classify
backedge-delimited blocks by proof and side-effect requirements, then build one
cold-fallback executor only if the cross-suite census passes before timing.
