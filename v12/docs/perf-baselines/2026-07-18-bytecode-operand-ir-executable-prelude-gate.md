# Bytecode operand-IR executable prelude gate

Date: 2026-07-18

## Decision

Reject the cached operand-prelude design and retain no VM, runtime, compiler,
canonical-stdlib, benchmark, fixture, or language change from it. Both bounded
implementations passed their correctness gates, but the refined one regressed
seven of eight unlike verified applications. Removing producer dispatch while
still materializing the same operand stack is not a useful VM architecture.

This result rejects only the compatibility-prelude approach. It does not
invalidate the preceding whole-function operand-IR static gate, because this
prototype did not introduce explicit source/destination registers, eliminate
`Dup`/`Pop`, or change semantic helpers to consume operands directly.

## Prototype contract

The opt-in `ABLE_BYTECODE_OPERAND_IR` path built a translation once per
bytecode program. Contiguous runs of `Const`, `ConstI32`, `LoadSlot`, and
`LoadSlotI32` were executed before the following semantic instruction without
entering the main opcode switch for each producer. Original bytecode remained
intact, disabled mode was the exact baseline path, and branches could enter a
transport run at any instruction.

The first implementation validated each cached run and then walked it again.
The second cached immutable instruction descriptors and walked each run once.
Both reused existing constant validation, raw-i32, slot-load, stack-append,
semantic-operation, call, error, concurrency, and return behavior. Neither
recognized a benchmark, source sequence, Array, regex, or named nominal type.

## Repeated refined A/B

One preserved binary supplied both modes; only the opt-in environment variable
changed. Each application ran in five alternating baseline/candidate pairs on
CPU 0 with `GOMAXPROCS=1`, `GOGC=50`, `GOMEMLIMIT=1GiB`, the canonical external
stdlib, catalog working directory and arguments, and a 55-second process cap.
All 80 outputs passed their public Ruby verifiers. The refined binary SHA-256
was `36f20031143b610418f8fd32338aeb3c63c463f48eb3a52a2dfdcd2407da2a17`.

| Application | Baseline mean | Candidate mean | Delta | Candidate wins |
| --- | ---: | ---: | ---: | ---: |
| Future Pipeline | 0.512 s | 0.504 s | -1.56% | 3/5 |
| Fixed Width 128 | 8.312 s | 8.324 s | +0.14% | 3/5 |
| Distance Field | 5.950 s | 6.416 s | +7.83% | 1/5 |
| Mandelbrot | 6.858 s | 6.990 s | +1.92% | 1/5 |
| Array Slice Window | 0.696 s | 0.710 s | +2.01% | 2/5 |
| Option/Result Config | 0.926 s | 1.046 s | +12.96% | 2/5 |
| Word Frequency | 1.584 s | 1.734 s | +9.47% | 1/5 |
| Reverse Complement | 4.340 s | 4.590 s | +5.76% | 1/5 |

Positive deltas are regressions. Full samples are in
`2026-07-18-bytecode-operand-ir-executable-ab-samples.tsv`.

The initial double-walk version also failed the broad bar: its aggregate
deltas were +15.52%, +2.83%, -0.30%, -15.16%, +10.67%, -6.49%, +6.34%, and
+5.84% in the table's application order. Its apparent short-process wins were
not stable enough to outweigh systematic regressions in longer guards. The
refinement removed the avoidable validation walk but did not change the
decision.

## Interpretation

The prelude removes entries through the large outer switch, but it retains the
work that dominates their semantics: descriptor iteration, slot/raw-sidecar
checks, constant checks, stack appends, and later stack reads. It also adds a
plan lookup to every main-loop iteration. Longer semantic-heavy controls show
that the saved dispatch is smaller than this compatibility cost.

The static gate modeled a materially different machine: semantic operations
consume explicit slot/register/immediate operands and write explicit result
registers, while `Dup` becomes another reference and `Pop` becomes a discarded
destination. That representation can remove stack materialization as well as
dispatch. This prelude cannot, so extending it would optimize a rejected
compatibility layer rather than implement the admitted architecture.

## Correctness and cleanup

- Focused plan, transport-dispatch, branch-entry, and cached-source tests
  passed for both prototypes.
- `ABLE_BYTECODE_OPERAND_IR=1 go test ./pkg/interpreter -run TestBytecode`
  passed before measurement.
- Verifier-backed fixture groups 01-14 passed with the opt-in path.
- After reverting the prototype, `go test ./pkg/interpreter -run
  TestBytecode -count=1 -timeout 60s` passed in 26.232 seconds.
- The environment switch, cached plans, executor, tests, and candidate binary
  are not retained.
- No canonical `able-stdlib` source needed a change.
- No WASM work was performed.

## Next recommendation

Implement a genuinely separate, opt-in operand IR for a deliberately small
but complete opcode subset, with explicit source operands and destination
registers and whole-function fallback when translation cannot be proved.

Why: the compatibility prelude proves that dispatch-only savings are
insufficient, while the prior static census still shows 30.09%-44.32%
representation traffic across all eight unlike applications. The remaining
testable hypothesis is removal of both dispatch and operand-stack
materialization; another wrapper around the existing stack helpers cannot test
it.

What it entails: define the operand/register representation and CFG stack-map
merge validator; translate constants, slot loads, `Dup`, `Pop`, and a bounded
set of common semantic operations; make those semantic executors read explicit
operands and write a destination without rebuilding the bytecode operand
stack; materialize ordinary runtime values at calls, dynamic operations,
allocation, errors, concurrency, and returns. Unsupported functions must fall
back before executing any effect. Keep it opt-in and retain it only after
focused parity plus repeated verifier-backed A/B improves at least three
unlike applications without a material broader regression.
