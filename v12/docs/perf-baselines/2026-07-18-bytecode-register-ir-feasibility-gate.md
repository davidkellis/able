# Bytecode register-IR feasibility gate

Date: 2026-07-18

## Decision

Reject an executable typed-block register-IR prototype and retain no bytecode
VM, runtime, compiler, canonical-stdlib, benchmark, fixture, or language
change. A temporary static translator plus dynamic census found enough
removable stack transport in only two unlike applications, short of the
predeclared three-application admission rule.

Mandelbrot and Future Pipeline clear the 15% dynamic-instruction threshold.
The next-best complete row is Fixed Width 128 at 6.46%, and the remaining five
complete controls are 0.11%-5.34%. Typed straight-line blocks are therefore a
real local opportunity, but not a broad enough VM architecture on the current
lowered programs.

## Conservative model

The temporary translator ran during bytecode-program finalization and marked
dispatches that a register-form block could erase while preserving the
existing semantic operations:

- `Const` and `LoadSlot` producers consumed by statically typed integer
  arithmetic or a local slot assignment;
- dead `Pop` instructions after a modeled local result or `StoreSlot`; and
- a materialized ordinary-stack value as a block input, without carrying a
  virtual value across the preceding boundary.

Jump targets restarted symbolic state. Calls, member/index/dynamic dispatch,
allocation and new-binding identity, generic unary/binary operations,
raise/rescue/ensure, yield/spawn/await, returns, unknown effects, and control
exits ended the modeled region. The arithmetic and store operations themselves
remained counted because a register VM would still execute their semantics;
only proven transport dispatches counted as removable. The model neither
scalar-replaced nominal values nor recognized a benchmark, source sequence,
Array, regex, or named type.

The first draft required both arithmetic operands to originate in the local
region. A focused test exposed that this undercounted valid blocks beginning
with one materialized input. The final model permits that input while marking
only locally supplied operand transport. Both versions were tested before the
final census; only the corrected model below is decision evidence.

## Dynamic census

Each complete process used the ordinary bytecode CLI, canonical external
stdlib, catalog run directory and arguments, CPU 0, `GOMAXPROCS=1`, `GOGC=50`,
`GOMEMLIMIT=1GiB`, main-only bytecode statistics, and a 55-second cap. Public
Ruby verifiers accepted every retained output. Counts are deterministic event
counts rather than timing samples, so one complete process per application is
sufficient.

| Application | Dynamic instructions | Modeled removals | Share | `LoadSlot` | `Const` | `Pop` |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Future Pipeline | 9,650,629 | 2,105,348 | **21.82%** | 1,048,576 | 532,484 | 524,288 |
| Fixed Width 128 | 60,446,436 | 3,904,057 | 6.46% | 841,547 | 2,000,000 | 1,062,510 |
| Distance Field | 78,000,165 | 2,000,072 | 2.56% | 24 | 24 | 2,000,024 |
| Mandelbrot | 260,984,057 | 65,297,146 | **25.02%** | 48,396,449 | 640,000 | 16,260,697 |
| Array Slice Window | 8,751,239 | 336,260 | 3.84% | 24,002 | 312,258 | 0 |
| Option/Result Config | 3,487,172 | 98,112 | 2.81% | 0 | 98,112 | 0 |
| Word Frequency | 17,591,889 | 938,577 | 5.34% | 644,347 | 141,083 | 153,147 |
| Reverse Complement | 60,602,670 | 66,677 | 0.11% | 2 | 33,339 | 33,336 |

The complete outputs have SHA-256 values `3db937fd...d98`,
`eceabf58...56a`, `cdaaf445...a94`, `d1560cad...73e`, `155f8912...03e`,
`28e46b27...112`, `7dc1dae3...d07`, and `db06a593...bb7`, respectively.
Regex Set was also attempted, but the full statistics observer reached the
55-second cap before producing output or a snapshot. Its incomplete run was
removed and supplies neither positive nor negative evidence.

## Interpretation

Mandelbrot's ordinary integer loop-control stack and Future Pipeline's numeric
workers retain enough short typed transport to benefit. Distance Field's hot
float work and several other numeric/Array loops are already represented by
fused opcodes, while nominal, text/map, iterator, and byte applications spend
their loads around semantic operations that the conservative model treats as
boundaries. Adding a second typed-block dispatcher would therefore repeat the
previous failure mode: it would help two shapes while leaving most application
instruction streams unchanged.

No executable baseline/candidate A/B is warranted because the static gate is
specifically intended to avoid paying that implementation cost without broad
reach.

## Correctness and cleanup

- Focused translator tests covered typed transport, materialized block inputs,
  hard boundaries, and jump-target restarts.
- Focused bytecode load/store, integer arithmetic, call-name, and return tests
  pass after the temporary instrumentation was removed.
- The translator, masks, counters, tests, census binary, full snapshots,
  stdout/stderr captures, and incomplete Regex run are not production state.
- No canonical `able-stdlib` source needed a change.
- No WASM work was performed.

## Next recommendation

Run one final static feasibility gate for an operand-addressed whole-function
register IR before closing the register architecture entirely.

Why: this tranche proves that typed straight-line regions alone are too narrow,
but the retained opcode censi still show broad `LoadSlot`, `Const`, and `Pop`
traffic around calls, member/index operations, generic arithmetic, and other
semantic boundaries. A normal register VM gives those existing operations
explicit operands and a result register; it does not need to remove or cross
the semantic operation to remove its surrounding transport dispatches.

What it entails: add a complete stack-effect/argument-count table and perform
static def-use translation across each function. Calls, dynamic operations,
allocation, errors, and concurrency remain ordinary semantic instructions and
materialization points, but may consume explicit slot/register/immediate
operands and write one explicit result. Count only original dispatches that
this representation makes unnecessary. Do not execute the IR, scalar-replace
identity-bearing values, or recognize an application/type/source pattern.
Admit implementation only if the same 15% threshold is reached in at least
three unlike verified applications; otherwise record that result and close
register-IR work in favor of a different architecture.
