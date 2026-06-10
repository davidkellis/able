# Bytecode await/concurrency attribution (2026-07-11)

## Purpose

Test the new `await_batch_i64_small` coverage fixture only as a guard against
an application-shaped concurrent workload. This asks whether a concrete
bytecode scheduler/runtime cost recurs in independent programs; it does not
authorize an optimization merely because all three programs use `spawn`.

## Method

- Used normal bytecode processes, not the repeated-`main` runtime benchmark.
  Reusing a process would retain executor-sensitive state and would therefore
  measure a different concurrency contract.
- Pinned execution to CPUs `2-3` with `GOMEMLIMIT=1GiB`, `GOGC=50`, and
  `GOMAXPROCS=2`; all three workloads explicitly selected the goroutine
  executor. `ABLE_GO_CPU_PROFILE` captured the full ordinary
  load/lower/execute path.
- `await_batch_i64_small` and `future_yield_i32_small` ran from their fixture
  directories. Running a fixture from the repository root correctly triggers
  the CLI's stdlib-collision protection because that root is itself named
  `able`; the fixture directory has no competing stdlib root.
- The await preflight printed `32835560`; Future-yield printed `163562`; and
  Channel-Rollup printed `16384:4828:502100`. Channel-Rollup captures also
  used its upstream Ruby verifier. Every capture process exited successfully.
- Three independently started await and Future-yield processes were merged
  solely to obtain a minimally useful normal-process sample. Channel-Rollup
  supplied two independent verified processes. No process ran `main` more
  than once.

The retained aggregate captures are:

- `.profiles/20260711_await_batch_goroutine_bytecode_process.cpu.pprof`
  (180 ms sampled)
- `.profiles/20260711_future_yield_goroutine_bytecode_process.cpu.pprof`
  (280 ms sampled)
- `.profiles/20260711_channel_rollup_goroutine_bytecode_process.cpu.pprof`
  (1.50 s sampled)

## Result

| Workload | Material sampled work | Concrete bytecode consequence |
| --- | --- | --- |
| Await batch | Loader 66.67%; cgo parser boundary 50.00%; `runResumable` 5.56% | The small application does not expose a scheduler leaf above bootstrap noise. |
| Future-yield | Loader 53.57%; cgo parser boundary 39.29%; `runResumable` 7.14% | Likewise, no material `future_yield` or task bookkeeping leaf is sampled. |
| Channel-Rollup | `runResumable` 42.67%; `GoroutineExecutor.runTask` 29.33%; call-opcode path 25.33%; atomic add 4.67% flat | Its material work is the real channel producer/worker/consumer path, not a cost repeated in both controls. |

`runResumable` and `runtime.cgocall` appear in all three captures, but they
are broad parents: the former contains distinct workload execution and the
latter is substantially loader/tree-sitter work in the short local controls.
They are not removable VM leaves. Channel-Rollup's task runner, atomic
bookkeeping, call dispatch, raw materialization, channel receive, and
type-match descendants do not recur materially in either await or
Future-yield. Conversely, the controls provide too little in-main work to
claim that an absent sample proves a scheduler cost is zero.

## Decision

Keep no runtime, compiler, or `able-stdlib` change. Do not tune
`GoroutineExecutor.runTask`, channel capacity, atomic bookkeeping, or a
particular spawn/await call shape from one Channel-Rollup profile. Enlarging a
local fixture merely to manufacture a profile target would also violate the
application-shaped benchmark rule.

## Next recommendation

Return to the verified external-scorecard misses rather than padding the
local concurrency controls. The next attribution should select two current
verified bytecode misses from different workload families plus a verified
neutral control, then profile normal processes only if a concrete leaf is
plausibly shared. The concurrency branch has no second application-level
reference workload yet, so another scheduler micro-tranche would add noise
rather than move the interpreter toward Python/Ruby parity.
