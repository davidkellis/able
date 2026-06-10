# Compiled N-body and I-Before-E Profile Refresh — 2026-07-13

## Decision

Keep no compiler, bridge, runtime, bytecode VM, canonical-stdlib, or benchmark
source change. Fresh verifier-backed compiled measurements confirm that N-body
is still a material Go-relative miss, while I-Before-E is a smaller one. Their
generated-main profiles overlap at environment swapping, but that boundary is
only 8.8% and 13.2% cumulative respectively and is already covered by the
rejected fixed-context ABI experiments. Their dominant descendants are
different: generated numeric calls in N-body and file/text allocation work in
I-Before-E.

Do not reopen the fixed-context ABI or add a math, file, String, benchmark, or
nominal-container-specific lowering rule from this evidence.

## Method

Fresh Go 1.26 references and current Able binaries used CPU 15,
`GOMEMLIMIT=1GiB`, `GOGC=50`, `GOMAXPROCS=1`, the canonical external
`able-stdlib`, and a 45-second process cap. Every completed process passed its
benchmark suite Ruby verifier and produced a stable output hash.

| Benchmark | Fresh Go mean | Current Able mean | Able / Go | Able runs |
| --- | ---: | ---: | ---: | ---: |
| N-body | 0.0331 s | 0.4367 s | 13.19x | 3/3 verified |
| I-Before-E | 0.0595 s | 0.1000 s | 1.68x | 3/3 verified |

Each binary was built once, then phase-profiled through the generated
launcher’s CPU-only hook (`ABLE_GO_PHASE_CPU_PROFILE_DIR`). This avoids the
allocation-snapshot collector and excludes bootstrap from attribution. The
separately verified `main.cpu.pprof` files were merged across 12 N-body and 80
I-Before-E launches.

| Benchmark | Verified phase launches | Merged main samples |
| --- | ---: | ---: |
| N-body | 12/12 | 3.97 s |
| I-Before-E | 80/80 | 2.72 s |

Retained merged profiles:

- `v12/interpreters/go/.profiles/20260713_compiled_miss_refresh_nbody_main_merged.cpu.pprof`
- `v12/interpreters/go/.profiles/20260713_compiled_miss_refresh_i_before_e_main_merged.cpu.pprof`

The rebuilt binaries, per-launch profiles, fresh reference reports, captured
outputs, and timing files remain under
`v12/tmp/perf/2026-07-13-compiled-nbody-i-before-e-profile-refresh/`; they are
cleanup-eligible.

## Attribution

| Benchmark | Material generated-main evidence | Interpretation |
| --- | --- | --- |
| N-body | generated `sqrt`: 2.59 s cumulative (65.2%); inline `abs`: 0.81 s (20.4%); `SwapEnvIfNeeded`: 0.35 s (8.8%) | Primitive numeric/package-call wall |
| I-Before-E | `is_valid`: 0.80 s cumulative (29.4%); `read_lines`: 0.74 s (27.2%); `String.contains`: 0.64 s (23.5%); `SwapEnvIfNeeded`: 0.36 s (13.2%) | File loading, String validation/search, and allocation/GC wall |

Both profiles pass through `bridge.SwapEnvIfNeeded` and restoration. That is a
real shared generated-call boundary, but it is not the dominating cost in
either workload. Its apparent `sync/atomic.StorePointer` overlap is not an
independent target: N-body reaches it entirely through bridge swaps, while
I-Before-E also reaches it through the GC write barrier caused by text/file
allocation.

The fixed-pointer execution-context ABI already targeted this family of
generated-call context work. Its broad default gate produced a verified 54.7%
N-body regression; the later allocation-free package-linkage refinement then
produced a repeated 16.6% K-Nucleotide regression. Retrying the same boundary
without a new semantic design and cross-family evidence would repeat a known
failed experiment. Even a perfect removal of N-body’s measured swap share
cannot close its 13.19x gap while generated `sqrt` remains 65.2% of the main
profile.

## Next Recommendation

Refresh the verifier-backed compiled generality scorecard with fresh Go
references, then profile only two or more unlike remaining misses that repeat
a concrete compiler/runtime leaf not already rejected by a broad gate. The
current pair resolves the proposed N-body/I-Before-E selection: its only
overlap is too small and already disqualified, while the large costs diverge.
The next tranche entails rebuilding the reference programs outside timing,
running current compiled Able binaries with canonical verifiers on the pinned
lane, and retaining CPU-only generated-main profiles only for the selected
cross-family misses.
