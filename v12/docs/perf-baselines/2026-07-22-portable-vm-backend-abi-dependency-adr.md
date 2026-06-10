# Portable VM backend ABI and dependency decision

Date: 2026-07-22

Status: accepted architecture closure

## Decision

**`close-portable-foreign-backend-under-current-runtime-contract`**.

Do not build a C-ABI engine, adopt a JIT dependency, or add direct machine-code
emitters against the current Go-owned bytecode/runtime object graph. No backend
variant satisfies the semantic, ownership, fallback, deployment, and prior-
rejection gates together, so no executable prototype or dependency is admitted.

This does not say native execution is too small. The predecessor model shows
compiled-equivalent reach clears the materiality gate in 5
of 5 phase-one families. It says the current
Go object graph is not a portable execution ABI, and selecting a code generator
does not solve that semantic boundary.

## Context

- **program-is-go-object-graph**: bytecodeInstruction embeds runtime.Value, strings, AST types/nodes, nested program pointers, and Go-managed metadata; bytecodeProgram embeds maps, slices, and runtime definition pointers.
- **value-is-go-object-graph**: runtime.Value implementations embed Go strings, slices, maps, interfaces, functions, AST pointers, environments, mutexes, contexts, condition variables, futures, and arbitrary host values.
- **execution-state-is-go-owned**: VM frames, environments, lookup caches, iterator and loop stacks, unwind state, Array ownership, raw sidecars, and async payloads are Go object graphs.
- **concurrency-is-go-runtime-backed**: production spawn uses goroutines, contexts, channels/conditions, and Go-owned Future state; deterministic execution resumes exact Go VM state after serial yields.
- **externs-are-go-runtime-backed**: extern functions are compiled as exact-toolchain Go plugins and exchange runtime.Value/NativeCallContext state.
- **cgo-handles-do-not-create-a-value-abi**: integer handles can keep Go values rooted across C, but each semantic operation still returns to Go and foreign memory cannot traverse the referenced Go object graph.

Go/C handles are suitable opaque roots, not a foreign value model. Foreign code
cannot retain or traverse the interface-, slice-, map-, function-, context-, and
pointer-bearing graphs represented by `runtime.Value`, `bytecodeProgram`, frames,
environments, or futures. Keeping those values Go-owned therefore requires an
exit or callback for the semantic operations that dominate real programs.

## Measured boundary density

| Application | Hot program share | Primitive share | Longest primitive span | Effect instructions in hot program | Contract eligible |
| --- | ---: | ---: | ---: | ---: | --- |
| `fixed_width_128` | 46.74% | 34.38% | 7 | 15125029 | no |
| `distance_field` | 56.41% | 35.90% | 4 | 22000058 | no |
| `concurrent_event_routing` | 38.50% | 35.16% | 4 | 5159680 | no |
| `word_frequency` | 40.23% | 35.09% | 4 | 4428635 | no |
| `array_slice_window` | 63.49% | 34.71% | 6 | 3564014 | no |
| `reverse_complement` | 53.41% | 43.40% | 4 | 18200223 | no |

No hot function is eligible, and the longest static primitive span is only 7
instructions. A safe Go-owned-root design would repeatedly cross the foreign
boundary for calls, collections, allocation, lookup, nominal operations, errors,
unwind, suspension, and externs. That is the completed native-leaf rejection,
not a whole-engine execution design.

## Alternatives

| Class / ownership | No per-op callback | One semantic owner | Identity without graph conversion | Exact fallback | Concrete distribution | Prior rejection? | Decision |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `whole-engine-c-abi` / `go-owned-values-with-effect-exits` | no | yes | yes | yes | no | yes | reject |
| `whole-engine-c-abi` / `foreign-owned-complete-runtime` | yes | no | no | no | no | no | reject |
| `portable-jit-library` / `jit-with-go-owned-runtime` | no | yes | yes | yes | no | yes | reject |
| `portable-jit-library` / `jit-with-foreign-runtime` | yes | no | no | no | no | no | reject |
| `direct-machine-code-generation` / `direct-codegen-with-go-owned-runtime` | no | yes | yes | yes | no | yes | reject |
| `direct-machine-code-generation` / `direct-codegen-with-foreign-runtime` | yes | no | no | no | no | no | reject |

### Whole-engine C ABI

With Go-owned values, C receives opaque handles and must return to Go for real
semantics; this repeats the dense side-exit architecture. With a C-owned heap, it
can remain native but must independently implement values, GC, environments,
dispatch, errors, cleanup, scheduling, futures, and externs. That creates two
semantic authorities and cannot fall back mid-activation without graph conversion.

### Portable JIT library

A JIT changes instruction selection and register allocation, not value ownership.
It therefore inherits the same safe-leaf limit with Go-owned values or the same
duplicate-runtime problem with a foreign heap, while adding dependency, license,
compile-latency, executable-memory, cache, and supported-platform decisions.

### Direct machine-code generation

Direct emitters inherit the same runtime fork. They additionally require per-OS/
architecture calling conventions, unwind behavior, instruction-cache handling,
W^X memory, code lifetime, and security maintenance. Removing a library does not
remove these obligations.

## Consequences

- Ordinary Go bytecode remains the only VM and exact fallback.
- No C/C++/Rust/JIT dependency, foreign heap, executable-memory path, or platform
  support promise is added.
- No `runtime.Value`, Go ABI, AST pointer, or current bytecode layout becomes a
  serialized/public ABI.
- Existing tree-walker/bytecode parity, source diagnostics, concurrency behavior,
  extern behavior, and established performance guards remain unchanged.
- Reopening one of these backend classes requires a backend-neutral semantic ABI
  that invalidates this decision; a different code generator alone does not.

## Next recommendation

Complete **`shared-runtime-semantic-abi-feasibility`**.

Determine whether Able can define one backend-neutral serialized program/value/effect ABI that both the Go reference runtime and a future fast engine can implement without per-operation foreign calls or duplicate language semantics. This is a feasibility/design tranche, not a runtime rewrite.

Required outputs:

- an immutable serialized instruction/type/source-identity format with no Go pointers
- a value, identity, mutation, root, and lifetime model covering every runtime.Value kind
- semantic operation and structured effect/resume contracts for dispatch, errors, cleanup, suspension, and externs
- an ownership decision for heap, GC, environments, futures, and host handles
- a migration and conformance plan preserving tree-walker and ordinary-bytecode fallback
- a target-budget model for at least three unlike whole hot functions

Why: the performance model says whole effectful functions are large enough, but
the current Go representation prevents any foreign backend from owning them
safely and cheaply. The next question is therefore whether Able can have one
implementation-independent semantic ABI—not which machine-code library to call.

Admission: Advance only if one design keeps a single semantic authority, avoids per-opcode foreign transitions, preserves identity/fallback without whole-graph conversion, and models at least 25% target-excess reduction in three unlike applications. Otherwise close foreign/native backends and return to the portable application frontier.

Exclusions: No WASM, foreign runtime implementation, JIT dependency, executable-memory code, benchmark branches, named-container rules, or non-primitive nominal special cases.

## Reproduction

```sh
python3 v12/bench_portable_vm_backend_adr_test.py
v12/bench_portable_vm_backend_adr --check
just bench-evidence-ledger --check
```
