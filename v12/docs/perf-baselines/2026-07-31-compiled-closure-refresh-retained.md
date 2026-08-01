# Compiled closure refresh retained

Date: 2026-07-31

## Decision

Retain the post-ownership-default compiler-wide baseline and advance all eleven
compiled performance closures. Reconcile and advance the shared cross-family
architecture-ownership closure last. Retain no additional compiler, runtime,
interpreter, bytecode VM, parser, stdlib, language, benchmark, dependency, or
WASM change.

The refresh measured every selected compiled application rather than a sampled
subset. All 66 strict applications compiled with the ordinary default nominal
ownership path and `--no-fallbacks`. Every timed Able and Go process passed its
public verifier.

## Measurement contract

- Go reference toolchain: `go1.26.5`.
- CPU pool: `12-15`, resolved per benchmark through the catalog's serial or
  goroutine executor contract.
- Primary cohort: five Able and five fresh Go processes for each of 66
  applications, or 660 verifier-backed processes.
- Crossing cohort: a second independent five Able and five fresh Go processes
  for Fib, I Before E, and Matrix Multiply, or 30 more verifier-backed
  processes.
- Per-process and per-build timeout: 55 seconds.
- Canonical stdlib source tree:
  `6a412c872ee66752de7c4417a5eda99806a7631da3b880c55574c6a640b82d9b`.
- All disposable builds and emitted modules lived under disk-backed
  `/var/tmp`.

The primary report contains 330/330 successful Able runs and 330/330
successful Go runs. The crossing report contains another 15/15 Able and 15/15
Go successes. There were no verification failures or timeouts.

## Current compiled result

The complete compiled corpus now has 9 target meets and 57 misses. The
geometric mean Able/Go ratio is 4.2187x. This aggregate is strongly affected by
short applications whose Go references complete in 3.8-17.3 ms, so sustained
rows remain the more useful lowering signal.

Eight of the eleven rows whose Go reference exceeds 100 ms meet the target:

| Application | Able (s) | Go (s) | Able/Go |
| --- | ---: | ---: | ---: |
| JSON | 0.560 | 1.5717 | 0.356x |
| Monte Carlo Pi | 0.140 | 0.2537 | 0.552x |
| Binary Trees | 7.374 | 11.9591 | 0.617x |
| QuickSort | 1.782 | 2.7145 | 0.656x |
| Pi Digits | 1.200 | 1.2436 | 0.965x |
| Fib | 3.416 | 3.5367 | 0.966x |
| Base64 | 2.356 | 2.4163 | 0.975x |
| Matrix Multiply | 1.040 | 1.0344 | 1.005x |

The sustained misses are TapeLang Alphabet at 1.143x, Sudoku Masks at 2.115x,
and Versioned Telemetry Pipeline at 9.665x. K-Nucleotide is also material
despite its shorter Go reference: 1.304 seconds Able versus 0.062 seconds Go,
or 21.032x.

Fib, I Before E, and Matrix Multiply crossed into the target set relative to
the prior promoted snapshot. Their second independent cohorts also meet:

| Application | Cohort A | Cohort B | Pooled Able/Go |
| --- | ---: | ---: | ---: |
| Fib | 0.966x | 1.044x | 1.003x |
| I Before E | 0.557x | 0.739x | 0.640x |
| Matrix Multiply | 1.005x | 0.986x | 0.996x |

All three are therefore promoted as established target guards, not treated as
single-cohort crossings. The combined frontier has 13 target meets and 119
misses, with 13 established guards and no unestablished snapshot meets.

## Closure reconciliation

| Closure | Rows | Misses | Excess (s) | Maximum ratio | Decision |
| --- | ---: | ---: | ---: | ---: | --- |
| `compiled-target-guards` | 6 | 0 | 0.0000 | 0.975x | retain guard |
| `compiled-current-control` | 3 | 1 | 0.2892 | 1.143x | closed, no shared leaf |
| `compiled-sudoku-quotient` | 1 | 1 | 0.8047 | 2.115x | closed, insufficient breadth |
| `compiled-float-numeric` | 4 | 4 | 0.1163 | 3.846x | rejected candidate remains closed |
| `compiled-wide-numeric` | 3 | 3 | 0.1438 | 13.125x | rejected candidate remains closed |
| `compiled-byte-output` | 2 | 2 | 0.0437 | 2.326x | closed, no shared leaf |
| `compiled-text-map` | 9 | 8 | 1.5502 | 21.032x | closed, no shared leaf |
| `compiled-regex` | 6 | 6 | 0.2506 | 12.830x | rejected candidate remains closed |
| `compiled-concurrency` | 23 | 23 | 0.7199 | 23.830x | rejected candidate remains closed |
| `compiled-iterator-control` | 9 | 9 | 2.0323 | 11.053x | closed, no shared leaf |
| `compiled-architecture-target-budget` | 6 | 6 | — | 21.032x | rejected candidate remains closed |

The ownership-default change does not justify reopening the historical rejected
routes. Its admitted rule is already general and compiler-only; the primary
cohort shows the remaining performance surface still divides among launch
floor, text/map work, telemetry, Sudoku arithmetic, regex state work,
concurrency/runtime services, and wide-number support. No new exact owner is
demonstrated across three unlike families by these timing data.

The shared cross-family architecture closure remains
`closed-no-shared-leaf`. All twelve of its compiled rows are current in the
primary cohort. Default nominal ownership can remove proven caller-owned
compiled allocations, but the bytecode product has no corresponding compiled
ownership path and its eleven bytecode-only closures remain current. The
compiler-only improvement therefore does not establish a common compiled and
bytecode execution owner.

## Interpreter boundary verification

The retained workdir contained all 66 emitted Go modules. `go list -mod=mod
-deps .` completed for every module under the pinned toolchain:

- 66/66 dependency graphs resolved;
- zero dependency checks failed; and
- zero graphs contained `able/interpreter-go/pkg/interpreter`.

This independently confirms that every refreshed timing row is fallback-free
and does not regain performance by crossing into the interpreter.

## Evidence products

- `2026-07-31-compiled-closure-refresh-go-references.{json,md}`: five fresh,
  verified Go samples for all 66 applications.
- `2026-07-31-compiled-closure-refresh-scorecard.{json,md}`: five verified
  default compiled samples for all 66 applications and matched Go references.
- `2026-07-31-compiled-closure-refresh-variance.{json,md}`: retained
  per-process variance data.
- `2026-07-31-compiled-closure-refresh-crossings-b.{json,md}` and its Go
  reference report: independent evidence for the three new target crossings.
- `2026-07-31-compiled-closure-refresh-stdlib-source-state.json`: exact
  canonical stdlib identity.
- `external-scoreboard-current.{json,md}`: one current 66-row compiled source
  plus the unchanged 66-row bytecode cohort.
- `2026-07-20-cross-mode-performance-frontier.{json,md}`: regenerated
  132-row frontier with zero actionable groups.

## Next recommendation

Profile K-Nucleotide, Versioned Telemetry Pipeline, and one unlike
allocation-heavy nominal/interface application with the current default
compiler, then select only an exact generated-code or runtime leaf that is
material in all three.

Why: the refreshed baseline shows that native-quality compiled performance is
already achieved for most sustained primitive kernels, while the largest
remaining sustained costs cluster around text/maps, telemetry
normalization/storage, and nominal/interface-heavy work.

What it entails: capture repeated main-only CPU and exact allocation profiles,
map every material leaf back to emitted Go and lowering provenance, reject
launch-floor and already-closed routes, and admit an A/B prototype only if one
general compiler/runtime owner repeats across all three unlike applications.

Why it matters: this is the shortest evidence-backed path toward removing the
next real compiled overhead without introducing benchmark-specific,
named-container, or non-primitive nominal special cases.
