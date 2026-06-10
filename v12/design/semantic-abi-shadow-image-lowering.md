# Semantic ABI shadow image lowering

Date: 2026-07-22

Status: retained; shared value/heap conformance contract next

## Decision

**`retain-execution-complete-shadow-lowering`**.

The stage-one semantic images now carry complete typed data flow and control
flow for three unlike whole functions without executing those images. Retain
the flow schema and lowerer. No tree-walker, bytecode VM, compiler, runtime
value, stdlib, benchmark source, production dispatch, application output,
foreign heap, dependency, executable memory, or WASM path changed.

This establishes that the codec can represent the selected functions as real
program graphs rather than preorder AST records. It does not establish runtime
value/heap behavior or execution performance.

## Schema evolution

The internal image is now format/semantic version 2. The eight-section shape is
unchanged, while the package/function section adds:

- a type ID for every function-local register;
- each function's register range;
- stable call-target records containing target kind, package, owner type, name,
  arity, and resolved return type when semantic metadata provides one.

The generated manifest retains the 31 exhaustive value kinds and grows from 23
stage-zero structural records to 38 total operations. Fifteen new generic flow
operations cover constants, globals, moves, unary/binary values, casts, member
reads, calls, type tests, jumps, branches, returns, raises, match failure, and
host-effect resume. Current bytecode opcode numbers and fused source shapes
remain private.

Each operation declares fixed operand classes, an optional variadic class,
written-register positions, and whether it terminates a block. The manifest
identity includes all of those contracts.

## Validator contract

A function receives `FunctionFlagFlowValidated` only after validation proves:

- its register and instruction ranges are in bounds;
- every register has a valid type ID;
- every call target, symbol, constant, type, block, capability, and source ID is
  in range;
- every call's variadic argument count equals its target arity;
- result-register types agree with instruction and call-result types, allowing
  only explicit `dynamic` joins;
- encoded blocks are dense, ordered, non-empty, and end in a declared
  terminator;
- no instruction follows a terminator;
- every declared block is reachable from entry;
- branch, jump, and host-resume successors are valid;
- every register read is definitely assigned on every predecessor path;
- type-test destinations and branch conditions are boolean-compatible.

Corruption tests independently reject missing terminators, unreachable blocks,
undefined register reads, call-arity mismatch, and result-type mismatch with
deterministic errors.

## Generic lowering rules

The flow lowerer recursively preserves Able evaluation order and emits:

- mutable local registers and typed temporaries;
- explicit literal constants and global loads;
- real loop headers, exits, backedges, condition branches, and merges;
- stable local/imported/member call identities;
- typed-pattern decision blocks, clause bindings, non-exhaustive match failure,
  and raise termination;
- host-effect terminators containing result destination, capability, result
  type, exact continuation block, and argument registers;
- implicit and explicit return terminators.

Imported and member-call return types remain `dynamic` unless ordinary semantic
metadata supplies a result type. The lowerer does not infer return types from
names such as `HashMap`, `Array`, `len`, `read_slot`, `new`, or `zero`. Resolved
extern/native metadata enters through a generic fully-qualified host-function
map; the current proof supplies `able.math.hypot -> f64`, while `print` is the
language/kernel host boundary.

This separation is important: the proof contains no named-container or non-
primitive nominal lowering rule.

## Representative graphs

| Application | Function | AST nodes | Instructions | Registers | Blocks | Call targets | Host effects | Image bytes |
| --- | --- | ---: | ---: | ---: | ---: | ---: | --- | ---: |
| Fixed Width 128 | `ordered_select_checksum` | 70 | 44 | 32 | 9 | 3 | none | 2,877 |
| Distance Field | `main` | 89 | 67 | 47 | 14 | 0 | `able.math.hypot`, `able.host.print` | 4,036 |
| Array Slice Window | `rolling_checksum` | 147 | 81 | 62 | 16 | 3 | none | 4,813 |

The Fixed Width graph resolves imported `UInt128.zero`, imported `UInt128.new`,
and local `words_less`. The Array graph resolves `len`, `slice`, and `read_slot`
as member identities owned by the parameter/pattern type `Array<i32>`, while
leaving their result types dynamic. Its graph contains two typed tests, explicit
match failure, and one raise terminator. Distance Field contains two exact host
effect/resume boundaries. Every graph contains conditional branches and loop
backedges.

All 306 canonical AST nodes are accepted with zero fallback. Repeated encodes
and decode/re-encodes are byte-identical. Checked evidence:
`v12/docs/perf-baselines/2026-07-22-semantic-abi-shadow-image-lowering.json`.

## Remaining boundary

The images are deliberately not executed. Registers currently describe values,
but no shared heap exists behind heap/host cells. Dynamic calls describe exact
identity and arity without yet proving dispatch against shared definitions.
There is no live root stack, object table, aliasing, mutation, cycle tracing,
closure environment, iterator state, structured Error object, or host-registry
lifetime implementation.

The next change must therefore address value ownership rather than add more
instructions or another backend.

## Next recommendation

Complete **`shared-value-heap-conformance-contract`** before migrating any live
`runtime.Value`.

Define generated, backend-neutral object layouts and trace/identity/mutation
contracts for all 20 shared-heap kinds and registry/lifetime contracts for the
four host kinds. Specify root frames, generation reuse, stale-ID failure,
cycles, environments/closures, definitions/types, arrays/maps, interfaces/
unions, errors, iterators, and host-held roots. Build deterministic model-level
conformance vectors shared by future Go bindings and any engine, but do not yet
replace production values.

Why: data/control representation is no longer the blocker. The largest
remaining architectural risk is whether one heap can preserve Able identity,
aliasing, mutation, cleanup, and host lifetimes across all execution modes.
Locking that contract down before a production migration prevents a second
runtime from emerging and provides the parity oracle required for the eventual
tree-walker/VM/compiler conversion.

Retain that next tranche only if every heap/host kind has exactly one owner and
trace/lifetime rule, cycles and stale IDs fail safely, identity/mutation tests
cover unlike values, and no named nominal/container special case or execution
dispatch change is introduced.

Exclusions remain: no foreign heap, cgo runtime, JIT/backend dependency,
executable memory, benchmark branch, named-container/non-primitive-nominal
lowering rule, or WASM.

## Reproduction

```sh
just bench-semantic-abi-flow-check
just bench-architecture-budget-check
just bench-evidence-ledger --check
```
