# Bytecode VM cross-workload profile refresh (2026-07-11)

## Purpose

Refresh the warmed bytecode view after the Array-ownership investigation before
selecting another VM optimization. The three controls deliberately exercise
different language features: text construction/split/join, lazy iterator
collection over a nominal container, and numeric Array mapping. This is an
attribution tranche, not a source, VM, or stdlib change.

## Method

Each profile used the direct bytecode runtime benchmark, which loads and warms
the program before profiling repeated `main()` calls. Every run used the
canonical `able-stdlib` source, `taskset -c 2`, and
`GOMEMLIMIT=1GiB GOGC=50 GOMAXPROCS=1`; type checking was skipped only for the
already-covered performance fixture path. The retained CPU profiles are:

- `20260711_string_split_join_post_ownership_5x.cpu.pprof`
- `20260711_iterator_collect_post_ownership_5x.cpu.pprof`
- `20260711_array_map_i32_post_ownership_20x.cpu.pprof`

under `v12/interpreters/go/.profiles/`. The longer numeric capture avoids
trying to infer a leaf from a short profile.

| Workload | Iterations | Result | CPU samples |
| --- | ---: | --- | ---: |
| `string_split_join_small` | 5 | 1,034,507,713 ns/op; 49,141,425 B/op; 549,673 allocs/op | 5.06 s |
| `linked_list_iterator_collect_i64_small` | 5 | 270,593,566 ns/op; 3,256,209 B/op; 29,091 allocs/op | 1.31 s |
| `array_map_i32_small` | 20 | 77,418,334 ns/op; 852,688 B/op; 332 allocs/op | 1.51 s |

## Attribution

`execCallOpcode` and `finishInlineReturn(...)` occur in all three captures,
but they are dispatcher/boundary parents rather than one shared actionable
descendant:

| Workload | `finishInlineReturn` cumulative cost | Material descendant / context |
| --- | ---: | --- |
| Split/join | 1.07 s (21.15%) | `coerceInlineProgramReturnValue` (9.29%), including cached canonical return coercion and text codec typed-pattern/type-match work. |
| Iterator collect | 0.22 s (16.79%) | Generator/member dispatch, `IteratorEnd`/`Self` checks, and `popCallFrameFields` (7.63%). |
| Numeric Array map | 0.13 s (8.61%) | Raw i32 slot transport, Array get/push calls, binary arithmetic, and `popCallFrameFields` (3.31%). |

The one exact low-level common function, `popCallFrameFields`, is only
3.75%, 7.63%, and 3.31% respectively. Its samples are pooled-frame hygiene,
sidecar restoration, and active-lookup restoration across different frame
kinds; there is no single guard or allocation to remove without changing the
existing frame contract. The previously rejected return-guard reordering is
therefore not reopened.

Likewise `bytecodeRawIntegerValueInfo(...)` is visible at 4.74%, 3.05%, and
2.65%, but the earlier direct raw-carrier bypass regressed the broad controls,
and the callers remain text-store/bitwise, iterator protocol, and numeric
transport paths. It is not a justified shared candidate. Type matching is
material only in the text and iterator controls, while numeric Array map is
almost entirely outside that semantic boundary.

## Decision

Keep no runtime, compiler, tree-walker, or `able-stdlib` code. The refreshed
profiles rule out treating return completion, raw extraction, or call dispatch
as one broadly profitable leaf. A change here would either repeat a rejected
micro-reordering or encode one workload's descendant into a generic VM path.

## Next recommendation

Use the bounded external bytecode scorecard to select two independently
verified application misses, then collect their profile pair plus an unrelated
guard before considering a candidate. This is preferable to another pass over
the same fixture trio because the fixture refresh has already separated the
apparent shared parents into distinct text, iterator, and numeric costs. The
selected work must retain the same CPU/OOM limits and upstream verifier gates;
only a concrete generic descendant repeated across both applications can justify
a VM/runtime or canonical-stdlib change.
