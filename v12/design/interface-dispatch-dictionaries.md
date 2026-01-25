# Interface Dispatch via Dictionaries (v11)

This document captures the decision to use dictionary dispatch for interface values, plus the resulting typechecker/runtime behavior and follow-ups.

## Goals
- Allow method calls on interface values, including default interface methods and generic interface methods.
- Preserve expressiveness (no object-safety restriction for interface method calls).
- Keep behavior consistent across interpreters and future compilers.

## Decisions
- Interface values use dictionary dispatch, not vtables.
- An interface value carries: the concrete value, the interface name, interface type arguments, and a method dictionary.
- Dictionary entries are bound to the receiver and can point to impl methods or interface default methods.
- Dictionaries include methods from base interfaces (interface aliases).
- Generic interface methods are resolved at call time; inferred type arguments are stored on the call node when omitted.
- Interface upcasts are allowed only when the target interface is in scope and implemented by the concrete type.

## Runtime behavior
- Wrapping a concrete value into an interface builds a dictionary by:
  - scanning base interfaces for method signatures and default methods
  - scanning impls for concrete overrides
  - preferring impl methods over defaults, and defaults over missing
- Dictionary entries can reference native functions (bound at access time).
- The dictionary is currently built per interface value; caching per (concrete type, interface) is a valid future optimization.

## Typechecker behavior
- Method lookup on interface values considers interface methods, including defaults, and base interfaces.
- Generic interface methods use call-site inference; missing type arguments are written back into the AST for runtime dispatch.
- Inferred type arguments are applied to the return type and constraint solving.

## Implementation summary (2026-01-18)
- TS runtime: interface values carry dictionaries; interface member access binds native methods; for-loops accept interface-wrapped iterators; iterator native methods exposed via interface dictionaries.
- Go runtime: interface coercion builds dictionaries with interface args; interface member access uses dictionaries.
- TS/Go typecheckers: include default interface methods as callable candidates; allow generic interface methods; infer and record missing type arguments; include base interfaces in method lookup; collect transitive impls/method sets from imports.
- Stdlib: `Iterable.map`/`filter_map` made explicitly generic; `collect` uses `C.default()`; `Extend` returns `Self`.
- Spec: `spec/full_spec_v11.md` updated to describe dictionary-based dynamic dispatch.

## Follow-ups
- Add tests that exercise default methods and generic methods on interface-typed values across package boundaries.
- Consider caching dictionaries per (concrete type, interface) to reduce allocation overhead.
- Ensure compiler backends mirror dictionary dispatch for dynamic calls.
