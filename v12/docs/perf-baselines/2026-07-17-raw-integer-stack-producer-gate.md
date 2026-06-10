# Raw-integer stack producer and consumer gate

Date: 2026-07-17

## Decision

Complete the raw-integer stack producer/consumer tranche and retain no runtime
or lowering change. Binary production dominates the signed-`i32` stack traffic
in the selected programs, but most results either escape immediately or split
across different expression shapes. The one two-operation syntax shape shared
by three primary programs produced neutral-to-narrow timing changes and no
material allocation reduction.

The temporary producer counters, transition counters, JSON output, runner
scripts, diagnostic binaries, benchmark outputs, fused opcode, lowering rule,
runtime helper, and candidate tests were removed. The compiler, VM, stdlib,
fixtures, benchmark sources, and language remain unchanged.

## Attribution protocol

Temporary main-only counters tagged the centralized
`appendRawIntegerStack(...)` and `replaceTop2RawIntegerUnchecked(...)` callers
as match value/snapshot, iterator, index, slot load, slot-constant binary,
ordinary arithmetic binary, bitwise binary, cast, or return. Only values that
actually entered the signed-`i32` stack path were counted. The first version
also counted type-`i32` intermediates outside the signed range; that accounting
was corrected before selection runs and none of those preliminary files were
used.

For every counted producer, the final census also recorded the next two
bytecode operations and their operators. This distinguished a value consumed
by another arithmetic operation from mandatory call, aggregate, match, store,
or return escape. Instrumentation never changed the produced value.

Rational Series, Array Slice Window, Array Map, Binary Trees, and split/join
ran twice. Their counter JSON and stdout were byte-identical between processes.
Both runs of each external program passed its Ruby verifier; all fixture runs
exited successfully. One bounded K-Nucleotide process passed its verifier.

## Producer result

| Workload | Signed-`i32` stack writes | Leading producers |
| --- | ---: | --- |
| Rational Series | 600,000 | slot-constant binary 300,000; arithmetic binary 300,000 |
| Array Slice Window | 336,519 | cast 300,003; slot-constant binary 24,257; arithmetic binary 12,259 |
| Array Map | 120,000 | slot-constant binary 42,000; arithmetic binary 42,000; match snapshot 36,000 |
| K-Nucleotide | 10,336,631 | slot-constant binary 8,233,467; match value 2,103,006 |
| Binary Trees | 679,959 | arithmetic binary 679,954 |
| split/join | 407,987 | arithmetic binary 209,496; slot-constant binary 198,491 |

Binary producers own all requests above 65,535 in Rational Series, Array Slice
Window, and Array Map. They also own essentially all Binary Trees traffic and
about half of split/join. K-Nucleotide's high-value traffic is 68.81%
slot-constant binary, 31.19% match value, and negligible ordinary arithmetic.

Consumer sequences show why the aggregate producer family is not one
optimization:

- Rational executes 300,000
  `BinaryIntMulSlotConst(*) > Const > Binary(%)` sequences.
- Array Slice executes 12,257
  `BinaryIntMulSlotConst(*) > Const > BinaryIntAdd(+)` sequences, while its
  300,003 casts feed Array calls or stores.
- Array Map executes 42,000
  `BinaryIntMulSlotConst(*) > Const > BinaryIntAdd(+)` sequences; its following
  arithmetic then splits between return and modulo.
- split/join sends 197,490 slot-constant results directly into a struct literal
  and return, while most arithmetic results feed calls, stores, or conditions.
- Binary Trees has almost no slot-constant pair to fuse.

K-Nucleotide's inline-program transitions could not be reliably read through
the active-program pointer, so they were not inferred. The candidate binary's
ordinary bytecode statistics subsequently reported exactly zero fused-opcode
executions in a verified K-Nucleotide process, proving that the selected shape
does not occur there.

## Candidate

The generic candidate fused these two language-level expression forms:

```text
(slot * integer_constant) + integer_constant
(slot * integer_constant) % integer_constant
```

One value-producing opcode carried both typed immediates. It used the existing
raw integer rules for both operations, preserved intermediate overflow and
Euclidean modulo behavior, and materialized only the final stack result. A cold
fallback evaluated the same two ordinary language operations for unsupported
runtime representations. The eligibility rule contained no program, function,
container, or stdlib name.

Focused lowering, `i32`, `i64`, modulo, and fallback guards passed. The full
`TestBytecodeVM` family passed in 16.9 seconds. Verified candidate processes for
all six workloads produced byte-identical baseline output.

## Alternating A/B result

The benchmark harness loaded and warmed each program before one measured
`main()` call. Processes used one logical CPU, `GOGC=50`, a 1 GiB memory limit,
and a 55-second cap. Primary programs received ten alternating baseline and
candidate processes; controls received five. Every sample remains in the
means.

| Workload | Baseline mean | Candidate mean | Change | Baseline/Candidate CV | Allocation result |
| --- | ---: | ---: | ---: | ---: | --- |
| Rational Series (10 pairs) | 3.7685 s | 3.7314 s | -0.98% | 4.30% / 3.47% | five fewer allocations; about 80 B less |
| Array Slice Window (10 pairs) | 0.5067 s | 0.5071 s | +0.07% | 6.60% / 2.42% | allocations unchanged; 384 B less |
| Array Map (10 pairs) | 0.0714 s | 0.0657 s | -8.05% | 13.16% / 4.59% | exactly unchanged |
| Binary Trees (5 pairs) | 1.2180 s | 1.2332 s | +1.25% | 5.95% / 5.18% | effectively unchanged |
| split/join (5 pairs) | 0.9886 s | 1.0112 s | +2.28% | 1.66% / 6.41% | normal small process variation |

Array Map's apparent win includes a 96.7 ms baseline spike; its other baseline
samples are much closer to the candidate cohort. Rational and Array Slice are
neutral relative to their measured variation. The controls are mixed, and the
candidate removes no recurring allocation wall. K-Nucleotide cannot run both
warmup and measurement under the one-minute test rule, and the verified opcode
census proves it has zero candidate executions, so no overlong timing process
was admitted.

This fails the broad-benefit gate. Retaining a new opcode and fallback for one
short volatile Array improvement would be precisely the benchmark-shaped
optimization the project rules prohibit.

## Restoration and verification

After the revert:

- no producer/transition counter or fused-opcode symbol remains;
- the full `TestBytecodeVM` family passes in 17.7 seconds;
- focused CLI run tests pass; and
- the benchmark-selection contract check passes.

## Next recommendation

Refresh clean bounded CPU and allocation profiles for Rational Series, Array
Slice Window, Array Map, and K-Nucleotide, then reconcile their cumulative
parents rather than continuing raw-integer stack or slot-constant fusions.

Why: three consecutive gates now show that the large raw-`i32` cache is useful,
its shared stack lookup is allocation-free, and the most obvious common
producer fusion is neutral outside one volatile fixture. The next material wall
must be elsewhere—most plausibly repeated raw extraction, call/return and type
matching, or map/member lookup—but the current evidence does not justify
choosing among them.

What it entails: create one-process profiles under the existing one-core,
1 GiB, and sub-minute guardrails; use longer repeated benchmark means only for
any candidate that emerges; and require the same generic cumulative parent in
at least three unlike workloads. Attribute allocation and CPU together so a
cheap cache lookup is not mistaken for boxing. Build no candidate if the
profiles remain split, and continue to defer WASM.
