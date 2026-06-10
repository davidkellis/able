# Bytecode register-frame attribution gate

Date: 2026-07-21

## Decision

The register-cell allocation hypothesis is confirmed, and the allocation is
removable, but removing it does not make the whole-function operand-register
engine competitive. Reject and fully remove both the bounded frame-reuse
prototype and the recovered register-native `MemberAccess` engine.

Five order-balanced steady-state pairs per application still make Future
Pipeline 14.01% slower, Concurrent Text Index 4.87% slower, and Word Frequency
7.08% slower after frame reuse removes essentially all of the candidate's
extra allocation traffic. No VM, runtime, compiler, stdlib, application,
fixture, reference, scorecard, language, or WASM change is retained.

## Profile protocol

The exact candidate from the preceding member-access gate was recovered. A
candidate test binary was frozen, the source was restored, and a separate
baseline test binary was built. `BenchmarkBytecodeProgramRuntime` loaded and
warmed each application once, suspended memory sampling during setup, and
measured repeated `main()` calls only.

Profiles and measurements used the canonical external stdlib, the declared
executor, `GOMAXPROCS=1`, `GOGC=50`, a 1-GiB memory limit, no bytecode event
statistics, and a 60-second per-process timeout. Future and Concurrent Text
used the goroutine executor; Word Frequency used the serial executor.

## Exact allocation attribution

Exact one-call memory profiles at `MemProfileRate=1` identify the same flat
owner in every candidate process:
`(*bytecodeVM).execBytecodeOperandRegisterIRFrom`. Its flat allocation-object
counts equal the previously measured dynamic register-entry counts.

| Application | Register entries | Baseline allocations | Candidate allocations | Delta | Delta bytes | Bytes / entry |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Future Pipeline | 32,783 | 655,833 | 688,627 | +32,794 | +6,036,416 | 184.13 |
| Concurrent Text Index | 66,926 | 377,315 | 444,253 | +66,938 | +13,000,504 | 194.25 |
| Word Frequency | 35,959 | 637,156 | 673,102 | +35,946 | +7,908,112 | 219.92 |

The allocation delta is 1.000 objects per register entry after rounding in all
three applications. The small object-count differences from exact equality
are ordinary surrounding runtime noise. Clean candidate CPU profiles also
place the register executor at 5.86%, 11.35%, and 2.81% cumulative CPU in
Future, Concurrent Text, and Word Frequency respectively. This clears the
predeclared same-owner/materiality gate.

## General frame-reuse prototype

The bounded prototype was VM-owned rather than application- or type-owned. It:

- reused cleared register-cell backing slices across whole-function entries;
- bounded each VM's free list to eight frames;
- transferred frame ownership through inline-call and serial-yield
  continuations;
- returned frames on ordinary return and errors;
- cleared abandoned continuations during non-yield unwinding and VM reset;
- preserved cold allocation when no suitably sized frame existed.

Focused register, member, call-name, inline continuation, return, and slotless
return tests passed. The prototype added no named nominal/container rule and no
benchmark-specific branch.

## Order-balanced A/B

The frame-reuse binary and untouched baseline binary received five independent
steady-state processes per application. Pair order alternated, and the table
uses arithmetic means as required for the shared workstation. Positive time
deltas are regressions.

| Application | Baseline mean | Reuse mean | Time delta | Baseline B/op | Reuse B/op | Baseline allocs/op | Reuse allocs/op |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Future Pipeline | 290,942,903 ns | 331,689,799 ns | +14.01% | 14,385,736 | 14,400,822 | 655,796 | 655,829 |
| Concurrent Text Index | 315,170,125 ns | 330,518,399 ns | +4.87% | 33,324,326 | 33,493,272 | 377,322 | 377,364 |
| Word Frequency | 1,134,433,598 ns | 1,214,769,849 ns | +7.08% | 48,417,048 | 48,529,931 | 637,158 | 637,178 |

Compared with the original candidate's exact profiles, the roughly one
allocation and 184-220 bytes per register entry are gone. The small residual
allocation differences are 20-42 objects across 377,000-655,000 total objects
and do not track the old entry counts. Allocation/zeroing was therefore a real
cost, but it was not the reason the register executor lost broadly.

## Interpretation

The remaining regression is architectural control overhead: admission/resume
probing in the ordinary interpreter loop, transition into a second dispatch
loop for approximately three semantic instructions, and reconciliation back
through the existing return machinery. This is an inference from the
allocation-neutral A/B plus the short admitted functions; it is not yet a
license to optimize a guessed helper.

Adding `StructLiteralNamedFast`, `CallMember`, or another semantic family now
would place more code behind the same unresolved entry/exit wall. Retaining
frame reuse without the register engine would have no consumer. Both are
therefore removed. The restored full `TestBytecode` family passes in 25.029s.

## Next recommendation

Run one bounded residual control-flow attribution gate with the same
allocation-neutral frame-reuse build across Future Pipeline, Concurrent Text
Index, and Word Frequency.

Why: allocation is now ruled out as the dominant cause, while clean profiles
already suggest that admission/resume checks, second-loop dispatch, and return
handoff are the remaining generic fixed costs. We need to identify which exact
control owner repeats before undertaking a deeper VM integration.

What it entails: recover the frame-reuse build, collect clean CPU profiles
against the frozen baseline, and use untimed counters to separate admitted
entry transitions, unadmitted fallback probes, continuation probes, semantic
dispatch, and return reconciliation. Normalize each owner by both total
bytecode instructions and register entries. Advance only if the same exact
non-semantic owner is material in all three applications; then test one generic
hoist/fusion that removes a redundant ordinary-loop transition. Re-run the
six-application reach and order-balanced A/B gates before adding any semantic
family. Continue to defer WASM.
