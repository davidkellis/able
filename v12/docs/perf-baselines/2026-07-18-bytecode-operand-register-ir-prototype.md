# Bytecode operand-register IR prototype

Date: 2026-07-18

## Decision (superseded 2026-07-19)

The retention decision below is superseded by
`2026-07-19-bytecode-operand-register-ir-admission-census.md`. A dynamic
admission observer subsequently proved that the opt-in path translated zero
function executions and executed zero semantic IR instructions in all eight
applications used by this timing gate. The measured deltas therefore cannot
be attributed to the IR and are workstation variance around an extra plan
lookup. The prototype, environment switch, tests, and observer were removed.

## Original decision

Retain the first true operand-register IR as an opt-in bytecode-VM path. Do not
make it the default yet.

The candidate passed focused semantics, the complete split exec-fixture parity
gate, all broad bytecode-named tests, and 80/80 verifier-backed application
processes. Five alternating pairs improved six unlike applications, left Fixed
Width 128 neutral, and regressed Array Slice Window 1.82%. It therefore clears
the predeclared requirement to improve at least three unlike applications
without a material broader-suite regression.

No compiler, canonical-stdlib, benchmark, fixture, language, or WASM change was
needed.

## Architecture

`ABLE_BYTECODE_OPERAND_REGISTER_IR=1` enables a once-per-program translation.
The translator performs a CFG worklist pass over the complete function and
tracks separate boxed-value and raw-i32 virtual stacks. It admits a function
only when every reachable instruction is supported and every control-flow
merge has exactly compatible operand state. Unsupported functions return no
plan and execute the unchanged bytecode VM before any effect occurs.

The first subset supports:

- `Const`, `LoadSlot`, `Dup`, and `Pop` as representation-only operands;
- `ConstI32` and `LoadSlotI32` as raw-i32 operands;
- generic binary work plus integer add/subtract/less-equal/div-cast and the
  integer slot/immediate binary family;
- raw-i32 add/subtract, unbox, box, and slot store;
- generic existing/new slot stores with the existing typed-coercion, raw
  integer, raw-i32 sidecar, float ownership, and self-slot behavior;
- ordinary and bool-slot conditional jumps, unconditional jumps, loop
  boundaries, proven no-op scope boundaries, and returns.

Semantic IR instructions read explicit immediate/slot/register operands and
write explicit destination registers. Supported execution never appends to or
reads from either existing VM operand stack. Slot loads remain direct operands
only while no store to that slot can invalidate a still-live reference; the
translator rejects the entire function otherwise. Constants retain runtime
integer validation and returns retain the existing inline/top-level coercion
and frame restoration paths.

The IR path is disabled whenever bytecode statistics are enabled so the
existing opcode observer remains an exact account of ordinary bytecode.
Resumable generator execution also retains the ordinary VM until suspension
state has an explicit IR contract.

## Repeated application gate

One preserved binary supplied both modes; only
`ABLE_BYTECODE_OPERAND_REGISTER_IR` changed. Each application ran in five
alternating baseline/candidate pairs on CPU 0 with `GOMAXPROCS=1`, `GOGC=50`,
`GOMEMLIMIT=1GiB`, the canonical external stdlib, catalog working directory
and arguments, and a 55-second cap. Every output passed its public Ruby
verifier. The binary SHA-256 was
`f9745f7901ecdc522e59e88b1485c72b109c751315a00c4b87589e97fb5e6edc`.

| Application | Baseline mean | Candidate mean | Delta | Candidate wins |
| --- | ---: | ---: | ---: | ---: |
| Future Pipeline | 0.520 s | 0.496 s | -4.62% | 3/5 |
| Fixed Width 128 | 8.434 s | 8.412 s | -0.26% | 3/5 |
| Distance Field | 6.228 s | 5.804 s | **-6.81%** | 4/5 |
| Mandelbrot | 6.726 s | 6.610 s | -1.72% | 4/5 |
| Array Slice Window | 0.660 s | 0.672 s | +1.82% | 1/5 |
| Option/Result Config | 0.912 s | 0.882 s | -3.29% | 2/5 |
| Word Frequency | 1.688 s | 1.576 s | **-6.64%** | 3/5 |
| Reverse Complement | 4.746 s | 4.402 s | **-7.25%** | 4/5 |

Negative deltas are improvements. The aggregate means clear the breadth rule;
paired direction is strongest in Distance Field, Mandelbrot, and Reverse
Complement. The short Future and Option rows remain noisier, so their gains
are supporting rather than decisive evidence. Full samples are in
`2026-07-18-bytecode-operand-register-ir-ab-samples.tsv`.

## Correctness

- Focused tests cover stack-free boxed arithmetic, raw-i32 register flow, CFG
  loops, stale-slot-reference rejection, and whole-function unsupported-op
  fallback.
- Focused operator, overflow/div-cast, bool-slot, and i32 guards pass with the
  path enabled.
- `ABLE_BYTECODE_OPERAND_REGISTER_IR=1 go test ./pkg/interpreter -run
  TestBytecode -count=1 -timeout 60s` passes in 24.932 seconds.
- The same restored baseline path passes in 24.744 seconds with the option
  disabled.
- Exec-fixture parity groups 02-04, 05-08, 09-11, 12, 13, and 14 all pass
  separately with the option enabled; every command completes in 2.505-28.824
  seconds.
- All eight application outputs verified in both modes for every pair.

## Interpretation

This result separates the successful architecture from the rejected prelude.
The prelude retained producer iteration, stack append/read work, and a plan
check at every ordinary bytecode instruction. The retained prototype performs
one function-entry plan lookup, executes only semantic IR instructions, and
keeps intermediate values in direct registers. Its gains across float
geometry, text/map, primitive-byte, concurrency, union/control, and numeric
programs are evidence for a general representation improvement rather than a
benchmark or named-type rule.

The subset remains intentionally conservative. It does not yet cover calls,
dynamic/member/index work, allocation, error regions, concurrency operations,
resumable generators, specialized fused stores/branches/returns, or CFG merges
that require phi-like registers. Those functions use the ordinary VM in full.

## Next recommendation

Run an operand-IR admission census that records translated executions,
semantic IR instructions, removed representation instructions, and the first
unsupported opcode for every rejected function across the selected benchmark
suite. Then add only the smallest missing semantic family that blocks at least
three unlike applications, followed by the same parity and repeated A/B gate.

Why: the opt-in core has demonstrated broad causal value, but enabling opcodes
blindly could grow a second VM around rare shapes or turn the small Array Slice
regression into a material one. A first-blocker census identifies coverage
that unlocks whole functions, which is the relevant unit under the safety
contract.

What it entails: add opt-in diagnostics that do not affect normal execution;
measure the current eight applications plus selected controls; rank rejected
functions by dynamic invocation count and first unsupported opcode; implement
one generic family such as fused slot stores/branches/returns only if the same
barrier repeats in at least three unlike applications; add focused translation
and semantic tests; rerun split parity and five alternating verifier-backed
pairs. Keep the IR opt-in until broader coverage has repeated gains and the
Array Slice guard is neutralized or proven inside ordinary variance.
