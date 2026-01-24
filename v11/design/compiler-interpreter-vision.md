# Compiler + Interpreter Vision (v11)

This document captures the long-term execution strategy: fast compiled programs with dynamic fallbacks, plus a faster interpreter. It should stay consistent with `spec/full_spec_v11.md` and the runtime semantics.

## Goals
- Keep full Able expressiveness (dynamic interface values, metaprogramming, concurrency).
- Compiled output should be close to host performance (Go first, then other targets).
- Interpreted execution should move toward the performance of modern interpreted languages.
- A single semantic core must drive both interpreter and compiler backends.

## Recommended execution model
- Build a typed core IR from the AST.
- Two backends:
  1) Bytecode VM interpreter (fast, portable).
  2) Host-language codegen (Go first), emitting calls into a shared runtime library.
- Compiled artifacts bundle the runtime and (optionally) the VM for dynamic fallbacks.
- Dynamic features route to runtime or VM entry points when static compilation is not possible.

## Interface dispatch in compiled + interpreted code
- Interface values carry dictionaries (see `v11/design/interface-dispatch-dictionaries.md`).
- Static code uses direct calls where the concrete type is known.
- Dynamic interface calls use dictionary dispatch; default methods can be inlined or invoked through the dictionary entry.
- Dictionaries are constructed at interface coercion time and can be cached by the runtime.

## Interpreter modernization direction
- Move from tree-walking to bytecode or SSA-based VM:
  - Lower AST -> typed IR -> bytecode.
  - Use inline caches for member/method lookups and dictionary dispatch.
  - Keep value representations shared with the runtime to reduce conversions.
- Start with a stack-based bytecode to minimize compiler complexity, then consider register-based once stabilized.
- Maintain determinism for concurrency semantics (`spawn` / `Future` handles) while improving throughput.

## Compiler direction (Go first)
- Emit Go code from typed IR.
- Generate explicit runtime calls for:
  - interface coercion + dictionary dispatch
  - dynamic metaprogramming (expr-eval)
  - concurrency scheduling (`spawn`)
- Keep the runtime ABI stable so the interpreter and compiler stay in lockstep.

## Immediate next steps
- Define a typed core IR and document it.
- Prototype a minimal bytecode VM (values, calls, control flow).
- Add conformance tests that run on both tree-walker and VM backends.
- Sketch the Go codegen layer for a small subset of IR instructions.
