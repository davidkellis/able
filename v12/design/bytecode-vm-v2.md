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
   - The reduced matrix f64 path now has direct float arithmetic, a fused
     add-mul slot update, a guarded fused `Array.get!` operand update, and a
     VM-owned raw accumulator cell for that fused update. The latest slice
     now feeds raw `f32`/`f64` operands out of canonical `Array.get!` while
     preserving nil/Error propagation. The next f64 lane should collapse the
     remaining exact operand proof/read path or move toward typed f64
     array/slot cells.
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
- the exact slot-backed one-arg `i32` recurrence shape used by aligned
  external `fib` now attaches a guarded native bytecode kernel for
  `if n <= c { return r }` followed by `self(n-a) + self(n-b)`, preserving
  checked `i32` overflow and boxing only at the bytecode boundary.
- canonical kernel `Array.read_slot(i32)` and `Array.write_slot(i32, T)` now
  have guarded tracked-array member fast paths, so real external quicksort
  source sites can bypass the generic kernel method body while preserving
  negative-index errors, out-of-bounds `nil` reads, growth-on-write, and boxed
  dynamic fallback behavior.
- ordinary non-safe `Array.read_slot(i32)` and `Array.write_slot(i32, T)` call
  sites now lower to a guarded `CallMemberArraySlot` opcode. The first normal
  member-resolution pass seeds a revision-guarded proof cache for canonical
  kernel slot methods; subsequent executions bypass the broader
  `execCallMember` dispatch shell and jump directly into the tracked-array
  fast body.
- cached `CallMemberArraySlot` hits now finish the tracked read/write body
  directly after validating the array receiver and cache identity, avoiding
  the broader cached-member fast-path switch and the old
  `canUseCanonicalArraySlotCallCache(...)` guard on every hot hit.
- slot-backed quicksort pivot-loop conditions of the form
  `arr.read_slot(index) <op> pivotSlot` now lower to
  `JumpIfArrayReadSlotCompareSlotFalse`, which reuses the guarded canonical
  `read_slot` proof cache and skips the standalone read result, boxed bool,
  and generic `JumpIfFalse` path when the proof holds.
- ordinary slot-backed identifier-vs-identifier integer comparison conditions
  now lower to `JumpIfIntCompareSlotFalse`, avoiding the load/load/boxed-bool
  / `JumpIfFalse` sequence for quicksort guards like `lo >= hi`, `i > j`,
  `i <= j`, `lo < j`, and `i < hi`.

This proves the typed operand lane, checked overflow behavior, boxed boundary,
VM reset behavior, declared slot metadata, compact self-fast frame restoration,
and focused parity tests without replacing the general boxed frame model.

Rejected experiments are part of the design record:

- A parallel typed `i32` slot side cache across recursive frames passed focused
  parity but regressed reduced `Fib30Bytecode`, so the next work should not
  retry that shape.
- Conservative untyped-local `i32` proof for quicksort partition locals passed
  focused parity, but it did not produce a stable reduced quicksort win. Raw
  declaration slots without an end-to-end typed update/call/return path are not
  enough; do not reattempt untyped-local inference as a standalone slice.

The next implementation tranche should not add another `fib`-only helper. The
native recurrence kernel moved external bytecode `fib(45)` to `3.7633s` over
`3/3` runs, close enough to Go that further recursion work should either:

- generalize the recurrence machinery to generic `Int` / other primitive
  widths while preserving checked overflow and boxed dynamic boundaries; or
- pivot to the remaining external timeout families and resume typed slots,
  native collection/string bytecodes, and quickening where fresh profiles put
  the wall.

After the first five quicksort pivots, the next collection/string slice should
not target condition-only jumps again unless a fresh profile reverses the
ranking. The reduced profile now puts the remaining wall in residual boxed
slot updates, slot-call dispatch, frame release, small-int boxing, and the
`swap` / recursive quicksort call path. The follow-up cached parameter
simple-check slice removed the string-dispatch part of primitive inline
argument checks. The follow-up discard-result store slice removed the
push-then-pop roundtrip for statement-position fused slot-const
self-assignments. The follow-up bracket-swap pattern opcode removes the
standalone get/cast/set/cast/set sequence for the exact local swap block used
by quicksort while keeping the same v12 index guards and generic fallback.
The follow-up small-index swap lane removes the remaining broad index
conversion/get/set ladder for the hot tracked-array swap shape. This still
does not change the larger structural picture: the next bounded test should
start from fresh profile evidence around fused array-index comparison, direct
call-frame setup, generic binary/modulo work in `build_data`, or a v12-safe
typed-loop lane while preserving the same boxed dynamic fallback behavior.

The first reduced matrix slice confirms the same VM-v2 direction for `f64`.
Canonical tracked-array `Array.get(i32)` success values with cached primitive
`f32`/`f64` element tokens and matching actual float results now skip an
immediately following postfix propagation opcode when the current method-cache
version proves that primitive type does not implement `Error`. This is
intentionally a boxed-boundary fusion: nil, stale non-float element shapes, and
active primitive-Error impls keep the old propagation path.
The kept profile removes propagation from the top reduced matrix profile, and
the follow-up direct boxed-float binary path removes the old
`evaluateArithmetic(Fast)` wall for primitive `+`, `-`, and `*`. The remaining
lesson is unchanged but sharper: the next bounded matrix slice should carry raw
`f64` values through expression arithmetic and slot updates before boxing at
array/dynamic/spec boundaries. The follow-up `StoreSlotFloatAddMul` opcode now
does this for the common `x = x + left * right` update while preserving
evaluation order and boxed fallback behavior. The follow-up fused-array-get
and raw-accumulator slices now feed canonical `Array.get(i32)!` operands into
that update and keep the accumulator in a VM-owned float cell until a visible
slot read. The raw-operand slice dropped reduced `matrixmultiply_f64_small` to
a `4.06-4.43s/op` kept band.

The follow-up native f64 dot-loop slice is the first deliberately broader
VM-v2-style matrix cut: lowering recognizes only the exact slot-backed loop
body `if k >= n { break }; s = s + ai.get(k)! * cj.get(k)!; k = k + 1`,
attaches a plan to the existing `LoopEnter`, and leaves the original loop
bytecode in place as the fallback. Runtime guards require the canonical
`Array.get` method, tracked arrays, valid `i32` index/bound slots, and actual
`f64` elements. When any guard fails, execution enters the original loop before
the unsupported iteration.

That tranche drops reduced `matrixmultiply_f64_small` to
`319.62-333.92ms/op`, with a traced/profiled confirmation at `405.57ms/op`.
Full external bytecode `matrixmultiply` now completes in `23.85s` instead of
timing out at `90s`, beating the Ruby and Python references for this benchmark
while still trailing Go by `27.10x`. The next f64 work should target the
remaining matrix construction/transpose calls and then generalize the typed f64
array/slot lane from this proof point, not add more boxed float helpers.

The f64 row-cache follow-up validates the next VM-v2 direction: keep raw typed
rows behind a guarded representation boundary instead of repeatedly proving
boxed values. Dynamic array states now carry a revision, writes and state
resyncs invalidate that revision, and the native dot loop caches each tracked
row/column as `[]float64` for the duration of a VM run. The reduced matrix band
falls again to `204.02-229.45ms/op`, with a profiled `236.52ms/op`
confirmation. Full external bytecode `matrixmultiply` moves to `3.08s`, about
`3.50x` the Go reference and faster than Ruby/Python. The next VM-v2 f64 work
should attack row construction/transpose allocation and member dispatch, then
consider a broader typed-array storage lane once the fallback boundary is as
clear as this row cache.

The small-integer float-cast follow-up is a smaller construction-side lesson:
before adding another opcode, remove representation churn at existing primitive
boundaries. Directly converting small boxed integers to `f32`/`f64` cuts reduced
matrix allocation volume from about `1.63M/op` to about `913k/op` and moves the
kept band to `184.04-212.11ms/op`; full external bytecode `matrixmultiply`
moves to `2.90s`, about `3.30x` Go. The remaining VM-v2 matrix work should
focus on collection construction dispatch and typed row storage, not on more
boxed numeric helper rewrites.

The tracked `Array.push` append helper is a small but useful reminder that the
external benchmark shape matters. The reduced fixture was neutral because it
still grows rows dynamically, but the external benchmark preallocates rows with
`Array.with_capacity(n)`. Skipping redundant capacity checks and using
unaliased tracked sync moves full external bytecode `matrixmultiply` to
`2.75s` over `3/3` runs, about `3.12x` Go. The next VM-v2 matrix work should
focus on remaining construction-time `Array.get`/`Array.slot` dispatch and GC
scan pressure rather than another push helper.

The adjacent-`Pop` push cleanup is useful mainly as a semantics guardrail for
future quickening: only the proven canonical push fast path may skip the
statement-result `Pop`; lowering still emits the `Pop` so generic fallback stack
behavior is preserved. It improves the reduced matrix fixture to
`176.45-186.98ms/op`, but external bytecode `matrixmultiply` is neutral at
`2.774s` over `5/5`. That makes the next VM-v2 target construction-time
`Array.get` reads and residual slot-call cache checks, not more push-specific
work.

The f64 dot-loop accumulator-store follow-up sharpens the typed-value boundary:
owned float cells help repeated slot mutation, but the native dot loop writes
the accumulator only once after it consumes the full row/column. Storing that
completed accumulator as a plain `FloatValue` removes one unamortized allocation
source without changing fallback semantics. Reduced matrix now lands at
`163.01-170.93ms/op`, around `822.9k allocs/op`; full external bytecode
`matrixmultiply` moves to `2.604s` over `5/5`, about `2.96x` Go. The next VM-v2
matrix target should be remaining boxed float arithmetic/cast allocation or a
real typed f64 row/storage lane, not broad owned-slot reuse.

The f64 affine `Array.push` try-fast path confirms the right VM-v2 shape for
construction-side numeric work: recognize a narrow slot-backed expression, emit
a guarded opcode before the original bytecode, and let every guard miss fall
through to the boxed path. For the matrix `build_matrix` expression, direct f64
append drops reduced matrix to `121.57-136.46ms/op` after warmup and moves full
external bytecode `matrixmultiply` to `2.130s` over `5/5`, about `2.42x` Go.
The next VM-v2 matrix step should reduce row/column storage allocation and
remaining construction/transpose array traffic, ideally with a guarded typed
f64 row/storage lane rather than more per-helper boxed arithmetic shaving.

The versioned-stdlib canonical proof follow-up fixes an important measurement
boundary rather than the macro runtime wall. Installed stdlib origins under
`.able/pkg/src/able/<version>/src/...` now validate as canonical stdlib paths,
and direct/bound canonical nullable `Array.get` functions are accepted by the
same proof as overload wrappers. This restores the runtime-only reduced matrix
harness to a warmed `117.29-122.68ms/op` band instead of falling into generic
`Array.get` fallback after warmup, while full external bytecode
`matrixmultiply` remains neutral at `2.1333s` over `3/3`. The next VM-v2 matrix
work is still typed row/storage allocation and construction/transpose traffic,
not more canonical-origin or `Array.get` proof-cache polishing.

The mono-f64 array storage tranche confirms the row-storage direction but also
shows where the next boundary is. Dynamic rows can now promote to guarded mono
f64 storage from the affine `Array.push` fast path, the native f64 dot-loop can
read mono f64 rows without building a boxed row cache, and canonical
`Array.get` fast paths read mono f64 handles without calling generic
`ArrayStoreRead`. Reduced runtime-only matrix keeps the warmed wall-clock band
while dropping allocation volume to about `21.8-22.0MB/op` and `193.1k
allocs/op`; full external bytecode `matrixmultiply` moves to `2.0400s` over
`3/3`, about `2.32x` Go. The remaining f64 matrix work is now boundary boxing:
`finishArrayGetMemberFast(...)` still materializes about one boxed float per
external result read, and the native dot-loop accumulator slot write still boxes
one `FloatValue` per completed cell. The next VM-v2 matrix step should target a
typed f64 cell/result boundary there, not another push/storage helper.

The guarded nested-get push tranche removes the transpose-side boxed f64
boundary for the exact canonical shape `ci.push(b.get(j)!.get(i)!)`. Lowering
emits a try opcode ahead of the original call and the VM only commits when both
`Array.get` calls and the destination `Array.push` are canonical, the propagated
outer value is a concrete Array that cannot implement `Error`, and the inner row
has an in-bounds raw f64 value; otherwise execution falls through to the
unchanged bytecode. This drops reduced runtime-only matrix allocation from the
prior `193k allocs/op` area to about `103k allocs/op` while keeping wall time in
the same warmed band. Same-session external control without the fusion landed at
`2.3533s`; the restored fused confirmation landed at `2.1840s` over `5/5`.
Because the older mono-f64 best was `2.0400s`, treat this as a shape/allocation
keep rather than proof that matrix wall time has reached a new low. The next
work should target the dot-loop accumulator box and repeated array/cache guard
costs, or graduate the matrix loop to a typed f64 cell/result boundary.

The owned f64 accumulator cell tranche is a smaller typed-cell step. Ordinary
float slot stores now seed/reuse the VM-owned float cell used by the fused
float update machinery, and the native f64 dot-loop stores its accumulator
through that cell. Slot loads still snapshot `*FloatValue` cells back to
ordinary `FloatValue`, so user-visible primitive value semantics are unchanged.
Reduced runtime-only matrix wall time moved to `108.75-114.94ms/op`, and full
external bytecode `matrixmultiply` moved to `2.0840s` over `5/5`. Allocation did
not drop because the box moved from the dot-loop accumulator write to the
following `bytecodeSlotReadValue(...)` for `di.push(s)`. The next VM-v2 matrix
step should therefore be a guarded slot-backed f64 push for that exact
`Array.push` shape, not a broad load-slot pointer exposure.

The reserved-capacity follow-up is an allocation-only keep from the full
external profile. Interpreted `Array.with_capacity(n)` no longer allocates a
dynamic `[]Value` backing immediately; it records logical capacity on the
dynamic handle and lets the first dynamic write allocate the reserved backing.
Rows that immediately promote through the guarded mono-f64 append path now skip
the discarded dynamic backing entirely. This preserves `Array.capacity`, sparse
writes, generic dynamic writes, and mono-f64 deopt behavior; `Array.new(n)` and
`ArrayStoreNewWithCapacity(n)` remain eager for compatibility with compiled and
runtime ABI paths that still observe `ArrayValue.Elements`.

The reduced fixture is neutral because it uses `Array.new`, but full
bytecode-runtime `matrixmultiply` allocation moves from the prior profiled
`125.83MB` total / `121MB/op` area to `90.89MB` profiled total and
`71.84MB/op` unprofiled. Full external bytecode `matrixmultiply` stays neutral
at `2.1000s` over `5/5`, with lower GC. The next matrix work should return to
the f64 result boundary (`bytecodeSlotReadValue(...)` feeding `di.push(s)`) or
a broader typed f64 result-row lane, not generic capacity reservation.

The native f64 dot-loop result append is the successful shape for that
boundary. Instead of adding another pre-call try opcode for `di.push(s)`, the
existing dot-loop plan now optionally spans the following discarded
statement-position push of the same accumulator. If the dot-loop, canonical
`Array.get`, canonical `Array.push`, and raw f64 append guards all pass, the VM
appends the accumulator directly and jumps past the fallback push bytecode. If
any guard fails, the original loop and push execute unchanged.

This removes `bytecodeSlotReadValue(...)` from the allocation top list and
moves full bytecode-runtime `matrixmultiply` to about `39.8MB/op` and `73.5k
allocs/op`; full external bytecode `matrixmultiply` lands at `2.0240s` over
`5/5`. The next matrix work should target mono-f64 append storage and growth,
especially `ArrayStoreAppendF64Promote(...)` / `appendMonoF64Value(...)`, not
more result-load boxing or standalone slot-push dispatch.
