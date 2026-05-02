# v12 Design Notes

v12 is Go-first (tree-walker + bytecode). The TypeScript interpreter was removed
from the active toolchain. Some design notes still reference TypeScript/Bun
workflows for historical context; treat those references as archived unless a
future non-Go runtime is revived.

Authority map:
- `compiler-go-lowering-spec.md` is the canonical Able -> Go lowering map.
- `compiler-go-lowering-plan.md` is the canonical ordered completion plan for
  the compiler.
- `compiler-native-lowering.md` is the short-form compiler contract and
  guardrail summary.
- `compiler-aot.md` defines the correctness-first AOT scope and boundary
  contract.
- `spec/full_spec_v12.md` is the language authority when a design note and the
  implementation diverge.

Interpretation rules for the rest of this directory:
- documents that describe bytecode, parser, runtime, stdlib, or concurrency
  mechanics remain useful design notes, but they do not override the compiler
  lowering spec/plan;
- documents that still discuss TypeScript/Bun are historical or future-runtime
  notes unless they have been explicitly refreshed for the current Go-first
  toolchain;
- named stdlib/container examples in compiler docs are proof cases for shared
  lowering machinery, not licenses to add nominal-type-specific compiler rules.

Compiler architecture references:
- `compiler-go-lowering-spec.md`: exhaustive Able -> Go lowering map
- `compiler-go-lowering-plan.md`: ordered path from current compiler to that target

Bytecode/runtime architecture references:
- `bytecode-vm-v2.md`: current bytecode VM v2 plan for typed cells,
  quickening, native Array/String bytecodes, and spec-preserving fallback
