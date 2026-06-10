# Bytecode f64 Operand Lane Rejection

Date: 2026-07-16

## Outcome

Complete the true operand-stack f64-lane experiment with no retained VM,
compiler, stdlib, fixture, or benchmark production change. The boxed operand
stack and its completed storage boundary remain the retained implementation.

The experiment proved that operand boxing is a broad allocation wall, but it
also proved that maintaining raw state independently at every stack operation
costs more CPU than the avoided allocation saves. Every candidate reduced
wall performance across Distance Field, RMS Norm, reduced NBody, and
`matrixmultiply_f64_small`, so the candidate fails the broad retention bar.

## Candidate forms

An exact boxed test binary was captured before any source mutation. Three
successively narrower representations were evaluated:

1. a lazy tagged side lane that admitted ordinary runtime floats as well as
   VM-private raw float carriers;
2. the same lane restricted to VM-private raw carriers, so already-boxed
   runtime floats did not acquire bookkeeping; and
3. compact singleton f32/f64 markers in the boxed stack plus a parallel
   `[]float64`, removing the per-entry tag/value struct.

All forms preserved the zero-copy boxed slice required by call-argument views.
Raw-aware arithmetic, comparisons, division, fused float stores, returns, and
iterator results consumed the lane directly. Dynamic consumers materialized
the value through the central stack accessors. Dedicated guards covered stale
tags, truncation, index reuse, raw copies, f32 normalization, and borrowed
argument stability. Focused float/return tests and the full `TestBytecodeVM`
family passed before the performance decision.

No eligibility depended on a benchmark, function, stdlib type, or source
name. It was a primitive `f32`/`f64` VM representation experiment.

## Broad gate

The most informative form was the raw-carrier-only lane. Each row is an
adjacent fresh-process boxed/candidate pair on the shared workstation.

| workload | boxed | raw-only lane | wall change | allocation change |
| --- | ---: | ---: | ---: | ---: |
| Distance Field | 6.851 s / 512,052,144 B / 38,000,142 allocs | 7.984 s / 496,052,136 B / 36,000,141 allocs | +16.5% | -3.1% bytes / -5.3% allocs |
| RMS Norm | 6.103 s / 592,051,712 B / 52,000,164 allocs | 7.544 s / 512,051,112 B / 42,000,156 allocs | +23.6% | -13.5% bytes / -19.2% allocs |
| reduced NBody | 1.708 s / 97,572,512 B / 6,161,057 allocs | 2.026 s / 73,090,328 B / 4,500,895 allocs | +18.6% | -25.1% bytes / -26.9% allocs |
| `matrixmultiply_f64_small` | 0.322 s / 47,636,112 B / 1,187,689 allocs | 0.363 s / 41,880,936 B / 828,287 allocs | +12.7% | -12.1% bytes / -30.3% allocs |

The workstation is noisy, but the allocation counts are stable and all four
wall comparisons moved in the same adverse direction. Later reverse-order
runs after guarding reconciliation and using compact markers still put the
candidate behind the boxed binary; RMS was 11.652 versus 11.006 seconds under
heavy background load. The primary numeric gate was already a uniform wall
failure, so running the non-float guard suite could not make the candidate
retainable.

The broader first form was worse: tagging ordinary runtime floats increased
Distance to 752,053,912 bytes and 52,000,169 allocations, and RMS to
672,052,216 bytes and 54,000,179 allocations. Restricting eligibility fixed
that mistake but did not fix CPU cost.

## Profile attribution

The candidate RMS CPU profile attributes its new cost directly to lane
maintenance rather than to a displaced application hotspot. Representative
flat/cumulative shares included:

- reconciliation: 3.88% flat;
- direct lane float reads: 3.08% flat, 4.15% cumulative;
- raw-float append: 2.14% flat, 3.48% cumulative;
- generic stack append: 2.01% flat, 6.29% cumulative;
- truncate: 2.01% flat, 4.55% cumulative; and
- raw-float replacement: 5.09% cumulative.

Alignment guards and the compact marker removed some representation work but
could not make the lane competitive. The design pays for tag inspection,
metadata coherence, and branch pressure at virtually every operand mutation;
those costs are more frequent than garbage-collection savings.

## Verification after revert

The lane types, fields, markers, reads, writes, copies, and lifecycle tests
were removed. A repository search finds no lane symbol. The restored runtime's
production symbol set and symbol sizes match the exact pre-experiment boxed
binary.

```text
go test ./pkg/interpreter -run 'TestBytecodeOperandStack|TestBytecodeVM_.*Float|TestBytecodeVM_.*Return' -count=1 -timeout 60s
ok  able/interpreter-go/pkg/interpreter

go test ./pkg/interpreter -run '^TestBytecodeVM' -count=1 -timeout 60s
ok  able/interpreter-go/pkg/interpreter  18.518s

go test ./pkg/interpreter -run '^TestBytecodeOperandStack' -count=1 -timeout 60s
ok  able/interpreter-go/pkg/interpreter  0.047s
```

## Next recommendation

Prototype statically proven typed-float evaluation regions, rather than
another per-value operand representation.

Why: the allocation reductions above confirm that keeping primitive floats
raw is valuable, while the uniform wall regressions show that deciding and
maintaining representation on every generic stack operation is too expensive.
The next design must amortize one type proof and one materialization decision
over a sequence of primitive operations.

What it entails: extend bytecode analysis to identify contiguous expression
or basic-block regions whose inputs, intermediate results, loop-carried slots,
and exit type are proven `f32`/`f64`; execute their internal primitive
operations in compact raw locals/registers; and materialize once at calls,
interfaces, aggregates, dynamic control-flow joins, or region exit. Begin with
the repeated arithmetic and loop-carried opcode shapes shared by the four
numeric programs, never with source or nominal-type names. Add mixed-float,
branch/join, error, recursion, call, alias, and materialization guards, then
use alternating binaries across the four numeric workloads plus split/join,
iterator collect, and numeric array/map controls. Retain only if wall time as
well as allocation improves broadly. Continue to defer WASM.
