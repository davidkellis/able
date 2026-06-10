# Normalized compiled generated-owner closure

Date: 2026-07-26

## Decision

Retain no compiler, generated-runtime, runtime, interpreter, bytecode VM,
canonical-stdlib, benchmark, reference, language, dependency, or WASM change
from this tranche.

Fresh strict generated source, assembly, three-process CPU profiles, and
three-process allocation evidence for normalized Tapelang Alphabet, normalized
Sudoku Masks, and Fib have no removable exact owner in common. All three
programs already stay in generated native Go for their measured work. The
remaining costs are:

- Fib's unreachable fallible-control result and checks;
- Tapelang's required checked `i32` mutations around an otherwise native flat
  dispatcher; and
- Sudoku's required division, multiplication, shift, and Array safety paths
  for functions whose unrestricted arguments can fail.

The Go backend already inlines Tapelang's universally total `Tape.get` and
`Tape.ensure` helpers and removes their nil-control results and checks. Their
generated-source control plumbing is therefore not a surviving machine-code
owner. A broad total-function ABI would remain material only in Fib, while
removing the surviving Tapelang or Sudoku checks would weaken Able semantics
without a stronger relational proof. The three-unlike-program admission gate
fails before implementation or A/B timing.

The compact machine-readable companion is
`2026-07-26-normalized-compiled-generated-owner-closure.json`.

## Frozen contract

The compiler SHA-256 is
`28697a5adf4f73918f3d83fbcddc211407dc7e539240f64e4127a4e3dd4ddcab`.
Every application was freshly emitted with `--no-fallbacks`. Each final graph
contains 96 packages and zero `pkg/interpreter` matches.

| Application | Able binary SHA-256 | Go binary SHA-256 |
| --- | --- | --- |
| Tapelang Alphabet | `4dfe122464875b339505e21b4afbda8fd3473d0ed0ab9871fc1bdc92596a6bfa` | `22f415a37b96c17880be702f3b6fbb2dcec7d8b17be718c288e9d9ca7cdf3475` |
| Sudoku Masks | `d13957f513bd6d59fae0147216b7a2e1c2904dc39b54e69c3c6fff85126e5090` | `fc3eb75dfb6c2b7ee57fb6ac5f2261868fd08df7ec80c97a41689a3c3ffab6b1` |
| Fib | `4f6f8014c6ff1907185e5587dcfead900cb8ead9cfd97b5d76ce03c4f56c5bbd` | `ccf44a10b99c79d12731e508ce9026764f49791710c08f0ddca6ee0cf48e479d` |

All fresh normal-binary smoke runs and all Able profile runs passed the public
verifiers. The Go profile harnesses copied the reference sources byte-for-byte,
ran one complete application workload per process, and asserted their results.
All work used Go 1.26.4, one logical CPU, disk-backed `/var/tmp`, and the
catalog working directories and inputs.

The immediately preceding five-run normalized cohort remains the governing
wall-time evidence: Tapelang is 3.7060 seconds versus Go at 2.9649 seconds
(1.2500x); Sudoku is 1.5740 seconds versus 0.7027 seconds (2.2399x). Fib
remains the equal-algorithm effect sentinel.

## CPU profiles

Three independently launched measured-main Able profiles and three
independently launched one-workload Go profiles were merged per application.

| Application | Able sampled CPU | Dominant Able generated owners | Go sampled CPU | Dominant Go owners |
| --- | ---: | --- | ---: | --- |
| Tapelang | 10.75 s | `execute` 62.14%; `Tape.inc` 27.16%; inlined `Tape.get` 8.19%; `Tape.move` 1.67% | 8.71 s | `execute` 61.88%; inlined `get` 16.42%; inlined `set` 14.24%; inlined `inc` 6.31% |
| Sudoku | 3.93 s | `find_best_empty` 35.37%; signed divmod 15.52%; `bit_count` 15.01%; checked multiply 12.47%; `square_index` 10.94% | 1.97 s | `findBestEmpty` 70.56%; inlined `bitCount` 19.29%; inlined `squareIndex` 5.58%; `solve` 4.57% |
| Fib | 10.69 s | generated `fib` 99.63% | 8.94 s | reference `fib` 100% |

No exact generated CPU symbol is material in all three applications. The only
three-way names are aggregate process/runtime parents. No bridge,
`runtime.Value`, wrapper, thunk, interpreter, or GC leaf owns material CPU in
the three generated profiles.

## Allocation evidence

The lightweight measured-main counter was repeated three times. Every Able
application returned the exact same byte/object/GC tuple in all three
processes. The Go values are one-workload `testing` allocation counters from
three independent processes.

| Application | Able main bytes / objects / GC | Go bytes / objects |
| --- | ---: | ---: |
| Tapelang | 6,160 / 48 / 0 | 4,456 / 13 |
| Sudoku | 618,616 / 15,011 / 0 | 61,293 mean / 1,026 mean |
| Fib | 144 / 6 / 0 | 0 / 0 |

The allocation shapes do not intersect: Tapelang allocates only its parsed
program/tape/output shape, Sudoku's remaining allocations are parsing and
output construction rather than its now-stack-resident search position, and
Fib has no application allocation. The exact allocation-snapshot lane was
also collected, but its start/end writer allocations intentionally make the
lightweight counters authoritative for totals.

## Generated assembly

Static instruction counts include calls and both conditional and unconditional
jumps.

| Function | Bytes | Instructions | Calls | Conditional jumps | Unconditional jumps |
| --- | ---: | ---: | ---: | ---: | ---: |
| Able Fib | 166 | 52 | 5 | 4 | 1 |
| Go Fib | 84 | 27 | 3 | 2 | 1 |
| Able Tapelang `execute` | 1,018 | 276 | 17 | 34 | 19 |
| Able `Tape.inc` | 295 | 82 | 4 | 10 | 6 |
| Able `Tape.move` | 248 | 66 | 4 | 5 | 3 |
| Go Tapelang `execute` | 1,465 | 321 | 19 | 38 | 20 |
| Able Sudoku `find_best_empty` | 784 | 198 | 10 | 13 | 5 |
| Able `bit_count` | 98 | 33 | 2 | 3 | 2 |
| Able `square_index` | 278 | 74 | 5 | 5 | 1 |
| Able `solve_with_masks` | 1,477 | 348 | 24 | 23 | 6 |
| Go Sudoku `findBestEmpty` | 346 | 98 | 5 | 10 | 6 |
| Go Sudoku `solve` | 958 | 186 | 12 | 14 | 4 |

Tapelang is the important reconciliation. Although generated source returns
`(value, *__ableControl)` from `Tape.get` and `Tape.ensure`, neither helper
survives as a callable symbol. `execute` contains no call to `Tape.get`, and
`Tape.move` contains no call to `Tape.ensure`; Go inlined their bodies and
constant-folded their nil controls. The surviving hot calls from `execute` are
`Tape.inc` and `Tape.move`, whose additions can overflow for valid arbitrary
Able inputs.

## Dynamic calls and checks

The normalized exact work counts expose where source-level controls survive:

- Fib executes 2,269,806,338 recursive child-result checks. Its generated body
  has no control-producing path, so all are unreachable ABI work.
- Tapelang executes 577,783,636 increments and 57,780,404 moves. Those
  635,564,040 surviving call-result checks guard mutations that can overflow.
  Its 577,783,654 hot `Tape.get` results and the 57,780,404 internal
  `Tape.ensure` results carry no machine-level control check after Go inlining.
- Sudoku executes 64,090,010 calls each to `square_index` and `bit_count`
  inside `find_best_empty`, 1,918,450 `find_best_empty` calls, 1,918,350
  nonterminal `square_index` calls, and 1,918,350 recursive solve calls. These
  contribute 133,935,170 checked static results before the smaller
  initialization and shift paths. None is universally total: negative
  `bit_count` input can overflow at minimum `i32`; arbitrary square-index
  coordinates can overflow; and Array operations can raise on invalid
  indices.

The fact that the benchmark's current values stay within safe ranges is not a
compiler proof. Eliding these controls from unrestricted functions would turn
valid Able errors into Go wraparound or unchecked indexing behavior.

## Candidate gate

| Candidate | Material reach | Disposition |
| --- | ---: | --- |
| universally total single-result ABI | Fib only | no three-program reach; Tapelang total helpers are already inlined away |
| remove successful-run control checks | all in generated source | unsafe; Tapelang/Sudoku controls remain reachable for unrestricted inputs |
| encourage Go inlining | Tapelang/Sudoku symptom | no semantic compiler rule; fallible cold paths are why the bodies remain callable |
| primitive constant division/multiply lowering | Sudoku only | previously closed single-family arithmetic route |
| general Array bounds elision | Tapelang/Sudoku only | no shared proof, no Fib reach, and invalid indices must still raise |
| allocation reduction | none shared | application parsing/output shapes differ; Fib has no application allocation |

No candidate was prototyped because none passes the required material
three-unlike-program reach. Consequently the five-or-more balanced A/B/Go gate
was not entered.

## Verification

- All three fresh strict builds succeeded with `--no-fallbacks`.
- All three final dependency graphs contain 96 packages and omit
  `pkg/interpreter`.
- All normal-binary smoke runs and 21 Able profile/counter processes passed
  their public verifiers.
- All nine Go profile/counter processes completed their full workload and
  result assertions.
- Generated source, symbols, disassembly, CPU profiles, and allocation
  counters agree on the no-shared-owner decision.
- The complete `./run_all_tests.sh` handoff passed every coverage, scorecard,
  selection, and threshold contract; every non-compiler package; all 32
  compiler batches; and the final bytecode fixture corpus in 85.495 seconds.
  Existing large aggregates included `pkg/interpreter` at 94.495 seconds and
  compiler batches 19, 28, and 29 at 179.320, 73.230, and 84.073 seconds.
  This tranche adds no test and does not change that sharding debt.

## Next recommendation

Run a report-only relational safety-proof feasibility census over the current
strict compiled corpus before implementing another lowering optimization.

Why: normalized Tapelang and Sudoku confirm that the remaining native gap is
mostly required overflow, Euclidean arithmetic, shift, and bounds behavior,
not interpreter crossing or boxed carriers. Matching Go requires proving those
operations safe, not deleting their semantics.

What it entails: extend the existing control census only as an observer so it
can track general static-Array lengths, index intervals, struct-field
intervals, and bitmask domains through direct calls and loops; weight every
newly dischargeable check by current profiles/dynamic counts; and require
material reach in at least three unlike applications before any production
compiler experiment. The analysis must be structural and type-driven, never
selected by benchmark, function, named container, or non-primitive nominal
type.

Why it is important: a successful proof can let generated code use the same
unchecked native Go operations when safety follows from program invariants,
while preserving Able errors everywhere else. A failed breadth/target-budget
census supplies an explicit stopping condition and prevents another
single-family optimization. Do not begin WASM work.
