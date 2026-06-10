# ArrayStore Synchronization Guard

This bounded tranche completed the generic process-wide ArrayStore safety
repair and checked its cost across independent program shapes. It adds no
benchmark-specific, named-container, compiler, or stdlib rule.

## Repair

The first tree-walker goroutine repair covered dynamic handles and
interpreter-owned array aliases. A focused race regression then showed that
the same registry remained unsafe for primitive monomorphic handles: the
constructors, primitive direct read/write paths, and primitive-read hot cache
all mutate or read process-global maps/caches.

The retained design gives the shared registry one ownership boundary:

- All dynamic and monomorphic handle publication, promotion, cloning, reserve,
  and mutation paths synchronize with `arrayStoreMu`.
- Primitive direct reads use the shared read boundary; mutations use the write
  boundary. The bytecode primitive-read cache is additionally serialized while
  it updates its process-global hot entries.
- Registry lookup is read-only. It no longer fills caches as a side effect,
  so a reader cannot mutate registry state while another task publishes a
  handle.
- Revision/type-name hot caches are protected on their own read paths. The
  cursor remains a cache hint, not a synchronization primitive for externally
  shared mutable values.

This is a generic runtime ownership repair; it applies to every primitive
carrier (`i32`, `i64`, `bool`, `char`, `u8`, `u32`, `u64`, `f64`) and dynamic
arrays. No Able program or stdlib container is recognized by name.

## Verification

Under `go test -race`, the regressions repeatedly create independent dynamic
handles plus all eight monomorphic handle types, write/read them directly, and
exercise `ArrayStoreMonoPrimitiveReadInfoInto`. The runtime concurrency set
passed five times; the tree-walker spawned-array and channel/Future subset
passed three times. The real tree-walker Channel-Rollup CLI still printed
`16384:4828:502100` with `ABLE_EXECUTOR=goroutine`.

Focused bytecode return/call-name regressions and the full runtime/interpreter
package tests also passed. A stale locally built benchmark binary exposed one
lock nesting in `ArrayStoreWrite`'s typed bridge; the bridge now uses the
already-held internal write helper, and a fresh bytecode runtime benchmark
completes normally.

## Cost guard

All sequential rows used the canonical external stdlib plus
`GOMEMLIMIT=1GiB`, `GOGC=50`, and `GOMAXPROCS=1`. Each bytecode-runtime row is
three warmed one-process `main()` invocations. Compiled rows build once and
run three processes; their coarse wall clock is context, not a microbenchmark.
The pre/post readings are same-session guard readings surrounding only this
completion of the synchronization repair, but are still too small to claim a
sub-percent win.

| Program | Mode | Before | After | Decision |
| --- | --- | ---: | ---: | --- |
| `array_map_i32_small` | Bytecode runtime | 59,893,425 ns/op; 910,808 B/op; 433 allocs/op | 62,383,009 ns/op; 910,808 B/op; 433 allocs/op | +4.16%; inside the 5% broad guard, with unchanged allocation shape. |
| `linked_list_iterator_collect_i64_small` | Bytecode runtime | 243,456,255 ns/op; 3,445,000 B/op; 29,336 allocs/op | 239,850,325 ns/op; 3,445,000 B/op; 29,336 allocs/op | -1.48%; normal run noise, no allocation change. |
| `array_map_i32_small` | Compiled | 0.0900 s | 0.0900 s | Unchanged at timer resolution. |
| `linked_list_iterator_collect_i64_small` | Compiled | 0.1233 s | 0.1267 s | +2.76%; within coarse three-run process timing. |
| Channel-Rollup | Bytecode, goroutine executor | 0.4100 s (earlier one-run context) | 0.4133 s (three-run average) | No material concurrent regression; output verified. |

The one-off Channel-Rollup bytecode run at 0.5300 seconds was rejected as
noise by the immediately repeated three-run row. It is not used for selection.
Neither profile nor application-specific mitigation is justified: the two
independent array/iterator bytecode workloads remain within the guard, the
compiled checks have no material shift, and concurrent Channel-Rollup remains
stable.

## Decision and next recommendation

Keep the generic ArrayStore synchronization repair. Do not weaken it for a
particular primitive type, executor, benchmark, or stdlib container.

Next, rebuild and re-run the isolated Channel-Rollup Docker publication after
this runtime change, then refresh a small multi-run external target-miss
ledger before choosing another optimization. This is needed because the
published Docker image predates the completed monomorphic-handle repair; fresh
all-mode output verification will keep the cross-language claims tied to the
actual runtime. The work entails rebuilding the three Able images, retaining
the Go/Python/Ruby reference rows as provenance, and collecting a CPU profile
only if a repeated, concrete leaf appears in at least two target-miss
applications. It should not introduce a scheduler, channel, container, or
benchmark-shaped fast path.
