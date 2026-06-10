# Struct Instance Allocation Notes

Status: 2026-07-08 investigation note.

## Context

After the raw generator/native boundary work, the
`linked_list_iterator_pipeline_i64_small` bytecode runtime allocation profile
still showed `runtime.NewStructInstancePositionalSized(...)` as the largest
remaining sampled allocation bucket. A line-level `alloc_space` profile for the
helper attributed the retained bucket to the escaping `&StructInstanceValue`
itself, not to map setup or positional backing-slice allocation.

The current constructor already keeps short positional payloads inline:

- `0..3` positional fields allocate only the escaping `StructInstanceValue`.
- `4+` positional fields allocate the instance plus a separate positional
  slice.
- named-field map storage is not part of this helper.

## Rejected Changes

`sync.Pool` reuse is not sound for this path. Struct instances are observable
Able values, may be mutable, and can escape through variables, collections,
interfaces, futures, closures, or host interop. The runtime has no ownership or
release point that can prove an instance is dead before reuse.

Reducing the inline positional capacity to favor two-field linked-list nodes
would be workload-shaped. It would reduce the size of every instance while
adding a new backing-slice allocation for ordinary three-field structs, so it is
not a generally applicable win.

Adding named-container or benchmark-specific construction paths is also out of
scope. Non-primitive nominal values must keep using the shared nominal
translation and runtime representation rules.

## Current Decision

No representation change was kept in this tranche. The benchmark-visible bucket
is currently the required allocation for each escaping nominal value. The kept
work is allocation regression coverage that locks in the generic property we
do want: short positional instances do not allocate extra helper storage beyond
the instance object.

Future reductions here require a broader representation design, such as
compiler-proven scalar replacement/unboxing for non-escaping values or a
language-level value/identity distinction. Those changes should be evaluated
across the benchmark corpus and fixtures, not introduced through a
linked-list-specific fast path.
