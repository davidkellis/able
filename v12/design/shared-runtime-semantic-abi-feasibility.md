# Shared runtime semantic-ABI feasibility

Date: 2026-07-22

Status: conditional feasibility; codec/layout spike admitted

## Decision

**`conditionally-feasible-admit-codec-layout-spike`**.

A backend-neutral ABI is feasible only as a shared semantic runtime, not as a
serialization veneer over the current Go object graph. The Go reference
interpreters, compiled applications, and any future fast engine must use the same
pointer-free value cells, stable object identities, heap, root protocol, semantic
operation manifest, and structured effect records.

This decision admits only a standalone pure-Go codec and layout spike. It does
not admit a foreign heap, runtime migration, backend, JIT dependency, executable
memory, or production dispatch change.

## Semantic authority and execution boundary

The v12 spec and canonical AST define language behavior. One shared semantic runtime owns value representation, heap objects, identity, mutation, dynamic dispatch primitives, type relations, and structured runtime errors. Generated bindings expose the same manifest to the Go reference interpreter, compiled applications, and a future fast engine.

A future engine and the shared semantic runtime execute in one native address space and call through an internal ABI, not Go/cgo per semantic opcode. Go is entered only for declared host effects. No backend is selected by this feasibility decision.

The Go tree-walker and ordinary Go bytecode VM remain behavioral references and use generated bindings over the same value cells and semantic runtime. Their performance is not used to justify the ABI; parity fixtures remain authoritative during migration.

The serialized ABI is internal and versioned. The current 144 live bytecodes—
especially fused source-shape instructions—remain Go implementation details and
are not frozen into the portable semantic instruction set.

## Pointer-free value cell

```text
AbleValueCell (16 bytes)
  tag     u32
  aux     u32
  payload u64
```

- The cell contains no Go pointer or native heap pointer.
- Primitive payload bits are stored directly; integer/float suffixes and flags use aux.
- Heap references use a generation-checked object identifier, never an address.
- Host references use a generation-checked host registry identifier and explicit retain/release.
- Zero is an invalid heap/host identifier and is safe as an uninitialized sentinel.

All 31 current `runtime.Kind` values map exactly once:

| Representation | Kinds | Count |
| --- | --- | ---: |
| Immediate or heap-overflow primitive | `KindBool`, `KindChar`, `KindNil`, `KindVoid`, `KindInteger`, `KindFloat`, `KindIteratorEnd` | 7 |
| Shared heap or immutable metadata | `KindString`, `KindArray`, `KindHashMap`, `KindHasher`, `KindFunction`, `KindFunctionOverload`, `KindStructDefinition`, `KindTypeRef`, `KindStructInstance`, `KindInterfaceDefinition`, `KindInterfaceValue`, `KindUnionDefinition`, `KindPackage`, `KindDynPackage`, `KindDynRef`, `KindError`, `KindBoundMethod`, `KindImplementationNamespace`, `KindIterator`, `KindPartialFunction` | 20 |
| Host-effect registry | `KindNativeFunction`, `KindHostHandle`, `KindNativeBoundMethod`, `KindFuture` | 4 |

Small primitive payloads live in the cell. Big integers and all identity-bearing
objects use generation-checked object IDs. Native functions, futures, and opaque
host objects use separate generation-checked registry IDs. Neither kind is a raw
Go or native pointer.

## Program image

Encoding: versioned canonical little-endian sectioned binary.

- header-and-semantic-abi-identity
- string-and-symbol-table
- type-and-nominal-definition-table
- source-file-span-and-callsite-table
- constant-and-immutable-object-table
- package-function-and-signature-table
- canonical-semantic-instructions-and-control-flow
- host-effect-import-and-capability-table

Instruction shape: `opcode:u16, flags:u16, source_id:u32, operands as validated u32 table/register/block indices`.

Validation and fallback rules:

- all section bounds, indices, opcodes, control-flow targets, types, arities, and capabilities validate before execution
- no AST node, Go interface, string header, slice, map, function, or pointer occurs in the image
- unsupported functions fall back before function entry; current source-shape fused opcodes remain backend-private
- a function is engine-eligible only when every node is lowered without AST evaluation fallback

This removes AST pointers from executable images. Match, rescue, ensure, calls,
definitions, and other currently AST-assisted behavior must lower to canonical
semantic instructions before a function becomes engine-eligible. Unsupported
functions stay in the ordinary Go VM from their entry point.

## Ownership

- **heap and gc**: The shared semantic runtime owns the object table/arena and tracing or equivalent collection. Object IDs are stable across movement; engine activations and Go bindings register explicit roots.
- **environments and closures**: Lexical environments, captures, overload sets, bound methods, type arguments, and mutable cells are shared-heap objects referenced by value cells.
- **definitions and types**: Immutable definition/type metadata lives in the validated program image or shared immutable tables and is addressed by stable IDs.
- **futures and scheduling**: Phase one keeps goroutines, executor scheduling, and Future state in the Go host registry. Spawn, await, cancellation, yield, channels, and mutexes are declared host effects and are excluded from the initial performance scope.
- **externs and host handles**: Native functions, externs, files, clocks, processes, and arbitrary host objects use capability-scoped registry IDs. The shared heap never contains a Go pointer.
- **root lifetime**: Go registries and engine activations explicitly retain/release shared cells. Host registry entries retain any shared cells reachable from host state; stale generation IDs fail deterministically.

The essential commitment is one canonical heap. Old Go identity-bearing values
and shared-heap identity values cannot coexist behind automatic conversion; that
would recreate the graph-conversion and aliasing failure from the backend ADR.

## Effects, control, and exact resume

Internal structured control remains inside the engine/runtime:

- return
- raise-rescue-ensure-rethrow
- propagation
- break-continue-and-cleanup
- iterator-close

Declared Go host effects:

- native-or-extern-call
- spawn-or-scheduler-operation
- await-or-yield-suspension
- channel-or-mutex-operation
- host-handle-or-operating-system-operation

Effect record: `effect_kind:u32, flags:u32, program_id:u64, function_id:u32, instruction_id:u32, source_id:u32, continuation_id:u32, argument-cell range, result destination, and structured error destination`.

The engine commits all prior effects once, returns before the host effect, and resumes the same continuation exactly once with a value or structured error. Ordinary-VM fallback is chosen before function entry; v1 never translates a live foreign frame into a Go frame.

Every instruction and effect references immutable source/callsite IDs so the existing diagnostic shape can be reproduced without retaining AST nodes.

Phase one excludes concurrency from the performance scope. Goroutines, futures,
channels, mutexes, cancellation, and yielding remain Go host effects until a
separate concurrency ABI clears its own parity and performance gates.

## Target-budget model

This is the existing optimistic compiled-equivalent hottest-function model, not
a timing promise. A shared heap/runtime is the missing condition that lets a fast
engine keep these whole effectful functions inside one native address space.

| Application | Family | Hottest function | Dynamic share | Target excess removed | Gate |
| --- | --- | --- | ---: | ---: | --- |
| `fixed_width_128` | wide-numeric | `ordered_select_checksum` | 46.74% | 48.57% | pass |
| `distance_field` | float-numeric | `main` | 56.41% | 59.77% | pass |
| `word_frequency` | text-map | `split` | 40.23% | 39.46% | pass |
| `array_slice_window` | array-iterator | `rolling_checksum` | 63.49% | 62.39% | pass |
| `reverse_complement` | byte-text | `reverse_complement_fasta` | 53.41% | 53.01% | pass |

All 5 serial/effectful rows clear the
25% materiality gate, satisfying the required three-unlike-family planning bar.

## Migration gates

| Stage | Scope | Production execution change? | Exit gate |
| ---: | --- | --- | --- |
| 0 | `codec-and-layout-spike` | no | Deterministic encode/decode, corruption rejection, exhaustive 31-kind mapping, source identity round trips, and three unlike function images pass without changing runtime execution. |
| 1 | `shadow-image-lowering` | no | Coverage reports prove complete non-AST lowering for selected whole functions and identify every host effect; ordinary execution remains unchanged. |
| 2 | `shared-value-and-heap-conformance` | yes | Tree-walker and ordinary bytecode parity cover identity, mutation, cycles, errors, cleanup, host roots, and GC before any fast engine runs. No old/new identity-bearing heap conversion remains. |
| 3 | `function-entry-engine-prototype` | no | Three unlike whole functions execute from the same image/heap, fall back before entry, preserve all guards, and remove at least 25% target excess in repeated averaged processes. |

## Next recommendation

Complete **`semantic-abi-codec-layout-spike`**.

Implement only a standalone pure-Go internal ABI package: the 16-byte pointer-free value cell, versioned program-image data structures, deterministic encoder/decoder, validator, source/callsite identity, and generated kind/op manifests. Shadow-encode three unlike existing whole functions without executing the images.

Why: the feasibility model now has an exhaustive value mapping, one ownership
model, exact function-entry fallback, and sufficient whole-function reach. The
smallest reversible way to test the design is the immutable data plane—layout,
encoding, validation, and source identity—before changing any live value or
execution path.

Governing applications: `fixed_width_128`, `distance_field`, `array_slice_window`.

Admission: Retain the spike only if all current runtime kinds map exactly once, malformed images fail deterministically, repeated encodes are byte-identical, all three functions lower without AST pointers/evaluation fallback, tests remain under one minute, and ordinary interpreter/compiler output is unchanged.

Exclusions: No foreign heap, cgo runtime, JIT/backend dependency, executable memory, production dispatch, runtime.Value migration, benchmark branch, named-container/non-primitive-nominal rule, or WASM.

## Reproduction

```sh
python3 v12/bench_shared_runtime_semantic_abi_test.py
v12/bench_shared_runtime_semantic_abi --check
just bench-evidence-ledger --check
```
