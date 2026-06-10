# Spawn-selected execution-context gate — 2026-07-14

## Decision

Keep no compiler, bridge, generated-runtime, bytecode VM, canonical-stdlib, or
benchmark-source change from this tranche. A temporary compiler candidate
selected the existing fixed-pointer execution-context ABI only when a
statically loaded user package contained Able's language-level `spawn` syntax.
It left serial programs and dormant canonical-stdlib declarations on the
compatibility ABI, so it was materially different from the earlier
program-wide fixed-context default experiment.

The candidate improved two independent async applications but regressed
`Mutex Ledger` by 10.0% in the same matched three-run screen. It was fully
reverted. Selecting an ABI from `spawn` is a language-level rule, but that does
not make a trade between public concurrency applications acceptable.

## Method

The candidate scanned only user-package ASTs for `SpawnExpression`; packages
named `able` or `able.*` were excluded so importing a dormant stdlib helper
could not alter a serial program's ABI. Temporary source guards proved:

- user `spawn` generated the context-aware task, static call, and kernel-helper
  entries;
- a serial program emitted no execution-context ABI; and
- a canonical-stdlib-only `spawn` declaration did not select the ABI.

The selector was then measured against a matching compatibility-ABI build with
the existing `bench_compare_external` harness. Every row used CPU 15,
`GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, the goroutine executor, a
45-second per-process cap, three independent compiled launches, and the
application's canonical Ruby verifier. Each baseline/candidate pair emitted
the same stdout digest.

| Application | Compatibility ABI | Spawn-selected ABI | Change |
| --- | ---: | ---: | ---: |
| Channel Rollup | 1.3133 s | 1.1933 s | -9.1% |
| Future Pipeline | 0.7133 s | 0.6100 s | -14.5% |
| Future Await Race | 0.1233 s | 0.1200 s | -2.7% |
| Await Channel Mux | 0.4000 s | 0.3933 s | -1.7% |
| Mutex Ledger | 0.5333 s | 0.5867 s | **+10.0%** |
| Mutex Await Journal | 0.4633 s | 0.4700 s | +1.4% |

## Why it is rejected

`bridge.currentGID` / `runtime.Stack` remains a real common compiled
concurrency wall, but its known general context remedy is not uniformly
beneficial. The spawn-selected rule avoids prior serial N-body and K-Nucleotide
losses, yet it still regresses an unlike application that also exercises the
language's public concurrency model. Adding a Mutex-, channel-, callback-, or
task-shape exception would violate the generality policy.

The temporary compiler files and the matched baseline/candidate workspaces
were removed after this decision. The external `able-stdlib` checkout did not
need a change.

## Next recommendation

Do not retry the unchanged fixed-context family, including program-wide,
payload-only, package-linkage, or spawn-selected variants. Return to the
verifier-backed suite only when a different concrete compiler or bytecode leaf
repeats across unlike applications; retain the existing scorecards as
regression gates in the meantime.
