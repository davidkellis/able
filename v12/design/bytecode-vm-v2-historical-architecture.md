# Bytecode VM v2: Historical Architecture and Staging Record

## Status

This is a historical summary of the 2026-04 VM-v2 proposal. Its staged order,
typed-cell sketches, and “next implementation” language are not current work.
The active contract is [bytecode-vm-v2.md](./bytecode-vm-v2.md).

## Original direction

The proposal correctly identified the lasting architectural concerns:

- retain the shared bytecode lowering, AST diagnostic nodes, runtime fallback,
  and tree-walker parity rather than create a second interpreter;
- use slot layouts and internal primitive representation only when exact
  materialization boundaries are known;
- use guarded quickening for stable call/member/index targets; and
- keep concurrency, yielding, unwinding, and dynamic values on the semantic
  runtime path until their state is proved safe to preserve.

It proposed a progression from typed layout metadata, through i32 slots and
stack cells, direct typed calls/returns, wider primitive lanes, quickening,
canonical Array/String paths, and compact frame records.

## What became current behavior

The current VM adopted parts of that direction incrementally: slot-indexed
layouts, raw operand helpers, i32 register/value-slot lanes, saved inline
frame state, cache invalidation, and canonical kernel fast paths. These are
implemented as guarded extensions of the existing runtime-value model, not the
single typed-frame ABI described by the proposal.

The proposal's central rule still holds: a raw cell is internal and must
materialize at a real semantic boundary. Its architecture does not permit a
named collection, benchmark, or source form to select a special VM path.

## Superseded staging conclusions

- Typed layout facts and typechecker proof metadata do not alone justify a
  frame-owned primitive representation.
- A direct, generally unboxed call/return ABI was not established by the
  register-frame work and is not pending implementation.
- Broader bool, f64, native storage, quickening, and compact-frame proposals
  require fresh cross-application evidence, not completion of an old list.
- Suspension/resume, errors, and dynamic features remain boundaries unless a
  change proves their complete save/restore and fallback contract.

For dated detail and measurements, use the repository history and the dated
entries in `PLAN.md`/`v12/LOG.md`. The current source and focused tests decide
whether a historical mechanism remains live.
