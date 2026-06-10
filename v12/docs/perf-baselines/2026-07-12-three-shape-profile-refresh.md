# Three-shape bytecode and compiled refresh (2026-07-12)

## Decision

Keep no VM, compiler, runtime, benchmark, fixture, or canonical-stdlib source
change from this tranche. Fresh profiles of text split/join, iterator
collection, and numeric Array map again show shared dispatcher parents but no
shared, material implementation leaf. The adjacent one-run compiled controls
also complete successfully, but are process controls rather than evidence for
a compiler optimization.

In particular, do not reopen either prior generic experiment:

- preserving a mixed-coercion direct-call raw cell; or
- reordering the slotless `finishInlineReturn(...)` guards.

Both were already neutral or worse on broad guards. This refresh shows why:
the costs below the apparent `execCallOpcode` and return parents differ by
program shape.

## Method

The bytecode profiles use a preloaded, warmed VM and repeated `main()` calls.
All commands used the canonical external stdlib,
`GOMEMLIMIT=1GiB`, `GOGC=50`, and `GOMAXPROCS=1`:

```text
cd v12/interpreters/go
ABLE_STDLIB_ROOT=/home/david/sync/projects/able-stdlib/src \
ABLE_BENCH_RUNTIME_TARGET=<fixture>/main.able \
ABLE_BENCH_RUNTIME_RUN_FROM=<fixture> \
go test ./pkg/interpreter -run '^$' -bench '^BenchmarkBytecodeProgramRuntime$' \
  -benchmem -count=1 -benchtime=<iterations>x -cpuprofile=<profile>
```

The retained CPU profiles are in `v12/interpreters/go/.profiles/`:

- `20260712_string_split_join_bytecode_refresh_5x.cpu.pprof`
- `20260712_iterator_collect_bytecode_refresh_20x.cpu.pprof`
- `20260712_array_map_i32_bytecode_refresh_75x.cpu.pprof`

The compiled process controls use `v12/bench_perf --source-root-only` with an
explicit `--run-from` fixture directory. That prevents the workspace module
from colliding with the canonical-stdlib module selected by the normal cache.
Their JSON records are in `v12/tmp/perf/`.

| Workload | Bytecode iterations | Bytecode result | CPU sample | Compiled process control |
| --- | ---: | ---: | ---: | ---: |
| `string_split_join_small` | 5 | 1,018,463,625 ns/op; 49,556,008 B/op; 555,709 allocs/op | 5.07 s | 0.41 s real; 0.28 s user; 7 GCs |
| `linked_list_iterator_collect_i64_small` | 20 | 320,679,221 ns/op; 3,207,429 B/op; 29,002 allocs/op | 6.38 s | 0.22 s real; 0.10 s user; 3 GCs |
| `array_map_i32_small` | 75 | 80,247,446 ns/op; 848,087 B/op; 323 allocs/op | 5.98 s | 0.16 s real; 0.06 s user; 3 GCs |

The compiled fixture controls have deterministic stdout hashes, but no
fixture verifier. They are therefore retained as bounded build-and-execution
sanity checks, not as Able-versus-Go ratios or variance claims.

## Attribution

| Workload | Shared-looking parent | Material descendant / context | Why it is not a generic candidate |
| --- | --- | --- | --- |
| Split/join | `execCallOpcode` 27.02% cumulative; `finishInlineReturn` 16.57% | Call-name lookup, string-key map lookup, and program-return coercion; `bytecodeRawIntegerValueInfo` is 6.11% flat. | Its cost is text/call-name and map traffic, not iterator or Array transport. |
| Iterator collect | `execCallOpcode` 57.37%; `execCallMember` 41.07% | `next` member dispatch, callable struct fields, exact native calls, typed-pattern checks, and frame-pop work; raw integer info is 6.58% flat. | The generic-iterator protocol and member resolution do not recur in the text or numeric leaf path. |
| Numeric Array map | `execCallOpcode` 34.45%; `finishInlineReturn` 6.35% | Raw `i32` slot extraction is 9.20% flat, plus Array get/push and arithmetic. | The scalar/Array lane has little return coercion and no iterator/text descendant to share. |

`runResumable` is expected to dominate all three profiles, and
`execCallOpcode`/`finishInlineReturn` are broad control-flow parents. Neither
names an operation that can be removed safely. The only repeated low-level
item, raw integer inspection, has a different carrier/caller mix in every
row, and its previously tested generic ordering change regressed broad
controls. The refreshed evidence does not justify a third attempt.

## Next recommendation

Do not repeat the already completed and rejected universal runtime-value
carrier design. Resume language/stdlib feature completion, starting with the
remaining reusable regex API surface: `RegexBuilder` and automata
introspection/export. Those are normal general-purpose text APIs, can be
implemented in the canonical external stdlib without a runtime or
benchmark-shaped branch, and will add new representative program shapes to
the corpus. Keep the same rule after that work: collect a profile only when a
newly exercised boundary recurs in independently verified applications, and
make a VM or AOT change only when its concrete descendant clears the full
generality guard matrix.
