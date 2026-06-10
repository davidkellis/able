# Bytecode register call/continuation ABI gate

Date: 2026-07-21

## Decision

Retain the call/continuation ABI design, but remove the executable opt-in
prototype. Do not run timing A/B cohorts from this slice.

The prototype correctly executed synchronous native `CallName` dispatch and
suspended/resumed an operand-register frame around an inline bytecode call.
It then reached 10,000,000 semantic register instructions in Distance Field
and 49,152 in Option/Result Config while preserving both public verifiers.
Those are only two materially reached unlike applications. Every other
verified application either admitted no function or executed only startup-
scale register work. The predeclared three-unlike-program reach gate therefore
failed before timing.

No compiler, VM, runtime, canonical stdlib, benchmark, fixture, language,
scorecard, or WASM performance change is retained.

## Register-frame ABI

The prototype established a usable compatibility contract with the ordinary
VM rather than duplicating call semantics:

1. A register frame owns its program, semantic PC, boxed/raw registers, and
   the destination register for a pending call result.
2. At `CallName`, explicit register operands are materialized once in source
   argument order. The existing `execCallName` remains authoritative for cache
   validation, direct/native dispatch, argument coercion, inline eligibility,
   frame construction, and generic fallback.
3. A synchronously completed call removes the ordinary VM result and places it
   directly in the destination register.
4. An inline call saves the caller register frame plus the expected return IP,
   then switches to the callee through the existing program-switch path.
   `finishInlineReturn` remains authoritative for return coercion, implicit
   receiver cleanup, sidecar/frame restoration, control stacks, ownership,
   and caller lookup-cache restoration. Its appended result is moved into the
   saved destination register before register execution resumes.
5. Ordinary runtime errors propagate through the existing unwinder. A serial
   executor yield preserves the current register continuation; non-yield exits
   discard it. Resumable entry may resume a saved register frame, but an
   unsupported fresh generator/function uses the unchanged stack VM before
   any effect occurs.

This boundary keeps reconciliation at calls and suspension points. It does not
reintroduce a boxed stack between semantic register instructions.

## Prototype scope and correctness

The recovered whole-function translator retained explicit boxed and raw-i32
operands for constants, slots, arithmetic, stores, branches, loops, and
returns. This tranche added `CallName`, inline register continuations, exact
native completion, return resumption, serial-yield state, admission/reach
counters, and the generic fused binary-comparison branch needed to test a
current lowering shape. Unsupported functions fell back cold before effects.

Focused tests proved:

- a native named call receives a materialized computed operand and returns to
  a register without leaving an operand-stack value;
- a caller register frame suspends around an inline bytecode function, lets
  the ordinary return path coerce/restore state, and resumes exactly once;
- the existing call-name, inline-return, minimal-return, slotless-return, and
  caller active-lookup-cache guards pass with the opt-in path enabled;
- the same focused guards pass after complete prototype removal.

## Dynamic reach census

Each reported process used the checked-in Able source, canonical external
stdlib, declared working directory/arguments, `GOMAXPROCS=1`, `GOGC=50`, a
1-GiB memory limit, a 55-second cap, and the public Ruby verifier. Concurrency
applications used the declared goroutine executor. Counts are deterministic
events, so repetition would not improve their evidentiary value. Bytecode
statistics materially slow hot dispatch and were used only for admission, not
wall-time claims.

| Application | Function attempts | Admitted | Register entries | Semantic IR | Inline suspend/resume | Result |
| --- | ---: | ---: | ---: | ---: | ---: | --- |
| Future Pipeline | 40,993 | 0 | 0 | 0 | 0 / 0 | verified |
| Fixed Width 128 | 3,000,016 | 1 | 5 | 5 | 4 / 4 | verified |
| Distance Field | 4,000,009 | 2,000,000 | 4,000,000 | 10,000,000 | 2,000,000 / 2,000,000 | verified |
| Mandelbrot | 80,018 | 4 | 8 | 11 | 4 / 4 | verified |
| Array Slice Window | 36,014 | 0 | 0 | 0 | 0 / 0 | verified |
| Option/Result Config | 293,789 | 24,576 | 49,152 | 49,152 | 24,576 / 24,576 | verified |
| Word Frequency | 673,776 | 1 | 1 | 2 | 0 / 0 | verified |
| Reverse Complement | 62 | 3 | 4 | 7 | 1 / 1 | verified |
| Concurrent Text Index | 117,493 | 1 | 1 | 2 | 0 / 0 | verified |
| Validated Job Pipeline | 182,030 | 0 | 0 | 0 | 0 / 0 | verified |
| Dependency Wave Validation | 120,703 | 0 | 0 | 0 | 0 / 0 | verified |
| Concurrent Event Routing | 763,196 | 1 | 1 | 2 | 0 / 0 | verified |
| Matrix Multiply | 4,014 | 0 | 0 | 0 | 0 / 0 | verified |
| Monte Carlo Pi | 13 | 1 | 11 | 11 | 10 / 10 | verified |

The fused comparison addition changed Distance Field's next rejection from
that branch to `StructLiteralNamedFast`, but did not increase its already
admitted hot count. It exposed 16,384 state/CFG rejections in Validated Job
Pipeline without admitting a function. Thus adding plausible first blockers
does not establish a third material application; merge state and later opcode
closure matter.

## Interpretation

The July 19 census is no longer a sufficient admission map. Many intervening
lowering and VM optimizations changed hot functions: current first blockers
include fused integer branches, member/Array calls, direct struct-field loads,
named struct literals, and typed-pattern returns. `CallName` is still a real
continuation boundary, and the ABI above works, but it is no longer the sole
coverage key across three current unlike applications.

The result also confirms why timing must follow reach. A plan lookup could be
timed in every process, while nearly all selected applications would still run
the ordinary VM. Such deltas would again measure workstation variance and
admission overhead rather than register execution.

## Next recommendation

Refresh the complete current unsupported-set and state/CFG admission census
before implementing another executor slice.

Why: first-blocker counts were insufficient here. Distance Field and
Option/Result prove the call ABI can carry hot work, while Validated Job proves
that removing its visible opcode blocker still leaves an unclassified merge or
state failure. We need the minimum complete coverage closure shared by at least
three unlike applications, not another guessed opcode.

What it entails: make the translator report every unsupported semantic family,
the precise transfer failure, and incompatible merge states for each rejected
function; weight those sets by dynamic invocation across the current selected
suite; compute the smallest cross-application closure that would admit three
material workloads; and compare that closure with a register-native lowering
design. Only then implement a new cold-fallback vertical slice, rerun reach,
and—if the three-program bar passes—run repeated arithmetic-mean A/B cohorts
plus all established guards. Continue to defer WASM.
