# Compiled equal-CPU concurrency reconciliation — 2026-07-17

## Scope

This tranche corrects the compiled concurrency comparison contract. The prior
Binary Trees row constrained Able to one Go scheduler P while its Go reference
reset `GOMAXPROCS` to twice the visible host CPU count. That result measured
different CPU budgets and could not decide whether recursive nominal
allocation was a product-level compiler priority.

The refreshed gate gives both binaries the same four logical CPUs (`0-3`). It
uses `GOMAXPROCS=4`, `GOGC=50`, `GOMEMLIMIT=1GiB`, Able's explicit
`goroutine` executor, a 60-second per-process cap, and five independent
processes per implementation. `taskset` remains the hard CPU-budget boundary
even when an application changes `GOMAXPROCS` internally. Every reported
process passed its application's Ruby verifier.

The applications are deliberately unlike:

- Binary Trees recursively allocates and traverses nominal nodes;
- Future Pipeline fans numeric work through Futures and channels;
- Channel Rollup streams file/text work through channel workers; and
- Await Channel Mux exercises await-arm selection over channel readiness.

No benchmark, reference implementation, compiler, runtime, stdlib, fixture,
specification, or WASM source changed.

## Equal-budget results

The ratio is Able elapsed time divided by Go elapsed time, so lower is better.
Means, medians, and ranges are from five verifier-backed processes.

| Application | Able mean | Able median (range) | Go mean | Go median (range) | Able / Go | Decision |
| --- | ---: | ---: | ---: | ---: | ---: | --- |
| Binary Trees | 10.0580s | 10.0800s (9.7900–10.3400) | 10.9153s | 10.9681s (10.5814–11.0915) | **0.921x** | passes the 95%-of-Go target; Able uses 7.9% less elapsed time |
| Future Pipeline | 0.3520s | 0.3500s (0.3000–0.4100) | 0.0056s | 0.0056s (0.0052–0.0058) | 62.86x | material compiled scheduler/runtime miss |
| Channel Rollup | 1.5160s | 1.5000s (1.2000–1.9600) | 0.0056s | 0.0054s (0.0051–0.0068) | 270.71x | material compiled scheduler/runtime miss |
| Await Channel Mux | 0.3920s | 0.3900s (0.3700–0.4300) | 0.0045s | 0.0046s (0.0042–0.0047) | 87.11x | material compiled scheduler/runtime miss |

All 20 Able launches and all 20 Go launches verified. Binary Trees therefore
changes classification: its earlier multi-fold miss was a CPU-contract
artifact, and recursive `Node` allocation is not the next compiler target.
The other three rows remain genuine misses. Their very small Go elapsed times
also mean that fixed process/runtime startup is part of the application
contract rather than evidence for a source-specific loop rewrite.

## Bounded profile attribution

One additional normal compiled process per scheduler-heavy application passed
the same verifier and wrote a main-process CPU profile. The samples converge
on one exact generic descendant:

| Application | CPU samples | `bridge.currentGID` / `runtime.Stack` cumulative |
| --- | ---: | ---: |
| Future Pipeline | 240ms | 220ms (91.7%) |
| Channel Rollup | 1.03s | 960ms (93.2%) |
| Await Channel Mux | 300ms | 290ms (96.7%) |

`bridge.currentGID` derives a goroutine identifier by asking `runtime.Stack`
for and parsing a stack header. The bridge uses that identity to recover the
per-task environment and call-frame state at generated/runtime boundaries.
Future completion, channel send/receive, and await selection are different
parents, but all repeatedly pay this same identity-discovery leaf.

This is strong attribution, but not a new implementation candidate. The
generic fixed execution-context ABI already removed this lookup. Its broad
default scorecard found a stable 54.7% N-body regression caused by allocating
new context objects at dense cross-package calls. Subsequent allocation-free
variants and normal-binary profile gates did not produce a broadly safe
replacement, and the project record explicitly closes retrying that ABI.
Removing `currentGID` locally without a proven task-local propagation contract
would risk nested compiled calls, dynamic/native compatibility entries,
cancellation payload isolation, and environment restoration.

## Decision

- Retain the equal-CPU measurement correction: Binary Trees currently meets
  the compiled 95%-of-Go objective under a matched four-CPU contract.
- Keep no production optimization. The three remaining gaps reproduce the
  already-rejected execution-context identity design, not an untried shared
  leaf.
- Do not add a `Node`, Future, Channel, await, application, task-count, or
  named-container special case.
- Do not change canonical `able-stdlib`: the measured wall is in the generic
  compiled runtime bridge.

## Commands

Go references were refreshed with `bench_refresh_go_refs --runs 5
--timeout 60 --cpu-affinity 0-3 --gomaxprocs 4`. Able comparisons used
`bench_compare_external --modes compiled --languages go --runs 5 --timeout 60
--cpu-affinity 0-3` with `GOMAXPROCS=4`, `GOGC=50`, and
`GOMEMLIMIT=1GiB`. The catalog selects the `goroutine` executor for these
applications. Profiles used one verifier-backed `bench_perf` compiled process
per application with the same environment and affinity.

## Next gate

Return to the bytecode product gap with a fresh equal-contract cross-language
screen over the current application catalog, then profile at least three
unlike interpreter misses. This is the best next use of effort because the
compiled Binary Trees concern is resolved and the only shared compiled
concurrency leaf has already failed its generality gate. A bytecode candidate
is eligible only if the same concrete VM descendant repeats materially across
three unlike verified programs and survives broad controls; parent dispatcher,
named-container, workload, and source-shape shortcuts remain ineligible.
