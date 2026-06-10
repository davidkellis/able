# Error-reconciliation scorecard eligibility

## Decision

Do not promote a new current scorecard and do not claim strict two-cohort
variance. Two bounded attempts proved that the former collection protocol
could not produce an eligible cohort:

- the 45-second default timed out valid Python/Ruby reference programs;
- after raising the per-process cap to 90 seconds, bytecode BinaryTrees timed
  out in all five processes;
- the historical scan-based Able Sudoku program did not finish even one
  compiled process under a separate 300-second calibration.

Keep no compiler, VM, or canonical-stdlib performance change from this
tranche. The completed measurements reproduce known, divergent application
costs rather than one new shared concrete descendant.

## Candidate-selection contract repair

The active catalog had treated two different Sudoku lanes as candidate
evidence:

- `sudoku` is the historical scan/backtracking Able source, while its Go and
  Python references use an exact-cover solver over the full corpus;
- `sudoku_masks` uses the same mask/MRV algorithm and fixed 100-solve workload
  in Able, Go, Python, and Ruby.

The first lane is useful historical diagnostic coverage, but it is neither an
algorithmically comparable implementation nor a bounded selection row. It is
now available explicitly as `legacy-sudoku` and is absent from `core`,
`full`, `generality`, and `coverage`. The fair `sudoku_masks` lane remains in
all active selection suites, so Array, integer, recursion, filesystem, and
mutation coverage is preserved. The portable selection catalog now contains
32 applications.

The full refresher's default per-process cap is now 90 seconds. This is not an
unbounded relaxation: fresh five-run reference means include Python
TapeLang at 50.91 seconds and Ruby TapeLang at 65.74 seconds, so the old
45-second default rejected valid controls by construction.

## Retained current-source evidence

The partial tagged run used five independent processes per row with
`GOMEMLIMIT=1GiB`, `GOGC=50`, and `GOMAXPROCS=1`. Its canonical stdlib state
is 69 source files at tree hash
`785a6fd058c179379b1a153529fb340151a11b96d9014394cc40dbd87e1882ab`,
Git head `219eff222c28406487231713753641bc49ee5b9a`, with the working tree visibly
dirty.

All fresh Go, Python, and Ruby references completed 5/5. All 15 compiled
generality applications also completed and verified 5/5. Selected compiler
ratios against fresh Go were:

| Application | Able | Able/Go | Reading |
| --- | ---: | ---: | --- |
| Fib | 3.272 s | 1.15x | close, but below the 95% target |
| MatrixMultiply | 1.048 s | 1.17x | close, but below target |
| QuickSort | 1.676 s | 0.73x | faster than Go in this cohort |
| JSON | 0.670 s | 0.52x | faster than Go |
| BinaryTrees | 27.066 s | 5.50x | recursive struct allocation miss |
| Sudoku Masks | 7.894 s | 15.27x | source temporary-array/allocation miss |
| K-Nucleotide | 3.206 s | 64.12x | map/text conversion miss |
| NBody | 0.386 s | 12.78x | floating-point generated-code miss |

The first bytecode group then produced:

| Application | Able | Able/Python | Able/Ruby | Status |
| --- | ---: | ---: | ---: | --- |
| Fib | 0.142 s | 0.0026x | 0.0034x | verified 5/5 |
| MatrixMultiply | 4.392 s | 0.0926x | 0.0940x | verified 5/5 |
| BinaryTrees | n/a | n/a | n/a | timeout 5/5 at 90 s |

Thus the VM already greatly exceeds the Python/Ruby target on scalar
recursion and dense numeric arrays, while recursive struct allocation remains
cap-bound. The matching historical 120-second timeout profiles attribute the
four old cap-bound applications to different concrete leaves: named-struct
allocation for BinaryTrees, i32 array boxing for QuickSort, raw-float
materialization for NBody, and evaluation-stack growth for TapeLang. Their
common dispatcher/allocation frames are parents, not a justified shared
candidate.

The retained source reports use the
`2026-07-15-error-reconcile-a-*` prefix. There is deliberately no aggregate
`*-refresh.json`: collection stopped as soon as strict eligibility became
impossible, and `external-scoreboard-current.*` is unchanged.

## Next recommendation

Reconcile full-status reporting with candidate-selection variance before
launching another all-application cohort. The full scorecard must continue to
retain timeout rows because they are important performance failures, while a
strict variance report needs an explicit, identical manifest of rows that can
produce five verifier-backed samples. This should be a two-tier protocol, not
an option that silently drops whatever timed out in a particular cohort.

The work should entail:

1. add an explicit mode-aware selection manifest to the refresh/variance
   tools;
2. keep every application/mode in the full status output, including timeout
   counts and source fingerprints;
3. require two cohorts to use the same reviewed selection manifest and retain
   exactly five successful samples for every selected row;
4. keep cap-bound rows eligible for focused profile work and automatically
   re-admit them when a bounded smoke run completes;
5. run the two independent cohorts only after this contract has fast tests.

Why: the current strict rule makes performance failure prevent all variance
analysis, while simply ignoring timeouts would hide the failures. An explicit
manifest preserves both honesty and usable repeated evidence. Only after that
reconciliation should a compiler/VM candidate be attempted, and only for a
concrete descendant repeated in at least three unlike selected applications.
