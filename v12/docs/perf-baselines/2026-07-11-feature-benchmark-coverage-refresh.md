# v12 feature-to-benchmark coverage refresh (2026-07-11)

## Purpose

Re-audit the active v12 feature surface after the external bytecode profile
pair found no shared optimization leaf. A benchmark is useful for performance
selection only when it is application-shaped and its comparison/verification
status is explicit; conformance coverage alone does not supply a timing target.

## Current corpus

The fresh `bench_bytecode_audit --suite corpus-full` lowering audit completed
with 92 programs, 352 lowered functions, and 16,815 instructions. The catalog
contains 15 general external applications, special verified text/collection/
channel applications, and 77 local fixtures. It continues to cover ordinary
control flow, primitive and nominal values, generic interfaces/iterators,
strings/bytes/files/host-backed stdlib calls, arrays/collections, numeric
kernels, and normal static-module imports.

The remaining feature-status matrix is:

| v12 feature family | Application/fixture coverage | Cross-language timing status |
| --- | --- | --- |
| Channels, `spawn`, Future values, and scheduler flush | Channel-Rollup, BinaryTrees, and channel/future fixtures | Channel applications have verified references; Future-specific paths are local-only. |
| `await` coordination | New `await_batch_i64_small` fixture | Local-only; no fair Go/Python/Ruby application is currently in the catalog. |
| Cancellation and mutex contention | Await-cancellation and mutex fixtures | Local-only. |
| `rescue`/`ensure`/rethrow | Conformance fixtures and error-path stdlib fixtures | Local-only. |
| Dynamic packages, `dynimport`, and late binding | Conformance fixtures | Local-only. |
| Static packages/imports and host interop | All external applications; file, codec, JSON, and OS paths | Broad, verifier-backed where the external suite supplies a verifier. |

The local-only rows are intentional. A synthetic busy loop would not make
`await`, cancellation, dynamic import, or exception handling comparable across
Go, Python, and Ruby: their schedulers and exception/module models differ. A
new cross-language row requires an independent application and a matching
reference/verifier, not merely an Able benchmark spelling.

## Added await application fixture

`fixtures/bench/await_batch_i64_small` is a deterministic batched-checksum
application. Each round spawns two checksum workers, spawns a collector that
awaits both Future values, flushes the scheduler, and reduces the completed
batch. It exercises `spawn`, Future completion, `await`, closure capture,
return/error propagation, and scheduler flushing without an Await-default
micro-loop.

The fixture is registered in `fixture-concurrency`, lowers to three functions
and 98 bytecode instructions, and has the checked output `32835560`. Under
`taskset -c 2` with `GOMEMLIMIT=1GiB GOGC=50 GOMAXPROCS=1`, it completed in
compiled (0.09s), tree-walker (0.18s), and bytecode (0.17s) modes. The
focused bytecode spawn/await and Future/cancellation suite also passes.

## Decision

Keep the new benchmark coverage and no runtime, compiler, or `able-stdlib`
optimization. One local await program is a semantic/performance guard, not
authorization for a scheduler fast path.

## Next recommendation

Use the await fixture only as one side of a guarded concurrency attribution:
compare it with the independently application-shaped Channel-Rollup and a
Future-yield or mutex control under the appropriate executor. Profile a change
only if the same concrete scheduler/runtime descendant is material in at least
two of those workloads and output remains deterministic. This preserves the
cross-program rule while letting the new feature coverage influence future
bytecode work.
