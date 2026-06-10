# Compiled async main-phase refresh — 2026-07-14

## Decision

Keep no compiler, bridge, generated-runtime, bytecode VM, canonical-stdlib, or
benchmark-source change. The refreshed generated-binary profiles repeat the
known goroutine-identity wall in all three unlike async applications, but the
only known general remedy has already failed independent broad-performance
guards. Re-running that rejected ABI experiment would add neither new evidence
nor a generally applicable optimization.

## Method

The compiler was built from the current tree and used to generate fresh
binaries for Channel Rollup, Future Pipeline, and Future Await Race. Each
binary used its catalog command-line arguments and working directory. Eight
independently launched processes per application ran on CPU 15 with a
45-second cap, `GOMEMLIMIT=1GiB`, and `GOGC=50`; `GOMAXPROCS` was deliberately
unset because a one-P cap changes the goroutine scheduling being attributed.

Each process emitted the generated binary's CPU-only post-bootstrap `main`
phase profile through `ABLE_GO_PHASE_CPU_PROFILE_DIR`. The eight profiles were
merged per application with `go tool pprof -proto`. Every completed execution
passed that application's exact canonical Ruby verifier. These profiles omit
compiler, parser, and package-registration startup work; they are not timing
claims and do not replace the CPU-pinned external scorecards.

## Result

| Application | Launches | Merged samples | `bridge.currentGID` cumulative | `runtime.Stack` cumulative | Task wrapper cumulative |
| --- | ---: | ---: | ---: | ---: | ---: |
| Channel Rollup | 8 | 11.83 s | 11.09 s (93.74%) | 10.98 s (92.81%) | 10.39 s (87.83%) |
| Future Pipeline | 8 | 5.37 s | 5.05 s (94.04%) | 5.02 s (93.48%) | 3.62 s (67.41%) |
| Future Await Race | 8 | 280 ms | 180 ms (64.29%) | 170 ms (60.71%) | 180 ms (64.29%) |

The shared concrete leaf is again `bridge.currentGID`, which calls
`runtime.Stack` to discover goroutine identity. Future Await Race also shows
130 ms (46.43%) in its generated await helpers, but those frames are callers
of the same identity path rather than a second repeated descendant. The other
two applications differ materially below the task wrapper, so no new shared
generated-runtime leaf emerged.

## Why no candidate is eligible

The generic fixed-pointer execution-context ABI remains the only known way to
remove this identity lookup. It is intentionally opt-in: its default-ABI gate
repeated a 54.7% N-body regression, and its allocation-free package-linkage
refinement was reverted after a separate six-run K-Nucleotide mean was 16.6%
slower. Those are independent verifier-backed application guards, not
concurrency timing noise. This refresh confirms that the cost is still shared;
it does not invalidate either disqualifying result. No prototype was rerun.

The profile artifacts and generated binaries were temporary, local diagnostic
outputs and were removed after this record was written. No `able-stdlib`
change was needed.

## Next recommendation

Do not invent a new portable application merely to avoid revisiting this
rejected concurrency route. The roadmap reconciliation found no unfinished v12
semantic boundary that can honestly add such a row; the active remaining work
is the separate WASM host ABI. Take its narrow stdout/stderr forwarding slice
next, with Node smoke coverage, and keep it outside performance selection.
Existing verifier-backed applications remain the only authority for a future
compiler or bytecode optimization.
