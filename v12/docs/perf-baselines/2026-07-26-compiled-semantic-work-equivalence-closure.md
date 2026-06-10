# Compiled semantic-work equivalence closure

Date: 2026-07-26

## Decision

Retain no compiler, generated-runtime, runtime, interpreter, bytecode VM,
canonical-stdlib, benchmark, reference, language, dependency, or WASM change
from this tranche.

Fresh source, assembly, dynamic-operation, allocation, and five-pair timing
evidence separates the three selected strict compiled misses:

- Fib performs the same recursive calls as Go, but its generated total
  function retains a fallible-control result and two nil checks per non-base
  call.
- Tapelang Alphabet uses a flat jump-table program while Go uses a recursive
  instruction tree. Able dispatches 1,268,239,314 operations where Go visits
  690,455,687 instruction nodes.
- Sudoku Masks performs the same search, but its Able source returns each best
  position as a dynamic three-element Array. Its 3,893,780 best updates cause
  exactly 7,787,560 allocations: one Array carrier and one backing slice per
  update. Go returns four scalars.

These are a compiler-effect ABI, an application algorithm/data-layout choice,
and an application result-representation choice. No concrete redundant
operation can be removed by one general compiler rule across the required
three unlike applications. The admission gate fails before implementation.

The compact machine-readable companion is
`2026-07-26-compiled-semantic-work-equivalence-closure.json`.

## Frozen contract

All generated applications used compiler SHA-256
`28697a5adf4f73918f3d83fbcddc211407dc7e539240f64e4127a4e3dd4ddcab`
and `--no-fallbacks`. All five final graphs contain 96 packages and omit
`able/interpreter-go/pkg/interpreter`.

Able and Go binaries were built once, then reused for five balanced,
alternating pairs on CPU 4. Every process used the catalog working directory,
arguments, input, and public verifier. The 50/50 executions passed.

| Application | Able binary SHA-256 | Go binary SHA-256 |
| --- | --- | --- |
| Tapelang Alphabet | `30721f2c769a678d740cce90096f295f2ff5f04ec97d9ca34d935a99ec43e93f` | `894b6fa87252a3632f663a412e0a1f8df92aea07325f5a76c1c544558dde6259` |
| Fib | `5b676673afd41c276526a9973665d6e4ee012bb6bde04a6d2dd7cb64d96274ff` | `5cd8b76f28156ea21945eb4a804425ff2aaf7c336979034e0650dc0145b33cee` |
| Sudoku Masks | `736474c9ef1f1c12c9c3d68c19c588cb98bac110feb3f230fe9ccfb95111058f` | `e80b51ecdf7731cc98baa1442d6ad7c5d6f51efda8a35765a631fa4716fbc83f` |
| Matrix Multiply | `9dde25d2ad0d1fde0d287ff52fda3f7dff075008c4c0555ea6ef3310ab75052e` | `81ddeffd210a5c5c5467cf9548ebc53b0b901082c24fe6dfd15bbe7990302b6a` |
| Pidigits | `4a51d18f92c6f3f6a3b576db232dc02666bef803bef025905a31d84e28f0e67f` | `02f02ab91722e1ebcc903821a42177ec67b83446715dc2ef9619aa852b2cac98` |

The raw timing TSV has SHA-256
`5be88fc7109ae87b2694b37fab64a2c9bb5413230ef07d6ae1d725dd2f1c4360`.
Disposable binaries, generated modules, counters, and disassembly remained
under disk-backed `/var/tmp`.

## Repeated exact-artifact comparison

| Application | Able mean | Go mean | Ratio | Able range | Go range |
| --- | ---: | ---: | ---: | ---: | ---: |
| Tapelang Alphabet | 3.740s | 1.924s | 1.944x | 3.22-4.20s | 1.85-1.96s |
| Fib | 3.360s | 3.128s | 1.074x | 3.31-3.39s | 3.00-3.27s |
| Sudoku Masks | 1.648s | 0.568s | 2.901x | 1.58-1.80s | 0.55-0.59s |
| Matrix Multiply control | 1.096s | 1.032s | 1.062x | 1.02-1.18s | 0.99-1.05s |
| Pidigits control | 1.074s | 1.176s | 0.913x | 1.02-1.11s | 1.10-1.22s |

Matrix's exact pair is 0.9 percentage points outside the 1.052632x target
amid its observed workstation range; the five-run full scorecard immediately
before this tranche measured 1.0126x. It remains a noise-aware control rather
than a regression claim. Matrix output differs only in formatting precision
and passes the shared numeric verifier.

## Fib: equal algorithm, single-family effect ABI

Both sources implement identical `fib(i32)` recursion and produce the same
2,269,806,339 calls for `fib(45)`, including 1,134,903,169 non-base calls.
The generated function uses native `int32` arithmetic and allocates no
application object.

The surviving difference is the result ABI:

```text
Able: (int32, *__ableControl)
Go:   int32
```

Every non-base generated call tests the returned control pointer after each
of its two recursive children. That is 2,269,806,338 dynamic control tests and
conditional branches. They are unreachable for this generated function,
whose body contains no control-producing path.

| Function | Code bytes | Instructions | Calls | Conditional/jump instructions |
| --- | ---: | ---: | ---: | ---: |
| generated Fib | 166 | 52 | 5 | 5 |
| reference Fib | 84 | 27 | 3 | 3 |

This is a real general effect-summary opportunity in principle, but it is
material only in Fib within the audited cohort. Tapelang's hot mutations can
raise overflow and preserve Array mutation semantics. Sudoku's hot helpers
contain checked arithmetic and fallible Array operations. Matrix and Pidigits
likewise retain reachable Array or arithmetic/extern controls.

The prior corpus control-effect and primitive-range censuses also found Fib as
the sole material control-free recursive row. A Fib-only ABI specialization
fails the three-unlike admission gate.

## Tapelang: unequal dynamic program representation

Both applications execute the same tape language and emit identical bytes,
but they do not perform equivalent interpreter work:

- Able parses to three flat structure-of-arrays vectors for opcode, operand,
  and jump target. Each loop iteration dispatches both `LOOP_START` and
  `LOOP_END`.
- Go parses to recursive `[]Op` blocks. Its range loop visits the `LOOP` node
  once per containing-block execution and calls the nested block for each
  nonzero iteration.

An instrumented copy of the unchanged Go reference measured:

| Event | Count |
| --- | ---: |
| instruction-node visits | 690,455,687 |
| recursive `_run` calls | 288,891,814 |
| increment operations | 603,786,263 |
| move operations | 57,780,405 |
| tape gets | 317,780,833 |
| dynamic loop-node visits | 28,888,993 |
| print operations | 27 |

The same tape actions imply the flat Able dispatcher executes:

| Flat operation | Count |
| --- | ---: |
| increments | 603,786,263 |
| moves | 57,780,405 |
| loop starts/tests | 317,780,806 |
| loop ends/backedges | 288,891,813 |
| prints | 27 |
| total flat dispatches | 1,268,239,314 |

Able therefore performs 577,783,627 more dispatches, or 83.68% more, before
considering required checked `i32` increment/move and Array read/write
semantics. This is not compiler-emitted redundant work: it follows directly
from the benchmark application's flat representation.

Static assembly reflects the different shapes but cannot make them equivalent:

| Function | Code bytes | Instructions | Calls | Conditional/jump instructions |
| --- | ---: | ---: | ---: | ---: |
| generated flat `execute` | 1,018 | 276 | 17 | 53 |
| generated `Tape.inc` | 295 | 82 | 4 | 16 |
| Go recursive `_run` | 977 | 207 | 13 | 26 |

A Tapelang-specific flat-dispatch or named-program optimization is forbidden.
Changing the benchmark representation is evidence normalization, not a
compiler optimization.

## Sudoku: equal search, unequal result representation

Both applications scan the board in row-major order, choose the empty cell
with the fewest candidates, try digits in ascending order, and update the same
row/column/square masks. An instrumented copy of the unchanged Go reference
measured:

| Event | Count |
| --- | ---: |
| `findBestEmpty` calls | 1,918,450 |
| recursive solve calls | 1,918,450 |
| board cells scanned | 155,394,450 |
| empty cells evaluated | 64,090,010 |
| bit-count calls | 64,090,010 |
| bit-count iterations | 166,923,250 |
| best-position updates | 3,893,780 |
| candidate digits tried | 1,918,350 |
| backtracks | 1,912,510 |

The Able allocation profile from the immediately preceding current-binary
owner refresh attributes exactly 7,787,560 objects to generated
`find_best_empty`. That equals two objects for every one of the 3,893,780
best-position updates.

The source explains the exact relation:

- Able creates `Array.with_capacity(3)`, pushes row, column, and choices, then
  returns `?Array i32`;
- Go retains row, column, choices, and found as scalars and returns four
  values.

| Function | Code bytes | Instructions | Calls | Conditional/jump instructions |
| --- | ---: | ---: | ---: | ---: |
| generated `find_best_empty` | 984 | 249 | 14 | 26 |
| Go `findBestEmpty` | 253 | 73 | 0 | 8 |
| generated `solve_with_masks` | 1,623 | 390 | 27 | 32 |
| Go `solve` | 536 | 130 | 6 | 12 |

Scalar-replacing this particular escaping dynamic Array, changing its source
to a positional record, or adding a Sudoku position rule would be
single-application work. No unlike application supplies the same hot
three-element dynamic-Array result shape.

## Control reconciliation

Matrix Multiply performs the same two 1,000x1,000 matrix generations, one
1,000x1,000 transpose, and 1,000,000,000 multiply-add kernel iterations.
Generated `matmul` is statically larger—398 instructions versus 189—but both
operate on native Go slices and the current repeated ratio is at the noisy
target boundary. It does not share Fib's unreachable control result or
Sudoku's escaping position Array.

Pidigits uses the same spigot sequence and the reference performs 22,454
candidate-digit attempts, 33,209 term iterations, 44,908 digit extractions,
and 10,000 eliminations. Its source work is not identical:

- Go performs two big-integer multiplications for the `3*numer` and
  `4*numer` digit extractions.
- Able computes `3*numer` once and obtains the second numerator with one
  addition of `numer`.
- Able's native `mul_i64` boundary also avoids the reference's reusable
  big-integer `y2` and `bigk` setup.

That lower expensive-operation count is consistent with Able's 0.913x control
ratio despite generated `next_term` being 153 instructions versus Go's 76.
It demonstrates why static generated-code size alone cannot select a compiler
optimization when semantic work differs.

## Candidate gate

| Candidate | Reach | Disposition |
| --- | ---: | --- |
| total-function no-control ABI | Fib only | not admitted; prior broad control/effect route also closed |
| flat tape dispatch fusion | Tapelang only | benchmark/application-specific |
| small escaping position Array scalar replacement | Sudoku only | one dynamic shape in one application |
| broad checked arithmetic removal | Tapelang and Sudoku, with different proofs | closed by prior mixed A/B and required Able semantics |
| generic Array bounds/result elision | Tapelang, Sudoku, Matrix | no shared proof; Matrix control already meets the target |
| generated function-size reduction | all five statically | not a semantic operation and contradicted by Pidigits |

No code was prototyped because no candidate passed the evidence gate.

## Verification

The exact five strict and five reference binaries built successfully. All
smoke executions, all 50 timing executions, and all three instrumented
reference executions passed their public verifiers. Every strict dependency
graph remains interpreter-free.

The complete `./run_all_tests.sh` handoff passed every coverage, scorecard,
selection, and threshold contract; every non-compiler package; all 32 bounded
compiler batches; and the final bytecode fixture corpus in 85.166 seconds.
The known large compiler aggregates included batch 19 at 179.523 seconds,
batch 28 at 72.179 seconds, and batch 29 at 84.153 seconds. They are existing
multi-test sharding debt. This tranche adds no production, test, benchmark,
reference, stdlib, or language source.

## Next recommendation

Normalize benchmark semantic-work equivalence before selecting another
compiled optimization. Begin with Tapelang Alphabet and Sudoku Masks, while
preserving Fib as the equal-algorithm compiler-effect sentinel.

Why: Tapelang's 1.944x and Sudoku's 2.901x ratios currently combine compiler
cost with 83.68% extra dispatch work and 7,787,560 source-chosen allocations.
Those rows cannot honestly measure native-lowering parity until both sides
perform the same algorithm and comparable data representation.

What it entails: define a checked operation/data-shape contract for each row;
choose one shared algorithmic representation for Able and Go; update only
benchmark/reference code and coverage metadata needed for equivalence; retain
the existing inputs, outputs, and public verifiers; and repeat five-or-more
balanced comparisons. Do not count benchmark normalization as a compiler
speedup, and do not add any benchmark, named-container, or non-primitive
nominal compiler rule.

Why it is important: the project goal is performance against equivalent Go
applications. Removing inequivalent work from the evidence base makes the
remaining ratio actionable: a miss then identifies compiler/runtime tax,
while a pass proves native-carrier parity. It also prevents large apparent
gaps from driving unsafe optimizations of required Able semantics. Do not
begin WASM work.
