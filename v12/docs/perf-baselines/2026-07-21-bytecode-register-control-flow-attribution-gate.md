# Bytecode register control-flow attribution gate

Date: 2026-07-21

## Decision

Reject and fully remove the residual continuation-probe hoist together with
the recovered allocation-neutral operand-register engine.

The hoist removes millions of provably fruitless continuation probes and
improves the experimental engine materially in Future Pipeline and Word
Frequency. It still leaves Word Frequency 5.75% slower than the ordinary VM in
the combined ten-process guard. The selected three-application gate therefore
fails before a six-application promotion run is warranted.

No VM, runtime, compiler, stdlib, application, fixture, reference, scorecard,
language, or WASM change is retained.

## Protocol

The exact register-native `MemberAccess` engine and the continuation-safe,
bounded frame reuse from the preceding attribution tranche were recovered.
Separate frozen test binaries represented:

- the restored ordinary bytecode VM;
- the allocation-neutral register engine before the control change; and
- the same engine with the continuation-probe hoist.

`BenchmarkBytecodeProgramRuntime` loaded and warmed each application once and
measured steady-state `main()` calls. Clean CPU processes did not enable
bytecode statistics. Separate one-call census processes enabled statistics but
were never used as timings. All runs used the canonical external stdlib,
declared executor, `GOMAXPROCS=1`, `GOGC=50`, a 1-GiB memory limit, and a
60-second per-process timeout.

## Untimed control census

The original register integration called
`resumeBytecodeOperandRegisterIR(...)` at the top of every ordinary VM loop
iteration while the feature was enabled. The census split loop/continuation
probes, hits, program-entry admission, register instructions, ordinary fallback
ops, and register-return handoffs.

| Application | Continuation probes | Hits | Ordinary fallback probes | Register entries | Fallback probes / entry | Return handoffs |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Future Pipeline | 9,535,891 | 0 | 9,503,108 | 32,783 | 289.88 | 32,783 |
| Concurrent Text Index | 1,724,942 | 0 | 1,658,016 | 66,926 | 24.77 | 66,926 |
| Word Frequency | 17,430,077 | 0 | 17,394,118 | 35,959 | 483.72 | 35,959 |

The exact owner repeats in all three applications and every probe misses.
Clean CPU profiles directly sample the helper at 2.82% flat in Future and
2.08% in Word. Concurrent Text has the smallest normalized probe load and no
direct helper sample, but still performs 1.66 million unnecessary fallback
probes. This was sufficient to test the narrowly bounded generic hoist.

## Candidate

The ordinary VM now consulted register continuation state only when one of the
following was true:

- execution was at a fresh program entry eligible for register admission;
- a serial-yield register continuation existed; or
- an inline-call register continuation existed.

The authoritative continuation matcher and all entry, suspension, return,
error, and fallback behavior remained unchanged. Existing register native and
inline call/resume, member, call-name, return, and slotless-return guards
passed. The candidate added no semantic family, named nominal/container rule,
application rule, or benchmark branch.

After full removal, the restored `TestBytecode` family passes in 24.017s.

## Three-way five-process gate

Each application received five order-rotated processes for the ordinary
baseline, pre-hoist frame-reuse engine, and hoisted engine. Arithmetic means
are used. Positive deltas are regressions.

| Application | Baseline mean | Pre-hoist mean | Hoist mean | Hoist vs pre-hoist | Hoist vs baseline |
| --- | ---: | ---: | ---: | ---: | ---: |
| Future Pipeline | 341,767,652 ns | 355,661,332 ns | 326,920,341 ns | -8.08% | -4.34% |
| Concurrent Text Index | 337,222,132 ns | 328,056,193 ns | 330,223,872 ns | +0.66% | -2.08% |
| Word Frequency | 1,096,718,183 ns | 1,264,009,744 ns | 1,175,304,129 ns | -7.02% | +7.17% |

The intended mechanism is visible: the two probe-heavy applications improve
7%-8% relative to the pre-hoist engine, while the lower-density Concurrent
Text workload is neutral. Allocation counts remain unchanged.

Because Word Frequency was volatile and was the deciding application, it
received a second order-balanced five-pair baseline/hoist batch. Across all ten
processes, baseline averages 1,101,354,229 ns and the hoist averages
1,164,731,082 ns: a 5.75% regression. Allocations average 637,163 versus
637,170 per call, confirming that the result is residual execution cost rather
than restored register-frame allocation.

## Interpretation

The continuation probe was a real, generic control-flow tax, and the hoist
removed it correctly. Together with the preceding frame-reuse result, two
large fixed costs of the separate whole-function executor have now been
eliminated. The engine still loses a major unlike application whose admitted
functions average roughly three semantic instructions.

That closes this whole-function register route for short functions on the
current corpus. Adding another semantic family or continuing to shave its
entry/return plumbing would optimize an architecture that has twice failed the
application gate. The ordinary bytecode VM remains the performance target.

## Next recommendation

Run a restored-VM main-dispatch line/opcode attribution gate across Future
Pipeline, Concurrent Text Index, and Word Frequency.

Why: current clean baseline profiles put `runResumable(...)` at 13.26%, 3.09%,
and 6.27% flat respectively, but their visible child costs differ. The separate
register loop did not amortize on real short functions. The next opportunity
must therefore come from removing a repeated cost inside the existing
authoritative dispatcher, not from another executor layer.

What it entails: collect multiple merged clean CPU profiles plus separate
untimed opcode/line censuses from the restored VM; attribute the flat dispatcher
samples to exact source lines and opcode families; exclude already closed
stack-carrier and raw-integer micro-variants; and require the same exact guard,
lookup, or stack transition to be material in all three applications. Advance
only one generic ordinary-loop hoist/fusion, then use repeated arithmetic-mean
application A/B with established bytecode target guards. Continue to defer
WASM.
