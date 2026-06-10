# Compiled concurrency fourteen-application refresh

Date: 2026-07-21

## Decision

Keep no compiler, bridge, runtime, canonical-stdlib, language, benchmark,
reference, or WASM change. Eleven current concurrency applications reproduce
`bridge.currentGID` through `runtime.Stack` as an exact shared wall, but the
existing opt-in fixed execution-context ABI still does not remove that wall at
task-entry and compatibility boundaries. It improves some programs while
materially regressing other concurrency programs and all three unlike serial
guards in the directly alternating gate.

This closes the refresh as `closed-rejected-candidate`. A future candidate
must carry one stable execution identity through generated calls, dynamically
entered/native goroutines, task creation, cancellation payload lookup, nested
calls, and environment/call-frame restoration without allocating or falling
back to stack parsing at hot boundaries. A Future-, Channel-, Mutex-,
application-, named-type-, benchmark-, or WASM-specific bypass is not eligible.

## Reproducibility contract

The current and opt-in binaries were built once per application, then frozen.
The current artifact set recorded source, generated-tree, binary,
execution-contract, verifier, and output hashes in its comparison JSON. Its
SHA-256 was
`76340725ccdea1afd5cfde6551187113a1efc7ed90a533d0984a6e66a0ca4094`;
the opt-in comparison JSON was
`a4535d3c735861e46d1116134bc6a504cbe6c6e3be716b6d34d9df402c3cfe87`.
Scratch artifacts were removed after the report and hashes were captured. Each
of the fourteen applications completed four verified current processes in two
forward/reverse cohorts. The eleven concurrency rows used four logical CPUs
(`0-3`) and the goroutine executor; NBody, Fib, and JSON used one CPU. Every
process used `GOGC=50`, `GOMEMLIMIT=1GiB`, and a 59-second ceiling.

Two additional verifier-backed main-only CPU processes and two exact
measured-main allocation processes were collected for every current binary.
The opt-in ABI received the same four-process broad screen, then five directly
alternating current/candidate pairs on two clear wins, two apparent
regressions, and all three serial guards. Workstation movement is retained in
the arithmetic means rather than selecting the fastest sample.

All 56 current broad-screen processes, 56 opt-in broad-screen processes, 70
paired processes, 28 current CPU processes, 28 current allocation processes,
8 opt-in CPU processes, and 14 opt-in allocation processes passed their Ruby
verifiers. No timing process failed or timed out.

## Current ownership

The two current CPU processes per row were merged. `currentGID` is absent from
the three serial controls and owns nearly all sampled work in every substantial
concurrency application:

| Application | Merged CPU | `bridge.currentGID` cumulative |
| --- | ---: | ---: |
| Await Channel Mux | 0.82 s | 87.80% |
| Channel Rollup | 1.77 s | 94.35% |
| Concurrent Event Routing | 9.29 s | 93.97% |
| Concurrent Text Index | 3.07 s | 94.46% |
| Dependency Wave Validation | 4.14 s | 93.48% |
| Future Await Race | 0.11 s | 72.73% |
| Future Pipeline | 0.68 s | 88.24% |
| Mutex Await Journal | 1.63 s | 93.25% |
| Mutex Ledger | 2.94 s | 91.16% |
| Mutex Work Queue | 8.54 s | 93.33% |
| Validated Job Pipeline | 9.86 s | 93.61% |
| NBody | 0.25 s | absent |
| Fib | 8.26 s | absent |
| JSON | 1.37 s | absent |

Exact allocation pairs are stable enough to separate mechanism from wall
noise. NBody allocates 1,096 bytes/41 objects in measured main, Fib 144
bytes/6 objects, and JSON 344,115,736 bytes/69 objects in both runs. The
concurrency means range from 1.12 MiB/17,432 objects in the short Future Await
Race to 77.81 MiB/1,502,343 objects in Event Routing and 60.49
MiB/1,669,406 objects in Validated Job. There is no second exact allocation
owner common to the serial controls.

## Existing ABI falsification

The opt-in `--experimental-execution-context` path is the only current general
candidate. It threads a fixed context pointer through statically generated
calls and child tasks while retaining compatibility entry points for dynamic
and native boundaries. The initial four-process screen was too volatile for a
decision: candidate cohort spreads reached 74% and two concurrency rows moved
against the candidate.

Five directly alternating, verifier-backed pairs resolved the sign:

| Application | Current samples (s) | Candidate samples (s) | Current mean | Candidate mean | Change |
| --- | --- | --- | ---: | ---: | ---: |
| Channel Rollup | 0.79, 0.50, 0.63, 0.51, 0.55 | 0.44, 0.33, 0.38, 0.33, 0.31 | 0.596 | 0.358 | -39.93% |
| Concurrent Event Routing | 3.24, 2.63, 2.94, 2.50, 2.52 | 3.54, 2.95, 3.05, 2.91, 2.86 | 2.766 | 3.062 | **+10.70%** |
| Future Pipeline | 0.38, 0.33, 0.33, 0.33, 0.31 | 0.22, 0.14, 0.14, 0.15, 0.13 | 0.336 | 0.156 | -53.57% |
| Mutex Ledger | 0.85, 0.72, 0.69, 0.65, 0.63 | 1.34, 0.75, 0.82, 0.84, 0.74 | 0.708 | 0.898 | **+26.84%** |
| NBody | 0.22, 0.14, 0.16, 0.14, 0.14 | 0.29, 0.14, 0.15, 0.13, 0.14 | 0.160 | 0.170 | **+6.25%** |
| Fib | 3.95, 3.14, 3.43, 3.27, 3.12 | 4.21, 3.26, 3.17, 3.64, 3.07 | 3.382 | 3.470 | **+2.60%** |
| JSON | 0.72, 0.69, 0.68, 0.64, 0.64 | 0.80, 0.71, 0.68, 0.66, 0.66 | 0.674 | 0.702 | **+4.15%** |

The allocation direction matches the wall result. The candidate reduces
Channel Rollup bytes/objects 20.31%/9.93% and Future Pipeline
46.90%/37.92%, but increases Event Routing 1.42%/3.35% and Mutex Ledger
3.61%/14.44%. It adds one measured-main allocation to both NBody and Fib and
three to JSON.

Most importantly, candidate profiles do not demonstrate the intended general
mechanism. `currentGID` remains 92.19% cumulative in Channel Rollup, 96.11% in
Event Routing, 90.48% in Future Pipeline, and 89.27% in Mutex Ledger. The
stacks pass through `__able_spawn_context`, `__able_run_compiled_task`,
`Runtime.Env`, and compatibility registration/entry helpers. The ABI saves
some generated-call context reconstruction but still parses goroutine stack
headers at the dominant task-local lookup boundary. Promoting it would keep
the selected wall, regress unlike programs, and encode an incomplete context
contract.

## Verification

`go test ./pkg/compiler/bridge -count=1 -timeout 60s` passes in 0.041 seconds.
A focused nine-test execution-context/goroutine compiler set passes in 10.325
seconds. The corresponding broad short-mode compiler selection passes in
2.421 seconds. A broader regex that accidentally included the full experimental
fixture-parity matrix hit the package-wide 60-second cap while its fixture
subtests were still progressing; it reported no assertion failure. That broad
matrix is excluded from the verification claim because the project requires
bounded tests and no production code changed in this tranche.

The benchmark verifier population supplies the semantic gate for every binary
actually measured. The canonical external stdlib was read but did not need a
change.

## Next recommendation

Refresh the `bytecode-concurrency` group across the same eleven selected
applications, with numeric and text controls. It is the next-largest stale
runtime group at 6.425 target-excess seconds, and its existing evidence is a
mixed set of targeted profiles rather than one same-artifact, same-contract
intersection.

This entails two bounded warmed-main CPU profiles and two exact allocation
processes per application from one frozen bytecode test binary, followed by
exact-symbol/caller intersection across scheduler, task, atomic, member,
call/return, and type-match descendants. A candidate advances only if one new
concrete VM-owned child is material in at least three unlike concurrency
families and remains absent or harmless in the controls. This is the best next
step because the compiler's largest remaining group is causally closed again,
while concurrency is still a meaningful bytecode product gap and can expose a
general scheduler/VM boundary without requiring benchmark-, facility-, named
stdlib-, or WASM-specific behavior.
