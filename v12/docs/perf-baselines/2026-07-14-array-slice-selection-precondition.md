# Array slice performance-selection precondition — 2026-07-14

## Decision

Keep `array_slice_window` as a portable **coverage** row and select no
`Array.slice` optimization. The completed API has one application carrier, not
the three semantically unlike, verifier-backed applications required to select
a general VM, compiler, runtime, or standard-library performance change.

## Source inventory

The following inventory intentionally prunes generated `target` trees:

```sh
find v12/examples/benchmarks ../benchmarks -type d -name target -prune -o \
  -type f -name '*.able' -print0 | xargs -0 rg -n '\.slice\('
```

It covers 63 non-generated Able source files and finds exactly two paths:

| Path | Role | Independent application? |
| --- | --- | --- |
| `v12/examples/benchmarks/array_slice_window/array_slice_window.able` | canonical Able source | yes |
| `../benchmarks/array-slice-window/run.able` | external Docker snapshot | no |

The files have the same SHA-256 and are deliberately kept byte-identical by
the benchmark harness. The strict execution fixture proves the API contract,
but is not an application benchmark. The unrelated `String.slice` call in the
process library does not exercise `Array.slice`.

## Consequence

Do not profile this one row as a selection exercise and do not manufacture two
slice-shaped workloads merely to satisfy a count. A profile could explain that
application's copy cost, but could not distinguish a reusable performance leaf
from work specific to its rolling-window shape. In particular, do not add an
Array view, storage lease, VM opcode, compiler lowering branch, or named
container exception from this evidence.

Selection may resume only when real portable applications independently need
copying range semantics and expose the same concrete, non-nominal descendant
in at least three unlike callers. At that point, collect bounded profiles with
the existing 1 GiB one-process guard, test one shared candidate against broad
compiled and bytecode controls, and retain it only if the controls hold.
