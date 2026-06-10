# Bytecode Typed-Float Register Gate

Date: 2026-07-16

## Outcome

Complete this tranche with no retained VM, compiler, stdlib, fixture, or
benchmark production change.

A conservative typed `f32`/`f64` frame experiment was implemented and then
reverted. It improved reduced NBody substantially, but increased Distance Field
bytes by 31.25% and regressed RMS Norm time, bytes, and allocations. That mixed
result fails the broad benchmark bar.

## Candidate

The experiment extended the existing frame-owned float sidecar with primitive
float slot kinds, seeded proven float parameters during ordinary inline calls,
kept discarded fused-float stores in the sidecar, and copied simple primitive
float returns into caller-owned carriers. Dynamic, native, aggregate, and
coercing boundaries continued to materialize runtime float values.

Reusable operand-stack cells were necessary to move float results through the
existing `[]runtime.Value` stack without allocating a new boxed float at every
operation. New ownership tests verified that callee parameters and returns were
copied before the caller's stack index could be reused. An application warmup
also exposed and repaired a missing sidecar read in slot-argument native calls.

The candidate remained architecture-wide: eligibility came from primitive
`f32`/`f64` frame facts, with no function, stdlib type, or benchmark-name
special case.

## Bounded gate

Each workload ran in a fresh process with one warmup and one measured
iteration. The allocation and byte totals are deterministic enough to reject
the candidate before a longer timing series: the scalar workloads acquire
large new byte/allocation costs even though reduced NBody benefits.

| workload | retained baseline | candidate | decision |
| --- | ---: | ---: | --- |
| Distance Field | 6.134s / 512.1 MB / 38.00M allocs | 6.141s / 672.1 MB / 38.00M allocs | time flat; bytes +31.25% |
| RMS Norm | 5.767s / 592.1 MB / 52.00M allocs | 6.496s / 992.1 MB / 54.00M allocs | time +12.65%; bytes +67.56%; allocs +3.85% |
| reduced NBody | 1.701s / 97.6 MB / 6.16M allocs | 1.627s / 82.6 MB / 4.17M allocs | time -4.34%; bytes -15.38%; allocs -32.31% |

After the revert, a fresh Distance process measured 5.785s / 512.1 MB /
38.00M allocations, confirming that the retained allocation shape was
restored. Workstation timing remains noisy; the hard byte/allocation regression
made further candidate timing repetitions unnecessary.

## Verification after revert

```text
go test ./pkg/interpreter -run 'TestBytecodeVM_(Float|.*Return|.*Inline.*Call|CallName|SlotOperands)|TestInlineCoercion|TestBytecode.*Float' -count=1 -timeout 55s
ok  able/interpreter-go/pkg/interpreter
```

The unsplit package-wide command reached its 55-second aggregate timeout in
`TestExecFixtureParity/09_05_method_set_generics_where`; it reported no
candidate-related assertion failure before the timeout. Focused verification
was therefore kept in bounded subsets.

## Lesson and next direction

The existing `[]runtime.Value` operand stack is the limiting boundary. A
pointer carrier can avoid one box, but its escape and snapshot costs move or
increase allocation in call-heavy scalar programs. Adding more sidecar reads
will repeat that trade rather than establish a true primitive register path.

The next candidate should begin with a bounded retained-state profile refresh,
then prototype a non-interface primitive-float operand lane only if float
boxing remains the same concrete descendant across Distance, RMS, reduced
NBody, and an unlike float guard. Such a lane would store scalar values and
tags by stack index, seed callee float frames by value, and materialize only at
dynamic/native/aggregate boundaries. It is a larger change, but it directly
removes the pointer escape that caused this gate to fail.
