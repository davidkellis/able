# Feature and scorecard coverage reconciliation — 2026-07-15

## Decision

Do not add a new portable benchmark application. The active v12 feature matrix
already has an honest cross-language application or an intentional local-only
semantic fixture for every current language/runtime boundary. The real gap was
measurement coverage: the 32-application portable catalog was broader than the
22-application current-scorecard workflow.

The scorecard refresher now covers all 32 portable applications, and the first
full refresh completed on 2026-07-15. It does not select a VM, compiler, or
canonical-stdlib change by itself; a future profile still needs the same
concrete material leaf in three unlike verifier-backed applications.

## Feature reconciliation

| Feature family | Portable application coverage | Local-only boundary, if any |
| --- | --- | --- |
| Control flow, recursion, patterns, numeric primitives, and BigInt | Fib, Binary Trees, Matrix Multiply, Quicksort, Sudoku, Monte Carlo, Mandelbrot, N-body, PiDigits, Fixed Width, Rational Series | Reduced numeric/algorithm fixtures guard small semantic edges. |
| Text, bytes, files, codecs, JSON, maps, and regex | Base64, JSON, I-Before-E, Reverse Complement, K-Nucleotide, Word Frequency, Document Audit, Lexical Rollup, and the three regex applications | Scanner/error-path details remain fixture guards. |
| Nominal values, generic collections, public iterators, and Array slicing | Binary Trees, K-Nucleotide, TapeLang, rollups, Array Slice Window, Dependency Plan | Local collection fixtures provide bounded parity and regression controls. |
| Spawn, Channel, Future, Awaitable, Mutex, and `ensure` | Binary Trees, Channel Rollup, Future Pipeline, Future Await Race, Await Channel Mux, Mutex Ledger, Mutex Await Journal | Cancellation and scheduler-policy timing remain intentionally local. |
| Packages, static imports, arguments, and canonical stdlib loading | Every portable application | Dynamic packages and user-authored extern bodies remain local-only because they lack a fair like-for-like foreign-runtime contract. |

No spec-defined portable family is absent. Adding a synthetic dynamic-import,
host-callback, or user-extern timing loop would misrepresent those language
boundaries and violate the generality policy.

## Scorecard repair

The catalog has 32 portable applications, while the previous refresh collected
only 16 stable-generality and six async rows. The remaining ten already have
Able, Go, Python, Ruby, run-directory, and verifier lanes, as checked by
`bench_catalog_check`:

`fixed_width_128`, `rational_series`, `word_frequency`, `document_audit`,
`lexical_rollup`, `regex_suffix_audit`, `regex_set_audit`,
`regex_stream_audit`, `array_slice_window`, and `dependency_plan`.

`bench_refresh_external_scorecard` now refreshes fresh Go/Python/Ruby
references for those ten rows, measures them in four bounded groups per Able
mode, writes a dated aggregate, and promotes the exact source cohort to
`external-scoreboard-current.*`. The scoreboard tool now reads the cohort
recorded in that current artifact instead of reverting ordinary regeneration
or CI checks to a hard-coded July 14 source list.

The checked-in current scoreboard is synchronized to the completed 32-
application July 15 cohort. It has 4 of 31 rankable compiled rows at the
95%-of-Go target and 3 of 23 rankable bytecode rows at both interpreter
targets. One bytecode row completed only two of three requested samples and
is retained as `incomplete`/`unranked`; seven bytecode and one compiled rows
timed out, and missing matched references remain unranked. The complete
aggregate is `2026-07-15-full-coverage-scorecard-refresh.*`.

## Verification

- `bash -n v12/bench_refresh_external_scorecard`
- `python3 -m py_compile v12/bench_external_scoreboard`
- `v12/bench_refresh_external_scorecard --dry-run` scheduled six reference
  refreshes and 24 bounded comparison groups, whose deduplicated union exactly
  matches the 32-entry `coverage` catalog.
- `v12/bench_external_scoreboard --check` passes after promoting the July 15
  verifier-backed 32-application source cohort.
- `v12/bench_catalog_check` continues to validate all 32 portable lanes and
  77 local fixtures.

## Next recommendation

Do not rerun Fixed Width, Rational Series, and K-Nucleotide as a newly selected
profile triad. The current paired and generated-main evidence already splits
their work into UInt128 member, generic call/frame, and map/conversion/GC
routes, so it cannot admit a shared implementation. Keep this full scorecard
as the broad guard and resume performance selection only after a material
language-wide compiler/runtime or semantic-portability change creates one
concrete non-nominal leaf in three unlike verifier-backed applications. Until
then, take the next unfinished semantic or portability roadmap boundary with
cross-runtime fixture parity first. See
`2026-07-15-full-coverage-scorecard-historical-reconciliation.md`.
