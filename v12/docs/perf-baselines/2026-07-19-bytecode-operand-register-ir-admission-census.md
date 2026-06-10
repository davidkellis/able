# Bytecode operand-register IR admission census

Date: 2026-07-19

## Decision

Remove the opt-in operand-register IR prototype. Do not add a fused
store/branch/return family and do not attempt a narrow `CallName` translation.

The dynamic census verified all eight applications from the original timing
gate but observed 8,124,677 rejected function entries, zero translated
executions, zero semantic IR instructions, and zero removed representation
instructions. Consequently, none of the earlier apparent timing gains were
caused by the prototype. They were ordinary workstation variance around an
extra once-cached admission lookup.

The complete unsupported-set census also rejects a small opcode expansion.
`CallName` is the only sole unsupported family that occurs in at least three
unlike applications. It is a sole blocker in six applications and accounts
for 2,024,585 dynamic entries. Correct support is not a local semantic opcode:
it must preserve cached resolution, direct/native dispatch, inline frame
switching, return coercion, errors and loop signals, goroutine-executor yield,
and resumable state. A helper-style call inside the IR would silently lose
those contracts or disable the existing inline path. The other sole blockers
do not clear the three-application generality bar.

The prototype, environment switch, CLI census sink, and focused prototype
tests were removed. No compiler, canonical stdlib, benchmark, fixture,
language, or WASM change was made.

## Method

The temporary observer ran only when both the operand IR and its census sink
were explicitly enabled. At each non-resumable function entry it recorded:

- admission or whole-function rejection;
- the first unsupported opcode in bytecode order;
- every distinct unsupported opcode in the rejected function;
- whether one opcode was the function's sole unsupported family;
- dynamically executed semantic IR and removed representation instructions.

Semantic and representation counts were attached to executable IR
instructions after CFG translation, so loops would have counted their actual
dynamic iterations rather than a static instruction estimate. Focused tests
covered a straight-line expression, a loop, and complete-set rejection before
the observer was used.

Each application ran once as a complete process with `GOMAXPROCS=1`,
`GOGC=50`, `GOMEMLIMIT=1GiB`, its catalog working directory, declared input,
executor policy, source-root policy, the canonical external stdlib, and a
55-second cap. Every stdout capture passed the public Ruby verifier. This was
an event census, not a timing comparison; deterministic dynamic counts do not
benefit from averaging repeated identical runs.

## Admission result

| Application | Function entries | Translated | Rejected | Semantic IR | Removed representation |
| --- | ---: | ---: | ---: | ---: | ---: |
| Future Pipeline | 40,993 | 0 | 40,993 | 0 | 0 |
| Fixed Width 128 | 3,000,016 | 0 | 3,000,016 | 0 | 0 |
| Distance Field | 4,000,009 | 0 | 4,000,009 | 0 | 0 |
| Mandelbrot | 80,018 | 0 | 80,018 | 0 | 0 |
| Array Slice Window | 36,014 | 0 | 36,014 | 0 | 0 |
| Option/Result Config | 293,789 | 0 | 293,789 | 0 | 0 |
| Word Frequency | 673,776 | 0 | 673,776 | 0 | 0 |
| Reverse Complement | 62 | 0 | 62 | 0 | 0 |
| **Total** | **8,124,677** | **0** | **8,124,677** | **0** | **0** |

The leading first-unsupported counts were `MemberAccess` (32,783) in Future,
`LoadSlotStructField` (2,000,002) in Fixed Width, `CallName` (2,000,000) and
`JumpIfBinaryCompareFalse` (2,000,000) in Distance Field,
`StoreSlotFloatAffine` (80,000) in Mandelbrot, `CallName` (12,003) in Array
Slice, `JumpIfNotNil` and `JumpIfNotTypedPattern` (73,728 each) in
Option/Result, `CallMemberArrayGet` (407,324) in Word Frequency, and `Cast`
(32) in Reverse Complement. This variation demonstrates why first-opcode rank
alone is insufficient: it is ordered by bytecode position and says nothing
about later blockers in the same function.

## Whole-function unlock rank

| Sole unsupported family | Dynamic entries | Unlike applications | Decision |
| --- | ---: | ---: | --- |
| `CallName` | 2,024,585 | 6 | Broad, but requires the complete call/continuation contract; not a small family |
| `DefineFunction` | 5 | 5 | Startup definition work, immaterial |
| `Propagation` | 2 | 2 | Below breadth bar |
| `StructLiteralNamedFast` | 1,000,004 | 1 | Below breadth bar |
| `CallMemberArrayGet` | 389,345 | 1 | Below breadth bar |
| `CallMember` | 35,958 | 1 | Below breadth bar |
| `ReturnBinaryIntAddI32` | 6,432 | 1 | Below breadth bar |

The complete unsupported-opcode counts reinforce the architectural result:
`CallName`, fused comparison branches, and named struct literals each occur
in all eight applications, but generally coexist in the same rejected
functions. Adding one of the latter families would still translate no such
function. Only call-name support crosses the whole-function unlock bar.

## Correctness and cleanup

After removing the prototype:

- focused call-name, inline direct-slot, return, and CLI bytecode-statistics
  tests pass;
- `go test ./pkg/interpreter -run 'TestBytecode' -count=1 -timeout 60s`
  passes in 25.666 seconds;
- split tree-walker/bytecode exec-fixture parity groups 02-04, 05-08, 09-11,
  12-13, and 14 pass in 11.522, 29.040, 7.466, 12.634, and 20.486 seconds;
- every command remains below the one-minute project ceiling.

Temporary binaries, stdout/stderr captures, and census JSON are not retained
in the repository. The compact tables above preserve the decision evidence.

## Next recommendation

Profile the existing ordinary VM call boundary in the six applications where
`CallName` alone blocked an otherwise translatable function, then look for one
shared cost inside the current call machinery rather than reviving a separate
IR executor.

Why: the census shows that calls are the cross-program architectural wall, but
it does not say whether the reusable cost is lookup validation, argument
transport, callee frame construction, inline return restoration, or return
coercion. Those operations already preserve exceptions and scheduler state;
improving them benefits ordinary programs immediately and avoids duplicating
the VM's continuation semantics.

What it entails: take bounded one-process CPU/allocation profiles for Fixed
Width, Distance Field, Mandelbrot, Option/Result, Word Frequency, and Reverse
Complement; reconcile samples with existing call-dispatch counters; require
the same concrete helper or frame operation to be material in at least three
unlike applications; implement at most one generic existing-VM candidate;
then run focused semantics, split parity, and repeated verifier-backed A/B
averages. Continue to defer WASM.
