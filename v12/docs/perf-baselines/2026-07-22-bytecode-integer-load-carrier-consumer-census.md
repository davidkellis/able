# Bytecode integer-load carrier/consumer census

Date: 2026-07-22

Decision: retain the opt-in diagnostic attribution, retain no execution-path
change, and reject the direct-bool comparison candidate.

## Census

The scalar-proof census was extended without adding work to ordinary VM runs.
When `ABLE_BYTECODE_STATS` is enabled, every verifier-proven integer
`LoadSlot` now records its concrete carrier and first source-enclosing consumer
opcode. Consumer attribution retains an explicit `unknown` category and the
exact opcode, preventing a broad category from satisfying the generality gate.

Each census application ran once in its own verifier-backed process with
`GOMAXPROCS=1`, `GOGC=50`, `GOMEMLIMIT=1GiB`, a 55-second cap, and
`ABLE_BYTECODE_STATS_MAIN_ONLY=1`. Every count reconciles exactly to the prior
proven-integer `LoadSlot` total; no program rows were dropped.

| Application | Family | Proven integer loads | Dominant carriers |
| --- | --- | ---: | --- |
| Fixed Width 128 | wide numeric | 5,000,006 | small value 2,000,006; raw i32 1,262,143; raw typed cell 999,999; small pointer 737,856 |
| Concurrent Event Routing | concurrency/text | 1,086,208 | raw i32 633,856; small value 452,352 |
| Word Frequency | text/map | 944,384 | raw i32 576,412; small value 367,897 |
| Array Slice Window | array/slice | 1,020,268 | raw i32 624,260; small value 396,008 |
| Reverse Complement | bio/text/array | 6,033,694 | small value 4,000,081; small pointer 1,771,211; raw i32 262,402 |
| Distance Field | float control | 0 | none; its corresponding scalar proof is float |

One exact carrier/consumer pair passed the material breadth gate cleanly:
raw-i32 slot values feeding `JumpIfBinaryCompareFalse` occurred 312,259 times
in Array Slice Window, 156,544 in Concurrent Event Routing, 262,145 in Reverse
Complement, and 133,011 in Word Frequency. Cast shapes crossed three or four
applications but were strongly concentrated in one or two. A small-integer
value feeding `CallMemberArraySlot` crossed three applications with 2,241,366
executions and remains the next unresolved exact shape.

## Candidate and repeated A/B

The branch already uses direct integer comparison. The only safely removable
work at that boundary was returning `runtime.BoolValue` through the generic
`runtime.Value` interface and immediately asserting it back. A temporary
shared bool-returning helper removed that round trip without changing opcode
selection, evaluation order, fallback semantics, or public value ABI.

Because sequential three-run means moved with workstation load, a reversed
control and second candidate cohort were collected. The pooled result contains
six independently verified processes per side:

| Application | Control mean | Candidate mean | Change |
| --- | ---: | ---: | ---: |
| Fixed Width 128 | 8.2650s | 7.8417s | -5.12% |
| Concurrent Event Routing | 3.0400s | 3.0100s | -0.99% |
| Word Frequency | 1.4150s | 1.4183s | +0.24% |
| Array Slice Window | 0.6733s | 0.7317s | +8.66% |
| Reverse Complement | 3.3317s | 3.3867s | +1.65% |
| Distance Field (zero-reach control) | 5.9000s | 5.5750s | -5.51% |

The candidate fails the broad bar: two materially reached applications
regress, while the largest apparent improvement is matched by a workload that
never reaches the integer shape. It was fully removed. No compiler,
tree-walker, stdlib, benchmark, language, or WASM code changed.

## Next recommendation

Next, attribute operand role and actual extraction work within the shared
small-integer-value to `CallMemberArraySlot` shape across Concurrent Event
Routing, Reverse Complement, and Word Frequency. This should distinguish the
receiver, index, and value operands and confirm whether the direct kernel
opcode still materializes a scalar unnecessarily.

Why: it is the largest remaining exact shape that is material in three unlike
applications, while the comparison round trip has now been rejected with a
zero-reach control. The work entails opt-in role counters, clean CPU sampling
at that opcode, and a candidate only if one operand role exposes the same
removable generic kernel-boundary operation in all three programs. This is a
language/kernel boundary investigation, not an Array nominal-type or
benchmark-specific lowering rule.
