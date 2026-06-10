# Truthiness/cast near-path closure refresh

Date: 2026-07-21

## Decision

The `compiled-iterator-control` and `bytecode-float-numeric` closures are
current against the post-truthiness/cast shared interpreter semantics. Keep no
compiler, VM, runtime, canonical-stdlib, benchmark, or reference change from
this tranche.

The refreshed timing retains every verifier-backed sample and confirms large
remaining target gaps. Exact reach evidence rules out the changed Error
truthiness fallback in all nine applications and rules out explicit casts in
all four bytecode applications. `option_result_config` reaches the compiled
explicit-cast bridge 24,384 times, but two current main CPU profiles sample
none of that path. Its measurable owners remain the previously separated
generic-union dispatch, allocation, lookup/call, and boundary-materialization
families. No new three-application shared leaf exists, so no candidate is
admitted.

## Frozen contract

- The v12 spec SHA-256 remains
  `4f0405b86c122993723e8617abd6f825d9a8ff858d4c72acaf4e33469452f080`.
- The canonical external stdlib source-tree SHA-256 remains
  `43ff2e68e59c8be7fb1024c86a1f61a0eea84596279b4f0e146511d66c5308d8`.
- Executables were built before timing. Every timed process passed its public
  verifier. Each process used the catalog run directory, arguments, CPU
  budget, and a 55-second cap.
- Arithmetic means retain all successful samples without outlier removal.
  Volatile rows received independent additional cohorts: Lexical Rollup has
  three five-process cohorts; Monte Carlo Pi and RMS Norm have two. Volatile
  Monte Carlo and Distance Field reference lanes also received second cohorts.
- Able executions use the runner's guarded `GOMEMLIMIT=1GiB`, `GOGC=50`, and
  catalog-resolved serial CPU policy.

## Repeated timing

| Application | Mode | Able samples | Able mean | Limiting reference | Reference samples | Reference mean | Ratio |
| --- | --- | ---: | ---: | --- | ---: | ---: | ---: |
| Array Slice Window | compiled | 5 | 0.086 s | Go | 5 | 0.004385 s | 19.614x |
| Dependency Plan | compiled | 5 | 0.072 s | Go | 5 | 0.004750 s | 15.158x |
| Document Audit | compiled | 5 | 0.096 s | Go | 5 | 0.004755 s | 20.188x |
| Lexical Rollup | compiled | 15 | 0.118 s | Go | 15 | 0.005873 s | 20.092x |
| Option/Result Configuration | compiled | 5 | 0.224 s | Go | 5 | 0.004691 s | 47.749x |
| Distance Field | bytecode | 5 | 6.726 s | Ruby | 10 | 0.425693 s | 15.800x |
| Mandelbrot | bytecode | 5 | 7.200 s | Python | 5 | 1.398018 s | 5.150x |
| Monte Carlo Pi | bytecode | 10 | 3.347 s | Ruby | 10 | 2.144855 s | 1.560x |
| RMS Norm | bytecode | 10 | 6.473 s | Ruby | 5 | 0.666460 s | 9.713x |

All 155 timing processes represented by these pooled decisions verified with
zero failures and zero timeouts: 70 compiled/Go processes and 85
bytecode/Python/Ruby processes. The extra cohorts are averaged with the first
cohorts rather than replacing them. Remaining workstation variance is reported
honestly: pooled Lexical Rollup Able CV is 22.47%, Monte Carlo Pi is 19.53%,
and RMS Norm is 28.60%.

## Exact reach and profile gate

Temporary opt-in counters were placed immediately at the changed semantic
boundaries, used only in untimed diagnostic processes, and then removed.

| Application | Mode | Census processes | Truthy checks/process | Changed Error fallback | Explicit casts/process | Cast failures |
| --- | --- | ---: | ---: | ---: | ---: | ---: |
| Array Slice Window | compiled | 1 | 0 | 0 | 0 | n/a |
| Dependency Plan | compiled | 1 | 0 | 0 | 0 | n/a |
| Document Audit | compiled | 1 | 0 | 0 | 0 | n/a |
| Lexical Rollup | compiled | 1 | 0 | 0 | 0 | n/a |
| Option/Result Configuration | compiled | 1 | 24,576 | 0 | 24,384 | n/a |
| Distance Field | bytecode | 2 | 0 | 0 | 0 | 0 |
| Mandelbrot | bytecode | 2 | 1,280,000 | 0 | 0 | 0 |
| Monte Carlo Pi | bytecode | 2 | 0 | 0 | 0 | 0 |
| RMS Norm | bytecode | 2 | 0 | 0 | 0 | 0 |

Every bytecode count reproduced exactly in its second process. Mandelbrot's
truth checks are primitive Boolean checks, which return before the corrected
non-primitive Error matcher. Therefore none of the four bytecode rows reaches
either changed path, and their current six-application ownership profiles
remain causally valid.

`option_result_config` alone crossed the compiled profile gate. Two verified
profiles of the uninstrumented current binary contain 100 ms and 130 ms of
main CPU samples. Across the combined 230 ms, `__able_cast`, `bridge.Cast`, and
the interpreter cast wrapper have zero flat and zero cumulative samples. One
profile is led by generic-union call dispatch, method lookup/calls, and memory
clearing; the other is led by allocation, interface unwrapping, boundary
materialization, and type matching. The reached cast wrapper is below current
sampling resolution and is not a candidate, while those concrete descendants
do not repeat in three unlike applications.

The concise combined census is retained in
`2026-07-21-truthiness-cast-near-path-closure-reach.json`; the raw compiled
telemetry report is retained in
`2026-07-21-truthiness-cast-near-path-closure-compiled-reach.json`. No
diagnostic counter, generated source, binary, or profile remains in production
code.

## Exact timing artifacts

- Initial compiled Able and Go references:
  `2026-07-21-truthiness-cast-near-path-closures-compiled.json`
  (`8a00a20ed60e7be446164a214f7974d95033dd9bbfb951f6e37e8fde891f1807`),
  `2026-07-21-truthiness-cast-near-path-closures-go-reference.json`
  (`b5583d1bc7e3aa2e48705a4bd2e79f515c84364a765fa8abdb56ab8dbe124efd`).
- Lexical cohorts two and three, Able/Go:
  `c4c466f5c0f43fd8b565d62283de8383b426256a048680a70dee8e4b0bc76b5b`,
  `c480d2c39b8c481be32e1a1f92f62e0311fddfda4f6f035423763fd1cbdb6b7a`,
  `9f2f4cd276cbcbdbdca443f204fc0d3c7d38f53d5cbcbeca9ee6c1ef4169db13`,
  and `cd414e9884efdaf9b0a79d30b8cda7c14d5076f9ddf39d24f65cbe4d55cfc26c`.
- Initial bytecode Able and interpreter references:
  `2026-07-21-truthiness-cast-near-path-closures-bytecode.json`
  (`3e27208085176756b2ca3023238bbcbad9be3b98685b6f147f0335246001ee99`),
  `2026-07-21-truthiness-cast-near-path-closures-interpreter-reference.json`
  (`2102cd653daa5a19a53746bc29297dc22496a1964f32974c66ad3e90aa47a013`).
- Additional volatile-row artifacts:
  `2026-07-21-truthiness-cast-near-path-closures-float-c2-bytecode.json`
  (`27b7d59147c3990978a5139e6cf45968227a47ec26bf298b9ab1ce85b5cd628a`),
  Monte Carlo references
  (`e7260bcdc8d3ba43fbce1db21df982238928d83eb74e8d5de98c7810b6c455e4`),
  and Distance Field Ruby references
  (`47bc2184e6d3d2992985b91186a88d30e641ee2477191bad746d8ffbb7677a46`).

## Next recommendation

Refresh `compiled-float-numeric` and `bytecode-wide-numeric` next.

Why: they are the next nearest invalidated numeric closures. The compiled
float family can exercise the same cast/type-matching boundary through unlike
numeric programs, while the bytecode wide-integer family tests whether the
catchable-cast alignment reaches raw/wide integer conversions that are not
represented by these float rows. This advances causal evidence without
reopening already closed aggregate allocation or benchmark-specific paths.

What it entails: reuse the frozen sources and current reference toolchains;
collect five verifier-backed processes per ordinary lane; add independent
cohorts where workstation variance remains high; run exact main-only
truthiness/cast reach first; and profile only rows with material changed-path
reach. Advance only those two closures. Build a candidate only if one concrete
generic leaf is material in at least three unlike applications and preserves
the current compiled and bytecode target guards.
