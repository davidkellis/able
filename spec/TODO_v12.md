# Able v12 Spec TODOs

This list tracks the remaining v12 items after audit; completed work should be removed.

## Parser gaps
- None currently tracked (cast `as` line breaks fixed 2026-02-06).

## Compiler AOT gaps
- Define the compiled runtime ABI for core types (Array, BigInt, Ratio, String, Channel, Mutex, Future).
- Specify compiler-only error behavior when static constructs cannot be compiled (no silent fallback).
- Document the compiled stdlib/kernel requirement and how compiled code resolves them.
- Define the compiled interface dispatch and overload resolution model (dictionary dispatch + monomorphization).
- Specify compiled ↔ dynamic value conversion rules and error surfaces in the dynamic boundary bridge.
