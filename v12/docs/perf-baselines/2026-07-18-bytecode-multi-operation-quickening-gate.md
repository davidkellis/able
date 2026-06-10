# Bytecode Multi-Operation Quickening Gate (2026-07-18)

## Decision

Retain one generic bytecode optimization: a primitive comparison immediately
followed by `JumpIfFalse` is quickened into a guarded comparison/branch opcode.
The direct path handles primitive integer and float comparisons without
materializing a temporary Able `Bool` on the value stack or dispatching the
second instruction. The original `JumpIfFalse` instruction remains in the
program, and every unsupported, dynamic, nominal, or overloaded comparison
falls through the existing `Binary` implementation and then executes that
unchanged branch.

This is an instruction-boundary optimization, not a benchmark or nominal-type
special case. It is admitted by the same dynamically executed sequence in
seven unlike applications and retains a material application win without a
greater-than-5% regression in the repeated guard set. No compiler, benchmark,
fixture, language, or canonical `able-stdlib` change is retained.

## Census method

Temporary opt-in counters recorded dynamically adjacent opcode pairs only
when both instructions were executed consecutively in the same bytecode
program. Call/return and branch transitions were deliberately excluded. A
second temporary counter split `Binary -> JumpIfFalse` by operator. The
observer ran one process at a time with the canonical external stdlib,
`GOMEMLIMIT=1GiB`, `GOGC=50`, the catalog CPU/executor contract, public output
verification, and a 55-second process cap.

These are frequency counts, not timings: the observer uses atomics and changes
runtime cost. Distance Field and Unicode Scalar Pipeline reached the cap, so
their partial snapshots were deleted and are not treated as zero or as
admission evidence. Eight verified applications completed:

| Application | Executed opcodes | Adjacent pairs | Distinct pairs | `Binary -> JumpIfFalse` |
| --- | ---: | ---: | ---: | ---: |
| Fixed Width 128 | 63,729,888 | 50,383,928 | 53 | 3,283,452 (6.52%) |
| Reverse Complement | 109,403,123 | 85,135,422 | 86 | 4,066,708 (4.78%) |
| Word Frequency | 17,882,116 | 14,158,435 | 124 | 290,227 (2.05%) |
| Rational Series | 53,913,733 | 42,661,610 | 78 | 1,500,005 (3.52%) |
| Array Slice Window | 9,075,499 | 7,466,700 | 85 | 324,260 (4.34%) |
| Option/Result Config | 3,538,629 | 2,527,339 | 83 | 51,457 (2.04%) |
| I-Before-E | 5,196,308 | 4,157,473 | 44 | 345,648 (8.31%) |
| Base64 control | 6,000,582 | 3,000,434 | 35 | 0 |

At a 2% materiality threshold, `Binary -> JumpIfFalse` repeats in seven
applications. The operator split independently confirms primitive comparison
traffic in unlike programs:

- Array Slice Window: 312,259 `>=` and 12,001 `>`.
- Option/Result Config: 26,832 `==` and 24,625 `>=`.
- I-Before-E: 172,824 `>` and 172,824 `>=`.
- Word Frequency: 157,022 `>`, 131,073 `==`, 1,938 `>=`, and 194 `!=`.

Other frequent pairs were not candidates. `LoadSlot -> LoadSlot` and
`Pop -> LoadSlot` are transport with no shared semantic boundary to remove.
`StoreSlotNew -> Pop` crosses allocation and identity semantics and reopens an
already rejected discard-store family. Specialized slot/constant store/jump
shapes likewise reopen previously tested lowering rather than expose a new
generic wall.

Raw completed snapshots are retained in
`2026-07-18-bytecode-sequence-census/`. All temporary pair and operator
counters were removed from the production runtime after the census.

## Retained implementation

Program finalization recognizes only adjacent generic `Binary` comparisons
with operators `<`, `<=`, `>`, `>=`, `==`, or `!=`. It changes the first
instruction to `JumpIfBinaryCompareFalse`, copies the existing branch target,
and does not alter instruction indices or remove the fallback branch.

At execution time:

1. Existing primitive integer and float comparison helpers are attempted.
2. A direct hit consumes both operands and updates the instruction pointer to
   either the branch target or the instruction after the retained branch.
3. A miss executes the existing generic `Binary` path; the retained original
   `JumpIfFalse` consumes its result on the following dispatch.

The fallback therefore preserves custom dispatch, nominal behavior, errors,
truthiness, and program branch targets. Tests cover finalization, direct true
and false integer branches, a direct float branch, and generic String fallback.

## Repeated application gate

Timing used separately built preserved baseline and candidate binaries with
stats disabled. Each row alternated baseline/candidate order, retained every
sample, used CPU 0 affinity and the catalog execution contract, and passed the
public Ruby verifier. Short or initially volatile rows received 15 samples per
binary; the two approximately seven-to-eight-second controls received three.

| Application | Samples / binary | Baseline mean | Candidate mean | Change |
| --- | ---: | ---: | ---: | ---: |
| Array Slice Window | 15 | 0.7020 s | 0.6606 s | -5.89% |
| I-Before-E | 15 | 0.5575 s | 0.5482 s | -1.67% |
| Fixed Width 128 | 3 | 8.1426 s | 8.0661 s | -0.94% |
| Reverse Complement | 3 | 7.0207 s | 6.9781 s | -0.61% |
| Base64 | 15 | 3.1024 s | 3.1269 s | +0.79% |
| Option/Result Config | 15 | 0.8784 s | 0.8860 s | +0.86% |
| Word Frequency | 15 | 1.5753 s | 1.6028 s | +1.74% |

The initial five-sample Option/Result and Word Frequency results moved in
opposite directions when expanded, which is why the complete expanded means,
including all early samples and outliers, govern the decision. Array Slice
Window remains a repeatable material win; the other six rows remain within
roughly two percent. Raw timing samples are retained in the sibling
`2026-07-18-bytecode-conditional-branch-ab-*.json` files.

## Verification

- Focused finalization, direct integer/float branch, generic fallback,
  `JumpIfFalse`, binary fast-path, and stats tests pass.
- The broader `TestBytecode` partition passes in 23.237 seconds.
- A clean post-observer binary completes and publicly verifies Array Slice
  Window.
- The shared fixture harness currently reports four pre-existing Error
  truthiness/interface-cast failures. The preserved pre-candidate binary and
  candidate binary produce identical output and exit status for all four, so
  they are recorded as existing worktree blockers rather than attributed to
  this opcode.
- `git diff --check` passes, and every touched Go source remains below 1,000
  lines.

## Next recommendation

Refresh the full selected bytecode scorecard before opening another VM
optimization. This retained opcode changes a sequence present in seven census
applications, while the current promoted scorecard predates it and only 3/27
selected bytecode rows meet both interpreter targets. Run every currently
bounded bytecode row repeatedly under its recorded CPU/executor contract,
verify every output, average all samples, and expand only volatile boundary
rows. Then profile the largest remaining misses and admit another candidate
only if the same concrete removable boundary repeats across at least three
unlike programs. This measures actual movement toward Python/Ruby parity and
prevents the single strong Array Slice result from steering the next tranche
by itself. Continue to defer WASM.
