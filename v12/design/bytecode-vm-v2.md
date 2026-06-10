# Bytecode VM: Active Contract

## Status and authority

This is the concise active contract for the Go bytecode VM. The name remains
for existing references; it is not a second runtime, a pending VM rewrite, or
a benchmark-specific implementation plan.

`spec/full_spec_v12.md` defines Able semantics. The Go tree-walker is the
behavioral reference. The bytecode lowering and VM source, focused regression
tests, and `PLAN.md` define executable behavior and performance selection.
The historical architecture and experiment records are linked below.

## Current execution model

The VM executes a linear `bytecodeProgram` with an operand stack of
`runtime.Value`-compatible cells. Slot-eligible functions use
`bytecodeFrameLayout` and a flat `[]runtime.Value` local frame; ineligible
functions retain normal environment lookup. Inline direct calls use full,
self-fast, or compact self-fast call frames, and preserve the caller's lookup,
scope, iterator, loop, ownership, and raw-lane state across return/unwind.

The VM has guarded raw primitive helpers where the existing lowering and frame
contract admit them:

- pooled i32 register lanes for eligible slot frames and separate value-slot
  i32 sidecars for other slot frames;
- raw integer/float operand or result cells where a checked opcode owns the
  operation; and
- canonical Array/String/kernel fast paths, lookup caches, and revision guards
  for already-proven language/kernel boundaries.

These are implementation caches over the runtime model. Generic loads,
coercions, dynamic calls, public results, diagnostics, and any uncertain path
use the canonical materialization and runtime-dispatch rules. The i32 frame
contract and inert typechecker proof metadata are defined in
`bytecode-i32-frame-abi-gate.md`.

## Semantic guardrails

- Preserve checked arithmetic, Euclidean integer division/modulo, truthiness,
  errors, diagnostics, mutable Array identity, UTF-8 String behavior, and all
  interface/union/dynamic semantics.
- Preserve `spawn`, Future, yielding, cancellation, iterator cleanup, and
  resume/unwind behavior. A path without an exact saved-state proof must use
  the existing generic representation.
- A guard failure must occur before observable mutation and execute the
  established fallback. Never replay an effect after a failed fast-path guard.
- Array and String receive VM-native treatment only at canonical
  language/kernel APIs. Other nominal types, including all collections, use
  shared runtime/member machinery; no named-container rule is allowed.
- Existing fused legacy patterns are regression-constrained implementation
  history, not templates for new source-shape, benchmark, or application
  optimizations.

## Caches and boundaries

Lookup, member, index, and call caches are version/identity guarded and are
invalidated through the existing runtime revision rules. A fast path may
bypass repeated lookup only after the normal semantic resolution has proved
the same target. It must retain argument, receiver, ownership, coercion, and
error behavior.

The VM may convert to raw internal cells only while the operation and frame
own that representation. It must materialize before an environment escape,
generic or dynamic dispatch, host/extern ABI edge, collection/nominal storage
without a separately proved kernel boundary, public result, or an error/
diagnostic/resume path that observes a runtime value.

## Register-IR feasibility

The 2026-07-18 typed-block feasibility census did not authorize an executable
register dispatcher. A conservative static model removed at least 15% of
dynamic instructions in Mandelbrot and Future Pipeline, but only 0.11%-6.46%
in six unlike controls. Typed straight-line regions therefore do not justify a
second execution engine. See
`../docs/perf-baselines/2026-07-18-bytecode-register-ir-feasibility-gate.md`.

The follow-up whole-function model clears that threshold in all eight unlike
applications measured, with 30.09%-44.32% of dynamic instructions attributable
only to `Const`, `ConstI32`, `LoadSlot`, `LoadSlotI32`, `Dup`, and `Pop`.
Calls, dynamic operations, allocation, stores, errors, control flow, returns,
and concurrency remain semantic instructions. This admits an opt-in executable
prototype, not a default path. See
`../docs/perf-baselines/2026-07-18-bytecode-operand-ir-feasibility-gate.md`.

Subsequent executable work supplied that evidence and closed the separate
whole-function executor. Register-native `MemberAccess` reached six unlike
applications but was neutral or slower in all six. Removing register-frame
allocation and millions of fruitless continuation probes did not repair the
broad Word Frequency guard. Those paths were removed; the ordinary VM remains
authoritative.

The current target-budget audit also closes a complete register
*representation by itself* as the next performance tranche. Six unlike
applications execute 30.09%-44.32% modeled transport operations. Even an
intentionally favorable equal-cost model that makes every one of those
operations free yields only 1.43x-1.80x, while the applications require
13.03x-117.50x to reach their current Python/Ruby budgets. All 143 live
opcodes are classified: six are representation transport and 137 retain
semantic work. See
`../docs/perf-baselines/2026-07-21-bytecode-architecture-target-budget.md`.

A future primary register IR is not forbidden, but it must be selected as part
of a separately evidenced semantic-operation or allocation reduction. Do not
reintroduce a parallel executor or implement complete register coverage merely
to erase stack transport.

The follow-up semantic-work audit normalizes the same six applications by hot
iterations, routed tasks, words, window elements, and sequence bases. It finds
no exact operation that is materially amplified in three unlike families.
Binding stores and call/return recur in four families but are amplified only in
the two stdlib-heavy rows; direct Array slot/member work is material in two;
typed Result control is material in two. Native-library use and allocation are
broad only as aggregate labels, with different contracts and concrete owners.
See
`../docs/perf-baselines/2026-07-21-bytecode-semantic-work-amplification.md`.

Consequently, do not infer a VM candidate from total semantic-operation or
allocation volume. The next admissible check is whether the same direct Array
storage/member boundary also owns material work in a third unlike application;
otherwise that route closes and selection returns to the compiled frontier.

That Array check is complete. Canonical `Array.push` is material in Array Slice
Window, Matrix Multiply, and Reverse Complement, but more than 99.999% of their
slot calls already hit the direct cache and none report a fast-path miss. The
push descendants divide among required independent slice backing/lease work,
monomorphic f64 store synchronization, and retained raw-u8 append work. A prior
generic validated-push wrapper also regressed broad guards. Even perfect removal
of the complete push subtree would yield only 1.046x and 1.087x in the two
target-missing rows. See
`../docs/perf-baselines/2026-07-21-bytecode-array-semantic-boundary.md`.

The opt-in stats snapshot now reports Array slot calls separately for `len`,
`read_slot`, `write_slot`, and `push`. Keep those diagnostic counters, but do
not reopen Array wrapper/storage generalization without a new exact shared leaf
and evidence that invalidates the prior broad performance gate.

The 2026-07-22 post-compiler three-shape refresh likewise leaves the VM
unchanged. Split/join text, linked-list iterator collection, and numeric Array
mapping still intersect only at aggregate call/return/type/map helpers. Their
raw-integer, string-map, inline-return, and typed-pattern children have unlike
semantic owners or already-rejected representations. Do not reopen those local
families from their shared parent names; see
`../docs/perf-baselines/2026-07-22-bytecode-three-shape-post-compiler-refresh.md`.

The follow-up generic-union cohort admits one exact leaf: static generic-union
method matching is material in three unlike applications and spans `Result`
and `Option`. Do not implement it with a global match map or recursive
unchanged-type identity reuse. The former removes the matcher but costs more
than it saves; the latter reduces owner allocations but regresses the unrelated
iterator guard. A future attempt must first prove instruction-local
monomorphism and stable scope/method-version guards with opt-in counters; see
`../docs/perf-baselines/2026-07-22-bytecode-generic-union-type-resolution-gate.md`.

## Change and performance gate

No typed-frame, quickening, raw-lane, host-boundary, array, text, recursion,
or float change is selected merely because it improves one benchmark. New work
must:

1. identify the same concrete non-nominal material leaf in at least three
   unlike verifier-backed applications;
2. use source and focused tests to prove parity, fallback, invalidation, and
   boundary behavior; and
3. clear the bounded full bytecode coverage/performance gate without a
   material regression outside its target.

Until that evidence exists, refresh profiles and improve feature coverage
instead of extending the historical typed-frame or source-shape proposals.

## Historical records

- [Architecture and staging record](./bytecode-vm-v2-historical-architecture.md)
  summarizes the 2026-04 VM-v2 proposal and its superseded tranches.
- [Performance experiment record](./bytecode-vm-v2-historical-performance.md)
  summarizes later register, array/text, host, nominal, and float work.
- [i32 register-frame gate](./bytecode-i32-frame-abi-gate.md) records the
  live raw-lane lifecycle and rejected frame-object candidate.
- [Performance competitiveness vision](./performance-competitiveness-vision.md)
  is the current cross-application selection policy.

The historical records retain rationale; they do not authorize their old next
steps. New language behavior belongs in the spec and fixtures first. New VM
performance work belongs in the verifier-backed selection process.
