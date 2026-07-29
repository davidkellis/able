# Compiled Go 1.26.5 scorecard refresh

Date: 2026-07-29

## Decision

Retain no compiler, generated-runtime, runtime, interpreter, VM, stdlib,
language, benchmark, fixture, dependency, nominal-special-case, or WASM
production change.

The complete strict compiled catalog was refreshed against Go 1.26.5, but the
new frontier still admits no general owner across three unlike applications.
The only target-status change is `i_before_e/compiled`, whose independent
stability cohort confirms that it is an established target guard rather than
an optimization lead.

## Measurement contract

- Measured all 63 selected compiled Able applications and all 63 matching Go
  references.
- Ran five verifier-backed processes per Able row and five per Go row: 315
  successful Able processes plus 315 successful Go processes, with zero
  timeouts and zero failures.
- Forced `GOTOOLCHAIN=go1.26.5` for both reference builds and generated Able
  builds. Every retained Go reference report records
  `go version go1.26.5 linux/amd64`.
- Passed `--no-fallbacks` through every compiled comparison. The compiler and
  generated-runtime source identities are unchanged, so the existing
  all-63-graph interpreter-dependency audit remains governing.
- Used the existing public verifier, input, argument, serial-execution, and
  single-logical-CPU contracts. Large build work stayed under the exact
  disk-backed
  `/var/tmp/able-v12-compiled-go1265-scorecard-20260729` workspace.
- Temporarily paused only the three exact Marketlab process groups that were
  repeatedly restarting during the cohort, installed a trap that resumed all
  three groups, and verified that none remained stopped after measurement.

An earlier partial attempt was discarded rather than promoted. It used Go
1.26.4 for root-level reference builds while module-local generated builds
selected Go 1.26.5, and recurring Marketlab work contaminated its first Able
rows. Its reports and build artifacts were deleted before the retained cohort
was started.

## Results

| Measure | Previous compiled frontier | Go 1.26.5 refresh |
| --- | ---: | ---: |
| target passes | 6 / 63 | 7 / 63 |
| target misses | 57 / 63 | 56 / 63 |
| geometric-mean Able/Go ratio | 5.575597x | 4.637116x |
| positive target excess | 5.675368 s | 5.353263 s |

Representative retained means:

| Application | Able mean | Go mean | Able/Go | Status |
| --- | ---: | ---: | ---: | --- |
| Binary Trees | 8.9780 s | 9.2644 s | 0.9691x | meets |
| QuickSort | 1.6260 s | 2.3153 s | 0.7023x | meets |
| Base64 | 1.9980 s | 2.2859 s | 0.8741x | meets |
| JSON | 0.6200 s | 1.3121 s | 0.4725x | meets |
| Monte Carlo Pi | 0.1360 s | 0.1876 s | 0.7250x | meets |
| Pi Digits | 1.0700 s | 1.1088 s | 0.9650x | meets |
| Fibonacci | 3.1180 s | 2.8996 s | 1.0753x | misses |
| Matrix Multiply | 0.9680 s | 0.9040 s | 1.0708x | misses |
| Sudoku Masks | 1.4940 s | 0.5939 s | 2.5156x | misses |

`i_before_e` measured 0.0620 s versus 0.0591 s in the full cohort
(1.049069x). A second matched five-process cohort measured 0.0520 s versus
0.0567 s (0.917108x). The pooled ten-process ratio is 0.984465x, so both
independent cohorts meet the 1.052632x target limit.

The largest remaining compiled target excesses are K-Nucleotide
(1.292947 s), Sudoku Masks (0.868842 s), Mutex Work Queue (0.480632 s),
TapeLang Alphabet (0.397263 s), and Mutex Await Journal (0.194000 s). They do
not expose one open compiler/runtime owner across three unlike programs.

## Frontier and verification

- Promoted
  `2026-07-29-compiled-go1265-refresh.{json,md}` to the authoritative external
  scorecard and regenerated the dated and fixed-name frontiers.
- The combined frontier contains 126 selected, fully sampled rows: 11
  established target guards, 115 misses, 226.856947 seconds of positive
  target excess, and zero actionable groups.
- Advanced only the `compiled-text-map` closure because the stable
  `i_before_e` target crossing changed that group's row semantics. The
  23-entry closure ledger remains current with zero invalidations.
- Regenerated the downstream architecture, native-tier, portable-backend, and
  shared-runtime checked artifacts. All 30 `bench_*_test.py` files pass with a
  60-second per-test ceiling.
- The external scorecard, frontier, closure ledger, scorecard-evidence check,
  and every affected checked generator reproduce exactly.

## Next

Harden the benchmark toolchain provenance contract before taking another
timing cohort.

Why: the discarded attempt demonstrated that the module `toolchain` directive
does not control root-level external Go reference builds. A report name can
therefore claim one patch release while the two sides silently use different
Go versions.

What it entails: add an explicit expected Go toolchain/version option to the
full-refresh driver and wrappers, propagate it to both reference and generated
Able build environments, record it in retained reports, reject mixed or
mislabeled implementations, and add fast shell/Python contract tests.

Why it matters: Able's 95%-of-Go goal is meaningful only when both sides use
the same recorded toolchain. This guard prevents noise or provenance drift
from manufacturing a false optimization opportunity or a false target pass.
