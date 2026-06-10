# Current iterator/control cross-mode profile refresh

Date: 2026-07-20

## Decision

Complete the current Document Audit, Dependency Plan, Lexical Rollup, and
Option/Result Configuration profile gate and retain no VM, compiler,
generated-runtime, canonical-stdlib, benchmark, fixture, language, or WASM
change.

Fresh generated-main profiles still separate into text/iterator, graph/Queue,
lexical iterator/channel, and generic-union/allocation work. The warmed
bytecode profiles do share one exact material operation,
`lookupCachedMemberMethodEntry`, but this is the already-completed dependency-
validated member-cache family. Its remaining policy has received a path census
and repeated broad A/B gates, and the current profiles expose no different
shared child or new invalidation contract that would justify reopening it.

## Current compiled-binary contract

The current compiler built each executable once with canonical
`../able-stdlib`. CPU-only generated-phase profiling used `GOMEMLIMIT=1GiB`,
`GOGC=50`, and `GOMAXPROCS=1`. Every launch was a separate process from the
catalog working directory with the catalog arguments, and every output passed
its public Ruby verifier. Only `main.cpu.pprof` files were merged.

| Application | Verified profiles | Main samples | Material exact ownership |
| --- | ---: | ---: | --- |
| Document Audit | 100/100 | 60 ms | UTF-8 validation 50.00% flat; substring search 33.33% flat; small generator/filter allocation residual |
| Dependency Plan | 100/100 | 110 ms total, 50 ms application | deployment resolver plus Queue/graph and common integer boxes; the remainder is CPU-profiler/GC startup |
| Lexical Rollup | 100/100 | 1.47 s | iterator next 16.33% cumulative; channel receive 8.84%; environment swap 4.08%; UTF-8 validation 2.04% |
| Option/Result Configuration | 50/50 | 6.71 s | static generic-union method calls 55.59% cumulative; call-value dispatch 42.03%; allocation 58.12% |

The two short profiles remain coarse, but their stacks agree with their prior
owners. No compiler-controlled exact leaf is material in three unlike
applications. Generic Go allocation and collector frames are consequences of
different Able operations and are not a lowering boundary.

Five further independent launches of each preserved binary supplied ordinary
workstation averages. All samples, including the Document Audit outlier, are
retained in the arithmetic means.

| Application | Samples (s) | Mean | Verification |
| --- | --- | ---: | ---: |
| Document Audit | 0.14, 0.09, 0.09, 0.09, 0.08 | 0.098 s | 5/5 |
| Dependency Plan | 0.09, 0.09, 0.09, 0.08, 0.09 | 0.088 s | 5/5 |
| Lexical Rollup | 0.10, 0.10, 0.09, 0.10, 0.10 | 0.098 s | 5/5 |
| Option/Result Configuration | 0.21, 0.25, 0.22, 0.21, 0.21 | 0.220 s | 5/5 |

## Current warmed-bytecode contract

One current interpreter test binary loaded and typechecked each program once,
warmed `main` once, and then measured repeated calls for at least five seconds.
Each process used the same memory, GC, and one-thread limits as the compiled
profiles. Program output is suppressed by this harness; the separate normal
process cohort below is verifier-backed.

| Application | Calls | ns/op | B/op | allocs/op | CPU samples |
| --- | ---: | ---: | ---: | ---: | ---: |
| Document Audit | 501 | 11,641,658 | 370,089 | 520 | 5.82 s |
| Dependency Plan | 22 | 291,172,863 | 1,987,705 | 28,815 | 6.39 s |
| Lexical Rollup | 49 | 138,593,248 | 2,200,968 | 14,701 | 6.76 s |
| Option/Result Configuration | 7 | 922,401,246 | 76,341,986 | 1,304,834 | 6.42 s |

The exact cached-member lookup is 32.65%, 12.05%, 20.86%, and 5.14%
cumulative respectively. It clears the breadth threshold, but not the novelty
or invalidation gate:

- the retained cross-transient-scope cache already validates the environment
  family, impl context, relevant name revisions, receiver identity, and exact
  lexical owners;
- the completed census found dependency validation repeated 11,962 times in
  Document Audit and 91,429 times in Lexical Rollup, disproving a cold-site
  admission hypothesis; and
- the safe same-parent shortcut changed Document Audit by +7.03%, Lexical
  Rollup by +1.23%, iterator collect by +4.42%, and split/join by +2.19% in
  alternating means, so it was removed.

The current descendants also differ. Document Audit spends 14.95% in complete
lexical-state discovery; Dependency Plan has map-64, Array, and return work;
Lexical Rollup has 11.69% lexical-state discovery plus static/member iterator
dispatch; Option/Result is led by GC, type matching, and typed-pattern return
work, with cached member lookup only 5.14%. No new exact child is material in
three applications. Broad dispatcher, Go map, and GC parents do not qualify.

## Normal bytecode controls

Five independent complete bytecode processes per application passed their
public verifiers and produced one stable output hash per application.

| Application | Samples (s) | Mean | Verification |
| --- | --- | ---: | ---: |
| Document Audit | 0.32, 0.35, 0.36, 0.27, 0.31 | 0.322 s | 5/5 |
| Dependency Plan | 0.48, 0.48, 0.47, 0.51, 0.52 | 0.492 s | 5/5 |
| Lexical Rollup | 0.43, 0.41, 0.45, 0.45, 0.42 | 0.432 s | 5/5 |
| Option/Result Configuration | 0.87, 0.86, 0.93, 0.88, 0.95 | 0.898 s | 5/5 |

## Verification and cleanup

- 350/350 compiled CPU-only phase launches verified;
- 20/20 additional compiled timing launches verified;
- 4/4 warmed bytecode benchmark/profile processes passed;
- 20/20 normal bytecode timing launches verified;
- focused member-cache/shadowing, profile-hook, and fixture-parity controls;
- `git diff --check`; and
- all raw profiles, binaries, generated trees, and captures remain cleanup-only
  under `/tmp` until the end-of-tranche cleanup.

## Next recommendation

Follow-up completed by
`2026-07-20-current-product-scorecard-and-frontier-refresh.md`. The promoted
scorecard now has five verified Able/reference samples for all 75 selected
rows, and the regenerated frontier has 8 meets, 67 misses, 121.969 seconds of
aggregate excess, and zero unclosed groups.

The completed recommendation was to regenerate the current cross-mode external
scorecard and performance frontier with five verifier-backed workstation runs
per Able row and fresh applicable Go, Python, and Ruby references.

Why: this tranche closes the last no-shared-leaf group whose compiled profiles
were explicitly pre-current, and it demonstrates that several ledger timings
are stale: the current Document, Dependency, Lexical, and Option means differ
from the frontier inputs in both modes. With every local profile group now
closed or guarded by a failed candidate, the next useful evidence is an
accurate product-level distance-to-target ranking, not another subdivision of
an exhausted cache/return/map parent.

What it entails: run the portable coverage applications in bounded independent
processes, keep every non-timeout sample in arithmetic means, verify every
output and fingerprint each source/input/verifier, refresh external references
under matching contracts, and regenerate the selected-row frontier. Rank by
both ratio and absolute seconds beyond the 95% budget. Reopen implementation
work only when changed current evidence invalidates a specific closure or an
exact generic descendant appears in at least three unlike applications. Do
not add named-container or benchmark special cases, and do not begin WASM work.
