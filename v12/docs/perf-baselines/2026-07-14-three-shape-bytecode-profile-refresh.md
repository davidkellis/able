# Three-shape bytecode VM profile refresh — 2026-07-14

## Decision

Keep no bytecode VM, compiler, runtime, canonical-stdlib, or benchmark-source
change. The refreshed split/join, iterator-collect, and numeric Array-map
profiles again share VM dispatch parents, but not one new concrete descendant.
The existing raw-integer, call-name/frame, return-guard, Array-slot-cache, and
member-dispatch experiments are therefore not reopened.

## Method

Each CPU run used the canonical external stdlib, CPU 15, `GOMEMLIMIT=1GiB`,
`GOGC=50`, and `GOMAXPROCS=1`. The bytecode-runtime benchmark loads and warms
one program before measuring it. A `45s` outer guard and Go `40s` test guard
bounded the normal profiles.

| Workload | Measured run | Result |
| --- | ---: | --- |
| split/join text | 5x | 1,027,600,654 ns/op; 49,454,032 B/op; 555,696 allocs/op |
| linked-list iterator collect | 20x | 434,414,436 ns/op; 8,094,335 B/op; 152,700 allocs/op |
| numeric Array map | 75x | 71,007,557 ns/op; 805,342 B/op; 94 allocs/op |

An independent ten-run iterator check was stable at 432,017,047 ns/op,
8,119,777 B/op, and 152,738 allocs/op. The retained CPU profiles are:

- `v12/interpreters/go/.profiles/20260714_three_shape_string_split_join.cpu.pprof`
- `v12/interpreters/go/.profiles/20260714_three_shape_iterator_collect.cpu.pprof`
- `v12/interpreters/go/.profiles/20260714_three_shape_array_map.cpu.pprof`

They are generated evidence and are cleanup-eligible.

One-run traces classified dispatch only; tracing takes a lock per recorded
call, so the trace timings were not used as performance measurements. Their
reports are cleanup-eligible under `v12/tmp/perf/`:

- `20260714_three_shape_string_split_join_trace.json`
- `20260714_three_shape_iterator_collect_trace.json`
- `20260714_three_shape_array_map_trace.json`

## Attribution

| Workload | Material route | Why it is not a shared candidate |
| --- | --- | --- |
| split/join | `execCallOpcode` 26.66% cumulative, named calls/inline return/type matching, tracked-Array reads in UTF-8 string helpers | The trace is led by `array_get_tracked_fast`, `read_byte`, `utf8_decode`, and character conversion. |
| iterator collect | `execEnsureStart` 84.50% and call dispatch 79.50% cumulative, enclosing iterator `next`/`yield`/inline work | `ensure` is an execution envelope, not repeated lowering: expression-program caching already caches the AST lowering. The trace is led by iterator `next`, `yield`, and visitor calls. |
| numeric Array map | call dispatch 34.42%, Array-slot calls 15.66%, raw integer extraction 6.20% | The trace is led by mono-`i32` Array `get`, `push`, and `len`, a different primitive-array route. |

The iterator `ensure` reading was checked against the unlike `mutex_counter`
control, whose `able.concurrency.with_lock` also uses `ensure`. A single bounded
run completed at 5,670,282,594 ns/op and showed `execEnsureStart` at 10.18%
cumulative. That confirms it is measuring the protected expression's work,
not a standalone lowering leaf: `execEnsureStart` calls the already-cached
nested bytecode program and contains no repeated compiler work to remove.
The retained control profile is
`v12/interpreters/go/.profiles/20260714_ensure_mutex_counter.cpu.pprof`.

`Array.push` appears in all three traces, but through tracked state in iterator
collect and numeric mapping and a handle-backed UTF-8-string path in split/join.
The CPU descendants and value representations differ; previous broad Array
slot/cache and raw-carrier variants also already rejected this parent-level
inference. No Array, iterator, text, mutex, or benchmark-shaped specialization
is authorized.

The three compiled binaries were also built and run once under the same CPU and
memory settings (0.23s split/join, 0.21s iterator, 0.09s numeric map). Their
fixture manifests have no external verifier, so these are build/run sanity
checks only and not comparative compiler evidence.

## Next recommendation

Refresh the verifier-backed external scorecard before selecting more runtime
work. This profile set exhausts the immediately proposed text/iterator/numeric
micro-candidates and shows that another VM edit would target a parent or one
family. A current scorecard will identify real misses across the whole
application suite; profile only a pair of unlike misses if they repeat a new,
concrete descendant that has not already failed broad guards.
