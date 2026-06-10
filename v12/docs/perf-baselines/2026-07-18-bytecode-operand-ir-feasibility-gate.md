# Bytecode operand-IR feasibility gate

Date: 2026-07-18

## Decision

Admit an executable operand-addressed bytecode-IR prototype as the next VM
architecture tranche. Retain no production VM, compiler, runtime,
canonical-stdlib, benchmark, fixture, or language change from this static
gate.

All eight unlike verifier-backed applications clear the predeclared 15%
dynamic-instruction-removal threshold. The smallest modeled reduction is
30.09% and the largest is 44.32%, so the result is not dependent on one
numeric kernel, source sequence, container, regex engine, or nominal type.

This is an instruction-count feasibility result, not a timing prediction.
Translation, register access, materialization, larger instructions, branches,
and Go dispatch costs may consume part of the theoretical gain. An executable
A/B remains required before any path becomes the default.

## Static translation contract

A temporary table described all 143 current bytecode opcodes as one of:

- operand transport;
- fixed-operand semantic work;
- `argCount` or receiver-plus-`argCount` semantic work;
- a materialization/effect boundary; or
- a control-flow boundary.

Focused tests failed if any opcode lacked a descriptor and checked variadic
member-call translation across a return boundary. Separate boxed-value and
raw-i32 stack effects were represented.

The modeled register IR removes only these representation instructions:

- `Const` and `ConstI32`, represented as immediate operands;
- `LoadSlot` and `LoadSlotI32`, represented as direct register/slot operands;
- `Dup`, represented by multiple references to one operand; and
- `Pop`, represented by a dead/discarded result destination.

Every lookup, call, member/index operation, generic or specialized arithmetic
operation, store/coercion, allocation, identity-bearing value construction,
raise/rescue/ensure operation, branch, return, yield, spawn/await operation,
and unknown/dynamic action remains one semantic IR instruction. Values at
effect and control boundaries remain ordinary runtime values; the model does
not scalar-replace, pool, replay, or bypass them. It also does not count a
store-to-destination fusion, so that possible reduction is outside the gate.

## Dynamic census

Each process used the ordinary bytecode CLI and existing main-only opcode
statistics, canonical external stdlib, catalog run directory and arguments,
CPU 0, `GOMAXPROCS=1`, `GOGC=50`, `GOMEMLIMIT=1GiB`, and a 55-second cap.
Every output passed its public Ruby verifier. These are deterministic event
counts, not timing samples, so one complete process per application is the
appropriate evidence unit.

| Application | Dynamic instructions | Transport removed | Share | `LoadSlot` | `Const` | `Pop` | Other transport |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Future Pipeline | 9,650,629 | 3,940,508 | **40.83%** | 2,203,692 | 565,289 | 1,155,134 | 16,393 |
| Fixed Width 128 | 60,446,436 | 19,221,015 | **31.80%** | 8,283,481 | 5,875,015 | 5,062,519 | 0 |
| Distance Field | 78,000,165 | 32,000,082 | **41.03%** | 22,000,024 | 8,000,029 | 2,000,029 | 0 |
| Mandelbrot | 260,984,057 | 90,471,490 | **34.67%** | 65,793,181 | 4,335,967 | 20,342,342 | 0 |
| Array Slice Window | 8,751,239 | 2,869,097 | **32.79%** | 2,004,545 | 372,531 | 480,020 | 12,001 |
| Option/Result Config | 3,487,172 | 1,228,520 | **35.23%** | 588,505 | 240,604 | 381,267 | 18,144 |
| Word Frequency | 17,591,889 | 7,797,373 | **44.32%** | 4,406,730 | 924,608 | 2,465,647 | 388 |
| Reverse Complement | 60,602,670 | 18,234,283 | **30.09%** | 14,100,754 | 66,765 | 4,066,764 | 0 |

`Other transport` is `Dup`, `ConstI32`, and `LoadSlotI32`. The complete output
SHA-256 values are `3db937fd...d98`, `eceabf58...56a`, `cdaaf445...a94`,
`d1560cad...73e`, `155f8912...03e`, `28e46b27...112`, `7dc1dae3...d07`,
and `db06a593...bb7`, respectively.

## Interpretation

The previous typed-block gate reached 15% in only Mandelbrot and Future
Pipeline because it materialized at calls and dynamic operations before
absorbing their operands. This whole-function model keeps those operations as
semantic instructions but gives them explicit operands and results. That is
the architectural distinction that exposes broad transport reduction:

- numeric and float applications are 31.80%-41.03%; and
- concurrency, wide nominal arithmetic, Array/iterator, union/control,
  text/map, and primitive-byte applications are all 30.09%-44.32%.

The repeated opportunity is therefore the VM's general stack representation,
not one downstream semantic helper. The static gate is strong enough to
justify implementation work, but not to skip an executable broad benchmark
gate.

## Correctness and cleanup

- The temporary opcode table covered all 143 opcodes and its focused tests
  passed.
- All eight measured applications verified under the normal public contracts.
- Focused load/call/binary/store/return tests pass after removing the model.
- The descriptor table, translator, tests, census binary, full snapshots, and
  stdout/stderr captures are removed after this record.
- No canonical `able-stdlib` source needed a change.
- No WASM work was performed.

## Next recommendation

Build an opt-in executable operand-IR prototype that actually replaces the
modeled stack instructions rather than wrapping the existing dispatcher.

Why: the static gate now supplies broad architectural reach, while the earlier
typed-block executor failed precisely because it retained the original stack
operations inside a second dispatcher. Only execution can determine whether
30%-44% fewer dispatches outweigh translation, register, branch, and
materialization overhead in Go.

What it entails: introduce a separate operand instruction representation with
explicit source operands and destination registers, translate each function
once with CFG stack maps/merge validation, and reuse the existing semantic
helpers for calls, dynamic operations, allocation, errors, concurrency, and
returns. Unsupported or unproved functions must use the existing bytecode VM
for the entire function initially; do not mix partially executed effects.
Keep the path opt-in until focused parity tests and repeated preserved-binary
A/B runs show improvement in at least three unlike applications without a
material regression across the broader selected suite. Do not add source,
benchmark, Array, regex, or named-nominal recognition.
