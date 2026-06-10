# Bytecode i32 Register Frames: Current Gate and Historical Record

## Status and authority

The bytecode VM has live, pooled i32 register frames and value-slot i32
sidecars. This document records their executable contract and the rejected
VM-v2 frame-object experiment; it is not an implementation queue for a new
typed-frame ABI.

The v12 spec remains semantic authority. `bytecodeFrameLayout`, the bytecode
VM, and their focused tests define execution. `PLAN.md` selects performance
work only after a concrete non-nominal leaf recurs across unlike verified
applications.

## Current VM contract

Slot-eligible functions retain `vm.slots []runtime.Value` for generic and
reference values. A syntactic lowering eligibility walk may also enable
`i32RegisterFrame`; for that active program the VM holds a pooled `[]int32`
and matching validity bitmap keyed by the program frame layout. A discarded
typed i32 store can therefore keep its raw value in the register lane without
writing a boxed value to its corresponding slot. A generic slot read
materializes the canonical raw i32 value only when the surrounding operation
requires a runtime value.

Programs without that register frame may use the separate value-slot i32
sidecar. It is attached to the active `[]runtime.Value` slot frame, has its own
validity bitmap and pool, and provides the same raw-value read/materialization
semantics. Neither mechanism changes Able's observable value, coercion, error,
or tracing behavior.

Full, self-fast, and compact self-fast inline call frames detach and restore
both the active register frame and value-slot sidecars. Program switching
activates or releases a register frame as required. Inline direct-call setup
may preseed an eligible typed i32 callee slot from the caller's raw lane; this
is safe frame setup, not a general unboxed typed call/return ABI. Normal generic,
dynamic, coercing, and public-result boundaries continue through the existing
runtime-value rules.

## Typechecker proof metadata

Successful program or standalone-module typechecking attaches immutable
`i32FrameProof` slot facts during lowering. The proof accepts concrete `i32`
facts for already typed slots, requires matching typed writes, and rejects
skipped checks, diagnostics, aliases, generic facts, and unannotated value
slots.

The current VM does not consume this metadata to choose storage or opcodes.
Its register-frame eligibility remains the conservative syntactic analysis.
The proof is diagnostic infrastructure for a possible future representation
study, not permission to infer a new raw lane or to change ordinary execution.

## Active safety and performance gate

- Preserve raw-lane values across calls, returns, program switches, reuse, and
  run cleanup; every generic observation must materialize through the existing
  canonical boundary.
- Keep generic/value-sidecar and register-frame paths behaviorally equivalent
  to the tree-walker and normal bytecode semantics.
- Do not add a QuickSort, Array, container, call-name, or source-shape special
  case. Existing register work is shared VM machinery, not a named-program
  optimization.
- Do not promote typechecker facts into a frame-owned representation or a
  broad typed call/return ABI from this record. A new candidate needs a fresh
  repeated material leaf across unlike verified applications, focused semantic
  proof, and the full bytecode coverage/performance gate.

## Historical VM-v2 staging record

The proof-plumbing stage completed on 2026-07-13 and deliberately made no VM
execution change. A subsequent leaf-only frame-object candidate used
typechecker-proven `[]int32` local storage while leaving calls, returns,
captures, collections, dynamic operations, and errors materialized. It was
reverted after the pinned three-run MatrixMultiply check averaged 5.2967s
versus 4.7867s restored (a 10.7% regression).

The prior proposed stages—replacing slots with a frame-owned primitive lane,
then adding a general direct i32 call/return ABI—are historical experiments,
not pending work. The later MatrixMultiply/Mandelbrot audit found no common
primitive-frame leaf that would justify reviving them. The broader VM-v2
chronology is retained in `bytecode-vm-v2.md`; the performance selection rule
is retained in `performance-competitiveness-vision.md`.

## Focused evidence

Focused tests prove typechecker proof attachment/fallback, register-frame raw
stores and generic reads, inline-call preservation, value-slot sidecar restore,
and register-frame activation/release on program switches. Those tests are a
regression gate for the current mechanism, not a benchmark for a new ABI.
