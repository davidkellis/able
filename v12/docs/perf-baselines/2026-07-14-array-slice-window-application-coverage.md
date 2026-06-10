# Array slice window application coverage — 2026-07-14

## Decision

Add `array_slice_window` to the portable external **coverage** catalog only.
It carries the completed `Array.slice(start, end) -> !Array T` semantics into
the Go/Python/Ruby comparison harness without changing the stable generality
scorecard or authorizing an Array-specific optimization.

## Comparable semantics

The application creates 257 deterministic event values, takes 12,000
overlapping end-exclusive windows of width 24, and reduces weighted checksums.
It also mutates a returned two-element window and prints a probe proving that
the source backing was not changed.

| Runtime | Range operation |
| --- | --- |
| Able | `Array.slice(start, end)` |
| Go | exact-length `make` plus `copy(values[start:end])` |
| Python | `values[start:end]` |
| Ruby | `values[start...end]` |

The explicit Go copy is required: a normal Go subslice aliases its source and
would not implement Able's new independent shallow-copy contract. Every
implementation and the verifier produce `137999:3648189002`.

## Verification

The catalog maps `array_slice_window` to the canonical Able source under
`examples/benchmarks/array_slice_window` and treats the sibling external
`run.able` as an excluded source root, so Docker and local harness runs cannot
load two same-package entry files. The two Able source copies are byte-identical.

Focused commands used the normal one-process 1 GiB guard:

```text
./v12/bench_refresh_go_refs --benchmarks array_slice_window --runs 1 --timeout 45
./v12/bench_refresh_interpreter_refs --benchmarks array_slice_window --runs 1 --timeout 45
GOMEMLIMIT=1GiB GOGC=50 GOMAXPROCS=1 ABLE_STDLIB_ROOT=../able-stdlib/src \
  ./v12/bench_compare_external --benchmarks array_slice_window \
  --modes compiled,bytecode,treewalker --languages go,python,ruby \
  --runs 1 --timeout 45
```

Fresh reference verification passed: Go `0.0036s`, Python `0.0267s`, and Ruby
`0.0582s`. The verifier-backed Able runs passed in compiled (`0.0900s`),
bytecode (`0.6600s`), and tree-walker (`5.9300s`) modes. These are one-run
coverage measurements, not a regression claim or a selection baseline.

## Selection result

The bytecode and compiled rows materially miss the project targets, but this
is one new container API carrier. It does not establish a repeated concrete
descendant across three unlike applications, so it selects no VM opcode,
ArrayStore/view representation, compiler lowering rule, canonical-stdlib fast
path, or benchmark-shape optimization. Future work remains gated on the same
repeated-leaf and broad-control policy as the existing scorecard.
