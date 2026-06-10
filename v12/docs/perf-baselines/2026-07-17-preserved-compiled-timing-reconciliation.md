# Preserved compiled timing reconciliation

Date: 2026-07-17

## Decision

Land the build-first preserved-binary comparison lane and retain no compiler,
generated-runtime, bytecode VM, canonical-stdlib, benchmark, fixture, or
language optimization from this tranche. The new lane shows that merely
preserving a compiled executable does not reproduce the anomalously fast
ad-hoc twenty-run measurements from the preceding profile gate.

The lane is selection evidence, not a replacement for the existing
build-and-run integration comparison. It builds every selected application
before timing starts, reuses each exact executable, reverses application order
in the second cohort, verifies every successful process, records individual
samples plus source/generated/binary fingerprints, and rejects a row when the
independent cohort means differ by more than 15% by default.

## Implementation

- `bench_perf --compiled-build-only` builds generated Go and the application
  executable without entering a timed run.
- `bench_perf --compiled-binary PATH` explicitly reuses an executable and
  skips both compiler-CLI and application rebuilding.
- `bench_compare_preserved_compiled` uses the external benchmark catalog for
  program arguments, working directories, verifier paths, CPU budgets,
  affinity, executor policy, and source-root policy. It completes the build
  phase for the whole selection before launching two forward/reverse cohorts.
- `bench_preserved_compiled_report.py` fingerprints the Able entry source,
  generated Go tree, executable, verifier, and declared input files. Its JSON
  retains every timed sample, cohort order, matched Go contract, pooled mean,
  cohort spread, and rejection reason.
- `just bench-preserved-compiled -- ...` exposes the lane. The focused report
  tests are part of `just bench-selection-check`.

The default remains five independent processes per cohort. The 15% gate uses
`(largest cohort mean - smallest cohort mean) / smallest cohort mean`; it does
not hide disagreement by pooling first. A rejected row can still show its
pooled descriptive mean, but it cannot support a performance-change decision.

## Four-application reconciliation

Fresh Go references and preserved Able binaries used catalog CPU 0 from pool
0-15, `GOMAXPROCS=1`, a 55-second process timeout, canonical external
`able-stdlib`, and public Ruby verifiers. All 20 Go processes and all 40 Able
processes completed and verified. Each Able row has two independently ordered
five-process cohorts.

| Application | Able pooled mean | Cohort means | Spread | Fresh Go mean | Able / Go | Stable evidence |
| --- | ---: | ---: | ---: | ---: | ---: | :---: |
| Reverse Complement | 0.1060 s | 0.1120 / 0.1000 s | 12.00% | 0.0160 s | 6.63x | yes |
| Rational Series | 0.1250 s | 0.1360 / 0.1140 s | 19.30% | 0.0130 s | 9.62x | no |
| Word Frequency | 0.2150 s | 0.2220 / 0.2080 s | 6.73% | 0.0053 s | 40.57x | yes |
| Array Slice Window | 0.0840 s | 0.0820 / 0.0860 s | 4.88% | 0.0043 s | 19.53x | yes |

Rational Series is deliberately rejected rather than promoted from its pooled
ten-run mean. The other three rows pass the cross-cohort variance rule.

The stable preserved results are close to the earlier build-and-run cohort for
Reverse Complement (6.67x), Word Frequency (36.84x), and, within the scale of
short process rows, Rational Series (11.67x). Array Slice Window improves from
29.51x to 19.53x but does not approach the prior ad-hoc 11.15x result. Across
all four, the new controlled lane does not reproduce the previous ad-hoc
3.73x/6.60x/26.25x/11.15x ratios. The executable lifecycle was therefore not
the explanation; those faster measurements were a non-reproducible workstation
cohort and must not drive compiler selection.

Machine-readable evidence:

- `2026-07-17-preserved-compiled-four-app-go-reference.json`
- `2026-07-17-preserved-compiled-four-app-cohorts.json`

## Verification and cleanup

- Shell syntax passes for both benchmark runners.
- `python -m unittest bench_preserved_compiled_report_test.py` passes the
  stable-cohort and material-disagreement cases.
- A one-application end-to-end smoke built Array Slice Window once and reused
  the exact executable in both verified cohorts.
- The four-application acceptance run completed 40/40 verified Able processes;
  its matched fresh reference refresh completed 20/20 verified Go processes.
- Temporary build trees and executables from the acceptance run were removed
  after the fingerprinted aggregate report was written.

## Next recommendation

Use the preserved lane for a focused compiled text/map family gate across Word
Frequency, Document Audit, and Dependency Plan, then admit a candidate only if
the same exact generic generated/runtime descendant is material in all three.

Why: Word Frequency remains the largest stable compiled gap in this tranche at
40.57x Go, and prior profiles put most of its time under nominal map lookup and
String conversion. The two other applications exercise related operations in
different program structures, so they can distinguish a generally useful
nominal-dispatch/String representation cost from a Word-Frequency-specific
shortcut. This respects the rule against named-container compiler lowering.

What it entails: refresh fresh Go references, collect two five-run preserved
cohorts for the three applications, reject unstable rows, and then take bounded
main-only CPU plus exact allocation profiles from the same fingerprinted
binaries. Merge only concrete descendants shared across all three. If the
shared owner is canonical String/byte conversion, test a general stdlib/runtime
change in `../able-stdlib`; if it is nominal call or value encoding, change the
general compiler pipeline. Retain a candidate only after repeated A/B cohorts
show a broad win without regression on unlike numeric/array controls.
