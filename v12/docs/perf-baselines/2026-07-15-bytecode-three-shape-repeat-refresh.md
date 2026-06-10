# Bytecode three-shape repeat refresh — 2026-07-15

## Decision

Keep no bytecode VM, compiler, canonical-stdlib, or benchmark-specific
performance change. Five independently launched warmed bytecode processes per
control confirm the current controls execute correctly, but their profiles do
not identify a new material concrete leaf shared by text, iterator, and
numeric work.

One correctness repair landed before measurement: dynamically materialized
interface default methods now retain their enclosing interface generic
parameters in their `MethodSet`. Without that metadata, a valid generic return
such as `Iterator<T>` could be coerced as though `T` were concrete. The fix is
covered both by the real iterator-collect warmup and a focused default-method
metadata regression; it does not add a performance fast path.

## Method

Every row used the canonical external stdlib, `--source-root-only`, one load
and warmup per process, and `GOMEMLIMIT=1GiB`, `GOGC=50`, and `GOMAXPROCS=1`.
No CPU affinity or quiet-host admission gate was used. `bench_perf --runs 5`
launches five independent benchmark processes and reports their average. The
single profile run for each shape is attribution-only and is not compared with
the repeated timing average.

| Workload | Calls | Independent runs | Average ns/op | B/op | allocs/op |
| --- | ---: | ---: | ---: | ---: | ---: |
| `string_split_join_small` | 5 | 5/5 | 937,221,187 | 49,385,712 | 555,697 |
| `linked_list_iterator_collect_i64_small` | 20 | 5/5 | 413,442,232 | 8,420,619 | 192,900 |
| `array_map_i32_small` | 75 | 5/5 | 60,996,096 | 805,651 | 113 |

Direct bytecode output checks completed before the timed screen: split/join
printed `191484`, iterator collect printed `382455000`, and numeric Array map
printed `1097192358`.

The generated runner records and CPU profiles are cleanup-eligible evidence in
`v12/interpreters/go/.profiles/`:

- `20260715_refresh_{string_split_join,iterator_collect,array_map_i32}_repeats.json`
- `20260715_refresh_{string_split_join,iterator_collect,array_map_i32}_profile.json`
- `20260715_refresh_{string_split_join,iterator_collect,array_map_i32}.cpu.pprof`

## Attribution

| Frame or leaf | Text | Iterator | Numeric | Decision |
| --- | ---: | ---: | ---: | --- |
| `execCallOpcode` cumulative | 23.19% | 51.99% | 37.47% | A dispatcher parent: named calls/type matching in text, generator/member dispatch in iterator work, and Array slot calls in numeric work. |
| `finishInlineReturn` cumulative | 19.26% | 8.69% | 10.11% | The shared parent reaches different returns. The broad slotless-return guard experiment was already neutral to mixed and remains closed. |
| `bytecodeRawIntegerValueInfo` flat | 4.38% | 2.24% | 8.74% | Different callers and carriers; the generic raw-extraction candidate regressed text and iterator guards despite numeric benefit. |
| `runtime.mapaccess2_faststr` cumulative | 9.19% | 5.87% | 4.37% | Callers mix `integerInfos`, known-type cache, environment lookup, and nominal matching. The only fixed common map, `integerInfos`, already lost to a generic switch replacement. |

Text retains named-call, return-coercion, and type-match work. Iterator collect
is mostly generator execution plus member-cache and static-member dispatch.
Numeric mapping is Array slot get/push, primitive transport, and binary work.
None is a language-wide concrete descendant that can support a single
optimization without specializing a benchmark, collection, or protocol.

## Next recommendation

Use the verifier-backed external scorecard to select a new pair or trio whose
dominant miss crosses a language boundary not represented by these controls,
then run the same repeated attribution gate. This is necessary because the
remaining local call/frame, raw-integer, and fixed-map ideas are either parent
frames or previously rejected generic candidates; another local rewrite would
optimize noise rather than real programs.
