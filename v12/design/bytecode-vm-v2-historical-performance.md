# Bytecode VM v2: Historical Performance Experiment Record

## Status

This record condenses the post-proposal VM experiment chronology formerly kept
in the monolithic VM-v2 note. It is regression context, not a work queue. The
active VM contract and selection gate are in
[bytecode-vm-v2.md](./bytecode-vm-v2.md).

## 2026 experiment families

| Family | Historical outcome | Current rule |
| --- | --- | --- |
| i32 slot/register work | Raw operand cells, register lanes, sidecars, and inline-frame save/restore landed; the later frame-owned typechecker-proven storage candidate regressed MatrixMultiply 10.7% and was reverted. | Preserve the live lane lifecycle; do not restart a typed-frame/call-return ABI without fresh shared evidence. |
| canonical Array/String and `u8` paths | Guarded kernel/host boundaries and cache paths landed for actual language/runtime APIs. | Extend only a shared language/kernel boundary, never a named collection or one benchmark. |
| extern and parser boundaries | Reusable primitive host conversion and parser-range work supplied normal runtime capabilities. | A host boundary needs a specified reusable contract and explicit conversion/error behavior. |
| nominal structs and inline calls | Exact nominal/frame and member shortcuts were tried to remove repeated semantic lookup where source proof existed. | Keep only shared nominal/frame mechanisms; do not add a structure-name or application rule. |
| raw float and matrix paths | Several guarded float/array kernels improved historical matrix rows, alongside rejected raw-slot ownership/cache variants. | Matrix measurements are status evidence; profile a recurring VM leaf before any further float/storage work. |

## Lessons retained

1. A focused benchmark win is insufficient when the same representation loses
   elsewhere; revert rather than expand it.
2. A cache or raw lane must preserve its owner, invalidation, materialization,
   call-frame, and error behavior as one contract.
3. Canonical API fast paths are allowed only when they remain generic to that
   language/kernel boundary and retain a pre-mutation fallback.
4. New performance work starts from fresh verifier-backed profiles, not from a
   historical top-level parent such as dispatch, slots, Array, or VM frames.

The detailed dates, commands, and measurements remain available in the dated
`PLAN.md` and `v12/LOG.md` entries and version history. They must not be used
to infer that any old “next slice” is still selected.
