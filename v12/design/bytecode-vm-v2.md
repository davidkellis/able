# Bytecode VM v2 Design

Date: 2026-04-30

## Purpose

The current Go bytecode VM is semantically useful but performance-limited. It
has slot-indexed frames, inline call frames, expression caching, and several
hot opcodes, but the common execution path still boxes most values as
`runtime.Value` and routes many static operations through dynamic lookup or
operator helpers.

VM v2 is an incremental upgrade of the existing bytecode interpreter. It is not
a new language runtime, not a fork of semantics, and not a benchmark-specific
shortcut. The v12 spec remains the authority.

## Spec Guardrails

Every VM v2 optimization must preserve these v12 requirements:

- Primitive integer arithmetic remains checked by default. Overflow in `+`,
  `-`, `*`, integer exponentiation, casts, and fixed-width division edge cases
  still raises the standard overflow error described in the spec.
- Integer `//`, `%`, and `/%` keep Euclidean division/remainder semantics,
  including division-by-zero and minimum-integer edge behavior.
- Truthiness is unchanged: only `false`, `nil`, and values implementing
  `Error` are falsy.
- `Array T` is a mutable indexed sequence with stable runtime identity.
  Native array bytecodes may optimize the canonical kernel/stdlib boundary, but
  they must preserve allocation, length/capacity, indexing, mutation, iteration,
  and `IndexError` behavior.
- `String` remains the canonical immutable UTF-8 text value. Native string
  bytecodes may avoid intermediate Able allocations but must preserve UTF-8,
  byte/grapheme API contracts, and immutable value behavior.
- Interface, union, nullable, `Error | T`, and dynamic values continue to use
  the existing semantic runtime representation at dynamic boundaries.
- `spawn`, `Future`, `await`, cancellation, `future_yield`, and deterministic
  serial-executor behavior stay compatible with the concurrency spec and the
  executor contract.
- Diagnostics must still attach the same AST-context information as the
  tree-walker and current bytecode VM.

If a typed or quickened path cannot prove that it preserves these rules, it
must box back to the existing `runtime.Value` path before proceeding.

## Current VM Shape

The active VM is stack based:

- `bytecodeProgram` owns a linear instruction stream and optional
  `bytecodeFrameLayout`.
- Slot-eligible functions use flat `[]runtime.Value` frames instead of map
  environment locals.
- The operand stack is `[]runtime.Value`.
- Calls can inline slot-eligible functions and have special self-recursive
  frame paths.
- Lookup/member/index/call caches sit beside the VM and preserve existing
  runtime semantics.
- Reduced recursion is currently helped by fused opcodes such as
  slot-const arithmetic, fused conditional jumps, and fused self calls.

This is a good migration base. VM v2 should reuse the lowering pipeline,
diagnostic nodes, slot eligibility, call frame stack, serial-yield resume path,
and runtime fallback helpers.

## Core Representation

Add typed cells as an internal VM representation:

```go
type bytecodeCellKind uint8

const (
    cellValue bytecodeCellKind = iota // boxed runtime.Value
    cellRef                           // non-primitive runtime.Value/reference
    cellBool
    cellI32
    cellF64
)
```

The first implementation should not introduce every lane at once. Start with
`cellI32`, then add `cellBool`, `cellF64`, and finally reference-specialized
lanes when the boxing contract is stable.

Typed cells have two mandatory operations:

- `boxCell(...) runtime.Value`: materializes the exact existing runtime value.
- `storeCellFromValue(...)`: accepts an existing `runtime.Value`, performs the
  same validation/coercion checks as today, and fills the typed lane only when
  safe.

Typed storage is an optimization cache, not a new public value model.

## Typed Frames And Stack

Each slot-eligible program gains an optional typed layout derived from declared
parameter/local facts and the existing slot analysis:

- `slotKinds []bytecodeCellKind`
- `paramKinds []bytecodeCellKind`
- `returnKind bytecodeCellKind`
- `hasTypedSlots bool`

The VM frame stores typed values either as parallel arrays or as a compact
cell slice. The first implementation should prefer the smallest scoped change
that makes `i32` recursion measurable:

- keep the existing `[]runtime.Value` slots for compatibility;
- add an optional parallel `[]int32` plus a per-slot kind/valid bit for
  `cellI32`;
- only read from the typed lane when the current slot kind is known and valid;
- clear or box the typed lane when a slot crosses a dynamic boundary.

The operand stack should move to typed cells once slot reads can produce typed
values. Until then, a typed frame that immediately boxes every load will not
move external `fib`. The first stack slice should support:

- `PushI32`, `PopI32`, `PeekI32`
- boxed fallback push/pop for every existing opcode
- `LoadSlotI32`, `StoreSlotI32`
- `ConstI32`
- `AddI32Checked`, `SubI32Checked`, `LessEqualI32`
- `JumpIfFalseBool` or fused `JumpIfI32LessEqualConstFalse`
- typed return for inline calls, boxing only when the caller expects boxed
  values or crosses a runtime boundary

## Dynamic Boundaries

A typed value must be boxed before any operation that can observe or require a
general `runtime.Value`:

- environment definition/assignment outside a typed frame;
- generic function calls or calls with type arguments;
- interface, union, nullable, result/error, and `any` coercion;
- dynamic member/index dispatch;
- `match`, `rescue`, `ensure`, propagation, raise/rethrow, and diagnostics
  paths that need semantic values;
- array/map/struct literals unless the opcode is explicitly native and still
  preserves runtime identity and error behavior;
- host extern boundaries unless the host wrapper has a proven native carrier
  path;
- suspension/resume paths for `future_yield` and `await` unless the frame state
  serializer preserves typed lanes exactly.

The fallback rule is simple: when uncertain, box and run the existing code.

## Quickening

Quickening should come after typed cells are in place. A quickened opcode is a
rewritten instruction guarded by the same semantic conditions that made the
first execution valid.

Initial quickening targets:

- `call_name`: static function, inherent method, and stable overload entries.
- `call_member`: primitive/kernel method and native interface default method
  cases that already have cache entries.
- `index_get` / `index_set`: canonical `Array T` paths and string/byte views.
- `member_access`: struct field, bound method, and primitive member paths.

Required guards:

- callee or method identity;
- receiver runtime shape or primitive kind;
- package/environment revision when the lookup can be affected by mutation or
  import/bootstrap state;
- argument count and type-argument absence/presence;
- fallback when a guard fails.

Quickening must never bypass overload resolution, implementation specificity,
safe navigation, implicit receiver rules, or dynamic interface dispatch unless
the same result has been proven by the guard.

## Native Array And String Bytecodes

Array and String are allowed VM-native treatment because they are core
language/stdlib boundary types in the v12 spec. This is not permission to add
benchmark-specific bytecodes for arbitrary nominal containers.

Array bytecodes should target:

- `Array.with_capacity`, `push`, `len`/`size`, `capacity`;
- indexed read/write with the same `IndexError` payloads;
- iteration over canonical arrays without per-element method lookup;
- mono primitive arrays for `u8`, `i32`, and `f64` after typed cells are stable.

String bytecodes should target:

- `len_bytes`, byte iteration, byte search, contains, replacement, and split;
- UTF-8 validation boundaries;
- zero-copy views only when the public API promises immutability and lifetime
  safety.

Any operation outside the canonical API shape falls back to method dispatch.

## Concurrency And Resume

The serial-yield and `await` paths must preserve typed state. Before enabling
typed lanes for bytecode programs that can yield:

- call frames must save typed slot arrays and typed operand-stack cells;
- `finishRunResumable` must release typed frames exactly once;
- unwound frames must not retain boxed references longer than the current
  `runtime.Value` implementation would;
- deterministic serial-executor tests must pass with typed lanes enabled.

The first `i32` recursion slice can reject programs with `spawn`, `await`,
`yield`, iterator literals, or ensure/rescue frames until resume coverage is
added.

## Implementation Tranches

1. **Typed layout metadata**
   - Extend `bytecodeFrameLayout` with slot kinds.
   - Infer only simple primitive slots from declared types and existing slot
     eligibility.
   - Add tests proving unsupported/dynamic locals keep `cellValue`.

2. **`i32` typed slots and stack**
   - Add internal typed cells for `i32`.
   - Lower/load/store `i32` parameters and locals without boxing in
     slot-eligible non-yielding functions.
   - Implement checked `i32` add/sub and `<=` using spec-compatible overflow
     and diagnostic behavior.
   - Keep all other opcodes boxing through the current VM path.
   - Benchmark reduced `Fib30Bytecode` and external `fib`.

3. **Typed inline calls and returns**
   - Pass `i32` args/results between inline slot frames without boxing.
   - Preserve coercion checks and generic return handling.
   - Re-run recursive, return, and unwound-frame finalization tests.

4. **Bool and `f64` lanes**
   - First bool branch slice is landed for declared slot-backed conditions:
     `if`, `elsif`, and `while` can use a bool-slot conditional jump.
   - Add broader bool cells and `f64` arithmetic/comparison lanes only after
     post-branch profiles justify them.
   - Target `sudoku` and `matrixmultiply` reduced/external profiles.

5. **Quickened dispatch**
   - Add guarded quickened call/member/index opcodes with cache invalidation.
   - Keep counters for hit/miss/guard-fail rates.
   - Target `i_before_e` and text-method traces.

6. **Native Array/String bytecodes**
   - Add canonical array/string fast paths behind exact API guards.
   - Target `quicksort`, `i_before_e`, and future `base64`/`json` coverage.

7. **Compact typed frames**
   - Replace remaining hot boxed slot frame churn with compact typed frame
     records once typed call/return behavior is proven.

## Verification

Every tranche needs three checks:

- **Semantic parity:** focused `go test ./pkg/interpreter` slices covering both
  bytecode and tree-walker fixture parity for the affected language construct.
- **Spec edge cases:** explicit overflow, division/modulo, index bounds,
  truthiness, nil/error, generic/dynamic fallback, and cancellation/yield tests
  where relevant.
- **Performance guardrail:** reduced benchmark first, then external scoreboard
  rows when the reduced signal is real.

Use feature flags or build-time toggles for large typed-lane changes until the
parity surface is broad enough. A failing guard must fall back to the existing
boxed VM path, not continue with partially typed state.

## Current Status And Next Coding Slice

The first three kept code slices are landed:

- literal-only final `i32` add/sub expressions use a raw `i32` operand stack
  and box back to `runtime.Value` before the existing return path;
- simple declared `i32` params and typed local identifier declarations now
  carry typed slot metadata, and safe final arithmetic can enter/leave the raw
  stack through `LoadSlotI32` / `StoreSlotI32`;
- the proven two-slot one-argument self-fast recursive frame shape can now
  reuse the current slot frame by saving/restoring slot 0, avoiding
  per-recursive-step two-slot frame acquire/release churn;
- the fused self-call opcode now has an early exact-shape compact branch for
  that same raw-immediate two-slot slot-0 recursive shape, avoiding the generic
  immediate/layout/return-name ladder while preserving the boxed fallback path;
- the direct small-`i32` return-add value-pair branch now lives inline in
  `execReturnBinaryIntAdd`, so the aligned recursive return-add edge no longer
  pays a helper call before boxing the checked result;
- the exact compact self-call branch now writes the slot-0 self-fast frame
  record directly, so the aligned recursive call edge no longer pays a helper
  call just to append the proven compact frame.
- the exact compact self-fast slot-0 frame now carries a raw `i32` lane beside
  the boxed semantic slot value, letting recursive subtract and base-case
  slot-const compare use proven raw `i32` state while all fallback/spec
  boundaries continue to observe the boxed `runtime.Value`.

This proves the typed operand lane, checked overflow behavior, boxed boundary,
VM reset behavior, declared slot metadata, compact self-fast frame restoration,
and focused parity tests without replacing the general boxed frame model.

One rejected experiment is now part of the design record: a parallel typed
`i32` slot side cache across recursive frames passed focused parity but
regressed reduced `Fib30Bytecode`, so the next work should not retry that
shape.

The next implementation tranche should target compact typed return/result
handoff rather than per-frame side arrays or more helper shuffling:

- carry small integer results through `execReturnBinaryIntAdd` and
  `finishInlineReturn` with less boxed-value probing, ideally by making the
  result handoff itself typed/raw and boxing only at the existing return
  boundary;
- avoid more helper-call rearrangement unless a fresh profile shows a specific
  helper has re-entered the hot path;
- keep `LoadSlotI32` / `StoreSlotI32` boxing-compatible at every dynamic/spec
  boundary, and keep unsupported shapes on the existing `runtime.Value` path.

That tranche should deliberately target the recursive `fib` shape because the
focused external scoreboard now shows bytecode `fib` completing at `67.8200s`
but still far behind Go and slower than Python/Ruby. The implementation must
remain general to checked integer operations and boxed dynamic boundaries.
